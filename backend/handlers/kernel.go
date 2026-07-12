package handlers

import (
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sing_panel/models"
	"sing_panel/services"

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
	info := models.SystemInfo{
		Platform:   runtime.GOOS,
		Arch:       runtime.GOARCH,
		Hostname:   getHostname(),
		KernelVersion: getKernelVersion(),
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info,
	})
}

func getHostname() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

func getKernelVersion() string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return runtime.GOOS + "/" + runtime.GOARCH
}

// GetVersions returns available versions
func (h *KernelHandler) GetVersions(c *gin.Context) {
	versions, err := h.service.GetVersions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    versions,
		"cacheTime": h.service.GetCacheTime(),
	})
}

// RefreshVersions manually refreshes the versions cache
func (h *KernelHandler) RefreshVersions(c *gin.Context) {
	h.service.RefreshVersions()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Cache refreshed",
		"cacheTime": h.service.GetCacheTime(),
	})
}

// Download starts downloading a kernel
func (h *KernelHandler) Download(c *gin.Context) {
	var req models.DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.Download(req); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Download started",
	})
}

// StopDownload stops the current download
func (h *KernelHandler) StopDownload(c *gin.Context) {
	if err := h.service.StopDownload(); err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Download stopped",
	})
}

// Remove removes the installed kernel
func (h *KernelHandler) Remove(c *gin.Context) {
	if err := h.service.Remove(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Kernel removed",
	})
}

// SwitchVersion switches to a specific version
func (h *KernelHandler) SwitchVersion(c *gin.Context) {
	var req models.SwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := h.service.SwitchVersion(req.Version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Version switch started",
	})
}
