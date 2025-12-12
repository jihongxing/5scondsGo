package handler

import (
	"net/http"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/service"
	"github.com/gin-gonic/gin"
)

// MonitoringHandler 监控处理器
type MonitoringHandler struct {
	monitoringService *service.MonitoringService
}

// NewMonitoringHandler 创建监控处理器
func NewMonitoringHandler(monitoringService *service.MonitoringService) *MonitoringHandler {
	return &MonitoringHandler{
		monitoringService: monitoringService,
	}
}

// GetRealtimeMetrics 获取实时指标
func (h *MonitoringHandler) GetRealtimeMetrics(c *gin.Context) {
	metrics, err := h.monitoringService.GetRealtimeMetrics(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

// GetHistoricalMetrics 获取历史指标
func (h *MonitoringHandler) GetHistoricalMetrics(c *gin.Context) {
	var query model.MetricsHistoryQuery
	// 不使用 ShouldBindQuery 因为 Page/PageSize 有 min=1 验证但可能不传
	query.TimeRange = c.DefaultQuery("time_range", "1h")
	query.Page = 1
	query.PageSize = 100

	snapshots, err := h.monitoringService.GetHistoricalMetrics(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": snapshots})
}
