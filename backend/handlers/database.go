package handlers

import (
	"net/http"
	"sing-panel/services"

	"github.com/gin-gonic/gin"
)

type DatabaseHandler struct {
	service       *services.Database
	configService *services.ConfigService
}

func NewDatabaseHandler(service *services.Database, configService *services.ConfigService) *DatabaseHandler {
	return &DatabaseHandler{service: service, configService: configService}
}

// ListBuckets returns all bucket names
func (h *DatabaseHandler) ListBuckets(c *gin.Context) {
	buckets, err := h.service.ListBuckets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": buckets})
}

// ListKeys returns all keys in a bucket
func (h *DatabaseHandler) ListKeys(c *gin.Context) {
	bucket := c.Query("bucket")
	if bucket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "bucket is required"})
		return
	}
	keys, err := h.service.ListKeys(bucket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": keys})
}

// GetValue returns the value for a bucket/key
func (h *DatabaseHandler) GetValue(c *gin.Context) {
	bucket := c.Query("bucket")
	key := c.Query("key")
	if bucket == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "bucket and key are required"})
		return
	}
	value, err := h.service.GetValue(bucket, key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": value})
}

// PutValue stores a value
func (h *DatabaseHandler) PutValue(c *gin.Context) {
	var req struct {
		Bucket string `json:"bucket" binding:"required"`
		Key    string `json:"key" binding:"required"`
		Value  string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := h.service.PutValue(req.Bucket, req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Value saved"})
}

// DeleteKey removes a key
func (h *DatabaseHandler) DeleteKey(c *gin.Context) {
	bucket := c.Query("bucket")
	key := c.Query("key")
	if bucket == "" || key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "bucket and key are required"})
		return
	}
	if err := h.service.DeleteKey(bucket, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Key deleted"})
}

// DeleteBucket removes an empty bucket
func (h *DatabaseHandler) DeleteBucket(c *gin.Context) {
	bucket := c.Query("bucket")
	if bucket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "bucket is required"})
		return
	}
	if err := h.service.DeleteBucket(bucket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Bucket deleted"})
}

// Export exports the entire database as JSON
func (h *DatabaseHandler) Export(c *gin.Context) {
	data, err := h.service.ExportAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// Import replaces the entire database with the uploaded JSON
func (h *DatabaseHandler) Import(c *gin.Context) {
	var req struct {
		Data map[string]map[string]string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.Data == nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "data is required"})
		return
	}
	if err := h.service.ImportAll(req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	// Reload in-memory caches that are backed by the database.
	if h.configService != nil {
		h.configService.Reload()
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Database imported"})
}
