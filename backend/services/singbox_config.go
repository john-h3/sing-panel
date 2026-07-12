package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sing_panel/models"

	"github.com/google/uuid"
)

type SingBoxConfigService struct {
	db *Database
}

func NewSingBoxConfigService(db *Database) *SingBoxConfigService {
	return &SingBoxConfigService{db: db}
}

// GetConfig returns the full sing-box configuration
func (s *SingBoxConfigService) GetConfig() (models.SingBoxConfig, error) {
	var config models.SingBoxConfig

	if err := s.db.Get("singbox", "config", &config); err != nil {
		// Return empty config if not exists
		return models.SingBoxConfig{
			Inbounds:  []models.Inbound{},
			Outbounds: []models.Outbound{},
		}, nil
	}

	return config, nil
}

// SaveConfig saves the full sing-box configuration
func (s *SingBoxConfigService) SaveConfig(config models.SingBoxConfig) error {
	return s.db.Put("singbox", "config", config)
}

// AddInbound adds a new inbound configuration
func (s *SingBoxConfigService) AddInbound(inbound models.Inbound) (models.Inbound, error) {
	config, err := s.GetConfig()
	if err != nil {
		return models.Inbound{}, err
	}

	// Generate ID if not provided
	if inbound.ID == "" {
		inbound.ID = uuid.New().String()
	}

	config.Inbounds = append(config.Inbounds, inbound)

	if err := s.SaveConfig(config); err != nil {
		return models.Inbound{}, err
	}

	return inbound, nil
}

// UpdateInbound updates an existing inbound configuration
func (s *SingBoxConfigService) UpdateInbound(inbound models.Inbound) (models.Inbound, error) {
	if inbound.ID == "" {
		return models.Inbound{}, fmt.Errorf("inbound ID is required")
	}

	config, err := s.GetConfig()
	if err != nil {
		return models.Inbound{}, err
	}

	for i, item := range config.Inbounds {
		if item.ID == inbound.ID {
			config.Inbounds[i] = inbound
			if err := s.SaveConfig(config); err != nil {
				return models.Inbound{}, err
			}
			return inbound, nil
		}
	}

	return models.Inbound{}, fmt.Errorf("inbound not found")
}

// DeleteInbound deletes an inbound configuration
func (s *SingBoxConfigService) DeleteInbound(id string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	for i, item := range config.Inbounds {
		if item.ID == id {
			config.Inbounds = append(config.Inbounds[:i], config.Inbounds[i+1:]...)
			return s.SaveConfig(config)
		}
	}

	return fmt.Errorf("inbound not found")
}

// AddOutbound adds a new outbound configuration
func (s *SingBoxConfigService) AddOutbound(outbound models.Outbound) (models.Outbound, error) {
	config, err := s.GetConfig()
	if err != nil {
		return models.Outbound{}, err
	}

	// Generate ID if not provided
	if outbound.ID == "" {
		outbound.ID = uuid.New().String()
	}

	config.Outbounds = append(config.Outbounds, outbound)

	if err := s.SaveConfig(config); err != nil {
		return models.Outbound{}, err
	}

	return outbound, nil
}

// UpdateOutbound updates an existing outbound configuration
func (s *SingBoxConfigService) UpdateOutbound(outbound models.Outbound) (models.Outbound, error) {
	if outbound.ID == "" {
		return models.Outbound{}, fmt.Errorf("outbound ID is required")
	}

	config, err := s.GetConfig()
	if err != nil {
		return models.Outbound{}, err
	}

	for i, item := range config.Outbounds {
		if item.ID == outbound.ID {
			config.Outbounds[i] = outbound
			if err := s.SaveConfig(config); err != nil {
				return models.Outbound{}, err
			}
			return outbound, nil
		}
	}

	return models.Outbound{}, fmt.Errorf("outbound not found")
}

// DeleteOutbound deletes an outbound configuration
func (s *SingBoxConfigService) DeleteOutbound(id string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	for i, item := range config.Outbounds {
		if item.ID == id {
			config.Outbounds = append(config.Outbounds[:i], config.Outbounds[i+1:]...)
			return s.SaveConfig(config)
		}
	}

	return fmt.Errorf("outbound not found")
}

// ExportConfig exports the sing-box configuration as JSON for the kernel
func (s *SingBoxConfigService) ExportConfig() (map[string]interface{}, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	// Build the sing-box config format
	result := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "info",
		},
		"inbounds":  s.buildInbounds(config.Inbounds),
		"outbounds": s.buildOutbounds(config.Outbounds),
		"experimental": map[string]interface{}{
			"cache_file": map[string]interface{}{},
			"clash_api": map[string]interface{}{
				"external_controller": "127.0.0.1:9090",
				"external_ui":        "",
				"secret":             "",
				"default_mode":       "",
			},
		},
	}

	return result, nil
}

func (s *SingBoxConfigService) buildInbounds(inbounds []models.Inbound) []map[string]interface{} {
	var result []map[string]interface{}
	for _, inbound := range inbounds {
		if !inbound.Enabled {
			continue
		}
		item := map[string]interface{}{
			"type":       inbound.Type,
			"tag":        inbound.Tag,
			"listen":     inbound.Listen,
			"listen_port": inbound.ListenPort,
		}
		// Merge custom options
		for k, v := range inbound.Options {
			item[k] = v
		}
		result = append(result, item)
	}
	return result
}

func (s *SingBoxConfigService) buildOutbounds(outbounds []models.Outbound) []map[string]interface{} {
	var result []map[string]interface{}
	for _, outbound := range outbounds {
		if !outbound.Enabled {
			continue
		}
		item := map[string]interface{}{
			"type": outbound.Type,
			"tag":  outbound.Tag,
		}
		// Merge custom options
		for k, v := range outbound.Options {
			item[k] = v
		}
		result = append(result, item)
	}
	return result
}

// GetDefaultInboundTypes returns available inbound types
func (s *SingBoxConfigService) GetDefaultInboundTypes() []map[string]string {
	return []map[string]string{
		{"type": "http", "name": "HTTP 代理"},
		{"type": "socks", "name": "SOCKS5 代理"},
		{"type": "mixed", "name": "Mixed 代理"},
		{"type": "tun", "name": "TUN 网卡"},
		{"type": "shadowsocks", "name": "Shadowsocks 服务端"},
	}
}

// GetDefaultOutboundTypes returns available outbound types
func (s *SingBoxConfigService) GetDefaultOutboundTypes() []map[string]string {
	return []map[string]string{
		{"type": "direct", "name": "直连"},
		{"type": "block", "name": "阻断"},
		{"type": "selector", "name": "选择器"},
		{"type": "urltest", "name": "自动测速"},
		{"type": "shadowsocks", "name": "Shadowsocks"},
		{"type": "vmess", "name": "VMess"},
		{"type": "vless", "name": "VLESS"},
		{"type": "trojan", "name": "Trojan"},
		{"type": "hysteria", "name": "Hysteria"},
	}
}

// GetTimestamp returns current timestamp for config versioning
func (s *SingBoxConfigService) GetTimestamp() int64 {
	return time.Now().Unix()
}

// ImportVMess imports a VMess link and returns an Outbound
func (s *SingBoxConfigService) ImportVMess(link string) (models.Outbound, error) {
	// VMess format: vmess://base64_encoded_json
	if len(link) < 9 || link[:8] != "vmess://" {
		return models.Outbound{}, fmt.Errorf("invalid vmess link format")
	}

	// Decode base64
	b64Data := link[8:]
	decoded, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		// Try URL-safe base64
		decoded, err = base64.RawStdEncoding.DecodeString(b64Data)
		if err != nil {
			return models.Outbound{}, fmt.Errorf("failed to decode vmess link: %w", err)
		}
	}

	// Parse JSON
	var vmessInfo map[string]interface{}
	if err := json.Unmarshal(decoded, &vmessInfo); err != nil {
		return models.Outbound{}, fmt.Errorf("failed to parse vmess json: %w", err)
	}

	// Extract fields
	server, _ := vmessInfo["add"].(string)
	port := 0
	if p, ok := vmessInfo["port"].(float64); ok {
		port = int(p)
	}
	userID, _ := vmessInfo["id"].(string)
	aid := 0
	if a, ok := vmessInfo["aid"].(float64); ok {
		aid = int(a)
	}
	net, _ := vmessInfo["net"].(string)
	host, _ := vmessInfo["host"].(string)
	path, _ := vmessInfo["path"].(string)
	tls, _ := vmessInfo["tls"].(string)
	sni, _ := vmessInfo["sni"].(string)
Remarks, _ := vmessInfo["ps"].(string)

	if server == "" || port == 0 || userID == "" {
		return models.Outbound{}, fmt.Errorf("missing required vmess fields (server, port, id)")
	}

	// Build tag
	tag := Remarks
	if tag == "" {
		tag = fmt.Sprintf("vmess-%s:%d", server, port)
	}

	// Build options
	options := map[string]interface{}{
		"server":     server,
		"server_port": port,
		"uuid":       userID,
	}

	if aid > 0 {
		options["alter_id"] = aid
	}

	// Transport
	if net != "" && net != "tcp" {
		transport := map[string]interface{}{
			"type": net,
		}
		if host != "" {
			transport["host"] = []string{host}
		}
		if path != "" {
			transport["path"] = path
		}
		options["transport"] = transport
	}

	// TLS
	if tls == "tls" {
		options["tls"] = map[string]interface{}{
			"enabled": true,
			"server_name": sni,
		}
	}

	return models.Outbound{
		ID:      uuid.New().String(),
		Type:    "vmess",
		Tag:     tag,
		Options: options,
		Enabled: true,
	}, nil
}

// ImportVLESS imports a VLESS link (vless://uuid@server:port?params#remark)
func (s *SingBoxConfigService) ImportVLESS(link string) (models.Outbound, error) {
	// VLESS format: vless://uuid@server:port?params#remark
	if !strings.HasPrefix(link, "vless://") {
		return models.Outbound{}, fmt.Errorf("invalid vless link format")
	}

	// Remove scheme
	rest := link[8:]

	// Split fragment
	parts := strings.SplitN(rest, "#", 2)
	remark := ""
	if len(parts) == 2 {
		remark = parts[1]
	}

	// Split query
	hostParams := strings.SplitN(parts[0], "?", 2)
	queryStr := ""
	if len(hostParams) == 2 {
		queryStr = hostParams[1]
	}

	// Parse query params
	params := make(map[string]string)
	if queryStr != "" {
		for _, param := range strings.Split(queryStr, "&") {
			kv := strings.SplitN(param, "=", 2)
			if len(kv) == 2 {
				params[kv[0]] = kv[1]
			}
		}
	}

	// Parse uuid@server:port
	addrParts := strings.SplitN(hostParams[0], "@", 2)
	if len(addrParts) != 2 {
		return models.Outbound{}, fmt.Errorf("invalid vless format: missing @")
	}

	userID := addrParts[0]
	serverAddr := addrParts[1]

	// Split server:port
	serverParts := strings.SplitN(serverAddr, ":", 2)
	server := serverParts[0]
	port := 443
	if len(serverParts) == 2 {
		fmt.Sscanf(serverParts[1], "%d", &port)
	}

	// Extract params
	security := params["security"]
	flow := params["flow"]
	sni := params["sni"]
	pbk := params["pbk"]
	sid := params["sid"]
	fp := params["fp"]
	alpn := params["alpn"]
	type_ := params["type"]
	host := params["host"]
	path := params["path"]
	serviceName := params["serviceName"]
	headerType := params["headerType"]

	// Build tag
	tag := remark
	if tag == "" {
		tag = fmt.Sprintf("vless-%s:%d", server, port)
	}

	// Build options (note: encryption is not used in sing-box outbound config)
	options := map[string]interface{}{
		"server":      server,
		"server_port": port,
		"uuid":        userID,
		"flow":        flow,
	}

	// TLS / Reality
	if security == "reality" {
		realityOpts := map[string]interface{}{
			"enabled":    true,
			"public_key": pbk,
		}
		if sid != "" {
			realityOpts["short_id"] = sid
		}

		tlsOpts := map[string]interface{}{
			"enabled":     true,
			"server_name": sni,
			"reality":     realityOpts,
		}
		if fp != "" {
			tlsOpts["utls"] = map[string]interface{}{
				"enabled":     true,
				"fingerprint": fp,
			}
		}
		if alpn != "" {
			tlsOpts["alpn"] = strings.Split(alpn, ",")
		}
		options["tls"] = tlsOpts
	} else if security == "tls" {
		tlsOpts := map[string]interface{}{
			"enabled":     true,
			"server_name": sni,
		}
		if fp != "" {
			tlsOpts["utls"] = map[string]interface{}{
				"enabled":     true,
				"fingerprint": fp,
			}
		}
		if alpn != "" {
			tlsOpts["alpn"] = strings.Split(alpn, ",")
		}
		options["tls"] = tlsOpts
	}

	// Transport (tcp is default, no need to set)
	if type_ != "" && type_ != "tcp" {
		transport := map[string]interface{}{
			"type": type_,
		}
		switch type_ {
		case "ws":
			if host != "" {
				transport["host"] = []string{host}
			}
			if path != "" {
				transport["path"] = path
			}
		case "grpc":
			if serviceName != "" {
				transport["service_name"] = serviceName
			}
		case "h2", "http":
			if host != "" {
				transport["host"] = host
			}
			if path != "" {
				transport["path"] = path
			}
		case "quic":
			if headerType != "" {
				transport["header"] = map[string]interface{}{
					"type": headerType,
				}
			}
		}
		options["transport"] = transport
	}

	return models.Outbound{
		ID:      uuid.New().String(),
		Type:    "vless",
		Tag:     tag,
		Options: options,
		Enabled: true,
	}, nil
}

// ImportLink auto-detects link type and imports
func (s *SingBoxConfigService) ImportLink(link string) (models.Outbound, error) {
	link = strings.TrimSpace(link)

	if strings.HasPrefix(link, "vmess://") {
		return s.ImportVMess(link)
	}
	if strings.HasPrefix(link, "vless://") {
		return s.ImportVLESS(link)
	}

	return models.Outbound{}, fmt.Errorf("unsupported link format, supported: vmess://, vless://")
}
