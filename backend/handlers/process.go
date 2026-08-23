package handlers

import (
	"encoding/json"
	"net/http"
	"sing-panel/services"

	"github.com/gin-gonic/gin"
)

type ProcessHandler struct {
	service *services.ProcessService
}

func NewProcessHandler(service *services.ProcessService) *ProcessHandler {
	return &ProcessHandler{service: service}
}

// GetStatus returns the current process status
func (h *ProcessHandler) GetStatus(c *gin.Context) {
	status := h.service.GetStatus()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    status,
	})
}

// Health reports whether the embedded sing-box kernel is running.
// This endpoint is intentionally independent of the regular API response
// envelope so it can be used directly by external health checkers.
func (h *ProcessHandler) Health(c *gin.Context) {
	if h.service.GetStatus().Running {
		c.Status(http.StatusOK)
		return
	}

	c.Status(http.StatusBadRequest)
}

// GetRuntimeConfig returns the config actually passed to the kernel
func (h *ProcessHandler) GetRuntimeConfig(c *gin.Context) {
	config, ok := h.service.GetRuntimeConfig()
	if !ok {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"running": false,
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"running": true,
		"data":    json.RawMessage(config),
	})
}

// Start starts the sing-box process
func (h *ProcessHandler) Start(c *gin.Context) {
	if err := h.service.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sing-Box started",
	})
}

// Stop stops the sing-box process
func (h *ProcessHandler) Stop(c *gin.Context) {
	if err := h.service.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sing-Box stopped",
	})
}

// Restart restarts the sing-box process
func (h *ProcessHandler) Restart(c *gin.Context) {
	if err := h.service.Restart(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Sing-Box restarted",
	})
}
