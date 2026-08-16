package models

import "time"

// KernelStatus represents the current kernel status
type KernelStatus struct {
	Version string `json:"version"`
}

// SystemInfo represents server system information
type SystemInfo struct {
	Platform      string `json:"platform"`
	Arch          string `json:"arch"`
	Hostname      string `json:"hostname"`
	KernelVersion string `json:"kernelVersion"`
}

// KernelState represents the runtime state stored in database
type KernelState struct {
	Version   string    `json:"version"`
	Path      string    `json:"path"`
	Installed bool      `json:"installed"`
	PID       int       `json:"pid"`
	StartTime time.Time `json:"startTime"`
}

// ProcessStatus represents the sing-box process status
type ProcessStatus struct {
	Running   bool      `json:"running"`
	PID       int       `json:"pid,omitempty"`
	Status    string    `json:"status"` // "running", "stopped"
	Version   string    `json:"version,omitempty"`
	StartTime time.Time `json:"startTime,omitempty"`
}

// MonitorStats represents Go runtime performance metrics
type MonitorStats struct {
	UptimeSeconds int64  `json:"uptimeSeconds"` // panel process uptime
	Goroutines    int    `json:"goroutines"`
	NumCPU        int    `json:"numCPU"`
	GOMAXPROCS    int    `json:"gomaxprocs"`
	HeapAlloc     uint64 `json:"heapAlloc"` // bytes of allocated heap objects
	HeapSys       uint64 `json:"heapSys"`   // bytes of heap memory obtained from OS
	HeapInuse     uint64 `json:"heapInuse"` // bytes in in-use spans
	HeapObjects   uint64 `json:"heapObjects"`
	Sys           uint64 `json:"sys"` // total bytes of memory obtained from OS
	NumGC         uint32 `json:"numGC"`
	LastGC        int64  `json:"lastGC"` // unix nanoseconds of last GC
	PauseTotalNs  uint64 `json:"pauseTotalNs"`
	LastPauseNs   uint64 `json:"lastPauseNs"`
	GCPercent     int    `json:"gcPercent"`
}
