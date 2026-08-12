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
	mu         sync.Mutex
	cmd        *exec.Cmd
	pid        int
	startTime  time.Time
	binaryPath string
	configPath string
	done       chan struct{}
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

// GetVersion returns the installed Mihomo core version.
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

// Start starts the Mihomo process. A missing or invalid binary is a hard
// error; production must never silently substitute a mock process.
func (pm *ProcessManager) Start() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning() {
		return fmt.Errorf("mihomo is already running (PID: %d)", pm.pid)
	}

	info, err := os.Stat(pm.binaryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("mihomo binary not found: %s", pm.binaryPath)
		}
		return fmt.Errorf("failed to stat mihomo binary: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("mihomo binary path is a directory: %s", pm.binaryPath)
	}
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("mihomo binary is not executable: %s", pm.binaryPath)
	}

	if pm.configPath == "" {
		return fmt.Errorf("mihomo config path is empty")
	}

	cfgDir := filepath.Dir(pm.configPath)
	if err := os.MkdirAll(cfgDir, 0750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	if _, err := os.Stat(pm.configPath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to inspect config file: %w", err)
		}
		minimalConfig := []byte("mode: rule\nport: 7890\n")
		if err := os.WriteFile(pm.configPath, minimalConfig, 0600); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
	}

	cmd := exec.Command(pm.binaryPath, "-d", cfgDir, "-f", pm.configPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mihomo: %w", err)
	}

	pm.cmd = cmd
	pm.pid = cmd.Process.Pid
	pm.startTime = time.Now()
	pm.done = make(chan struct{})
	done := pm.done

	// Wait exactly once for every exec.Cmd. Stop synchronizes through done
	// instead of calling Wait a second time.
	go func(c *exec.Cmd, finished chan struct{}) {
		_ = c.Wait()
		close(finished)
	}(cmd, done)

	return nil
}

// Stop stops the Mihomo process and waits for the single Wait goroutine.
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
	if err == nil && pgid > 0 {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else if cmd.Process != nil {
		_ = cmd.Process.Signal(syscall.SIGTERM)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		if pgid > 0 {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		} else if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()
	if pm.cmd == cmd {
		pm.cmd = nil
		pm.pid = 0
		pm.startTime = time.Time{}
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
	pm.mu.Unlock()

	versionStr := "unknown"
	vInfo, err := pm.GetVersion()
	if err == nil && vInfo != nil {
		versionStr = vInfo.Version
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
