package models

// DashboardConfig represents a single dashboard configuration
type DashboardConfig struct {
	Name    string `json:"name"`    // Dashboard name
	URL     string `json:"url"`     // Dashboard URL
	Enabled *bool  `json:"enabled"` // Whether this dashboard is enabled (pointer for default true)
}

// AppConfig represents the application configuration
type AppConfig struct {
	AccelerateDomain  string            `json:"accelerateDomain"`  // Mirror domain for downloads, e.g. "https://mirror.ghproxy.com"
	AccelerateDomains []string          `json:"accelerateDomains"` // Domains to match for acceleration, e.g. ["github.com"]
	DashboardURL      string            `json:"dashboardURL"`      // Deprecated: use Dashboards instead
	Dashboards        []DashboardConfig `json:"dashboards"`        // Multiple dashboard configurations
}

// ConfigUpdateRequest represents a config update request
// Pointer fields allow distinguishing "not provided" (nil) from "provided as empty"
type ConfigUpdateRequest struct {
	AccelerateDomain  *string           `json:"accelerateDomain"`
	AccelerateDomains *[]string         `json:"accelerateDomains"`
	DashboardURL      *string           `json:"dashboardURL"`
	Dashboards        *[]DashboardConfig `json:"dashboards"`
}
