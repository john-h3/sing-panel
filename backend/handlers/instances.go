package handlers

import (
	"net/http"
	"sing_panel/models"
	"sing_panel/services"

	"github.com/gin-gonic/gin"
)

type InstanceHandler struct {
	service *services.InstanceService
}

func NewInstanceHandler(service *services.InstanceService) *InstanceHandler {
	return &InstanceHandler{service: service}
}

// List returns all managed instances
func (h *InstanceHandler) List(c *gin.Context) {
	instances, err := h.service.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": instances})
}

// Create adds a new managed instance
func (h *InstanceHandler) Create(c *gin.Context) {
	var req models.ManagedInstance
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	inst, err := h.service.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": inst})
}

// Update modifies an existing managed instance
func (h *InstanceHandler) Update(c *gin.Context) {
	var req models.InstanceUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	inst, err := h.service.Update(c.Param("id"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": inst})
}

// Delete removes a managed instance
func (h *InstanceHandler) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "实例已删除"})
}

// GetStatus checks a single instance live
func (h *InstanceHandler) GetStatus(c *gin.Context) {
	inst, err := h.service.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	status := h.service.CheckInstance(inst)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": status})
}

// CheckAll checks every managed instance live
func (h *InstanceHandler) CheckAll(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": h.service.CheckAll()})
}

// Sync pushes or pulls configuration for one instance
func (h *InstanceHandler) Sync(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	inst, err := h.service.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}

	var syncErr error
	switch req.Action {
	case "push":
		syncErr = h.service.SyncPush(inst)
	case "pull":
		syncErr = h.service.SyncPull(inst)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "action 必须为 push 或 pull"})
		return
	}
	if syncErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": syncErr.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "同步成功"})
}

// SyncAll pushes the local configuration to every managed instance
func (h *InstanceHandler) SyncAll(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": h.service.SyncPushAll()})
}

// LocalInfo returns this panel's info and config fingerprint
func (h *InstanceHandler) LocalInfo(c *gin.Context) {
	info := h.service.LocalPanelInfo()
	fingerprint, err := h.service.LocalFingerprint()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
		"info":        info,
		"fingerprint": fingerprint,
		"syncToken":   h.service.SyncToken(),
	}})
}

// SetSyncToken updates this panel's sync token (empty clears protection)
func (h *InstanceHandler) SetSyncToken(c *gin.Context) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.service.SetSyncToken(req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "同步令牌已保存"})
}

// GetConfigDiff returns the configuration differences between local and remote
func (h *InstanceHandler) GetConfigDiff(c *gin.Context) {
	inst, err := h.service.Get(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		return
	}
	diff, err := h.service.GetConfigDiff(inst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": diff})
}
