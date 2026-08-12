package mihomo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type ProcessManager struct {
	mu          sync.Mutex
	cmd         *exec.Cmd
	pid         int
	startTime   time.Time
	binaryPath  string
	configPath  string
	isSimulated bool
	done        chan struct{}
}

var globalPM *ProcessManager
var pmOnce sync.Once

func GetProcessManager(binary, config string) *ProcessManager {
	pmOnce.Do(func() {
		globalPM = &ProcessManager{
			binaryPath: binary,
			configPath: config,
		}
	})
	return globalPM
}

// GetVersion returns the Mihomo core version.
func (pm *ProcessManager) GetVersion() (*VersionInfo, error) {
	if _, err := os.Stat(pm.binaryPath); os.IsNotExist(err) {
		return &VersionInfo{
			Version: "v1.18.1-meta (Simulated Mock)",
			Commit:  "70f900a",
			Build:   "Go1.24.0",
		}, nil
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

// Start starts the Mihomo process.
func (pm *ProcessManager) Start() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning() {
		return fmt.Errorf("mihomo is already running (PID: %d)", pm.pid)
	}

	if pm.configPath != "" {
		cfgDir := filepath.Dir(pm.configPath)
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if _, err := os.Stat(pm.configPath); os.IsNotExist(err) {
			minimalConfig := []byte("mode: rule\nport: 7890\n")
			if err := os.WriteFile(pm.configPath, minimalConfig, 0644); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		}
	}

	var cmd *exec.Cmd
	if _, err := os.Stat(pm.binaryPath); os.IsNotExist(err) {
		// Keep the simulation path for development/tests, but never hide a real
		// filesystem error. Production installations should always ship Mihomo.
		pm.isSimulated = true
		cmd = exec.Command("sleep", "3600")
	} else {
		pm.isSimulated = false
		cfgDir := filepath.Dir(pm.configPath)
		cmd = exec.Command(pm.binaryPath, "-d", cfgDir, "-f", pm.configPath)
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pm.cmd = cmd
	pm.pid = cmd.Process.Pid
	pm.startTime = time.Now()
	pm.done = make(chan struct{})
	done := pm.done

	// Wait exactly once. os/exec.Cmd.Wait must not be called concurrently or
	// more than once; Stop waits on this completion channel instead.
	go func(c *exec.Cmd, finished chan struct{}) {
		_ = c.Wait()
		close(finished)
	}(cmd, done)

	return nil
}

// Stop stops the Mihomo process.
func (pm *ProcessManager) Stop() error {
	pm.mu.Lock()
	if !pm.isRunning() {
		pm.mu.Unlock()
		return fmt.Errorf("mihomo is not running")
	}

	cmd := pm.cmd
	pid := pm.pid
	done := pm.done
	pm.mu.Unlock()

	pgid, err := syscall.Getpgid(pid)
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = cmd.Process.Kill()
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		if pgid > 0 {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else {
			_ = cmd.Process.Kill()
		}
		<-done
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.cmd == cmd {
		pm.cmd = nil
		pm.pid = 0
		pm.done = nil
	}
	return nil
}

// Restart restarts the Mihomo process.
func (pm *ProcessManager) Restart() error {
	if pm.IsRunning() {
		if err := pm.Stop(); err != nil {
			return fmt.Errorf("failed to stop before restart: %w", err)
		}
	}
	return pm.Start()
}

// IsRunning is a thread-safe helper checking if the process is running.
func (pm *ProcessManager) IsRunning() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.isRunning()
}

func (pm *ProcessManager) isRunning() bool {
	if pm.cmd == nil || pm.cmd.Process == nil || pm.pid == 0 {
		return false
	}
	process, err := os.FindProcess(pm.pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

// Status returns current status.
func (pm *ProcessManager) Status() (*StatusResponse, error) {
	pm.mu.Lock()
	running := pm.isRunning()
	pid := pm.pid
	startTime := pm.startTime
	simulated := pm.isSimulated
	pm.mu.Unlock()

	versionStr := "unknown"
	vInfo, err := pm.GetVersion()
	if err == nil && vInfo != nil {
		versionStr = vInfo.Version
		if simulated {
			versionStr += " (Simulated)"
		}
	}

	uptime := "0s"
	if running && !startTime.IsZero() {
		uptime = formatDuration(time.Since(startTime))
	}

	return &StatusResponse{
		Running: running,
		Version: versionStr,
		PID:     pid,
		Uptime:  uptime,
	}, nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
