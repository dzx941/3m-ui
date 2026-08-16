package system

import (
	"math"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

var (
	netMu    sync.Mutex
	lastRecv uint64
	lastSent uint64
	lastTime time.Time
)

// GetSystemStats returns the parsed dynamic system performance statistics
func GetSystemStats() *SystemStats {
	// 1. CPU Usage
	cpuPercent := 0.0
	cpuPercents, err := cpu.Percent(time.Duration(100*time.Millisecond), false)
	if err == nil && len(cpuPercents) > 0 {
		cpuPercent = math.Round(cpuPercents[0]*10) / 10
	}

	// 2. Memory Usage
	var memoryInfo MemoryInfo
	vMem, err := mem.VirtualMemory()
	if err == nil {
		memoryInfo = MemoryInfo{
			Used:    float64(vMem.Used) / (1024 * 1024),  // to MB
			Total:   float64(vMem.Total) / (1024 * 1024), // to MB
			Percent: math.Round(vMem.UsedPercent*10) / 10,
		}
	}

	// 3. Disk Usage
	var diskInfo DiskInfo
	dUsage, err := disk.Usage("/")
	if err == nil {
		diskInfo = DiskInfo{
			Used:    float64(dUsage.Used) / (1024 * 1024 * 1024),  // to GB
			Total:   float64(dUsage.Total) / (1024 * 1024 * 1024), // to GB
			Percent: math.Round(dUsage.UsedPercent*10) / 10,
		}
	}

	// 4. Network Rates (upload & download rates in bytes/sec)
	var networkInfo NetworkInfo
	netIO, err := net.IOCounters(false)
	if err == nil && len(netIO) > 0 {
		netMu.Lock()
		now := time.Now()
		currRecv := netIO[0].BytesRecv
		currSent := netIO[0].BytesSent

		if !lastTime.IsZero() {
			duration := now.Sub(lastTime).Seconds()
			if duration > 0 {
				recvDelta := currRecv - lastRecv
				sentDelta := currSent - lastSent

				networkInfo.Download = float64(recvDelta) / duration
				networkInfo.Upload = float64(sentDelta) / duration
			}
		}

		lastRecv = currRecv
		lastSent = currSent
		lastTime = now
		netMu.Unlock()
	}

	return &SystemStats{
		CPU: CPUInfo{
			Percent: cpuPercent,
		},
		Memory:  memoryInfo,
		Disk:    diskInfo,
		Network: networkInfo,
	}
}
