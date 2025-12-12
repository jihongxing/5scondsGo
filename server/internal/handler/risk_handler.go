package handler

import (
	"net/http"
	"strconv"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/service"
	"github.com/gin-gonic/gin"
)

// RiskHandler 风控处理器
type RiskHandler struct {
	riskService *service.RiskControlService
}

// NewRiskHandler 创建风控处理器
func NewRiskHandler(riskService *service.RiskControlService) *RiskHandler {
	return &RiskHandler{
		riskService: riskService,
	}
}

// ListRiskFlags 列表风控标记
// GET /api/admin/risk-flags
func (h *RiskHandler) ListRiskFlags(c *gin.Context) {
	var query model.RiskFlagListQuery
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

	flags, total, err := h.riskService.ListFlags(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"flags": flags,
		"total": total,
		"page":  query.Page,
		"size":  query.PageSize,
	})
}

// GetRiskFlag 获取风控标记详情
// GET /api/admin/risk-flags/:id
func (h *RiskHandler) GetRiskFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flag id"})
		return
	}

	flag, err := h.riskService.GetFlag(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	// 解析详情
	details, _ := h.riskService.GetFlagDetails(flag)

	c.JSON(http.StatusOK, gin.H{
		"flag":    flag,
		"details": details,
	})
}

// ReviewRiskFlag 审核风控标记
// POST /api/admin/risk-flags/:id/review
func (h *RiskHandler) ReviewRiskFlag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid flag id"})
		return
	}

	var req model.ReviewRiskFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("user_id")

	if err := h.riskService.ReviewFlag(c.Request.Context(), id, req.Action, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "flag reviewed"})
}
