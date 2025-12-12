package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// InvitationStatus 邀请状态
type InvitationStatus string

const (
	InvitationPending  InvitationStatus = "pending"
	InvitationAccepted InvitationStatus = "accepted"
	InvitationDeclined InvitationStatus = "declined"
	InvitationExpired  InvitationStatus = "expired"
)

// RoomInvitation 房间邀请
type RoomInvitation struct {
	ID         int64            `json:"id" db:"id"`
	RoomID     int64            `json:"room_id" db:"room_id"`
	FromUserID int64            `json:"from_user_id" db:"from_user_id"`
	ToUserID   int64            `json:"to_user_id" db:"to_user_id"`
	Status     InvitationStatus `json:"status" db:"status"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	ExpiresAt  time.Time        `json:"expires_at" db:"expires_at"`

	// 关联数据（查询时填充）
	RoomName     string          `json:"room_name,omitempty"`
	BetAmount    decimal.Decimal `json:"bet_amount,omitempty"`
	PlayerCount  int             `json:"player_count,omitempty"`
	FromUsername string          `json:"from_username,omitempty"`
}

// InviteLink 邀请链接
type InviteLink struct {
	ID        int64     `json:"id" db:"id"`
	RoomID    int64     `json:"room_id" db:"room_id"`
	Code      string    `json:"code" db:"code"`
	CreatedBy int64     `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	UseCount  int       `json:"use_count" db:"use_count"`
	MaxUses   *int      `json:"max_uses,omitempty" db:"max_uses"`
}

// SendInvitationReq 发送邀请请求
type SendInvitationReq struct {
	ToUserID int64 `json:"to_user_id" binding:"required"`
}

// CreateInviteLinkReq 创建邀请链接请求
type CreateInviteLinkReq struct {
	MaxUses *int `json:"max_uses,omitempty"`
}

// InviteLinkResponse 邀请链接响应
type InviteLinkResponse struct {
	Code      string    `json:"code"`
	Link      string    `json:"link"`
	ExpiresAt time.Time `json:"expires_at"`
}
