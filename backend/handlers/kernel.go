package handlers

import (
	"net/http"

	"sing-panel/services"

	"github.com/gin-gonic/gin"
)

type KernelHandler struct {
	service *services.KernelService
}

func NewKernelHandler(service *services.KernelService) *KernelHandler {
	return &KernelHandler{service: service}
}

// GetStatus returns the current kernel status
func (h *KernelHandler) GetStatus(c *gin.Context) {
	status := h.service.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// GetSystemInfo returns system platform and architecture info
func (h *KernelHandler) GetSystemInfo(c *gin.Context) {
	info := h.service.GetSystemInfo()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

// GetMonitor returns Go runtime performance metrics
func (h *KernelHandler) GetMonitor(c *gin.Context) {
	stats := h.service.GetMonitorStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
