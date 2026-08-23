package services

import (
	"log/slog"
	"strings"
	"sync"

	"sing-panel/models"
)

type ConfigService struct {
	mu   sync.RWMutex
	conf models.AppConfig
	db   *Database
}

func NewConfigService(db *Database) *ConfigService {
	s := &ConfigService{db: db}
	s.loadFromDB()
	return s
}

func (s *ConfigService) Get() models.AppConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.conf
}

// Reload reloads the config from the database. It should be called after an
// external database change (e.g. database import).
func (s *ConfigService) Reload() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadFromDB()
}

func (s *ConfigService) Update(req models.ConfigUpdateRequest) error {
	var logLevel string
	if req.LogLevel != nil {
		var err error
		logLevel, err = NormalizeLogLevel(*req.LogLevel)
		if err != nil {
			return err
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if req.AccelerateDomain != nil {
		s.conf.AccelerateDomain = *req.AccelerateDomain
	}
	if req.AccelerateDomains != nil {
		s.conf.AccelerateDomains = normalizeAccelerateDomains(*req.AccelerateDomains)
	}
	if req.DashboardURL != nil {
		s.conf.DashboardURL = *req.DashboardURL
	}
	if req.Dashboards != nil {
		s.conf.Dashboards = *req.Dashboards
	}
	if req.AutoStartKernel != nil {
		s.conf.AutoStartKernel = *req.AutoStartKernel
	}
	if req.CustomizedFeaturesEnabled != nil {
		s.conf.CustomizedFeaturesEnabled = *req.CustomizedFeaturesEnabled
		if !s.conf.CustomizedFeaturesEnabled {
			s.disableCustomizedOutbounds()
		}
	}
	if req.LogLevel != nil {
		if err := GetBoxService().SetLogLevel(logLevel); err != nil {
			return err
		}
		s.conf.LogLevel = logLevel
	}
	s.saveToDB()
	return nil
}

// disableCustomizedOutbounds disables fork-only outbounds when the panel
// feature switch is turned off through the app configuration API.
func (s *ConfigService) disableCustomizedOutbounds() {
	var config models.SingBoxConfig
	if err := s.db.Get("singbox", "config", &config); err != nil {
		return
	}
	disableCustomizedOutbounds(&config)
	if err := s.db.Put("singbox", "config", config); err != nil {
		slog.Error("customized outbound state save failed", "error", err)
	}
}

func (s *ConfigService) loadFromDB() {
	var conf models.AppConfig
	if err := s.db.Get("config", "app_config", &conf); err != nil {
		s.conf.LogLevel = DefaultLogLevel
		_ = ApplyLogLevel(DefaultLogLevel)
		return
	}
	s.conf = conf
	s.conf.AccelerateDomains = normalizeAccelerateDomains(conf.AccelerateDomains)
	level, err := NormalizeLogLevel(conf.LogLevel)
	if err != nil {
		s.conf.LogLevel = DefaultLogLevel
		level = DefaultLogLevel
	}
	if err := GetBoxService().SetLogLevel(level); err != nil {
		s.conf.LogLevel = DefaultLogLevel
		_ = ApplyLogLevel(DefaultLogLevel)
	}
}

// normalizeAccelerateDomains removes blank entries and duplicate domains.
// Domain matching is intentionally opt-in: an empty list must not match all
// URLs.
func normalizeAccelerateDomains(domains []string) []string {
	result := make([]string, 0, len(domains))
	seen := make(map[string]struct{}, len(domains))
	for _, domain := range domains {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			continue
		}
		if _, ok := seen[domain]; ok {
			continue
		}
		seen[domain] = struct{}{}
		result = append(result, domain)
	}
	return result
}

func (s *ConfigService) saveToDB() {
	if err := s.db.Put("config", "app_config", s.conf); err != nil {
		slog.Error("config save failed", "error", err)
	}
}
