package handlers

import (
	"net/http"
	"sing_panel/services"

	"github.com/gin-gonic/gin"
)

type DatabaseHandler struct {
	service *services.Database
}

func NewDatabaseHandler(service *services.Database) *DatabaseHandler {
	return &DatabaseHandler{service: service}
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
