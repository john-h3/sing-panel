package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"sing_panel/models"
)

const (
	tmpConfigPath  = "/tmp/sing-box-config.json"
	singBoxLogPath = "/tmp/sing-box.log"
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

	pid := state.PID
	if pid == 0 {
		return models.ProcessStatus{
			Running: false,
			Status:  "stopped",
		}
	}

	// Check if process is still running
	if s.isProcessRunning(pid) {
		return models.ProcessStatus{
			Running:   true,
			PID:       pid,
			Status:    "running",
			Version:   state.Version,
			StartTime: state.StartTime,
		}
	}

	// Process not running, but don't clear PID here
	// Let Start() handle cleanup before starting new process
	return models.ProcessStatus{
		Running: false,
		PID:     pid,
		Status:  "stopped",
	}
}

// Start starts the sing-box process
func (s *ProcessService) Start() error {
	// Check if already running
	status := s.GetStatus()
	if status.Running {
		return fmt.Errorf("already running (PID: %d)", status.PID)
	}

	// Check if installed
	state, err := s.kernelService.loadState()
	if err != nil || !state.Installed {
		return fmt.Errorf("kernel not installed")
	}

	// Clear old PID if process is not running
	s.kernelService.clearPID()

	// Generate config
	if err := s.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	binPath := filepath.Join(singBoxBinDir, singBoxBinName)

	// Check config before starting
	slog.Info("checking config", "cmd", binPath+" check -c "+tmpConfigPath)
	checkCmd := exec.Command(binPath, "check", "-c", tmpConfigPath)
	var checkStdout, checkStderr bytes.Buffer
	checkCmd.Stdout = &checkStdout
	checkCmd.Stderr = &checkStderr

	if err := checkCmd.Run(); err != nil {
		errMsg := checkStderr.String()
		if errMsg == "" {
			errMsg = checkStdout.String()
		}
		return fmt.Errorf("config check failed:\n%s", errMsg)
	}
	slog.Info("config check passed")

	// Start process
	cmd := exec.Command(binPath, "run", "-c", tmpConfigPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	slog.Info("starting sing-box", "config", tmpConfigPath)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	// Save PID and start time
	pid := cmd.Process.Pid
	startTime := time.Now()
	slog.Info("sing-box started", "pid", pid)
	s.kernelService.savePID(pid)
	s.kernelService.saveStartTime(startTime)
	s.statsService.SetStartTime(startTime)

	// Reap child process to prevent zombie
	go func() {
		cmd.Wait()
		slog.Info("sing-box process exited", "pid", pid)
	}()

	// Wait and check if still running
	time.Sleep(1000 * time.Millisecond)
	if !s.isProcessRunning(pid) {
		s.kernelService.clearPID()
		return fmt.Errorf("process exited immediately after start")
	}

	return nil
}

// Stop stops the sing-box process
func (s *ProcessService) Stop() error {
	status := s.GetStatus()
	if !status.Running {
		return fmt.Errorf("not running")
	}

	// Kill process group (negative PID = process group)
	// sing-box is started with Setsid, so it's the group leader
	slog.Info("stopping sing-box", "pid", status.PID)
	syscall.Kill(-status.PID, syscall.SIGTERM)

	// Wait for process to exit (max 5 seconds)
	for i := 0; i < 50; i++ {
		if !s.isProcessRunning(status.PID) {
			s.kernelService.clearPID()
			slog.Info("sing-box stopped", "pid", status.PID)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Force kill process group if still running
	slog.Warn("force killing sing-box", "pid", status.PID)
	syscall.Kill(-status.PID, syscall.SIGKILL)
	s.kernelService.clearPID()
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

// generateConfig generates the sing-box config.json in /tmp
func (s *ProcessService) generateConfig() error {
	config, err := s.configService.ExportConfig()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tmpConfigPath, data, 0644)
}

// isProcessRunning checks if a process with given PID is running
func (s *ProcessService) isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Try to find the process
	if runtime.GOOS == "windows" {
		// Windows: use tasklist
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid)).Output()
		if err != nil {
			return false
		}
		return len(out) > 0
	}

	// Unix: check /proc or use kill -0
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Send signal 0 to check if process exists
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
