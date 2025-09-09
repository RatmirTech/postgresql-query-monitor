package collectors

import (
	"log"
	"runtime"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

// SysMetricsCollector collects system-level metrics
type SysMetricsCollector struct{}

// NewSysMetricsCollector creates new system metrics collector
func NewSysMetricsCollector() *SysMetricsCollector {
	return &SysMetricsCollector{}
}

// SystemMetric represents a single system metric
type SystemMetric struct {
	Name        string            `json:"name"`
	Value       interface{}       `json:"value"`
	Labels      map[string]string `json:"labels"`
	Description string            `json:"description"`
	Timestamp   time.Time         `json:"timestamp"`
}

// SystemMetrics represents all collected system metrics
type SystemMetrics struct {
	CPUCores   int       `json:"cpu_cores"`
	CPULoad    float64   `json:"cpu_load"`
	RAMTotal   uint64    `json:"ram_total"`
	RAMUsed    uint64    `json:"ram_used"`
	RAMFree    uint64    `json:"ram_free"`
	DiskTotal  uint64    `json:"disk_total"`
	DiskUsed   uint64    `json:"disk_used"`
	DiskFree   uint64    `json:"disk_free"`
	Goroutines int       `json:"goroutines"`
	GCPauses   uint64    `json:"gc_pauses"`
	HeapAlloc  uint64    `json:"heap_alloc"`
	HeapSys    uint64    `json:"heap_sys"`
	StackInUse uint64    `json:"stack_in_use"`
	Timestamp  time.Time `json:"timestamp"`
}

// Collect gathers system metrics
func (c *SysMetricsCollector) Collect() SystemMetrics {
	now := time.Now()
	var metrics SystemMetrics

	// CPU metrics
	metrics.CPUCores = runtime.NumCPU()
	metrics.CPULoad = float64(runtime.NumGoroutine()) / float64(metrics.CPUCores)

	// Memory metrics — TRY SYSTEM MEMORY FIRST
	if vmStat, err := mem.VirtualMemory(); err == nil {
		metrics.RAMTotal = vmStat.Total
		metrics.RAMUsed = vmStat.Used
		metrics.RAMFree = vmStat.Available
	} else {
		log.Printf("Warning: failed to read system memory via gopsutil: %v. Falling back to Go runtime stats.", err)

		// Fallback: use Go process memory stats (NOT system RAM!)
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		metrics.RAMTotal = memStats.Sys
		metrics.RAMUsed = memStats.Alloc
		metrics.RAMFree = memStats.Sys - memStats.Alloc

		// Также заполняем Go-specific поля
		metrics.HeapAlloc = memStats.HeapAlloc
		metrics.HeapSys = memStats.HeapSys
		metrics.StackInUse = memStats.StackInuse
		metrics.GCPauses = memStats.PauseTotalNs
	} 

	// If we didn't fallback, we still need Go runtime stats for Go-specific metrics
	if metrics.HeapAlloc == 0 { // simple check to avoid double-read if already done
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		metrics.HeapAlloc = memStats.HeapAlloc
		metrics.HeapSys = memStats.HeapSys
		metrics.StackInUse = memStats.StackInuse
		metrics.GCPauses = memStats.PauseTotalNs
	}

	// Garbage collection & goroutines
	metrics.Goroutines = runtime.NumGoroutine()

	// Disk metrics (using syscall for root filesystem)
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/", &stat); err == nil {
		blockSize := uint64(stat.Bsize)
		metrics.DiskTotal = stat.Blocks * blockSize
		metrics.DiskUsed = (stat.Blocks - stat.Bavail) * blockSize
		metrics.DiskFree = stat.Bavail * blockSize
	} else {
		log.Printf("Warning: failed to read disk stats: %v", err)
	}

	metrics.Timestamp = now
	return metrics
}

// CollectDetailed returns detailed system metrics as separate metric objects
func (c *SysMetricsCollector) CollectDetailed() []SystemMetric {
	metrics := c.Collect()
	now := time.Now()

	return []SystemMetric{
		{
			Name:        "system_cpu_cores",
			Value:       metrics.CPUCores,
			Description: "Number of CPU cores",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
		{
			Name:        "system_cpu_load_ratio",
			Value:       metrics.CPULoad,
			Description: "CPU load ratio (goroutines per core)",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
		{
			Name:        "system_memory_total_bytes",
			Value:       metrics.RAMTotal,
			Description: "Total system memory in bytes (from OS)",
			Labels:      map[string]string{"type": "total"},
			Timestamp:   now,
		},
		{
			Name:        "system_memory_used_bytes",
			Value:       metrics.RAMUsed,
			Description: "Used system memory in bytes (from OS)",
			Labels:      map[string]string{"type": "used"},
			Timestamp:   now,
		},
		{
			Name:        "system_memory_free_bytes",
			Value:       metrics.RAMFree,
			Description: "Free system memory in bytes (from OS)",
			Labels:      map[string]string{"type": "free"},
			Timestamp:   now,
		},
		{
			Name:        "system_disk_total_bytes",
			Value:       metrics.DiskTotal,
			Description: "Total disk space in bytes",
			Labels:      map[string]string{"mountpoint": "/", "type": "total"},
			Timestamp:   now,
		},
		{
			Name:        "system_disk_used_bytes",
			Value:       metrics.DiskUsed,
			Description: "Used disk space in bytes",
			Labels:      map[string]string{"mountpoint": "/", "type": "used"},
			Timestamp:   now,
		},
		{
			Name:        "system_disk_free_bytes",
			Value:       metrics.DiskFree,
			Description: "Free disk space in bytes",
			Labels:      map[string]string{"mountpoint": "/", "type": "free"},
			Timestamp:   now,
		},
		{
			Name:        "go_goroutines",
			Value:       metrics.Goroutines,
			Description: "Number of goroutines",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
		{
			Name:        "go_gc_pause_total_ns",
			Value:       metrics.GCPauses,
			Description: "Total garbage collection pause time in nanoseconds",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
		{
			Name:        "go_memory_heap_alloc_bytes",
			Value:       metrics.HeapAlloc,
			Description: "Bytes allocated on the heap (Go runtime)",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
		{
			Name:        "go_memory_heap_sys_bytes",
			Value:       metrics.HeapSys,
			Description: "Bytes obtained from system for heap (Go runtime)",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
		{
			Name:        "go_memory_stack_inuse_bytes",
			Value:       metrics.StackInUse,
			Description: "Bytes in use by the stack allocator (Go runtime)",
			Labels:      map[string]string{},
			Timestamp:   now,
		},
	}
}