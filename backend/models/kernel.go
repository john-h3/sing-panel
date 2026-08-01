package models

import "time"

// KernelStatus represents the current kernel status
type KernelStatus struct {
	Installed bool      `json:"installed"`
	Version   string    `json:"version"`
	Path      string    `json:"path"`
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
	Status    string    `json:"status"` // "running", "stopped", "not_installed"
	Version   string    `json:"version,omitempty"`
	StartTime time.Time `json:"startTime,omitempty"`
}
