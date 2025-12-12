package ws

import (
	"sync"

	"github.com/fiveseconds/server/internal/model"
)

// Conn 抽象连接
type Conn interface {
	WriteJSON(v interface{}) error
	Close() error
}

// Hub 管理房间到连接的映射
type Hub struct {
	mu      sync.RWMutex
	rooms   map[int64]map[int64]Conn // roomID -> userID -> conn
	userMap map[int64]Conn           // userID -> conn (便于私发)
}

func NewHub() *Hub {
	return &Hub{
		rooms:   make(map[int64]map[int64]Conn),
		userMap: make(map[int64]Conn),
	}
}

func (h *Hub) AddConn(roomID, userID int64, c Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[int64]Conn)
	}
	h.rooms[roomID][userID] = c
	h.userMap[userID] = c
}

func (h *Hub) RemoveConn(roomID, userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if m, ok := h.rooms[roomID]; ok {
		delete(m, userID)
		if len(m) == 0 {
			delete(h.rooms, roomID)
		}
	}
	delete(h.userMap, userID)
}

// BroadcastToRoom 广播
func (h *Hub) BroadcastToRoom(roomID int64, msg *model.WSMessage) {
	h.mu.RLock()
	// 复制连接映射，避免长时间持有锁
	conns := make(map[int64]Conn, len(h.rooms[roomID]))
	for k, v := range h.rooms[roomID] {
		conns[k] = v
	}
	h.mu.RUnlock()

	for userID, c := range conns {
		if err := c.WriteJSON(msg); err != nil {
			// 记录错误但不中断广播
			// 连接可能已断开，后续会被清理
			_ = userID // 避免未使用变量警告，实际生产中应记录日志
		}
	}
}

// SendToUser 私发
func (h *Hub) SendToUser(userID int64, msg *model.WSMessage) {
	h.mu.RLock()
	c := h.userMap[userID]
	h.mu.RUnlock()
	if c != nil {
		// 发送失败时连接可能已断开，后续会被清理
		_ = c.WriteJSON(msg)
	}
}
