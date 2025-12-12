package model

import "time"

// ChatMessage 聊天消息
type ChatMessage struct {
	ID        int64     `json:"id" db:"id"`
	RoomID    int64     `json:"room_id" db:"room_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Username  string    `json:"username" db:"username"`
	Content   string    `json:"content" db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ChatMessageReq 发送聊天消息请求
type ChatMessageReq struct {
	Content string `json:"content" binding:"required,max=200"`
}

// ChatHistoryQuery 聊天历史查询
type ChatHistoryQuery struct {
	RoomID int64 `form:"room_id" binding:"required"`
	Limit  int   `form:"limit"`
}

// EmojiReaction 表情反应
type EmojiReaction struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
	RoomID   int64  `json:"room_id"`
}

// 预定义表情列表
var PredefinedEmojis = []string{
	"happy", "sad", "angry", "surprised",
	"thumbs_up", "thumbs_down", "clap", "fire",
	"heart", "laugh", "cry", "cool",
}

// IsValidEmoji 检查是否为有效表情
func IsValidEmoji(emoji string) bool {
	for _, e := range PredefinedEmojis {
		if e == emoji {
			return true
		}
	}
	return false
}
