package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const clashAPIBase = "http://127.0.0.1:9090"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ProxyHandler struct {
	httpClient    *http.Client
	trafficMu     sync.RWMutex
	trafficCache  []TrafficPoint
	maxCacheSize  int
}

type TrafficPoint struct {
	Up   int64
	Down int64
	Time time.Time
}

func NewProxyHandler() *ProxyHandler {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
	h := &ProxyHandler{
		httpClient:   &http.Client{Transport: transport},
		trafficCache: make([]TrafficPoint, 0, 300),
		maxCacheSize: 300, // Keep last 5 minutes of data (300 * 1s)
	}
	// Start background traffic fetcher
	go h.trafficFetcher()
	return h
}

// trafficFetcher continuously fetches traffic from Clash API and caches it
func (h *ProxyHandler) trafficFetcher() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.fetchAndCacheTraffic()
	}
}

func (h *ProxyHandler) fetchAndCacheTraffic() {
	streamClient := &http.Client{Timeout: 3 * time.Second}
	resp, err := streamClient.Get(clashAPIBase + "/traffic")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var traffic struct {
		Up   int64 `json:"up"`
		Down int64 `json:"down"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&traffic); err != nil {
		return
	}

	h.trafficMu.Lock()
	h.trafficCache = append(h.trafficCache, TrafficPoint{
		Up:   traffic.Up,
		Down: traffic.Down,
		Time: time.Now(),
	})
	// Trim old data
	if len(h.trafficCache) > h.maxCacheSize {
		h.trafficCache = h.trafficCache[len(h.trafficCache)-h.maxCacheSize:]
	}
	h.trafficMu.Unlock()
}

// getAverageTraffic returns average traffic over the last N seconds
func (h *ProxyHandler) getAverageTraffic(seconds int) (up, down int64) {
	h.trafficMu.RLock()
	defer h.trafficMu.RUnlock()

	if len(h.trafficCache) == 0 {
		return 0, 0
	}

	cutoff := time.Now().Add(-time.Duration(seconds) * time.Second)
	var totalUp, totalDown int64
	var count int

	for _, p := range h.trafficCache {
		if p.Time.After(cutoff) {
			totalUp += p.Up
			totalDown += p.Down
			count++
		}
	}

	if count == 0 {
		return 0, 0
	}

	return totalUp / int64(count), totalDown / int64(count)
}

// ProxyHTTP proxies regular HTTP requests to Clash API
func (h *ProxyHandler) ProxyHTTP(c *gin.Context) {
	path := c.Param("path")
	targetURL := clashAPIBase + path

	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)
}

// WSProxyTraffic handles WebSocket for traffic - returns averaged data
func (h *ProxyHandler) WSProxyTraffic(c *gin.Context) {
	slog.Info("WSProxyTraffic: upgrading connection")

	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WSProxyTraffic: upgrade failed", "error", err)
		return
	}
	defer clientConn.Close()
	slog.Info("WSProxyTraffic: WebSocket connected")

	// Read interval from query param, default 1 second
	interval := 1
	if iv := c.Query("interval"); iv != "" {
		switch iv {
		case "1":
			interval = 1
		case "3":
			interval = 3
		case "5":
			interval = 5
		case "10":
			interval = 10
		case "30":
			interval = 30
		}
	}

	slog.Info("WSProxyTraffic: started", "interval", interval)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		up, down := h.getAverageTraffic(interval)
		data := map[string]interface{}{
			"up":   up,
			"down": down,
		}
		if err := clientConn.WriteJSON(data); err != nil {
			slog.Error("WSProxyTraffic: write failed", "error", err)
			return
		}
	}
}

// WSProxyConnections handles WebSocket for connections - polls at interval
func (h *ProxyHandler) WSProxyConnections(c *gin.Context) {
	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	// Read interval from query param, default 1 second
	interval := 1
	if iv := c.Query("interval"); iv != "" {
		switch iv {
		case "1":
			interval = 1
		case "3":
			interval = 3
		case "5":
			interval = 5
		case "10":
			interval = 10
		case "30":
			interval = 30
		}
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		targetURL := clashAPIBase + "/connections"
		resp, err := h.httpClient.Get(targetURL)
		if err != nil {
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			continue
		}

		if err := clientConn.WriteMessage(websocket.TextMessage, body); err != nil {
			return
		}
	}
}

// GenericProxy handles any Clash API endpoint
func (h *ProxyHandler) GenericProxy(c *gin.Context) {
	path := c.Param("path")

	if isWebSocketUpgrade(c) {
		switch path {
		case "/traffic":
			h.WSProxyTraffic(c)
			return
		case "/connections":
			h.WSProxyConnections(c)
			return
		}
	}

	h.ProxyHTTP(c)
}

func isWebSocketUpgrade(c *gin.Context) bool {
	return strings.EqualFold(c.GetHeader("Connection"), "upgrade") &&
		strings.EqualFold(c.GetHeader("Upgrade"), "websocket")
}
