package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"sing-panel/services"

	"github.com/gin-gonic/gin"
)

type LogHandler struct{ ring *services.MemoryLogRing }

const (
	defaultLogTail = 100
	maxLogTail     = services.MemoryLogCapacity
)

func NewLogHandler(ring *services.MemoryLogRing) *LogHandler { return &LogHandler{ring: ring} }

func (h *LogHandler) List(c *gin.Context) {
	limit := defaultLogTail
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "limit must be a positive integer"})
			return
		}
		limit = parsed
		if limit > maxLogTail {
			limit = maxLogTail
		}
	}
	var entries []services.MemoryLogEntry
	if value := c.Query("after"); value != "" {
		after, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "after must be a non-negative sequence"})
			return
		}
		entries = h.ring.RecentAfter(after, limit, c.Query("level"), c.Query("source"))
	} else {
		entries = h.ring.Recent(limit, c.Query("level"), c.Query("source"))
	}
	count, dropped := h.ring.Stats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"entries":  entries,
			"count":    count,
			"capacity": services.MemoryLogCapacity,
			"dropped":  dropped,
		},
	})
}

func (h *LogHandler) Clear(c *gin.Context) {
	h.ring.Clear()
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *LogHandler) Stream(c *gin.Context) {
	var after uint64
	if value := c.Query("after"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "after must be a non-negative sequence"})
			return
		}
		after = parsed
	}
	client, unsubscribe, replay := h.ring.SubscribeFilteredAfter(after, c.Query("level"), c.Query("source"))
	defer unsubscribe()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	if flusher, ok := c.Writer.(http.Flusher); ok {
		fmt.Fprint(c.Writer, "retry: 3000\n\nevent: ready\ndata: {}\n\n")
		flusher.Flush()
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}
	for _, entry := range replay {
		if err := writeLogEvent(c.Writer, flusher, entry); err != nil {
			return
		}
	}
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case entry, open := <-client:
			if !open {
				return
			}
			if err := writeLogEvent(c.Writer, flusher, entry); err != nil {
				return
			}
		case <-heartbeat.C:
			if _, err := fmt.Fprint(c.Writer, ": heartbeat\n\n"); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func writeLogEvent(writer http.ResponseWriter, flusher http.Flusher, entry services.MemoryLogEntry) error {
	payload, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(writer, "event: log\ndata: %s\n\n", payload); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
