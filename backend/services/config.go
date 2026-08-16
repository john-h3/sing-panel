package services

import (
	"log/slog"
	"strings"
	"sync"

	"sing_panel/models"
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

func (s *ConfigService) Update(req models.ConfigUpdateRequest) {
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
	s.saveToDB()
}

func (s *ConfigService) loadFromDB() {
	var conf models.AppConfig
	if err := s.db.Get("config", "app_config", &conf); err != nil {
		slog.Debug("no config in database", "error", err)
		return
	}
	s.conf = conf
	s.conf.AccelerateDomains = normalizeAccelerateDomains(conf.AccelerateDomains)
	slog.Info("config loaded", "accelerateDomain", conf.AccelerateDomain, "accelerateDomains", conf.AccelerateDomains)
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
