package services

import (
	"log/slog"
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

func (s *ConfigService) Update(req models.ConfigUpdateRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if req.AccelerateDomain != nil {
		s.conf.AccelerateDomain = *req.AccelerateDomain
	}
	if req.AccelerateDomains != nil {
		s.conf.AccelerateDomains = *req.AccelerateDomains
	}
	if req.DashboardURL != nil {
		s.conf.DashboardURL = *req.DashboardURL
	}
	if req.Dashboards != nil {
		s.conf.Dashboards = *req.Dashboards
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
	slog.Info("config loaded", "accelerateDomain", conf.AccelerateDomain, "accelerateDomains", conf.AccelerateDomains)
}

func (s *ConfigService) saveToDB() {
	if err := s.db.Put("config", "app_config", s.conf); err != nil {
		slog.Error("config save failed", "error", err)
	}
}
