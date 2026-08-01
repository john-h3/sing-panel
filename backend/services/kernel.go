package services

import (
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"sing_panel/models"
)

const singBoxVersion = "1.14.0-alpha.43"

type KernelService struct {
	db *Database
}

func NewKernelService(db *Database) *KernelService {
	return &KernelService{db: db}
}

// GetStatus returns the current kernel status
func (s *KernelService) GetStatus() models.KernelStatus {
	var state models.KernelState
	if err := s.db.Get("state", "kernel", &state); err != nil {
		return models.KernelStatus{
			Installed: false,
			Version:   singBoxVersion,
		}
	}

	return models.KernelStatus{
		Installed: state.Installed,
		Version:   singBoxVersion,
		Path:      state.Path,
	}
}

// GetSystemInfo returns system platform and architecture info
func (s *KernelService) GetSystemInfo() models.SystemInfo {
	hostname, _ := os.Hostname()
	return models.SystemInfo{
		Platform:      runtime.GOOS,
		Arch:          runtime.GOARCH,
		Hostname:      hostname,
		KernelVersion: getKernelVersion(),
	}
}

func getKernelVersion() string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}

// saveState saves kernel state to database
func (s *KernelService) saveState(state models.KernelState) error {
	return s.db.Put("state", "kernel", state)
}

// loadState loads kernel state from database
func (s *KernelService) loadState() (models.KernelState, error) {
	var state models.KernelState
	err := s.db.Get("state", "kernel", &state)
	return state, err
}

// saveStartTime saves the process start time to state
func (s *KernelService) saveStartTime(startTime time.Time) {
	s.db.UpdateKernelState(func(state *models.KernelState) {
		state.StartTime = startTime
	})
	slog.Info("kernel start time saved", "time", startTime)
}

// clearPID clears the PID from state
func (s *KernelService) clearPID() {
	s.db.UpdateKernelState(func(state *models.KernelState) {
		state.PID = 0
	})
}

// SetInstalled marks the kernel as installed (for embedded mode)
func (s *KernelService) SetInstalled(installed bool) {
	s.db.UpdateKernelState(func(state *models.KernelState) {
		state.Installed = installed
		state.Version = singBoxVersion
	})
}
