package models

import "time"

// KernelStatus represents the current kernel installation status
type KernelStatus struct {
	Installed    bool      `json:"installed"`
	Version      string    `json:"version"`
	Path         string    `json:"path"`
	LastUpdated  time.Time `json:"lastUpdated"`
	DownloadType string    `json:"downloadType"` // "latest", "stable", "custom"
	Active       bool      `json:"active"`       // true if download is in progress
	Progress     float64   `json:"progress"`     // 0-100, download progress
	Status       string    `json:"status"`       // "idle", "downloading", "completed", "failed"
	StatusMsg    string    `json:"statusMsg"`    // status message
}

// VersionInfo represents a sing-box release version
type VersionInfo struct {
	Version    string    `json:"version"`
	Tag        string    `json:"tag"`
	Prerelease bool      `json:"prerelease"`
	PublishedAt time.Time `json:"publishedAt"`
	Assets     []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	DownloadURL        string `json:"downloadUrl"`
	Size               int64  `json:"size"`
	DownloadCount      int    `json:"downloadCount"`
}

// DownloadRequest represents a download request
type DownloadRequest struct {
	Type     string `json:"type" binding:"required"` // "latest", "stable", "custom"
	Version  string `json:"version"`                  // For custom URL
	URL      string `json:"url"`                      // Custom download URL
}

// DownloadProgress represents download progress
type DownloadProgress struct {
	Active    bool    `json:"active"`
	Progress  float64 `json:"progress"`
	Status    string  `json:"status"`
	Version   string  `json:"version"`
	Error     string  `json:"error,omitempty"`
}

// SwitchRequest represents a version switch request
type SwitchRequest struct {
	Version string `json:"version" binding:"required"`
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
	Version      string    `json:"version"`
	Path         string    `json:"path"`
	Installed    bool      `json:"installed"`
	LastUpdated  time.Time `json:"lastUpdated"`
	DownloadType string    `json:"downloadType"` // "latest", "stable", "custom"
	PID          int       `json:"pid"`
	StartTime    time.Time `json:"startTime"`
}

// ProcessStatus represents the sing-box process status
type ProcessStatus struct {
	Running   bool      `json:"running"`
	PID       int       `json:"pid,omitempty"`
	Status    string    `json:"status"` // "running", "stopped", "not_installed"
	Version   string    `json:"version,omitempty"`
	StartTime time.Time `json:"startTime,omitempty"`
}
