package handler

import (
	"fmt"
	"net/http"

	"github.com/fiveseconds/server/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// WalletHandler 钱包处理器
type WalletHandler struct {
	walletService *service.WalletService
}

// NewWalletHandler 创建钱包处理器
func NewWalletHandler(walletService *service.WalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

// GetWallet 获取钱包信息
func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID := GetUserID(c)
	wallet, err := h.walletService.GetWallet(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// GetTransactions 获取交易历史
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	userID := GetUserID(c)
	
	page := 1
	pageSize := 20
	if p, ok := c.GetQuery("page"); ok {
		if v, err := parseInt(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps, ok := c.GetQuery("page_size"); ok {
		if v, err := parseInt(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	records, total, err := h.walletService.GetTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": records,
		"total": total,
	})
}

// GetEarnings 获取收益统计
func (h *WalletHandler) GetEarnings(c *gin.Context) {
	userID := GetUserID(c)
	earnings, err := h.walletService.GetEarnings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, earnings)
}

// TransferEarningsReq 收益转余额请求
type TransferEarningsReq struct {
	Amount string `json:"amount" binding:"required"` // 使用字符串避免浮点精度问题
}

// TransferEarnings 房主收益转可提现余额
func (h *WalletHandler) TransferEarnings(c *gin.Context) {
	userID := GetUserID(c)

	var req TransferEarningsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount format"})
		return
	}
	
	if amount.LessThanOrEqual(decimal.Zero) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be positive"})
		return
	}

	if err := h.walletService.TransferEarningsToBalance(c.Request.Context(), userID, amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

// parseInt 解析整数
func parseInt(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid character: %c", c)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
