package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// AlertType 告警类型
type AlertType string

const (
	AlertTypeNegativeBalance    AlertType = "negative_balance"
	AlertTypeNegativeCustody    AlertType = "negative_custody"
	AlertTypeLargeTransaction   AlertType = "large_transaction"
	AlertTypeDailyVolumeExceed  AlertType = "daily_volume_exceed"
	AlertTypeSettlementFailed   AlertType = "settlement_failed"
	AlertTypeConservationFailed AlertType = "conservation_failed"
	AlertTypeRiskFlagCreated    AlertType = "risk_flag_created"
)

// AlertSeverity 告警严重程度
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus 告警状态
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
)

// Alert 告警记录
type Alert struct {
	ID             int64         `json:"id" db:"id"`
	AlertType      AlertType     `json:"alert_type" db:"alert_type"`
	Severity       AlertSeverity `json:"severity" db:"severity"`
	Title          string        `json:"title" db:"title"`
	Details        string        `json:"details" db:"details"`
	Status         AlertStatus   `json:"status" db:"status"`
	AcknowledgedBy *int64        `json:"acknowledged_by,omitempty" db:"acknowledged_by"`
	AcknowledgedAt *time.Time    `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	CreatedAt      time.Time     `json:"created_at" db:"created_at"`
}

// AlertDetails 告警详情
type AlertDetails struct {
	UserID            *int64          `json:"user_id,omitempty"`
	Username          string          `json:"username,omitempty"`
	RoomID            *int64          `json:"room_id,omitempty"`
	RoomName          string          `json:"room_name,omitempty"`
	Amount            decimal.Decimal `json:"amount,omitempty"`
	Balance           decimal.Decimal `json:"balance,omitempty"`
	Difference        decimal.Decimal `json:"difference,omitempty"`
	FailureCount      int             `json:"failure_count,omitempty"`
	RiskFlagID        *int64          `json:"risk_flag_id,omitempty"`
	RiskFlagType      string          `json:"risk_flag_type,omitempty"`
	AdditionalInfo    string          `json:"additional_info,omitempty"`
}

// AlertListQuery 告警列表查询
type AlertListQuery struct {
	AlertType *AlertType    `form:"alert_type"`
	Severity  *AlertSeverity `form:"severity"`
	Status    *AlertStatus  `form:"status"`
	Page      int           `form:"page" binding:"min=1"`
	PageSize  int           `form:"page_size" binding:"min=1,max=100"`
}

// AcknowledgeAlertReq 确认告警请求
type AcknowledgeAlertReq struct {
	Remark string `json:"remark"`
}

// WSAlert WebSocket 告警通知
type WSAlert struct {
	ID        int64         `json:"id"`
	AlertType AlertType     `json:"alert_type"`
	Severity  AlertSeverity `json:"severity"`
	Title     string        `json:"title"`
	Details   string        `json:"details"`
	CreatedAt int64         `json:"created_at"` // Unix毫秒
}
