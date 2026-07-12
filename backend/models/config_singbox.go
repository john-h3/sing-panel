package models

// Inbound represents a sing-box inbound configuration
type Inbound struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "http", "socks", "tun", "shadowsocks", etc.
	Tag         string                 `json:"tag"`
	Listen      string                 `json:"listen,omitempty"`
	ListenPort  int                    `json:"listenPort"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// Outbound represents a sing-box outbound configuration
type Outbound struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // "direct", "block", "selector", "urltest", "shadowsocks", "vmess", etc.
	Tag         string                 `json:"tag"`
	Options     map[string]interface{} `json:"options,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// SingBoxConfig represents the full sing-box configuration
type SingBoxConfig struct {
	Inbounds  []Inbound  `json:"inbounds"`
	Outbounds []Outbound `json:"outbounds"`
}

// ConfigRequest represents a config update request
type ConfigRequest struct {
	Action  string      `json:"action"` // "add", "update", "delete"
	Type    string      `json:"type"`   // "inbound", "outbound"
	Data    interface{} `json:"data"`
}

// InboundRequest represents an inbound CRUD request
type InboundRequest struct {
	Action string  `json:"action"` // "add", "update", "delete"
	Data  Inbound `json:"data"`
}

// OutboundRequest represents an outbound CRUD request
type OutboundRequest struct {
	Action string   `json:"action"` // "add", "update", "delete"
	Data  Outbound `json:"data"`
}
