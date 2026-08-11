package system_test

import (
	"testing"

	"github.com/dzx941/3m-ui/backend/internal/system"
)

func TestGetSystemStats(t *testing.T) {
	stats := system.GetSystemStats()

	if stats == nil {
		t.Fatal("expected non-nil stats response")
	}

	if stats.CPU.Percent < 0 {
		t.Fatalf("unexpected CPU value: %f", stats.CPU.Percent)
	}

	if stats.Memory.Percent < 0 || stats.Memory.Percent > 100 {
		t.Fatalf("unexpected Memory percent: %f", stats.Memory.Percent)
	}

	if stats.Disk.Percent < 0 || stats.Disk.Percent > 100 {
		t.Fatalf("unexpected Disk percent: %f", stats.Disk.Percent)
	}

	if stats.Network.Upload < 0 || stats.Network.Download < 0 {
		t.Fatal("unexpected Network rate metrics")
	}
}
