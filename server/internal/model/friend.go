package model

import (
	"time"
)

// FriendRequestStatus 好友请求状态
type FriendRequestStatus string

const (
	FriendRequestPending  FriendRequestStatus = "pending"
	FriendRequestAccepted FriendRequestStatus = "accepted"
	FriendRequestRejected FriendRequestStatus = "rejected"
)

// FriendRequest 好友请求
type FriendRequest struct {
	ID         int64               `json:"id" db:"id"`
	FromUserID int64               `json:"from_user_id" db:"from_user_id"`
	ToUserID   int64               `json:"to_user_id" db:"to_user_id"`
	Status     FriendRequestStatus `json:"status" db:"status"`
	CreatedAt  time.Time           `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at" db:"updated_at"`

	// 关联数据（查询时填充）
	FromUser *UserPublicInfo `json:"from_user,omitempty"`
	ToUser   *UserPublicInfo `json:"to_user,omitempty"`
}

// Friend 好友关系
type Friend struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	FriendID  int64     `json:"friend_id" db:"friend_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// FriendInfo 好友信息（包含在线状态）
type FriendInfo struct {
	ID         int64    `json:"id"`
	Username   string   `json:"username"`
	Role       UserRole `json:"role"`
	IsOnline   bool     `json:"is_online"`
	CurrentRoom *int64  `json:"current_room,omitempty"`
	RoomName   *string  `json:"room_name,omitempty"`
}

// SendFriendRequestReq 发送好友请求
type SendFriendRequestReq struct {
	ToUserID int64 `json:"to_user_id" binding:"required"`
}

// FriendRequestActionReq 好友请求操作
type FriendRequestActionReq struct {
	RequestID int64 `json:"request_id" binding:"required"`
}
