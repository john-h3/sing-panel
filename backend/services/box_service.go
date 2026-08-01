package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/include"
	"github.com/sagernet/sing-box/option"
)

// BoxService wraps sing-box as an embedded library
type BoxService struct {
	mu     sync.RWMutex
	box    *box.Box
	cancel context.CancelFunc
}

var globalBoxService *BoxService

func GetBoxService() *BoxService {
	if globalBoxService == nil {
		globalBoxService = &BoxService{}
	}
	return globalBoxService
}

// Start starts the embedded sing-box instance
func (s *BoxService) Start(configJSON []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.box != nil {
		return fmt.Errorf("sing-box already running")
	}

	slog.Info("starting embedded sing-box", "config_size", len(configJSON))

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	// Parse config using standard JSON unmarshal
	var options option.Options
	if err := json.Unmarshal(configJSON, &options); err != nil {
		cancel()
		slog.Error("failed to parse config", "error", err)
		return fmt.Errorf("failed to parse config: %w", err)
	}

	slog.Info("config parsed successfully")

	// Set up context with all registries using include package
	ctx = include.Context(ctx)

	slog.Info("registries set up")

	// Create sing-box instance
	b, err := box.New(box.Options{
		Context: ctx,
		Options: options,
	})
	if err != nil {
		cancel()
		slog.Error("failed to create sing-box", "error", err)
		return fmt.Errorf("failed to create sing-box: %w", err)
	}

	slog.Info("sing-box instance created")

	// Start the instance
	if err := b.Start(); err != nil {
		cancel()
		slog.Error("failed to start sing-box", "error", err)
		return fmt.Errorf("failed to start sing-box: %w", err)
	}

	s.box = b
	slog.Info("sing-box embedded service started")
	return nil
}

// Stop stops the embedded sing-box instance
func (s *BoxService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.box == nil {
		return nil
	}

	err := s.box.Close()
	s.cancel()
	s.box = nil
	s.cancel = nil

	if err != nil {
		slog.Error("sing-box stop error", "error", err)
		return err
	}

	slog.Info("sing-box embedded service stopped")
	return nil
}

// Restart restarts the embedded sing-box instance
func (s *BoxService) Restart(configJSON []byte) error {
	if err := s.Stop(); err != nil {
		return fmt.Errorf("failed to stop: %w", err)
	}
	time.Sleep(100 * time.Millisecond)
	return s.Start(configJSON)
}

// IsRunning returns whether sing-box is running
func (s *BoxService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.box != nil
}
