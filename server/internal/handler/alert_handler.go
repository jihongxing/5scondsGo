package handler

import (
	"net/http"
	"strconv"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/service"
	"github.com/gin-gonic/gin"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	alertManager *service.AlertManager
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(alertManager *service.AlertManager) *AlertHandler {
	return &AlertHandler{
		alertManager: alertManager,
	}
}

// ListAlerts 列表告警
// GET /api/admin/alerts
func (h *AlertHandler) ListAlerts(c *gin.Context) {
	var query model.AlertListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	alerts, total, err := h.alertManager.ListAlerts(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"total":  total,
		"page":   query.Page,
		"size":   query.PageSize,
	})
}

// GetAlert 获取告警详情
// GET /api/admin/alerts/:id
func (h *AlertHandler) GetAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert id"})
		return
	}

	alert, err := h.alertManager.GetAlert(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "alert not found"})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// AcknowledgeAlert 确认告警
// POST /api/admin/alerts/:id/acknowledge
func (h *AlertHandler) AcknowledgeAlert(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert id"})
		return
	}

	userID := c.GetInt64("user_id")

	if err := h.alertManager.AcknowledgeAlert(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "alert acknowledged"})
}

// GetAlertSummary 获取告警摘要
// GET /api/admin/alerts/summary
func (h *AlertHandler) GetAlertSummary(c *gin.Context) {
	activeCount, err := h.alertManager.GetActiveAlertCount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bySeverity, err := h.alertManager.GetAlertSummary(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_count": activeCount,
		"by_severity":  bySeverity,
	})
}
