package model

// WSMessageType WebSocket 消息类型
type WSMessageType string

const (
	// 客户端 -> 服务端
	WSTypeHeartbeat      WSMessageType = "heartbeat"
	WSTypeJoinRoom       WSMessageType = "join_room"
	WSTypeLeaveRoom      WSMessageType = "leave_room"
	WSTypeSetAutoReady   WSMessageType = "set_auto_ready"
	WSTypeJoinAsSpectator WSMessageType = "join_as_spectator"
	WSTypeSwitchToParticipant WSMessageType = "switch_to_participant"

	// 服务端 -> 客户端
	WSTypeError          WSMessageType = "error"
	WSTypeRoomState      WSMessageType = "room_state"
	WSTypePhaseChange    WSMessageType = "phase_change"
	WSTypePlayerJoin     WSMessageType = "player_join"
	WSTypePlayerLeave    WSMessageType = "player_leave"
	WSTypePlayerUpdate   WSMessageType = "player_update"
	WSTypeBettingDone    WSMessageType = "betting_done"
	WSTypeRoundResult    WSMessageType = "round_result"
	WSTypeRoundFailed    WSMessageType = "round_failed"
	WSTypeRoomLocked     WSMessageType = "room_locked"
	WSTypeBalanceUpdate  WSMessageType = "balance_update"
	WSTypeTimerSync      WSMessageType = "timer_sync"
	
	// 观战者相关
	WSTypeSpectatorJoin   WSMessageType = "spectator_join"
	WSTypeSpectatorLeave  WSMessageType = "spectator_leave"
	WSTypeSpectatorSwitch WSMessageType = "spectator_switch"

	// 聊天相关
	WSTypeChatMessage WSMessageType = "chat_message"
	WSTypeChatHistory WSMessageType = "chat_history"
	WSTypeSendChat    WSMessageType = "send_chat"

	// 表情相关
	WSTypeEmojiReaction WSMessageType = "emoji_reaction"
	WSTypeSendEmoji     WSMessageType = "send_emoji"

	// 好友相关
	WSTypeFriendRequest   WSMessageType = "friend_request"
	WSTypeFriendAccepted  WSMessageType = "friend_accepted"
	WSTypeFriendOnline    WSMessageType = "friend_online"
	WSTypeFriendOffline   WSMessageType = "friend_offline"

	// 邀请相关
	WSTypeRoomInvitation  WSMessageType = "room_invitation"
	WSTypeInviteResponse  WSMessageType = "invite_response"
	WSTypeSendInvite      WSMessageType = "send_invite"
	WSTypeRespondInvite   WSMessageType = "respond_invite"

	// 主题相关
	WSTypeThemeChange     WSMessageType = "theme_change"

	// 增量状态更新
	WSTypePhaseTick       WSMessageType = "phase_tick"

	// 玩家资格相关
	WSTypePlayerDisqualified WSMessageType = "player_disqualified"
	WSTypeRoundCancelled     WSMessageType = "round_cancelled"

	// 告警相关（管理员）
	WSTypeAlert           WSMessageType = "alert"
	WSTypeMetricsUpdate   WSMessageType = "metrics_update"
)

// WSMessage WebSocket 通用消息
type WSMessage struct {
	Type    WSMessageType `json:"type"`
	Payload interface{}   `json:"payload,omitempty"`
}

// WSError 错误消息
type WSError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// WSJoinRoom 加入房间请求
type WSJoinRoom struct {
	RoomID int64 `json:"room_id"`
}

// WSSetAutoReady 设置自动准备
type WSSetAutoReady struct {
	AutoReady bool `json:"auto_ready"`
}

// WSRoomState 房间完整状态
type WSRoomState struct {
	RoomID         int64                       `json:"room_id"`
	RoomName       string                      `json:"room_name"`
	BetAmount      string                      `json:"bet_amount"`
	WinnerCount    int                         `json:"winner_count"`
	MaxPlayers     int                         `json:"max_players"`
	MaxSpectators  int                         `json:"max_spectators"`
	Phase          GamePhase                   `json:"phase"`
	PhaseEndTime   int64                       `json:"phase_end_time"` // Unix毫秒
	CurrentRound   int                         `json:"current_round"`
	Players        map[int64]*WSPlayerState    `json:"players"`
	Spectators     map[int64]*WSSpectatorState `json:"spectators,omitempty"`
	PoolAmount     string                      `json:"pool_amount,omitempty"`
	IsSpectator    bool                        `json:"is_spectator,omitempty"` // 当前用户是否为观战者
}

// WSPlayerState 玩家状态
type WSPlayerState struct {
	UserID           int64  `json:"user_id"`
	Username         string `json:"username"`
	Balance          string `json:"balance"`
	AutoReady        bool   `json:"auto_ready"`
	IsOnline         bool   `json:"is_online"`
	Disqualified     bool   `json:"disqualified,omitempty"`      // 是否被取消资格
	DisqualifyReason string `json:"disqualify_reason,omitempty"` // 取消资格原因
}

// WSPhaseChange 阶段变化
type WSPhaseChange struct {
	Phase        GamePhase `json:"phase"`
	PhaseEndTime int64     `json:"phase_end_time"` // Unix毫秒
	Round        int       `json:"round,omitempty"`
}

// WSPlayerJoin 玩家加入
type WSPlayerJoin struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// WSPlayerLeave 玩家离开
type WSPlayerLeave struct {
	UserID int64 `json:"user_id"`
}

// WSPlayerUpdate 玩家状态更新
type WSPlayerUpdate struct {
	UserID           int64   `json:"user_id"`
	Balance          string  `json:"balance,omitempty"`
	AutoReady        *bool   `json:"auto_ready,omitempty"`
	IsOnline         *bool   `json:"is_online,omitempty"`
	Disqualified     *bool   `json:"disqualified,omitempty"`
	DisqualifyReason *string `json:"disqualify_reason,omitempty"`
}

// WSBettingDone 下注完成
type WSBettingDone struct {
	PoolAmount   string  `json:"pool_amount"`
	Participants []int64 `json:"participants"`
	Skipped      []int64 `json:"skipped"`
}

// WSRoundResult 回合结果
type WSRoundResult struct {
	RoundID        int64    `json:"round_id"`
	Winners        []int64  `json:"winners"`
	WinnerNames    []string `json:"winner_names"`
	PrizePerWinner string   `json:"prize_per_winner"`
	RevealSeed     string   `json:"reveal_seed"`
	CommitHash     string   `json:"commit_hash"`
}

// WSRoundFailed 回合失败
type WSRoundFailed struct {
	Reason   string  `json:"reason"`
	Refunded []int64 `json:"refunded,omitempty"`
}

// WSRoomLocked 房间锁定
type WSRoomLocked struct {
	Reason string `json:"reason"`
}

// WSBalanceUpdate 余额更新
type WSBalanceUpdate struct {
	Balance       string `json:"balance"`
	FrozenBalance string `json:"frozen_balance"`
}

// WSTimerSync 计时器同步
type WSTimerSync struct {
	ServerTime   int64 `json:"server_time"`   // Unix毫秒
	PhaseEndTime int64 `json:"phase_end_time"` // Unix毫秒
}

// WSSendChat 发送聊天消息请求
type WSSendChat struct {
	Content string `json:"content"`
}

// WSChatMessage 聊天消息广播
type WSChatMessage struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"` // Unix毫秒
}

// WSChatHistory 聊天历史
type WSChatHistory struct {
	Messages []*WSChatMessage `json:"messages"`
}

// WSSendEmoji 发送表情请求
type WSSendEmoji struct {
	Emoji string `json:"emoji"`
}

// WSEmojiReaction 表情反应广播
type WSEmojiReaction struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}


// WSFriendRequest 好友请求通知
type WSFriendRequest struct {
	RequestID  int64  `json:"request_id"`
	FromUserID int64  `json:"from_user_id"`
	FromUsername string `json:"from_username"`
}

// WSFriendAccepted 好友请求被接受通知
type WSFriendAccepted struct {
	FriendID   int64  `json:"friend_id"`
	FriendName string `json:"friend_name"`
}

// WSFriendOnline 好友上线通知
type WSFriendOnline struct {
	FriendID   int64  `json:"friend_id"`
	FriendName string `json:"friend_name"`
	RoomID     *int64 `json:"room_id,omitempty"`
}

// WSFriendOffline 好友下线通知
type WSFriendOffline struct {
	FriendID int64 `json:"friend_id"`
}

// WSRoomInvitation 房间邀请通知
type WSRoomInvitation struct {
	InvitationID int64  `json:"invitation_id"`
	RoomID       int64  `json:"room_id"`
	RoomName     string `json:"room_name"`
	BetAmount    string `json:"bet_amount"`
	PlayerCount  int    `json:"player_count"`
	FromUserID   int64  `json:"from_user_id"`
	FromUsername string `json:"from_username"`
}

// WSSendInvite 发送邀请请求
type WSSendInvite struct {
	RoomID   int64 `json:"room_id"`
	ToUserID int64 `json:"to_user_id"`
}

// WSRespondInvite 响应邀请请求
type WSRespondInvite struct {
	InvitationID int64 `json:"invitation_id"`
	Accept       bool  `json:"accept"`
}

// WSInviteResponse 邀请响应通知
type WSInviteResponse struct {
	InvitationID int64  `json:"invitation_id"`
	Accepted     bool   `json:"accepted"`
	FromUserID   int64  `json:"from_user_id"`
	FromUsername string `json:"from_username"`
}

// WSThemeChange 主题变更通知
type WSThemeChange struct {
	RoomID    int64  `json:"room_id"`
	ThemeName string `json:"theme_name"`
}

// WSPhaseTick 增量状态更新（只发送变化的字段）
type WSPhaseTick struct {
	ServerTime     int64   `json:"server_time"`                // 服务器时间戳（Unix毫秒）
	PhaseEndTime   int64   `json:"phase_end_time,omitempty"`   // 阶段结束时间（Unix毫秒）
	TimeRemaining  int64   `json:"time_remaining,omitempty"`   // 剩余时间（毫秒）
	Phase          *string `json:"phase,omitempty"`            // 当前阶段（仅变化时发送）
	PoolAmount     *string `json:"pool_amount,omitempty"`      // 奖池金额（仅变化时发送）
	PlayerCount    *int    `json:"player_count,omitempty"`     // 玩家数量（仅变化时发送）
	SpectatorCount *int    `json:"spectator_count,omitempty"`  // 观战者数量（仅变化时发送）
}

// WSPlayerDisqualified 玩家被取消资格（余额不足等）
type WSPlayerDisqualified struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Reason   string `json:"reason"` // insufficient_balance, offline, etc.
}

// WSRoundCancelled 回合取消（人数不足）
type WSRoundCancelled struct {
	Reason              string                  `json:"reason"`
	DisqualifiedPlayers []WSPlayerDisqualified  `json:"disqualified_players"`
	MinPlayersRequired  int                     `json:"min_players_required"`
	CurrentPlayers      int                     `json:"current_players"`
}
