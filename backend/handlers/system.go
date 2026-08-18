package handlers

import (
	"net/http"
	"os/exec"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SystemHandler exposes operating-system level actions such as restarting
// the sing-panel systemd/openrc service or rebooting the host machine.
type SystemHandler struct {
	mu sync.Mutex
}

func NewSystemHandler() *SystemHandler {
	return &SystemHandler{}
}

// detectInitSystem mirrors backend/service.go:detectInitSystem but lives
// here so the handler is self-contained for HTTP requests.
func detectInitSystem() string {
	if _, err := exec.LookPath("systemctl"); err == nil {
		return "systemd"
	}
	if _, err := exec.LookPath("rc-service"); err == nil {
		return "openrc"
	}
	return ""
}

// RestartService restarts the sing-panel system service (systemd or openrc).
// The response is sent before the service actually restarts so the HTTP
// client receives it; the panel process will be respawned by the init system.
func (h *SystemHandler) RestartService(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	initSystem := detectInitSystem()
	if initSystem == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "未检测到支持的初始化系统 (systemd/openrc)",
		})
		return
	}

	var cmd *exec.Cmd
	switch initSystem {
	case "systemd":
		cmd = exec.Command("systemctl", "restart", "sing-panel")
	case "openrc":
		cmd = exec.Command("rc-service", "sing-panel", "restart")
	}

	// Detach the command so it survives the imminent process exit.
	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "重启服务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "sing-panel 服务正在重启...",
	})

	// Give the response time to flush, then release the command so the
	// init system can take over. We don't wait for completion because
	// restarting ourselves terminates this process.
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = cmd.Process.Release()
	}()
}

// RebootMachine reboots the host operating system. Root/sudo privileges are
// required. The response is returned before the reboot actually happens.
func (h *SystemHandler) RebootMachine(c *gin.Context) {
	h.mu.Lock()
	defer h.mu.Unlock()

	initSystem := detectInitSystem()
	var cmd *exec.Cmd
	switch initSystem {
	case "systemd":
		cmd = exec.Command("systemctl", "reboot")
	case "openrc":
		cmd = exec.Command("reboot")
	default:
		// Generic fallback: try `reboot` which works on most Linux systems
		// when invoked as root.
		cmd = exec.Command("reboot")
	}

	if err := cmd.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "重启机器失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "机器正在重启...",
	})
}

// GetInitSystem returns the detected init system type, used by the frontend
// to decide whether the restart buttons should be shown.
func (h *SystemHandler) GetInitSystem(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"init_system": detectInitSystem(),
		},
	})
}
