package handlers

import (
	"net/http"
	"sing_panel/services"

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
