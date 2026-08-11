package mihomo_test

import (
	"os"
	"testing"
	"time"

	"github.com/dzx941/3m-ui/backend/internal/mihomo"
)

func TestProcessManagerStartStop(t *testing.T) {
	binaryPath := "/tmp/non_existent_binary_dummy"
	configPath := "/tmp/3m-ui-config-test/config.yaml"

	// Ensure clean slate
	_ = os.RemoveAll("/tmp/3m-ui-config-test")

	pm := mihomo.GetProcessManager(binaryPath, configPath)

	// Fetch status initially - should not be running
	status, err := pm.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.Running {
		t.Fatal("expected status.Running to be false initially")
	}

	// Start the process (it should run in simulated/sleep mode)
	err = pm.Start()
	if err != nil {
		t.Fatalf("failed to start process: %v", err)
	}

	// Verify it is running
	status, err = pm.Status()
	if err != nil {
		t.Fatalf("failed to get status after start: %v", err)
	}

	if !status.Running {
		t.Fatal("expected status.Running to be true after start")
	}

	if status.PID <= 0 {
		t.Fatalf("expected positive PID, got %d", status.PID)
	}

	// Sleep for a short while to check status uptime
	time.Sleep(1 * time.Second)
	status, err = pm.Status()
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}
	if status.Uptime == "0s" {
		t.Fatal("expected uptime to update from 0s")
	}

	// Stop the process
	err = pm.Stop()
	if err != nil {
		t.Fatalf("failed to stop process: %v", err)
	}

	// Verify it is stopped
	status, err = pm.Status()
	if err != nil {
		t.Fatalf("failed to get status after stop: %v", err)
	}

	if status.Running {
		t.Fatal("expected status.Running to be false after stop")
	}

	_ = os.RemoveAll("/tmp/3m-ui-config-test")
}
