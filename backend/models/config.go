package models

// AppConfig represents the application configuration
type AppConfig struct {
	AccelerateDomain string `json:"accelerateDomain"` // Mirror domain for downloads, e.g. "https://mirror.ghproxy.com"
}

// ConfigUpdateRequest represents a config update request
type ConfigUpdateRequest struct {
	AccelerateDomain string `json:"accelerateDomain"`
}
