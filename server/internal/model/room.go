package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// RoomStatus 房间状态
type RoomStatus string

const (
	RoomStatusActive RoomStatus = "active"
	RoomStatusPaused RoomStatus = "paused"
	RoomStatusLocked RoomStatus = "locked"
)

// Room 房间模型
type Room struct {
	ID         int64  `json:"id" db:"id"`
	OwnerID    int64  `json:"owner_id" db:"owner_id"`
	OwnerName  string `json:"owner_name,omitempty" db:"-"`
	Name       string `json:"name" db:"name"`
	InviteCode string `json:"invite_code" db:"invite_code"`

	// 房间配置
	BetAmount              decimal.Decimal `json:"bet_amount" db:"bet_amount"`
	WinnerCount            int             `json:"winner_count" db:"winner_count"`
	MaxPlayers             int             `json:"max_players" db:"max_players"`
	OwnerCommissionRate    decimal.Decimal `json:"owner_commission_rate" db:"owner_commission_rate"`
	PlatformCommissionRate decimal.Decimal `json:"platform_commission_rate" db:"platform_commission_rate"`
	Password               *string         `json:"-" db:"password"`

	// 状态
	Status RoomStatus `json:"status" db:"status"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RoomListItem 房间列表项（带当前玩家数）
type RoomListItem struct {
	*Room
	CurrentPlayers int    `json:"current_players"`
	HasPassword    bool   `json:"has_password"`
	OwnerName      string `json:"owner_name,omitempty"`
}

// TotalCommissionRate 总抽成比例
func (r *Room) TotalCommissionRate() decimal.Decimal {
	return r.OwnerCommissionRate.Add(r.PlatformCommissionRate)
}

// CreateRoomReq 创建房间请求
type CreateRoomReq struct {
	Name                   string `json:"name" binding:"required"`
	BetAmount              string `json:"bet_amount" binding:"required"`              // 使用字符串避免浮点精度问题
	WinnerCount            int    `json:"winner_count" binding:"required,min=1"`
	MaxPlayers             int    `json:"max_players" binding:"required,min=2,max=100"`
	OwnerCommissionRate    string `json:"owner_commission_rate"`                      // 使用字符串避免浮点精度问题
	PlatformCommissionRate string `json:"platform_commission_rate"`                   // 使用字符串避免浮点精度问题
	Password               string `json:"password"`
}

// GetBetAmountDecimal 获取下注金额的 Decimal 类型
func (r *CreateRoomReq) GetBetAmountDecimal() decimal.Decimal {
	d, err := decimal.NewFromString(r.BetAmount)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// GetOwnerCommissionRateDecimal 获取房主佣金率的 Decimal 类型
func (r *CreateRoomReq) GetOwnerCommissionRateDecimal() decimal.Decimal {
	if r.OwnerCommissionRate == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(r.OwnerCommissionRate)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// GetPlatformCommissionRateDecimal 获取平台佣金率的 Decimal 类型
func (r *CreateRoomReq) GetPlatformCommissionRateDecimal() decimal.Decimal {
	if r.PlatformCommissionRate == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(r.PlatformCommissionRate)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// UpdateRoomReq 更新房间请求
type UpdateRoomReq struct {
	Name        string          `json:"name"`
	BetAmount   decimal.Decimal `json:"bet_amount"`
	WinnerCount int             `json:"winner_count"`
	MaxPlayers  int             `json:"max_players"`
}

// UpdateRoomStatusReq 管理端更新房间状态请求
type UpdateRoomStatusReq struct {
	Status RoomStatus `json:"status" binding:"required"`
}

// RoomListQuery 房间列表查询
type RoomListQuery struct {
	OwnerID   *int64      `form:"owner_id"`
	Status    *RoomStatus `form:"status"`
	InvitedBy *int64      `form:"-"` // 内部使用，用于过滤关联房主
	Page      int         `form:"page" binding:"min=1"`
	PageSize  int         `form:"page_size" binding:"min=1,max=100"`
}

// JoinRoomReq 加入房间请求
type JoinRoomReq struct {
	Password string `json:"password"`
}

// RoomPlayer 房间内玩家
type RoomPlayer struct {
	RoomID    int64      `json:"room_id" db:"room_id"`
	UserID    int64      `json:"user_id" db:"user_id"`
	AutoReady bool       `json:"auto_ready" db:"auto_ready"`
	JoinedAt  time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt    *time.Time `json:"left_at,omitempty" db:"left_at"`
}

