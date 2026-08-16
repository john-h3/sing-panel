package handlers

import (
	"net/http"
	"sing_panel/services"

	"github.com/gin-gonic/gin"
)

// PanelHandler serves basic information about this panel instance. It is
// consumed by other panels for multi-instance management.
type PanelHandler struct {
	instanceService *services.InstanceService
}

func NewPanelHandler(instanceService *services.InstanceService) *PanelHandler {
	return &PanelHandler{instanceService: instanceService}
}

// Info returns this panel's basic information.
func (h *PanelHandler) Info(c *gin.Context) {
	info := h.instanceService.LocalPanelInfo()
	c.JSON(http.StatusOK, gin.H{"success": true, "data": info})
}
