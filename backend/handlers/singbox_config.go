package handlers

import (
	"net/http"

	"sing_panel/models"
	"sing_panel/services"

	"github.com/gin-gonic/gin"
)

type SingBoxConfigHandler struct {
	service *services.SingBoxConfigService
}

func NewSingBoxConfigHandler(service *services.SingBoxConfigService) *SingBoxConfigHandler {
	return &SingBoxConfigHandler{service: service}
}

// GetConfig returns the full sing-box configuration
func (h *SingBoxConfigHandler) GetConfig(c *gin.Context) {
	config, err := h.service.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// GetInbounds returns all inbound configurations
func (h *SingBoxConfigHandler) GetInbounds(c *gin.Context) {
	config, err := h.service.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config.Inbounds,
	})
}

// AddInbound adds a new inbound
func (h *SingBoxConfigHandler) AddInbound(c *gin.Context) {
	var req models.Inbound
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	inbound, err := h.service.AddInbound(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inbound,
	})
}

// UpdateInbound updates an existing inbound
func (h *SingBoxConfigHandler) UpdateInbound(c *gin.Context) {
	var req models.Inbound
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	inbound, err := h.service.UpdateInbound(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inbound,
	})
}

// DeleteInbound deletes an inbound
func (h *SingBoxConfigHandler) DeleteInbound(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	if err := h.service.DeleteInbound(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Inbound deleted",
	})
}

// GetOutbounds returns all outbound configurations
func (h *SingBoxConfigHandler) GetOutbounds(c *gin.Context) {
	config, err := h.service.GetConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config.Outbounds,
	})
}

// AddOutbound adds a new outbound
func (h *SingBoxConfigHandler) AddOutbound(c *gin.Context) {
	var req models.Outbound
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	outbound, err := h.service.AddOutbound(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    outbound,
	})
}

// UpdateOutbound updates an existing outbound
func (h *SingBoxConfigHandler) UpdateOutbound(c *gin.Context) {
	var req models.Outbound
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	outbound, err := h.service.UpdateOutbound(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    outbound,
	})
}

// DeleteOutbound deletes an outbound
func (h *SingBoxConfigHandler) DeleteOutbound(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	if err := h.service.DeleteOutbound(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Outbound deleted",
	})
}

// GetInboundTypes returns available inbound types
func (h *SingBoxConfigHandler) GetInboundTypes(c *gin.Context) {
	types := h.service.GetDefaultInboundTypes()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    types,
	})
}

// GetOutboundTypes returns available outbound types
func (h *SingBoxConfigHandler) GetOutboundTypes(c *gin.Context) {
	types := h.service.GetDefaultOutboundTypes()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    types,
	})
}

// ExportConfig exports the sing-box configuration
func (h *SingBoxConfigHandler) ExportConfig(c *gin.Context) {
	config, err := h.service.ExportConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// ImportLink imports an outbound from a link (vmess://, vless://)
func (h *SingBoxConfigHandler) ImportLink(c *gin.Context) {
	var req struct {
		Link string `json:"link" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	outbound, err := h.service.ImportLink(req.Link)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Save to config
	saved, err := h.service.AddOutbound(outbound)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    saved,
	})
}
