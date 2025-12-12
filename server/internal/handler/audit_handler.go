package handler

import (
	"net/http"
	"strconv"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/service"
	"github.com/gin-gonic/gin"
)

// AuditHandler 审计处理器
type AuditHandler struct {
	auditService *service.AuditService
	fundService  *service.FundService
}

// NewAuditHandler 创建审计处理器
func NewAuditHandler(auditService *service.AuditService, fundService *service.FundService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		fundService:  fundService,
	}
}

// RunConservationCheck 执行资金守恒检查
func (h *AuditHandler) RunConservationCheck(c *gin.Context) {
	check, err := h.auditService.RunGlobalConservationCheck(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, check)
}

// RunPeriodicAudit 执行定期审计
func (h *AuditHandler) RunPeriodicAudit(c *gin.Context) {
	periodType := c.DefaultQuery("period_type", "2h")
	if periodType != "2h" && periodType != "daily" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_type must be '2h' or 'daily'"})
		return
	}

	if err := h.auditService.RunPeriodicAudit(c.Request.Context(), periodType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "audit completed", "period_type": periodType})
}

// RunFullAudit 执行完整审计
func (h *AuditHandler) RunFullAudit(c *gin.Context) {
	result, err := h.auditService.RunFullAudit(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetAuditHistory 获取审计历史
func (h *AuditHandler) GetAuditHistory(c *gin.Context) {
	var query model.FundConservationHistoryQuery

	if scope := c.Query("scope"); scope != "" {
		query.Scope = &scope
	}
	if ownerID := c.Query("owner_id"); ownerID != "" {
		if id, err := strconv.ParseInt(ownerID, 10, 64); err == nil {
			query.OwnerID = &id
		}
	}
	if periodType := c.Query("period_type"); periodType != "" {
		query.PeriodType = &periodType
	}

	query.Page = 1
	query.PageSize = 20
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			query.PageSize = ps
		}
	}

	items, total, err := h.auditService.GetAuditHistory(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// GetFundSummary 获取资金摘要
func (h *AuditHandler) GetFundSummary(c *gin.Context) {
	var userID *int64
	if uid := c.Query("user_id"); uid != "" {
		if id, err := strconv.ParseInt(uid, 10, 64); err == nil {
			userID = &id
		}
	}

	summary, err := h.fundService.GetFundSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}
