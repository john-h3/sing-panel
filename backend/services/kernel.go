package services

import (
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"sing_panel/models"
)

// singBoxVersion returns the sing-box kernel version compiled into this
// binary, resolved from the module build info so it always matches the
// actually linked kernel version (no manual version bumping needed).
func singBoxVersion() string {
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range bi.Deps {
			if dep.Path == "github.com/sagernet/sing-box" && dep.Version != "" {
				return dep.Version
			}
		}
	}
	return "unknown"
}

type KernelService struct {
	db           *Database
	processStart time.Time
}

func NewKernelService(db *Database) *KernelService {
	return &KernelService{db: db, processStart: time.Now()}
}

// GetStatus returns the current kernel status
func (s *KernelService) GetStatus() models.KernelStatus {
	// The kernel is embedded into the binary; there is no separate
	// install/download concept, so the version is always the compiled one.
	return models.KernelStatus{
		Version: singBoxVersion(),
	}
}

// GetMonitorStats returns Go runtime performance metrics
func (s *KernelService) GetMonitorStats() models.MonitorStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// The most recent GC pause duration (ring buffer of recent pauses)
	var lastPause uint64
	if m.NumGC > 0 {
		idx := (m.NumGC + 255) % 256
		lastPause = m.PauseNs[idx]
	}

	return models.MonitorStats{
		UptimeSeconds: int64(time.Since(s.processStart).Seconds()),
		Goroutines:    runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		GOMAXPROCS:    runtime.GOMAXPROCS(0),
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapInuse:     m.HeapInuse,
		HeapObjects:   m.HeapObjects,
		Sys:           m.Sys,
		NumGC:         m.NumGC,
		LastGC:        int64(m.LastGC),
		PauseTotalNs:  m.PauseTotalNs,
		LastPauseNs:   lastPause,
		GCPercent:     getGCPercent(),
	}
}

// getGCPercent returns the GOGC target percentage (default 100)
func getGCPercent() int {
	if v := os.Getenv("GOGC"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			return p
		}
	}
	return 100
}

// GetSystemInfo returns system platform and architecture info
func (s *KernelService) GetSystemInfo() models.SystemInfo {
	hostname, _ := os.Hostname()
	return models.SystemInfo{
		Platform:      runtime.GOOS,
		Arch:          runtime.GOARCH,
		Hostname:      hostname,
		KernelVersion: getKernelVersion(),
		BuildTime:     BuildTime,
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
		state.Version = singBoxVersion()
	})
}
