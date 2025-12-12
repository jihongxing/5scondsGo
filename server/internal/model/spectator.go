package model

import "time"

// SpectatorState 观战者状态
type SpectatorState struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

// RoomSpectator 房间观战者数据库模型
type RoomSpectator struct {
	ID       int64     `json:"id" db:"id"`
	RoomID   int64     `json:"room_id" db:"room_id"`
	UserID   int64     `json:"user_id" db:"user_id"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// WSSpectatorState 观战者状态（WebSocket）
type WSSpectatorState struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// WSSpectatorJoin 观战者加入
type WSSpectatorJoin struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// WSSpectatorLeave 观战者离开
type WSSpectatorLeave struct {
	UserID int64 `json:"user_id"`
}

// WSSpectatorSwitch 观战者切换为参与者
type WSSpectatorSwitch struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// JoinAsSpectatorReq 以观战者身份加入房间请求
type JoinAsSpectatorReq struct {
	Password string `json:"password"`
}

// SwitchToParticipantReq 切换为参与者请求
type SwitchToParticipantReq struct{}
