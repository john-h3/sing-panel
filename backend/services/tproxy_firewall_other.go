//go:build !linux

package services

import "sing_panel/models"

// setupTproxyFirewall is a no-op on non-Linux platforms where tproxy firewall
// traffic steering is unsupported.
func (s *ProcessService) setupTproxyFirewall(_ []models.Inbound) {}

// cleanupTproxyFirewall is a no-op on non-Linux platforms.
func (s *ProcessService) cleanupTproxyFirewall() {}