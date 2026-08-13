package mihomo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessManager struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	pid        int
	startTime  time.Time
	binaryPath string
	configPath string
	done       chan struct{}
	logs       []string
	desired    bool
	external   bool
}

var globalPM *ProcessManager
var pmOnce sync.Once

func GetProcessManager(binary, config string) *ProcessManager {
	pmOnce.Do(func() {
		globalPM = &ProcessManager{
			binaryPath: binary,
			configPath: config,
			logs:       make([]string, 0, 200),
		}
	})

	return globalPM
}

func (pm *ProcessManager) GetVersion() (*VersionInfo, error) {
	info, err := os.Stat(pm.binaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("mihomo binary not found: %s", pm.binaryPath)
		}

		return nil, fmt.Errorf("failed to stat mihomo binary: %w", err)
	}

	if info.IsDir() {
		return nil, fmt.Errorf("mihomo binary path is a directory: %s", pm.binaryPath)
	}

	cmd := exec.Command(pm.binaryPath, "-v")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run mihomo -v: %w", err)
	}

	output := strings.TrimSpace(out.String())
	parts := strings.Fields(output)

	version := "unknown"
	if len(parts) >= 2 {
		version = parts[1]
	}

	return &VersionInfo{
		Version: version,
		Commit:  "official-build",
		Build:   output,
	}, nil
}

func (pm *ProcessManager) ValidateConfig() error {
	pm.mu.Lock()
	binaryPath := pm.binaryPath
	configPath := pm.configPath
	pm.mu.Unlock()

	if binaryPath == "" || configPath == "" {
		return fmt.Errorf("mihomo binary or config path is empty")
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		return fmt.Errorf("mihomo binary unavailable: %w", err)
	}

	if info.IsDir() || info.Mode()&0111 == 0 {
		return fmt.Errorf("mihomo binary is not executable: %s", binaryPath)
	}

	cfgInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("mihomo config not found: %s", configPath)
	}

	if cfgInfo.IsDir() || cfgInfo.Size() == 0 {
		return fmt.Errorf("mihomo config is empty: %s", configPath)
	}

	cmd := exec.Command(
		binaryPath,
		"-t",
		"-d",
		filepath.Dir(configPath),
		"-f",
		configPath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(out.String())

		if detail != "" {
			return fmt.Errorf(
				"mihomo configuration validation failed: %s",
				detail,
			)
		}

		return fmt.Errorf(
			"mihomo configuration validation failed: %w",
			err,
		)
	}

	return nil
}

// findExistingProcesses 查找已经运行的、使用相同 Mihomo
// binary + config 的进程。
//
// 不能依赖 ps，因为 Alpine/BusyBox 的 ps 参数并不统一。
// 直接读取 /proc 更可靠。
func (pm *ProcessManager) findExistingProcesses() []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	binary, err := filepath.EvalSymlinks(pm.binaryPath)
	if err != nil {
		return nil
	}

	config := filepath.Clean(pm.configPath)

	configDir := filepath.Dir(config)

	var pids []int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 || pid == os.Getpid() {
			continue
		}

		base := filepath.Join("/proc", entry.Name())

		cmdline, err := os.ReadFile(
			filepath.Join(base, "cmdline"),
		)
		if err != nil {
			continue
		}

		args := strings.Split(
			string(cmdline),
			"\x00",
		)

		if len(args) < 2 {
			continue
		}

		exe, err := filepath.EvalSymlinks(
			filepath.Join(base, "exe"),
		)
		if err != nil {
			continue
		}

		if exe != binary {
			continue
		}

		matchedDir := false
		matchedConfig := false

		for i := 1; i < len(args); i++ {
			switch args[i] {
			case "-d":
				if i+1 < len(args) {
					matchedDir =
						filepath.Clean(args[i+1]) == configDir
				}

			case "-f":
				if i+1 < len(args) {
					matchedConfig =
						filepath.Clean(args[i+1]) == config
				}
			}
		}

		if matchedDir && matchedConfig {
			pids = append(pids, pid)
		}
	}

	return pids
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	return process.Signal(syscall.Signal(0)) == nil
}

// adoptExistingLocked 接管已经存在的 Mihomo。
// 如果发现多个实例，只保留最早的 PID，其余发送 SIGTERM。
//
// 注意：这个函数必须在 pm.mu 已经锁定的情况下调用。
func (pm *ProcessManager) adoptExistingLocked(
	pids []int,
) int {
	if len(pids) == 0 {
		return 0
	}

	keep := pids[0]

	for _, pid := range pids[1:] {
		if pid < keep {
			keep = pid
		}
	}

	for _, pid := range pids {
		if pid == keep {
			continue
		}

		process, err := os.FindProcess(pid)
		if err == nil {
			_ = process.Signal(syscall.SIGTERM)
		}

		pm.appendLogLocked(
			fmt.Sprintf(
				"stopped duplicate Mihomo process PID %d; keeping PID %d",
				pid,
				keep,
			),
		)
	}

	pm.pid = keep
	pm.startTime = time.Now()
	pm.desired = true
	pm.external = true

	return keep
}

func (pm *ProcessManager) Start() error {
	/*
		第一阶段：

		先拿内部锁。

		避免：

		    HTTP 请求 A -> Start()
		    HTTP 请求 B -> Start()

		同时进入。
	*/

	pm.mu.Lock()

	if pm.isRunning() {
		pid := pm.pid

		pm.mu.Unlock()

		return fmt.Errorf(
			"mihomo is already running (PID: %d)",
			pid,
		)
	}

	binaryPath := pm.binaryPath
	configPath := pm.configPath

	pm.mu.Unlock()

	/*
		检查 binary。
	*/

	info, err := os.Stat(binaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"mihomo binary not found: %s",
				binaryPath,
			)
		}

		return fmt.Errorf(
			"failed to stat mihomo binary: %w",
			err,
		)
	}

	if info.IsDir() || info.Mode()&0111 == 0 {
		return fmt.Errorf(
			"mihomo binary is not executable: %s",
			binaryPath,
		)
	}

	/*
		检查 config。
	*/

	cfgInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf(
			"mihomo config not found: %s",
			configPath,
		)
	}

	if cfgInfo.IsDir() || cfgInfo.Size() == 0 {
		return fmt.Errorf(
			"mihomo config is empty: %s",
			configPath,
		)
	}

	/*
		第二阶段：

		重新获得锁。

		这是非常重要的。

		在上面的文件检查过程中，另一个请求可能已经启动
		Mihomo。

		所以不能只在函数开始时检查一次。
	*/

	pm.mu.Lock()

	if pm.isRunning() {
		pid := pm.pid

		pm.mu.Unlock()

		return fmt.Errorf(
			"mihomo is already running (PID: %d)",
			pid,
		)
	}

	/*
		检查系统中是否已经存在 Mihomo。

		这主要解决：

		3m-ui 重启
		     ↓
		Mihomo 子进程没有退出
		     ↓
		新的 3m-ui 启动
		     ↓
		Start()
		     ↓
		不能再启动第二个 Mihomo
	*/

	existing := pm.findExistingProcesses()

	if len(existing) > 0 {
		pid := pm.adoptExistingLocked(existing)

		pm.mu.Unlock()

		return fmt.Errorf(
			"mihomo is already running (PID: %d); existing process adopted",
			pid,
		)
	}

	pm.mu.Unlock()

	/*
		真正启动 Mihomo。
	*/

	cmd := exec.Command(
		binaryPath,
		"-d",
		filepath.Dir(configPath),
		"-f",
		configPath,
	)

	/*
		让 Mihomo 成为独立 process group。

		这样 Stop 时可以连同 Mihomo 的子进程一起清理。
	*/

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	writer := &processLogWriter{
		pm: pm,
	}

	cmd.Stdout = writer
	cmd.Stderr = writer

	if err := cmd.Start(); err != nil {
		return fmt.Errorf(
			"failed to start mihomo: %w",
			err,
		)
	}

	/*
		启动成功后立即记录 PID。
	*/

	pm.mu.Lock()

	pm.cmd = cmd
	pm.pid = cmd.Process.Pid
	pm.startTime = time.Now()
	pm.done = make(chan struct{})
	pm.desired = true
	pm.external = false

	done := pm.done

	pm.mu.Unlock()

	go pm.waitProcess(cmd, done)

	/*
		给 Mihomo 一小段时间确认没有立即退出。
	*/

	select {
	case <-done:
		pm.mu.Lock()

		if pm.cmd == cmd {
			pm.cmd = nil
			pm.pid = 0
			pm.startTime = time.Time{}
			pm.done = nil
			pm.desired = false
		}

		pm.mu.Unlock()

		return fmt.Errorf(
			"mihomo exited immediately; check logs",
		)

	case <-time.After(500 * time.Millisecond):
		return nil
	}
}

func (pm *ProcessManager) waitProcess(
	cmd *exec.Cmd,
	done chan struct{},
) {
	err := cmd.Wait()

	close(done)

	pm.mu.Lock()

	if err != nil {
		pm.appendLogLocked(
			fmt.Sprintf(
				"process exited: %v",
				err,
			),
		)
	} else {
		pm.appendLogLocked(
			"process exited",
		)
	}

	/*
		只有当前 cmd 仍然是 manager 管理的实例，
		并且 desired=true，才允许自动恢复。
	*/

	restart :=
		pm.desired &&
			pm.cmd == cmd

	pm.mu.Unlock()

	if !restart {
		return
	}

	time.Sleep(2 * time.Second)

	if pm.IsRunning() {
		return
	}

	if err := pm.ValidateConfig(); err != nil {
		pm.mu.Lock()

		pm.appendLogLocked(
			fmt.Sprintf(
				"automatic restart blocked by config validation: %v",
				err,
			),
		)

		pm.mu.Unlock()

		return
	}

	if err := pm.Start(); err != nil {
		pm.mu.Lock()

		pm.appendLogLocked(
			fmt.Sprintf(
				"automatic restart failed: %v",
				err,
			),
		)

		pm.mu.Unlock()
	}
}

func (pm *ProcessManager) Stop() error {
	pm.mu.Lock()

	if !pm.isRunning() {
		pm.desired = false

		pm.mu.Unlock()

		return fmt.Errorf(
			"mihomo is not running",
		)
	}

	pid := pm.pid
	cmd := pm.cmd
	done := pm.done

	/*
		必须先关闭 desired。

		否则 waitProcess 可能认为这是异常退出，
		然后又自动 Start()。
	*/

	pm.desired = false

	pm.mu.Unlock()

	/*
		优先杀整个 process group。
	*/

	if pgid, err := syscall.Getpgid(pid); err == nil &&
		pgid > 0 {

		_ = syscall.Kill(
			-pgid,
			syscall.SIGTERM,
		)

	} else if process, err := os.FindProcess(pid); err == nil {
		_ = process.Signal(syscall.SIGTERM)
	}

	/*
		如果是当前 3m-ui 创建的 cmd，
		等待 Wait() 完成。
	*/

	if done != nil {
		select {
		case <-done:

		case <-time.After(5 * time.Second):

			if pgid, err := syscall.Getpgid(pid); err == nil &&
				pgid > 0 {

				_ = syscall.Kill(
					-pgid,
					syscall.SIGKILL,
				)
			}

			if cmd != nil &&
				cmd.Process != nil {

				_ = cmd.Process.Kill()
			}

			<-done
		}

	} else {
		/*
			这是接管的旧 Mihomo。
		*/

		deadline :=
			time.Now().Add(5 * time.Second)

		for processAlive(pid) &&
			time.Now().Before(deadline) {

			time.Sleep(
				100 * time.Millisecond,
			)
		}

		if processAlive(pid) {
			if process, err := os.FindProcess(pid); err == nil {
				_ = process.Kill()
			}
		}
	}

	pm.mu.Lock()

	if pm.pid == pid {
		pm.cmd = nil
		pm.pid = 0
		pm.startTime = time.Time{}
		pm.done = nil
		pm.external = false
	}

	pm.mu.Unlock()

	return nil
}

func (pm *ProcessManager) Restart() error {
	/*
		Restart 必须严格执行：

		    Stop
		     ↓
		    等待退出
		     ↓
		    Validate
		     ↓
		    Start

		不能直接 Start。
	*/

	if pm.IsRunning() {
		if err := pm.Stop(); err != nil {
			return fmt.Errorf(
				"failed to stop before restart: %w",
				err,
			)
		}
	}

	if err := pm.ValidateConfig(); err != nil {
		return err
	}

	return pm.Start()
}

func (pm *ProcessManager) IsRunning() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.isRunning()
}

func (pm *ProcessManager) isRunning() bool {
	if pm.pid == 0 {
		return false
	}

	if pm.done != nil {
		select {
		case <-pm.done:
			return false

		default:
		}
	}

	return processAlive(pm.pid)
}

func (pm *ProcessManager) Status() (*StatusResponse, error) {
	pm.mu.Lock()

	running := pm.isRunning()
	pid := pm.pid
	startTime := pm.startTime

	pm.mu.Unlock()

	versionStr := "unknown"

	if vInfo, err := pm.GetVersion(); err == nil &&
		vInfo != nil {

		versionStr = vInfo.Version
	}

	uptime := "0s"

	if running && !startTime.IsZero() {
		uptime =
			formatDuration(
				time.Since(startTime),
			)
	}

	return &StatusResponse{
		Running: running,
		Version: versionStr,
		PID:     pid,
		Uptime:  uptime,
	}, nil
}

func (pm *ProcessManager) Logs() []string {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	out := make(
		[]string,
		len(pm.logs),
	)

	copy(
		out,
		pm.logs,
	)

	return out
}

func (pm *ProcessManager) appendLogLocked(line string) {
	line = strings.TrimSpace(line)

	if line == "" {
		return
	}

	pm.logs = append(
		pm.logs,
		time.Now().Format(time.RFC3339)+" "+line,
	)

	if len(pm.logs) > 200 {
		pm.logs =
			pm.logs[len(pm.logs)-200:]
	}
}

type processLogWriter struct {
	pm *ProcessManager
}

func (w *processLogWriter) Write(
	p []byte,
) (int, error) {

	lines :=
		strings.Split(
			string(p),
			"\n",
		)

	w.pm.mu.Lock()
	defer w.pm.mu.Unlock()

	for _, line := range lines {
		w.pm.appendLogLocked(line)
	}

	return len(p), nil
}

func formatDuration(
	d time.Duration,
) string {

	d = d.Round(time.Second)

	h := d / time.Hour
	d -= h * time.Hour

	m := d / time.Minute
	d -= m * time.Minute

	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf(
			\"%dh %dm %ds\",
			h,
			m,
			s,
		)
	}

	if m > 0 {
		return fmt.Sprintf(
			\"%dm %ds\",
			m,
			s,
		)
	}

	return fmt.Sprintf(
		\"%ds\",
		s,
	)
}
