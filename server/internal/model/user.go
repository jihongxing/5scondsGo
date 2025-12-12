package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// UserRole 用户角色
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleOwner  UserRole = "owner"
	RolePlayer UserRole = "player"
)

// UserStatus 用户状态
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

// User 用户模型
type User struct {
	ID           int64      `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Role         UserRole   `json:"role" db:"role"`
	InvitedBy    *int64     `json:"invited_by,omitempty" db:"invited_by"`
	InviteCode   *string    `json:"invite_code,omitempty" db:"invite_code"`

	// 玩家余额
	Balance        decimal.Decimal `json:"balance" db:"balance"`
	FrozenBalance  decimal.Decimal `json:"frozen_balance" db:"frozen_balance"`
	BalanceVersion int64           `json:"balance_version" db:"balance_version"`

	// 房主专属字段
	OwnerRoomBalance   decimal.Decimal `json:"owner_room_balance,omitempty" db:"owner_room_balance"`     // 房主佣金收益
	OwnerMarginBalance decimal.Decimal `json:"owner_margin_balance,omitempty" db:"owner_margin_balance"` // 保证金（固定不变，防跑路担保）

	// 用户偏好设置
	Language string `json:"language" db:"language"`

	Status    UserStatus `json:"status" db:"status"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// UpdateLanguageReq 更新语言偏好请求
type UpdateLanguageReq struct {
	Language string `json:"language" binding:"required"`
}

// IsAdmin 是否是管理员
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsOwner 是否是房主
func (u *User) IsOwner() bool {
	return u.Role == RoleOwner
}

// IsPlayer 是否是玩家
func (u *User) IsPlayer() bool {
	return u.Role == RolePlayer
}

// CanInvite 是否可以邀请他人
func (u *User) CanInvite() bool {
	return u.Role == RoleAdmin || u.Role == RoleOwner
}

// TotalBalance 玩家总余额(可用+冻结)
func (u *User) TotalBalance() decimal.Decimal {
	return u.Balance.Add(u.FrozenBalance)
}

// OwnerTotalBalance 房主总余额（可用余额 + 佣金收益）
func (u *User) OwnerTotalBalance() decimal.Decimal {
	return u.Balance.Add(u.OwnerRoomBalance)
}

// UserPublicInfo 用户公开信息(用于广播)
type UserPublicInfo struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Role     UserRole `json:"role"`
}

// ToPublicInfo 转换为公开信息
func (u *User) ToPublicInfo() UserPublicInfo {
	return UserPublicInfo{
		ID:       u.ID,
		Username: u.Username,
		Role:     u.Role,
	}
}

// PlayerStat 房主名下玩家统计
type PlayerStat struct {
	ID               int64           `json:"id" db:"id"`
	Username         string          `json:"username" db:"username"`
	Balance          decimal.Decimal `json:"balance" db:"balance"`
	FrozenBalance    decimal.Decimal `json:"frozen_balance" db:"frozen_balance"`
	TotalDeposit     decimal.Decimal `json:"total_deposit" db:"total_deposit"`
	TotalWithdraw    decimal.Decimal `json:"total_withdraw" db:"total_withdraw"`
	RegistrationTime time.Time       `json:"registration_time" db:"created_at"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code" binding:"required,len=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// Role 角色类型别名 (为了兼容)
type Role = UserRole

const (
	RoleAdminAlias  = RoleAdmin
	RoleOwnerAlias  = RoleOwner
	RolePlayerAlias = RolePlayer
)

// RegisterReq 注册请求
type RegisterReq struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Password   string `json:"password" binding:"required,min=6"`
	InviteCode string `json:"invite_code"`
	Role       string `json:"role"` // "player" or "owner", default "player"
}

// LoginReq 登录请求
type LoginReq struct {
	Username          string `json:"username" binding:"required"`
	Password          string `json:"password" binding:"required"`
	DeviceFingerprint string `json:"device_fingerprint"` // 设备指纹（可选）
}

// LoginResp 登录响应
type LoginResp struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

// CreateOwnerReq 创建房主请求
type CreateOwnerReq struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserListQuery 用户列表查询
type UserListQuery struct {
	Role     *UserRole `form:"role"`
	Search   *string   `form:"search"`
	Page     int       `form:"page" binding:"min=1"`
	PageSize int       `form:"page_size" binding:"min=1,max=100"`
}
