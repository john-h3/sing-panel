package models

import "time"

// InstanceUpdate is used to modify a managed instance. Pointer fields allow
// distinguishing "not provided" (nil) from "provided as empty" (e.g. to
// clear the token).
type InstanceUpdate struct {
	Name    string  `json:"name"`
	URL     string  `json:"url"`
	Token   *string `json:"token"`
	Timeout int     `json:"timeout"`
}

// ManagedInstance represents a remote panel instance managed by this panel.
type ManagedInstance struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Token     string    `json:"token,omitempty"` // optional sync token (X-Sync-Token)
	Timeout   int       `json:"timeout"`         // request timeout in seconds, 0 = default (10s)
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// PanelInfo describes a panel instance's basic information.
// It is served by GET /api/panel/info and used for cross-panel status.
type PanelInfo struct {
	Version          string `json:"version"`          // sing-box/panel version
	Hostname         string `json:"hostname"`         // machine hostname
	Platform         string `json:"platform"`         // GOOS
	Arch             string `json:"arch"`             // GOARCH
	KernelVersion    string `json:"kernelVersion"`    // OS kernel version
	SingboxRunning   bool   `json:"singboxRunning"`   // embedded sing-box running state
	UptimeSeconds    int64  `json:"uptimeSeconds"`    // panel process uptime
	DBSize           int64  `json:"dbSize"`           // database file size in bytes
	SyncTokenEnabled bool   `json:"syncTokenEnabled"` // whether remote sync requires a token
}

// InstanceStatus is the live status of a managed instance after a check.
type InstanceStatus struct {
	Instance    ManagedInstance `json:"instance"`
	Reachable   bool            `json:"reachable"`
	LatencyMs   int64           `json:"latencyMs"`
	Error       string          `json:"error,omitempty"`
	Info        *PanelInfo      `json:"info,omitempty"`
	Fingerprint string          `json:"fingerprint,omitempty"` // remote config fingerprint
	InSync      *bool           `json:"inSync"`                // nil when fingerprint unknown
	CheckedAt   time.Time       `json:"checkedAt"`
}
