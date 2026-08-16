package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"sing_panel/models"
)

type ProcessService struct {
	db            *Database
	configService *SingBoxConfigService
	kernelService *KernelService
	statsService  *StatsService
	mu            sync.RWMutex
	runtimeConfig json.RawMessage
}

func NewProcessService(db *Database, configService *SingBoxConfigService, kernelService *KernelService, statsService *StatsService) *ProcessService {
	return &ProcessService{
		db:            db,
		configService: configService,
		kernelService: kernelService,
		statsService:  statsService,
	}
}

// GetRuntimeConfig returns the exact config JSON that was passed to the
// embedded kernel at startup, together with whether the kernel is running.
func (s *ProcessService) GetRuntimeConfig() (json.RawMessage, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !GetBoxService().IsRunning() || s.runtimeConfig == nil {
		return nil, false
	}
	return bytes.Clone(s.runtimeConfig), true
}

// GetStatus returns the current process status
func (s *ProcessService) GetStatus() models.ProcessStatus {
	// The kernel is embedded into the binary, so it is always available.
	// Only the running state of the embedded sing-box instance matters.
	boxService := GetBoxService()
	if boxService.IsRunning() {
		state, _ := s.kernelService.loadState()
		version := state.Version
		if version == "" {
			version = singBoxVersion()
		}
		return models.ProcessStatus{
			Running:   true,
			Status:    "running",
			Version:   version,
			StartTime: state.StartTime,
		}
	}

	return models.ProcessStatus{
		Running: false,
		Status:  "stopped",
	}
}

// Start starts the sing-box process
func (s *ProcessService) Start() error {
	// Check if already running
	status := s.GetStatus()
	if status.Running {
		return fmt.Errorf("already running")
	}

	// Clear old PID if process is not running
	s.kernelService.clearPID()

	// Generate config
	configJSON, err := s.generateConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	s.mu.Lock()
	s.runtimeConfig = configJSON
	s.mu.Unlock()
	// Start embedded sing-box
	boxService := GetBoxService()
	if err := boxService.Start(configJSON); err != nil {
		return fmt.Errorf("failed to start embedded sing-box: %w", err)
	}

	// Save start time (no PID needed for embedded mode)
	startTime := time.Now()
	slog.Info("sing-box embedded started")
	s.kernelService.saveStartTime(startTime)
	s.statsService.SetStartTime(startTime)

	return nil
}

// Stop stops the sing-box process
func (s *ProcessService) Stop() error {
	status := s.GetStatus()
	if !status.Running {
		return fmt.Errorf("not running")
	}

	// Stop embedded sing-box
	boxService := GetBoxService()
	if err := boxService.Stop(); err != nil {
		slog.Error("failed to stop embedded sing-box", "error", err)
		return err
	}

	s.kernelService.clearPID()
	slog.Info("sing-box embedded stopped")
	s.mu.Lock()
	s.runtimeConfig = nil
	s.mu.Unlock()
	return nil
}

// Restart restarts the sing-box process
func (s *ProcessService) Restart() error {
	// Stop if running
	if status := s.GetStatus(); status.Running {
		if err := s.Stop(); err != nil {
			return fmt.Errorf("failed to stop: %w", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Start
	return s.Start()
}

// generateConfigJSON generates the sing-box config as JSON bytes
func (s *ProcessService) generateConfigJSON() ([]byte, error) {
	config, err := s.configService.ExportConfig()
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(config, "", "  ")
}
