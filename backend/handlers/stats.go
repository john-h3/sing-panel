package handlers

import (
	"net/http"

	"sing-panel/services"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	service *services.StatsService
}

func NewStatsHandler(service *services.StatsService) *StatsHandler {
	return &StatsHandler{service: service}
}

// GetServiceInfo returns service information
func (h *StatsHandler) GetServiceInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"startTime": h.service.GetStartTime(),
		},
	})
}
