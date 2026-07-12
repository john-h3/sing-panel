package services

import (
	"log/slog"
	"time"
)

// StatsService handles sing-box runtime info
type StatsService struct {
	startTime time.Time
}

func NewStatsService(db *Database) *StatsService {
	s := &StatsService{}
	s.loadStartTimeFromDB(db)
	return s
}

func (s *StatsService) loadStartTimeFromDB(db *Database) {
	var state struct {
		StartTime time.Time `json:"startTime"`
		PID       int       `json:"pid"`
	}
	if err := db.Get("state", "kernel", &state); err != nil {
		return
	}
	if state.PID > 0 && !state.StartTime.IsZero() {
		s.startTime = state.StartTime
		slog.Info("loaded start time from DB", "startTime", s.startTime)
	}
}

// SetStartTime records the process start time
func (s *StatsService) SetStartTime(t time.Time) {
	s.startTime = t
}

// GetStartTime returns the recorded start time
func (s *StatsService) GetStartTime() time.Time {
	return s.startTime
}
