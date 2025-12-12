package handler

import (
	"net/http"

	"github.com/fiveseconds/server/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LogLevelHandler handles log level management endpoints
type LogLevelHandler struct {
	logger *logger.Logger
}

// NewLogLevelHandler creates a new log level handler
func NewLogLevelHandler(log *logger.Logger) *LogLevelHandler {
	return &LogLevelHandler{logger: log}
}

// GetLogLevel returns the current log level
// GET /api/admin/log-level
func (h *LogLevelHandler) GetLogLevel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"level": h.logger.GetLevel(),
	})
}

// SetLogLevelRequest is the request body for setting log level
type SetLogLevelRequest struct {
	Level string `json:"level" binding:"required"`
}

// SetLogLevel changes the log level at runtime
// PUT /api/admin/log-level
func (h *LogLevelHandler) SetLogLevel(c *gin.Context) {
	var req SetLogLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate log level
	if !logger.IsValidLevel(req.Level) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "invalid log level",
			"valid_levels": logger.ValidLevels(),
		})
		return
	}

	oldLevel := h.logger.GetLevel()
	if err := h.logger.SetLevel(req.Level); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log the level change
	h.logger.WithContext(c.Request.Context()).Info("log level changed",
		zap.String("old_level", oldLevel),
		zap.String("new_level", req.Level),
	)

	c.JSON(http.StatusOK, gin.H{
		"old_level": oldLevel,
		"new_level": req.Level,
	})
}
