package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"sing_panel/models"
)

type ProcessService struct {
	db            *Database
	configService *SingBoxConfigService
	kernelService *KernelService
	statsService  *StatsService
}

func NewProcessService(db *Database, configService *SingBoxConfigService, kernelService *KernelService, statsService *StatsService) *ProcessService {
	return &ProcessService{
		db:            db,
		configService: configService,
		kernelService: kernelService,
		statsService:  statsService,
	}
}

// GetStatus returns the current process status
func (s *ProcessService) GetStatus() models.ProcessStatus {
	state, err := s.kernelService.loadState()
	if err != nil || !state.Installed {
		return models.ProcessStatus{
			Running: false,
			Status:  "not_installed",
		}
	}

	// Check embedded service
	boxService := GetBoxService()
	if boxService.IsRunning() {
		return models.ProcessStatus{
			Running:   true,
			Status:    "running",
			Version:   state.Version,
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

	// Check if installed
	state, err := s.kernelService.loadState()
	if err != nil || !state.Installed {
		return fmt.Errorf("kernel not installed")
	}

	// Clear old PID if process is not running
	s.kernelService.clearPID()

	// Generate config
	configJSON, err := s.generateConfigJSON()
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

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


