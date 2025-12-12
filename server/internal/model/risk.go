package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// RiskFlagType 风控标记类型
type RiskFlagType string

const (
	RiskFlagConsecutiveWins RiskFlagType = "consecutive_wins"
	RiskFlagHighWinRate     RiskFlagType = "high_win_rate"
	RiskFlagMultiAccount    RiskFlagType = "multi_account"
	RiskFlagLargeTransaction RiskFlagType = "large_transaction"
)

// RiskFlagStatus 风控标记状态
type RiskFlagStatus string

const (
	RiskFlagStatusPending   RiskFlagStatus = "pending"
	RiskFlagStatusReviewed  RiskFlagStatus = "reviewed"
	RiskFlagStatusConfirmed RiskFlagStatus = "confirmed"
	RiskFlagStatusDismissed RiskFlagStatus = "dismissed"
)

// RiskFlag 风控标记
type RiskFlag struct {
	ID         int64          `json:"id" db:"id"`
	UserID     int64          `json:"user_id" db:"user_id"`
	FlagType   RiskFlagType   `json:"flag_type" db:"flag_type"`
	Details    string         `json:"details" db:"details"`
	Status     RiskFlagStatus `json:"status" db:"status"`
	ReviewedBy *int64         `json:"reviewed_by,omitempty" db:"reviewed_by"`
	ReviewedAt *time.Time     `json:"reviewed_at,omitempty" db:"reviewed_at"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

// RiskFlagDetails 风控标记详情
type RiskFlagDetails struct {
	ConsecutiveWins int             `json:"consecutive_wins,omitempty"`
	WinRate         float64         `json:"win_rate,omitempty"`
	TotalRounds     int             `json:"total_rounds,omitempty"`
	DeviceFingerprint string        `json:"device_fingerprint,omitempty"`
	RelatedUserIDs  []int64         `json:"related_user_ids,omitempty"`
	TransactionAmount decimal.Decimal `json:"transaction_amount,omitempty"`
}

// RiskFlagListQuery 风控标记列表查询
type RiskFlagListQuery struct {
	UserID   *int64          `form:"user_id"`
	FlagType *RiskFlagType   `form:"flag_type"`
	Status   *RiskFlagStatus `form:"status"`
	Page     int             `form:"page" binding:"min=1"`
	PageSize int             `form:"page_size" binding:"min=1,max=100"`
}

// ReviewRiskFlagReq 审核风控标记请求
type ReviewRiskFlagReq struct {
	Action string `json:"action" binding:"required,oneof=confirm dismiss"`
	Remark string `json:"remark"`
}

// RiskConfig 风控配置
type RiskConfig struct {
	ConsecutiveWinThreshold int     `json:"consecutive_win_threshold"` // 连续获胜阈值
	WinRateThreshold        float64 `json:"win_rate_threshold"`        // 胜率阈值
	WinRateMinRounds        int     `json:"win_rate_min_rounds"`       // 胜率检测最小回合数
	LargeTransactionAmount  decimal.Decimal `json:"large_transaction_amount"` // 大额交易阈值
	DailyVolumeThreshold    decimal.Decimal `json:"daily_volume_threshold"`   // 日交易量阈值
}

// DefaultRiskConfig 默认风控配置
var DefaultRiskConfig = RiskConfig{
	ConsecutiveWinThreshold: 10,
	WinRateThreshold:        0.8,
	WinRateMinRounds:        50,
	LargeTransactionAmount:  decimal.NewFromInt(10000),
	DailyVolumeThreshold:    decimal.NewFromInt(100000),
}
