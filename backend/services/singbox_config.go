package services

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"sing_panel/models"

	"github.com/google/uuid"
)

type SingBoxConfigService struct {
	db              *Database
	geoTreeCache    map[string]interface{}
	geoTreeTime     time.Time
	commonTreeCache map[string]interface{}
	commonTreeTime  time.Time
}

func NewSingBoxConfigService(db *Database) *SingBoxConfigService {
	s := &SingBoxConfigService{
		db: db,
	}
	// Load cached trees from DB on startup
	s.loadTreeCacheFromDB()
	return s
}

// Tree cache types
type treeCacheEntry struct {
	Data      map[string]interface{} `json:"data"`
	UpdatedAt time.Time              `json:"updated_at"`
}

const treeCacheTTL = 1 * time.Hour

// loadTreeCacheFromDB loads cached trees from database on startup
func (s *SingBoxConfigService) loadTreeCacheFromDB() {
	var geoEntry treeCacheEntry
	if err := s.db.Get("singbox", "geo_tree_cache", &geoEntry); err == nil {
		s.geoTreeCache = geoEntry.Data
		s.geoTreeTime = geoEntry.UpdatedAt
		slog.Info("geo tree cache loaded from db", "updated_at", geoEntry.UpdatedAt.Format("15:04:05"))
	}
	var commonEntry treeCacheEntry
	if err := s.db.Get("singbox", "common_tree_cache", &commonEntry); err == nil {
		s.commonTreeCache = commonEntry.Data
		s.commonTreeTime = commonEntry.UpdatedAt
		slog.Info("common ruleset tree cache loaded from db", "updated_at", commonEntry.UpdatedAt.Format("15:04:05"))
	}
}

func (s *SingBoxConfigService) saveTreeCacheToDB(key string, data map[string]interface{}, updatedAt time.Time) error {
	entry := treeCacheEntry{Data: data, UpdatedAt: updatedAt}
	return s.db.Put("singbox", key, entry)
}

func (s *SingBoxConfigService) fetchGitHubTree(url string) (map[string]interface{}, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回状态码: %d", resp.StatusCode)
	}
	body = bytes.TrimSpace(body)
	if len(body) == 0 || body[0] != '{' {
		return nil, fmt.Errorf("返回了非 JSON 响应（可能触发了速率限制）")
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	return data, nil
}

// RefreshGeoTree fetches and caches the GeoIP tree
func (s *SingBoxConfigService) RefreshGeoTree() {
	data, err := s.fetchGitHubTree("https://api.github.com/repos/SagerNet/sing-geoip/git/trees/rule-set")
	if err != nil {
		slog.Error("refresh geo tree failed", "error", err)
		return
	}
	s.geoTreeCache = data
	s.geoTreeTime = time.Now()
	s.saveTreeCacheToDB("geo_tree_cache", data, s.geoTreeTime)
	slog.Info("geo tree refreshed", "count", len(data["tree"].([]interface{})))
}

// RefreshCommonRulesetTree fetches and caches the common ruleset tree
func (s *SingBoxConfigService) RefreshCommonRulesetTree() {
	data, err := s.fetchGitHubTree("https://api.github.com/repos/DustinWin/ruleset_geodata/git/trees/sing-box-ruleset")
	if err != nil {
		slog.Error("refresh common ruleset tree failed", "error", err)
		return
	}
	s.commonTreeCache = data
	s.commonTreeTime = time.Now()
	s.saveTreeCacheToDB("common_tree_cache", data, s.commonTreeTime)
	slog.Info("common ruleset tree refreshed", "count", len(data["tree"].([]interface{})))
}

// RefreshAllTrees refreshes both geo and common trees
func (s *SingBoxConfigService) RefreshAllTrees() {
	s.RefreshGeoTree()
	s.RefreshCommonRulesetTree()
}

// GetGeoTree returns the cached geo tree, refreshing if expired
func (s *SingBoxConfigService) GetGeoTree() (map[string]interface{}, error) {
	if s.geoTreeCache != nil && time.Since(s.geoTreeTime) < treeCacheTTL {
		return s.geoTreeCache, nil
	}
	// Try to refresh
	data, err := s.fetchGitHubTree("https://api.github.com/repos/SagerNet/sing-geoip/git/trees/rule-set")
	if err != nil {
		// Return stale cache if available
		if s.geoTreeCache != nil {
			slog.Warn("returning stale geo tree cache", "age", time.Since(s.geoTreeTime))
			return s.geoTreeCache, nil
		}
		return nil, fmt.Errorf("%v", err)
	}
	s.geoTreeCache = data
	s.geoTreeTime = time.Now()
	s.saveTreeCacheToDB("geo_tree_cache", data, s.geoTreeTime)
	return data, nil
}

// GetCommonRulesetTree returns the cached common ruleset tree, refreshing if expired
func (s *SingBoxConfigService) GetCommonRulesetTree() (map[string]interface{}, error) {
	if s.commonTreeCache != nil && time.Since(s.commonTreeTime) < treeCacheTTL {
		return s.commonTreeCache, nil
	}
	// Try to refresh
	data, err := s.fetchGitHubTree("https://api.github.com/repos/DustinWin/ruleset_geodata/git/trees/sing-box-ruleset")
	if err != nil {
		// Return stale cache if available
		if s.commonTreeCache != nil {
			slog.Warn("returning stale common ruleset tree cache", "age", time.Since(s.commonTreeTime))
			return s.commonTreeCache, nil
		}
		return nil, fmt.Errorf("%v", err)
	}
	s.commonTreeCache = data
	s.commonTreeTime = time.Now()
	s.saveTreeCacheToDB("common_tree_cache", data, s.commonTreeTime)
	return data, nil
}

// StartTreeRefreshLoop starts a background goroutine to refresh trees periodically
func (s *SingBoxConfigService) StartTreeRefreshLoop() {
	go func() {
		// Initial refresh on startup
		s.RefreshAllTrees()

		ticker := time.NewTicker(treeCacheTTL)
		defer ticker.Stop()
		for range ticker.C {
			s.RefreshAllTrees()
		}
	}()
}

// GetConfig returns the full sing-box configuration
func (s *SingBoxConfigService) GetConfig() (models.SingBoxConfig, error) {
	var config models.SingBoxConfig

	if err := s.db.Get("singbox", "config", &config); err != nil {
		// Return empty config if not exists
		return models.SingBoxConfig{
			Inbounds:   []models.Inbound{},
			Outbounds:  []models.Outbound{},
			Rulesets:   []models.Ruleset{},
			RouteRules: []models.RouteRule{},
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

// AddRuleset adds a new ruleset configuration
func (s *SingBoxConfigService) AddRuleset(ruleset models.Ruleset) (models.Ruleset, error) {
	config, err := s.GetConfig()
	if err != nil {
		return models.Ruleset{}, err
	}

	if ruleset.ID == "" {
		ruleset.ID = uuid.New().String()
	}

	config.Rulesets = append(config.Rulesets, ruleset)

	if err := s.SaveConfig(config); err != nil {
		return models.Ruleset{}, err
	}

	return ruleset, nil
}

// AddRulesets adds multiple ruleset configurations
func (s *SingBoxConfigService) AddRulesets(rulesets []models.Ruleset) ([]models.Ruleset, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	for i := range rulesets {
		if rulesets[i].ID == "" {
			rulesets[i].ID = uuid.New().String()
		}
	}

	config.Rulesets = append(config.Rulesets, rulesets...)

	if err := s.SaveConfig(config); err != nil {
		return nil, err
	}

	return rulesets, nil
}

// UpdateRuleset updates an existing ruleset configuration
func (s *SingBoxConfigService) UpdateRuleset(ruleset models.Ruleset) (models.Ruleset, error) {
	if ruleset.ID == "" {
		return models.Ruleset{}, fmt.Errorf("ruleset ID is required")
	}

	config, err := s.GetConfig()
	if err != nil {
		return models.Ruleset{}, err
	}

	for i, item := range config.Rulesets {
		if item.ID == ruleset.ID {
			config.Rulesets[i] = ruleset
			if err := s.SaveConfig(config); err != nil {
				return models.Ruleset{}, err
			}
			return ruleset, nil
		}
	}

	return models.Ruleset{}, fmt.Errorf("ruleset not found")
}

// DeleteRuleset deletes a ruleset configuration
func (s *SingBoxConfigService) DeleteRuleset(id string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	for i, item := range config.Rulesets {
		if item.ID == id {
			config.Rulesets = append(config.Rulesets[:i], config.Rulesets[i+1:]...)
			return s.SaveConfig(config)
		}
	}

	return fmt.Errorf("ruleset not found")
}

// DeleteRulesets deletes multiple ruleset configurations
func (s *SingBoxConfigService) DeleteRulesets(ids []string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	var remaining []models.Ruleset
	for _, item := range config.Rulesets {
		if !idSet[item.ID] {
			remaining = append(remaining, item)
		}
	}

	config.Rulesets = remaining
	return s.SaveConfig(config)
}

// AddRouteRule adds a new route rule configuration
func (s *SingBoxConfigService) AddRouteRule(rule models.RouteRule) (models.RouteRule, error) {
	config, err := s.GetConfig()
	if err != nil {
		return models.RouteRule{}, err
	}

	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}

	config.RouteRules = append(config.RouteRules, rule)

	if err := s.SaveConfig(config); err != nil {
		return models.RouteRule{}, err
	}

	return rule, nil
}

// UpdateRouteRule updates an existing route rule configuration
func (s *SingBoxConfigService) UpdateRouteRule(rule models.RouteRule) (models.RouteRule, error) {
	if rule.ID == "" {
		return models.RouteRule{}, fmt.Errorf("route rule ID is required")
	}

	config, err := s.GetConfig()
	if err != nil {
		return models.RouteRule{}, err
	}

	for i, item := range config.RouteRules {
		if item.ID == rule.ID {
			config.RouteRules[i] = rule
			if err := s.SaveConfig(config); err != nil {
				return models.RouteRule{}, err
			}
			return rule, nil
		}
	}

	return models.RouteRule{}, fmt.Errorf("route rule not found")
}

// DeleteRouteRule deletes a route rule configuration
func (s *SingBoxConfigService) DeleteRouteRule(id string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	for i, item := range config.RouteRules {
		if item.ID == id {
			config.RouteRules = append(config.RouteRules[:i], config.RouteRules[i+1:]...)
			return s.SaveConfig(config)
		}
	}

	return fmt.Errorf("route rule not found")
}

// ReorderRouteRules reorders route rules by the given ID list
func (s *SingBoxConfigService) ReorderRouteRules(ids []string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	idMap := make(map[string]models.RouteRule)
	for _, rule := range config.RouteRules {
		idMap[rule.ID] = rule
	}

	reordered := make([]models.RouteRule, 0, len(ids))
	for _, id := range ids {
		if rule, ok := idMap[id]; ok {
			reordered = append(reordered, rule)
		}
	}

	config.RouteRules = reordered
	return s.SaveConfig(config)
}

// GetRouteConfig returns the route configuration
func (s *SingBoxConfigService) GetRouteConfig() (*models.RouteConfig, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	if config.RouteConfig == nil {
		return &models.RouteConfig{}, nil
	}
	return config.RouteConfig, nil
}

// UpdateRouteConfig updates the route configuration
func (s *SingBoxConfigService) UpdateRouteConfig(rc models.RouteConfig) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.RouteConfig = &rc
	return s.SaveConfig(config)
}

// GetDNS returns the DNS configuration
func (s *SingBoxConfigService) GetDNS() (*models.DNSConfig, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	if config.DNS == nil {
		return &models.DNSConfig{
			Servers: []models.DNSServer{},
			Rules:   []models.DNSRule{},
			FakeIP:  map[string]interface{}{},
		}, nil
	}
	return config.DNS, nil
}

// UpdateDNS updates the DNS configuration
func (s *SingBoxConfigService) UpdateDNS(dns models.DNSConfig) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.DNS = &dns
	return s.SaveConfig(config)
}

// GetServices returns all service configurations
func (s *SingBoxConfigService) GetServices() ([]models.Service, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	return config.Services, nil
}

// AddService adds a new service configuration
func (s *SingBoxConfigService) AddService(svc models.Service) (models.Service, error) {
	config, err := s.GetConfig()
	if err != nil {
		return models.Service{}, err
	}

	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}

	config.Services = append(config.Services, svc)

	if err := s.SaveConfig(config); err != nil {
		return models.Service{}, err
	}

	return svc, nil
}

// UpdateService updates an existing service configuration
func (s *SingBoxConfigService) UpdateService(svc models.Service) (models.Service, error) {
	if svc.ID == "" {
		return models.Service{}, fmt.Errorf("service ID is required")
	}

	config, err := s.GetConfig()
	if err != nil {
		return models.Service{}, err
	}

	for i, item := range config.Services {
		if item.ID == svc.ID {
			config.Services[i] = svc
			if err := s.SaveConfig(config); err != nil {
				return models.Service{}, err
			}
			return svc, nil
		}
	}

	return models.Service{}, fmt.Errorf("service not found")
}

// DeleteService deletes a service configuration
func (s *SingBoxConfigService) DeleteService(id string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	for i, item := range config.Services {
		if item.ID == id {
			config.Services = append(config.Services[:i], config.Services[i+1:]...)
			return s.SaveConfig(config)
		}
	}

	return fmt.Errorf("service not found")
}

// GetHTTPClients returns all HTTP client configurations
func (s *SingBoxConfigService) GetHTTPClients() ([]models.HTTPClient, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	return config.HTTPClients, nil
}

// AddHTTPClient adds a new HTTP client configuration
func (s *SingBoxConfigService) AddHTTPClient(client models.HTTPClient) (models.HTTPClient, error) {
	config, err := s.GetConfig()
	if err != nil {
		return models.HTTPClient{}, err
	}

	if client.ID == "" {
		client.ID = uuid.New().String()
	}

	config.HTTPClients = append(config.HTTPClients, client)

	if err := s.SaveConfig(config); err != nil {
		return models.HTTPClient{}, err
	}

	return client, nil
}

// UpdateHTTPClient updates an existing HTTP client configuration
func (s *SingBoxConfigService) UpdateHTTPClient(client models.HTTPClient) (models.HTTPClient, error) {
	if client.ID == "" {
		return models.HTTPClient{}, fmt.Errorf("http client ID is required")
	}

	config, err := s.GetConfig()
	if err != nil {
		return models.HTTPClient{}, err
	}

	for i, item := range config.HTTPClients {
		if item.ID == client.ID {
			config.HTTPClients[i] = client
			if err := s.SaveConfig(config); err != nil {
				return models.HTTPClient{}, err
			}
			return client, nil
		}
	}

	return models.HTTPClient{}, fmt.Errorf("http client not found")
}

// DeleteHTTPClient deletes an HTTP client configuration
func (s *SingBoxConfigService) DeleteHTTPClient(id string) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}

	for i, item := range config.HTTPClients {
		if item.ID == id {
			config.HTTPClients = append(config.HTTPClients[:i], config.HTTPClients[i+1:]...)
			return s.SaveConfig(config)
		}
	}

	return fmt.Errorf("http client not found")
}

// GetExperimental returns the experimental configuration
func (s *SingBoxConfigService) GetExperimental() (*models.ExperimentalConfig, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}
	if config.Experimental == nil {
		return &models.ExperimentalConfig{}, nil
	}
	return config.Experimental, nil
}

// UpdateExperimental updates the experimental configuration
func (s *SingBoxConfigService) UpdateExperimental(exp models.ExperimentalConfig) error {
	config, err := s.GetConfig()
	if err != nil {
		return err
	}
	config.Experimental = &exp
	return s.SaveConfig(config)
}

// ExportConfig exports the sing-box configuration as JSON for the kernel
func (s *SingBoxConfigService) ExportConfig() (map[string]interface{}, error) {
	config, err := s.GetConfig()
	if err != nil {
		return nil, err
	}

	// Build the sing-box config format
	routeSection := map[string]interface{}{
		"rule_set": s.buildRulesets(config.Rulesets),
		"rules":    s.buildRouteRules(config.RouteRules),
	}
	if config.RouteConfig != nil && config.RouteConfig.Final != "" {
		routeSection["final"] = config.RouteConfig.Final
	}
	if config.RouteConfig != nil && config.RouteConfig.DefaultHttpClient != "" {
		routeSection["default_http_client"] = config.RouteConfig.DefaultHttpClient
	}

	result := map[string]interface{}{
		"log": map[string]interface{}{
			"level": "info",
		},
		"inbounds":  s.buildInbounds(config.Inbounds),
		"outbounds": s.buildOutbounds(config.Outbounds),
		"route":     routeSection,
	}

	// Build experimental section
	if config.Experimental != nil {
		experimental := map[string]interface{}{}
		if config.Experimental.CacheFile != nil {
			cacheFile := map[string]interface{}{}
			cf := config.Experimental.CacheFile
			if cf.Enabled != nil {
				cacheFile["enabled"] = *cf.Enabled
			}
			if cf.Path != "" {
				cacheFile["path"] = cf.Path
			}
			if cf.CacheID != "" {
				cacheFile["cache_id"] = cf.CacheID
			}
			experimental["cache_file"] = cacheFile
		}
		if config.Experimental.ClashAPI != nil {
			clashAPI := map[string]interface{}{}
			ca := config.Experimental.ClashAPI
			if ca.ExternalController != "" {
				clashAPI["external_controller"] = ca.ExternalController
			}
			if ca.ExternalUI != "" {
				clashAPI["external_ui"] = ca.ExternalUI
			}
			if ca.ExternalUIDownloadURL != "" {
				clashAPI["external_ui_download_url"] = ca.ExternalUIDownloadURL
			}
			if ca.ExternalUIDownloadDetour != "" {
				clashAPI["external_ui_download_detour"] = ca.ExternalUIDownloadDetour
			}
			if len(ca.AccessControlAllowOrigin) > 0 {
				clashAPI["access_control_allow_origin"] = ca.AccessControlAllowOrigin
			}
			if ca.AccessControlAllowPrivateNetwork != nil {
				clashAPI["access_control_allow_private_network"] = *ca.AccessControlAllowPrivateNetwork
			}
			experimental["clash_api"] = clashAPI
		}
		if len(experimental) > 0 {
			result["experimental"] = experimental
		}
	} else {
		result["experimental"] = map[string]interface{}{
			"cache_file": map[string]interface{}{},
		}
	}

	if config.DNS != nil {
		result["dns"] = s.buildDNS(config.DNS)
	}

	if len(config.Services) > 0 {
		result["services"] = s.buildServices(config.Services)
	}

	if len(config.HTTPClients) > 0 {
		result["http_clients"] = s.buildHTTPClients(config.HTTPClients)
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
			"type": inbound.Type,
			"tag":  inbound.Tag,
		}
		if inbound.Type != "tun" {
			item["listen"] = inbound.Listen
			item["listen_port"] = inbound.ListenPort
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

func (s *SingBoxConfigService) buildRulesets(rulesets []models.Ruleset) []map[string]interface{} {
	var result []map[string]interface{}
	for _, ruleset := range rulesets {
		if !ruleset.Enabled {
			continue
		}
		item := map[string]interface{}{
			"type": ruleset.Type,
			"tag":  ruleset.Tag,
		}

		switch ruleset.Type {
		case "remote":
			if v, ok := ruleset.Options["url"]; ok {
				item["url"] = v
			}
			if v, ok := ruleset.Options["format"]; ok {
				item["format"] = v
			}
			if v, ok := ruleset.Options["download_detour"]; ok {
				item["download_detour"] = v
			}
			if v, ok := ruleset.Options["update_interval"]; ok {
				item["update_interval"] = v
			}
		case "local":
			if v, ok := ruleset.Options["path"]; ok {
				item["path"] = v
			}
			if v, ok := ruleset.Options["format"]; ok {
				item["format"] = v
			}
		case "inline":
			// Inline: rules are stored directly in config
			if v, ok := ruleset.Options["rules"]; ok {
				item["rules"] = v
			}
		}

		result = append(result, item)
	}
	return result
}

func contains(arr []string, s string) bool {
	for _, v := range arr {
		if v == s {
			return true
		}
	}
	return false
}

func (s *SingBoxConfigService) buildRouteRules(rules []models.RouteRule) []map[string]interface{} {
	var result []map[string]interface{}
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		item := map[string]interface{}{}

		// Merge options first (allows overriding base fields via options)
		for k, v := range rule.Options {
			item[k] = v
		}

		// Set action
		if rule.Action != "" {
			item["action"] = rule.Action
		}

		// Set outbound for route action
		if rule.Action == "route" && rule.Outbound != "" {
			item["outbound"] = rule.Outbound
		}

		// Set inbound (only if not already in options)
		if len(rule.Inbound) > 0 {
			item["inbound"] = rule.Inbound
		}

		// Set rule_set (only if not already in options)
		if len(rule.RuleSet) > 0 {
			item["rule_set"] = rule.RuleSet
		}

		result = append(result, item)
	}
	return result
}

func (s *SingBoxConfigService) buildDNS(dns *models.DNSConfig) map[string]interface{} {
	result := map[string]interface{}{}

	if len(dns.Servers) > 0 {
		var servers []map[string]interface{}
		for _, srv := range dns.Servers {
			item := map[string]interface{}{
				"type": srv.Type,
				"tag":  srv.Tag,
			}
			servers = append(servers, item)
		}
		result["servers"] = servers
	}

	if len(dns.Rules) > 0 {
		var rules []map[string]interface{}
		for _, r := range dns.Rules {
			item := map[string]interface{}{
				"rule_set": r.RuleSet,
				"server":   r.Server,
			}
			if r.Disabled {
				item["disabled"] = true
			}
			rules = append(rules, item)
		}
		result["rules"] = rules
	}

	if dns.Final != "" {
		result["final"] = dns.Final
	}
	if dns.Strategy != "" {
		result["strategy"] = dns.Strategy
	}
	if dns.DisableCache {
		result["disable_cache"] = true
	}
	if dns.DisableExpire {
		result["disable_expire"] = true
	}
	if dns.IndependentCache {
		result["independent_cache"] = true
	}
	if dns.CacheCapacity > 0 {
		result["cache_capacity"] = dns.CacheCapacity
	}
	if dns.Optimistic != nil {
		result["optimistic"] = dns.Optimistic
	}
	if dns.Timeout != "" {
		result["timeout"] = dns.Timeout
	}
	if dns.ReverseMapping {
		result["reverse_mapping"] = true
	}
	if dns.ClientSubnet != "" {
		result["client_subnet"] = dns.ClientSubnet
	}
	if len(dns.FakeIP) > 0 {
		result["fakeip"] = dns.FakeIP
	}

	return result
}

func (s *SingBoxConfigService) buildServices(services []models.Service) []map[string]interface{} {
	var result []map[string]interface{}
	for _, svc := range services {
		if !svc.Enabled {
			continue
		}
		item := map[string]interface{}{
			"type": svc.Type,
			"tag":  svc.Tag,
		}
		if svc.Listen != "" {
			item["listen"] = svc.Listen
		}
		if svc.ListenPort > 0 {
			item["listen_port"] = svc.ListenPort
		}
		for k, v := range svc.Options {
			item[k] = v
		}
		result = append(result, item)
	}
	return result
}

func (s *SingBoxConfigService) buildHTTPClients(clients []models.HTTPClient) []map[string]interface{} {
	var result []map[string]interface{}
	for _, client := range clients {
		item := map[string]interface{}{
			"tag": client.Tag,
		}
		for k, v := range client.Options {
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
