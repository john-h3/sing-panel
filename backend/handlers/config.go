package handlers

import (
	"net/http"
	"sing_panel/models"
	"sing_panel/services"

	"github.com/gin-gonic/gin"
)

type ConfigHandler struct {
	service *services.ConfigService
}

func NewConfigHandler(service *services.ConfigService) *ConfigHandler {
	return &ConfigHandler{service: service}
}

// Get returns the current config
func (h *ConfigHandler) Get(c *gin.Context) {
	conf := h.service.Get()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    conf,
	})
}

// Update modifies the config
func (h *ConfigHandler) Update(c *gin.Context) {
	var req models.ConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.Update(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Config updated",
		"data":    h.service.Get(),
	})
}
