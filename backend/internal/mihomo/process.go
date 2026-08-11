package mihomo

import (
	"bytes"
	"fmt"
	"log"
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

// GetVersion returns the Mihomo core version
func (pm *ProcessManager) GetVersion() (*VersionInfo, error) {
	if _, err := os.Stat(pm.binaryPath); os.IsNotExist(err) {
		// Fallback for simulated environment
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
	// Typically: Mihomo v1.18.1 linux amd64 ...
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

// Start starts the Mihomo process
func (pm *ProcessManager) Start() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isRunning() {
		return fmt.Errorf("mihomo is already running (PID: %d)", pm.pid)
	}

	// Create configuration parent directories if they don't exist
	if pm.configPath != "" {
		cfgDir := filepath.Dir(pm.configPath)
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		// Write a minimal mock config if the file doesn't exist
		if _, err := os.Stat(pm.configPath); os.IsNotExist(err) {
			minimalConfig := []byte("mode: rule\nport: 7890\n")
			_ = os.WriteFile(pm.configPath, minimalConfig, 0644)
		}
	}

	var cmd *exec.Cmd
	if _, err := os.Stat(pm.binaryPath); os.IsNotExist(err) {
		log.Printf("Mihomo binary not found at %s. Running in simulated mode.", pm.binaryPath)
		pm.isSimulated = true
		// Use "sleep 3600" as a dummy background process
		cmd = exec.Command("sleep", "3600")
	} else {
		pm.isSimulated = false
		// Run real Mihomo core. Standard flags: -d (directory) and -f (config)
		cfgDir := filepath.Dir(pm.configPath)
		cmd = exec.Command(pm.binaryPath, "-d", cfgDir, "-f", pm.configPath)
	}

	// Prevent zombie processes and set process group
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	pm.cmd = cmd
	pm.pid = cmd.Process.Pid
	pm.startTime = time.Now()

	// Handle process exit asynchronously to clean up
	go func(c *exec.Cmd) {
		_ = c.Wait()
		pm.mu.Lock()
		defer pm.mu.Unlock()
		if pm.cmd == c {
			pm.pid = 0
			pm.cmd = nil
		}
	}(cmd)

	return nil
}

// Stop stops the Mihomo process
func (pm *ProcessManager) Stop() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.isRunning() {
		return fmt.Errorf("mihomo is not running")
	}

	// Kill process group
	pgid, err := syscall.Getpgid(pm.pid)
	if err == nil {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
	} else {
		_ = pm.cmd.Process.Kill()
	}

	// Wait up to 3 seconds for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- pm.cmd.Wait()
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		// Force kill if it takes too long
		if pgid == 0 {
			_ = pm.cmd.Process.Kill()
		} else {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}

	pm.cmd = nil
	pm.pid = 0
	return nil
}

// Restart restarts the Mihomo process
func (pm *ProcessManager) Restart() error {
	if pm.IsRunning() {
		if err := pm.Stop(); err != nil {
			return fmt.Errorf("failed to stop before restart: %w", err)
		}
	}
	return pm.Start()
}

// IsRunning is a thread-safe helper checking if the process is running
func (pm *ProcessManager) IsRunning() bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	return pm.isRunning()
}

func (pm *ProcessManager) isRunning() bool {
	if pm.cmd == nil || pm.cmd.Process == nil || pm.pid == 0 {
		return false
	}

	// Send signal 0 to check if process exists and is running
	process, err := os.FindProcess(pm.pid)
	if err != nil {
		return false
	}
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// Status returns current status
func (pm *ProcessManager) Status() (*StatusResponse, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	running := pm.isRunning()
	pid := pm.pid

	versionStr := "unknown"
	vInfo, err := pm.GetVersion()
	if err == nil && vInfo != nil {
		versionStr = vInfo.Version
		if pm.isSimulated {
			versionStr += " (Simulated)"
		}
	}

	uptime := "0s"
	if running && !pm.startTime.IsZero() {
		dur := time.Since(pm.startTime)
		uptime = formatDuration(dur)
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
