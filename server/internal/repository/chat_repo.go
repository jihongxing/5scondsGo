package repository

import (
	"context"

	"github.com/fiveseconds/server/internal/model"
)

// ChatRepo 聊天消息仓库
type ChatRepo struct{}

// NewChatRepo 创建聊天仓库
func NewChatRepo() *ChatRepo {
	return &ChatRepo{}
}

// Create 创建聊天消息
func (r *ChatRepo) Create(ctx context.Context, msg *model.ChatMessage) error {
	sql := `INSERT INTO chat_messages (room_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return DB.QueryRow(ctx, sql, msg.RoomID, msg.UserID, msg.Content).Scan(&msg.ID, &msg.CreatedAt)
}

// GetHistory 获取聊天历史（最新的N条）
func (r *ChatRepo) GetHistory(ctx context.Context, roomID int64, limit int) ([]*model.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	sql := `SELECT cm.id, cm.room_id, cm.user_id, u.username, cm.content, cm.created_at
		FROM chat_messages cm
		JOIN users u ON cm.user_id = u.id
		WHERE cm.room_id = $1
		ORDER BY cm.created_at DESC
		LIMIT $2`

	rows, err := DB.Query(ctx, sql, roomID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.ChatMessage
	for rows.Next() {
		msg := &model.ChatMessage{}
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.UserID, &msg.Username, &msg.Content, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	// 反转顺序，使最旧的在前
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// DeleteOldMessages 删除超过指定数量的旧消息
func (r *ChatRepo) DeleteOldMessages(ctx context.Context, roomID int64, keepCount int) error {
	sql := `DELETE FROM chat_messages
		WHERE room_id = $1 AND id NOT IN (
			SELECT id FROM chat_messages
			WHERE room_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		)`
	_, err := DB.Exec(ctx, sql, roomID, keepCount)
	return err
}
