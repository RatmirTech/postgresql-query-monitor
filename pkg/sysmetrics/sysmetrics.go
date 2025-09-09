package sysmetrics

import (
	"runtime"
	"syscall"
)

type SystemMetrics struct {
	CPUCores  int
	CPULoad   float64
	RAMTotal  uint64
	RAMUsed   uint64
	DiskTotal uint64
	DiskUsed  uint64
}

func GetSystemMetrics() SystemMetrics {
	var metrics SystemMetrics

	// CPU Cores
	metrics.CPUCores = runtime.NumCPU()

	// CPU Load (approximated using runtime.NumGoroutine as a simple proxy; for accurate load, consider using gopsutil)
	metrics.CPULoad = float64(runtime.NumGoroutine()) / float64(metrics.CPUCores)

	// RAM Metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.RAMTotal = memStats.Sys
	metrics.RAMUsed = memStats.Alloc

	// Disk Metrics (using syscall for root filesystem)
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err == nil {
		metrics.DiskTotal = stat.Blocks * uint64(stat.Bsize)
		metrics.DiskUsed = (stat.Blocks - stat.Bavail) * uint64(stat.Bsize)
	}

	return metrics
}
