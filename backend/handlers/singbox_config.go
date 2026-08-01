package handlers

import (
	"net"
	"net/http"

	"sing_panel/models"
	"sing_panel/services"

	"github.com/gin-gonic/gin"
)

type SingBoxConfigHandler struct {
	service       *services.SingBoxConfigService
	configService *services.ConfigService
}

func NewSingBoxConfigHandler(service *services.SingBoxConfigService, configService *services.ConfigService) *SingBoxConfigHandler {
	return &SingBoxConfigHandler{service: service, configService: configService}
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

// GetRulesets returns all ruleset configurations
func (h *SingBoxConfigHandler) GetRulesets(c *gin.Context) {
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
		"data":    config.Rulesets,
	})
}

// AddRuleset adds a new ruleset
func (h *SingBoxConfigHandler) AddRuleset(c *gin.Context) {
	var req models.Ruleset
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ruleset, err := h.service.AddRuleset(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ruleset,
	})
}

// AddRulesets adds multiple rulesets
func (h *SingBoxConfigHandler) AddRulesets(c *gin.Context) {
	var req struct {
		Rulesets []models.Ruleset `json:"rulesets" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	rulesets, err := h.service.AddRulesets(req.Rulesets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rulesets,
	})
}

// UpdateRuleset updates an existing ruleset
func (h *SingBoxConfigHandler) UpdateRuleset(c *gin.Context) {
	var req models.Ruleset
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ruleset, err := h.service.UpdateRuleset(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    ruleset,
	})
}

// DeleteRuleset deletes a ruleset
func (h *SingBoxConfigHandler) DeleteRuleset(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	if err := h.service.DeleteRuleset(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Ruleset deleted",
	})
}

// DeleteRulesets deletes multiple rulesets
func (h *SingBoxConfigHandler) DeleteRulesets(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.DeleteRulesets(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Rulesets deleted",
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

// GetRouteRules returns all route rule configurations
func (h *SingBoxConfigHandler) GetRouteRules(c *gin.Context) {
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
		"data":    config.RouteRules,
	})
}

// AddRouteRule adds a new route rule
func (h *SingBoxConfigHandler) AddRouteRule(c *gin.Context) {
	var req models.RouteRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	rule, err := h.service.AddRouteRule(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rule,
	})
}

// UpdateRouteRule updates an existing route rule
func (h *SingBoxConfigHandler) UpdateRouteRule(c *gin.Context) {
	var req models.RouteRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	rule, err := h.service.UpdateRouteRule(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rule,
	})
}

// DeleteRouteRule deletes a route rule
func (h *SingBoxConfigHandler) DeleteRouteRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	if err := h.service.DeleteRouteRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Route rule deleted",
	})
}

// ReorderRouteRules reorders route rules
func (h *SingBoxConfigHandler) ReorderRouteRules(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.ReorderRouteRules(req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Route rules reordered",
	})
}

// GetRouteConfig returns the route configuration
func (h *SingBoxConfigHandler) GetRouteConfig(c *gin.Context) {
	rc, err := h.service.GetRouteConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    rc,
	})
}

// UpdateRouteConfig updates the route configuration
func (h *SingBoxConfigHandler) UpdateRouteConfig(c *gin.Context) {
	var req models.RouteConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.UpdateRouteConfig(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Route config updated",
	})
}

// GetDNS returns the DNS configuration
func (h *SingBoxConfigHandler) GetDNS(c *gin.Context) {
	dns, err := h.service.GetDNS()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    dns,
	})
}

// UpdateDNS updates the DNS configuration
func (h *SingBoxConfigHandler) UpdateDNS(c *gin.Context) {
	var req models.DNSConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.UpdateDNS(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "DNS config updated",
	})
}

// GetServices returns all service configurations
func (h *SingBoxConfigHandler) GetServices(c *gin.Context) {
	services, err := h.service.GetServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    services,
	})
}

// AddService adds a new service
func (h *SingBoxConfigHandler) AddService(c *gin.Context) {
	var req models.Service
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	svc, err := h.service.AddService(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    svc,
	})
}

// UpdateService updates an existing service
func (h *SingBoxConfigHandler) UpdateService(c *gin.Context) {
	var req models.Service
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	svc, err := h.service.UpdateService(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    svc,
	})
}

// DeleteService deletes a service
func (h *SingBoxConfigHandler) DeleteService(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	if err := h.service.DeleteService(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Service deleted",
	})
}

// GetHTTPClients returns all HTTP client configurations
func (h *SingBoxConfigHandler) GetHTTPClients(c *gin.Context) {
	clients, err := h.service.GetHTTPClients()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    clients,
	})
}

// AddHTTPClient adds a new HTTP client
func (h *SingBoxConfigHandler) AddHTTPClient(c *gin.Context) {
	var req models.HTTPClient
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	client, err := h.service.AddHTTPClient(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    client,
	})
}

// UpdateHTTPClient updates an existing HTTP client
func (h *SingBoxConfigHandler) UpdateHTTPClient(c *gin.Context) {
	var req models.HTTPClient
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	client, err := h.service.UpdateHTTPClient(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    client,
	})
}

// DeleteHTTPClient deletes an HTTP client
func (h *SingBoxConfigHandler) DeleteHTTPClient(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "id is required",
		})
		return
	}

	if err := h.service.DeleteHTTPClient(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "HTTP client deleted",
	})
}

// FetchGeoTree returns the cached GeoIP rule-set tree
func (h *SingBoxConfigHandler) FetchGeoTree(c *gin.Context) {
	data, err := h.service.GetGeoTree()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// FetchCommonRulesetTree returns the cached common ruleset tree
func (h *SingBoxConfigHandler) FetchCommonRulesetTree(c *gin.Context) {
	data, err := h.service.GetCommonRulesetTree()
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetNetworkInterfaces returns available network interfaces on the system
func (h *SingBoxConfigHandler) GetNetworkInterfaces(c *gin.Context) {
	ifaces, err := net.Interfaces()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	type ifaceInfo struct {
		Name        string `json:"name"`
		DisplayName string `json:"display_name"`
	}

	var result []ifaceInfo
	for _, iface := range ifaces {
		// Skip loopback interface
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		result = append(result, ifaceInfo{
			Name:        iface.Name,
			DisplayName: iface.HardwareAddr.String(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// GetExperimental returns the experimental configuration
func (h *SingBoxConfigHandler) GetExperimental(c *gin.Context) {
	exp, err := h.service.GetExperimental()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    exp,
	})
}

// UpdateExperimental updates the experimental configuration
func (h *SingBoxConfigHandler) UpdateExperimental(c *gin.Context) {
	var req models.ExperimentalConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.UpdateExperimental(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Experimental config updated",
	})
}
