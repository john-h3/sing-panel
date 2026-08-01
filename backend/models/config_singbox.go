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

// Ruleset represents a sing-box ruleset configuration
type Ruleset struct {
	ID      string                 `json:"id"`
	Tag     string                 `json:"tag"`
	Type    string                 `json:"type"` // "inline", "local", "remote"
	Options map[string]interface{} `json:"options,omitempty"`
	Enabled bool                   `json:"enabled"`
}

// RouteRule represents a route rule configuration
type RouteRule struct {
	ID       string                 `json:"id"`
	Action   string                 `json:"action"`   // "route", "bypass", "reject"
	Inbound  []string               `json:"inbound"`  // tags of inbounds
	RuleSet  []string               `json:"rule_set"` // tags of rulesets
	Outbound string                 `json:"outbound"` // tag of outbound (only for route)
	Options  map[string]interface{} `json:"options,omitempty"`
	Enabled  bool                   `json:"enabled"`
}

// DNSServer represents a DNS server entry
type DNSServer struct {
	Type string `json:"type"` // currently only "local"
	Tag  string `json:"tag"`
}

// DNSRule represents a DNS rule
type DNSRule struct {
	RuleSet  []string `json:"rule_set"`
	Server   string   `json:"server"`
	Disabled bool     `json:"disabled,omitempty"`
}

// DNSConfig represents the sing-box DNS configuration
type DNSConfig struct {
	Servers          []DNSServer            `json:"servers"`
	Rules            []DNSRule              `json:"rules"`
	Final            string                 `json:"final"`
	Strategy         string                 `json:"strategy"`
	DisableCache     bool                   `json:"disable_cache"`
	DisableExpire    bool                   `json:"disable_expire"`
	IndependentCache bool                   `json:"independent_cache"`
	CacheCapacity    int                    `json:"cache_capacity"`
	Optimistic       interface{}            `json:"optimistic"`
	Timeout          string                 `json:"timeout"`
	ReverseMapping   bool                   `json:"reverse_mapping"`
	ClientSubnet     string                 `json:"client_subnet"`
	FakeIP           map[string]interface{} `json:"fakeip"`
}

// RouteConfig represents the sing-box route configuration
type RouteConfig struct {
	Final            string `json:"final"`             // tag of default outbound
	DefaultHttpClient string `json:"default_http_client"` // tag of default HTTP client
}

// Service represents a sing-box service configuration (e.g. API)
type Service struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // "api"
	Tag        string                 `json:"tag"`
	Listen     string                 `json:"listen,omitempty"`
	ListenPort int                    `json:"listenPort"`
	Options    map[string]interface{} `json:"options,omitempty"`
	Enabled    bool                   `json:"enabled"`
}

// HTTPClient represents a sing-box HTTP client configuration
type HTTPClient struct {
	ID      string                 `json:"id"`
	Tag     string                 `json:"tag"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// CacheFileConfig represents the cache_file experimental configuration
type CacheFileConfig struct {
	Enabled *bool  `json:"enabled,omitempty"`
	Path    string `json:"path,omitempty"`
	CacheID string `json:"cache_id,omitempty"`
}

// ClashAPIConfig represents the clash_api experimental configuration
type ClashAPIConfig struct {
	ExternalController            string   `json:"external_controller,omitempty"`
	ExternalUI                    string   `json:"external_ui,omitempty"`
	ExternalUIDownloadURL         string   `json:"external_ui_download_url,omitempty"`
	ExternalUIDownloadDetour      string   `json:"external_ui_download_detour,omitempty"`
	AccessControlAllowOrigin      []string `json:"access_control_allow_origin,omitempty"`
	AccessControlAllowPrivateNetwork *bool `json:"access_control_allow_private_network,omitempty"`
}

// ExperimentalConfig represents the experimental configuration
type ExperimentalConfig struct {
	CacheFile *CacheFileConfig `json:"cache_file,omitempty"`
	ClashAPI  *ClashAPIConfig  `json:"clash_api,omitempty"`
}

// SingBoxConfig represents the full sing-box configuration
type SingBoxConfig struct {
	Inbounds     []Inbound          `json:"inbounds"`
	Outbounds    []Outbound         `json:"outbounds"`
	Rulesets     []Ruleset          `json:"rulesets"`
	RouteConfig  *RouteConfig       `json:"route_config"`
	RouteRules   []RouteRule        `json:"route_rules"`
	DNS          *DNSConfig         `json:"dns"`
	Services     []Service          `json:"services"`
	HTTPClients  []HTTPClient       `json:"http_clients"`
	Experimental *ExperimentalConfig `json:"experimental,omitempty"`
}

// ConfigRequest represents a config update request
type ConfigRequest struct {
	Action string      `json:"action"` // "add", "update", "delete"
	Type   string      `json:"type"`   // "inbound", "outbound"
	Data   interface{} `json:"data"`
}

// InboundRequest represents an inbound CRUD request
type InboundRequest struct {
	Action string  `json:"action"` // "add", "update", "delete"
	Data   Inbound `json:"data"`
}

// OutboundRequest represents an outbound CRUD request
type OutboundRequest struct {
	Action string   `json:"action"` // "add", "update", "delete"
	Data   Outbound `json:"data"`
}

// RulesetRequest represents a ruleset CRUD request
type RulesetRequest struct {
	Action string   `json:"action"` // "add", "update", "delete"
	Data   Ruleset  `json:"data"`
}
