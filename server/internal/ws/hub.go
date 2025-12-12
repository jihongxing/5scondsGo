package ws

import (
	"sync"
	"time"

	"github.com/fiveseconds/server/internal/model"
)

// Conn 抽象连接
type Conn interface {
	WriteJSON(v interface{}) error
	Close() error
}

// connInfo 连接信息，包含连接和最后活跃时间
type connInfo struct {
	conn       Conn
	lastActive time.Time
}

// Hub 管理房间到连接的映射
type Hub struct {
	mu      sync.RWMutex
	rooms   map[int64]map[int64]*connInfo // roomID -> userID -> connInfo
	userMap map[int64]*connInfo           // userID -> connInfo (便于私发)
	
	// 用于记录需要清理的死连接
	deadConns []struct {
		roomID int64
		userID int64
	}
}

func NewHub() *Hub {
	h := &Hub{
		rooms:   make(map[int64]map[int64]*connInfo),
		userMap: make(map[int64]*connInfo),
	}
	
	// 启动定期清理任务
	go h.cleanupLoop()
	
	return h
}

// cleanupLoop 定期清理不活跃的连接
func (h *Hub) cleanupLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		h.cleanupDeadConnections()
	}
}

// cleanupDeadConnections 清理超时的连接
func (h *Hub) cleanupDeadConnections() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	now := time.Now()
	timeout := 90 * time.Second // 90秒无活动视为死连接
	
	for roomID, users := range h.rooms {
		for userID, info := range users {
			if now.Sub(info.lastActive) > timeout {
				// 关闭连接
				info.conn.Close()
				delete(users, userID)
				delete(h.userMap, userID)
			}
		}
		// 清理空房间
		if len(users) == 0 {
			delete(h.rooms, roomID)
		}
	}
}

func (h *Hub) AddConn(roomID, userID int64, c Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 如果用户已有连接，先关闭旧连接
	if oldInfo, ok := h.userMap[userID]; ok {
		oldInfo.conn.Close()
	}

	info := &connInfo{
		conn:       c,
		lastActive: time.Now(),
	}

	if _, ok := h.rooms[roomID]; !ok {
		h.rooms[roomID] = make(map[int64]*connInfo)
	}
	h.rooms[roomID][userID] = info
	h.userMap[userID] = info
}

func (h *Hub) RemoveConn(roomID, userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if m, ok := h.rooms[roomID]; ok {
		if info, exists := m[userID]; exists {
			info.conn.Close()
			delete(m, userID)
		}
		if len(m) == 0 {
			delete(h.rooms, roomID)
		}
	}
	delete(h.userMap, userID)
}

// UpdateActivity 更新连接活跃时间
func (h *Hub) UpdateActivity(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if info, ok := h.userMap[userID]; ok {
		info.lastActive = time.Now()
	}
}

// BroadcastToRoom 广播
func (h *Hub) BroadcastToRoom(roomID int64, msg *model.WSMessage) {
	h.mu.RLock()
	// 复制连接映射，避免长时间持有锁
	conns := make(map[int64]Conn, len(h.rooms[roomID]))
	for k, v := range h.rooms[roomID] {
		conns[k] = v.conn
		v.lastActive = time.Now() // 更新活跃时间
	}
	h.mu.RUnlock()

	var failedUsers []int64
	for userID, c := range conns {
		if err := c.WriteJSON(msg); err != nil {
			// 记录失败的连接，稍后清理
			failedUsers = append(failedUsers, userID)
		}
	}
	
	// 清理发送失败的连接
	if len(failedUsers) > 0 {
		h.mu.Lock()
		for _, userID := range failedUsers {
			if m, ok := h.rooms[roomID]; ok {
				if info, exists := m[userID]; exists {
					info.conn.Close()
					delete(m, userID)
				}
			}
			delete(h.userMap, userID)
		}
		h.mu.Unlock()
	}
}

// SendToUser 私发
func (h *Hub) SendToUser(userID int64, msg *model.WSMessage) {
	h.mu.RLock()
	info := h.userMap[userID]
	h.mu.RUnlock()
	
	if info != nil {
		if err := info.conn.WriteJSON(msg); err != nil {
			// 发送失败，关闭连接
			h.mu.Lock()
			if currentInfo, ok := h.userMap[userID]; ok && currentInfo == info {
				info.conn.Close()
				delete(h.userMap, userID)
				// 从所有房间中移除
				for _, users := range h.rooms {
					delete(users, userID)
				}
			}
			h.mu.Unlock()
		} else {
			// 更新活跃时间
			h.mu.Lock()
			info.lastActive = time.Now()
			h.mu.Unlock()
		}
	}
}

// GetStats 获取连接统计信息
func (h *Hub) GetStats() (totalConns int, totalRooms int) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return len(h.userMap), len(h.rooms)
}

// GetRoomUserCount 获取房间用户数
func (h *Hub) GetRoomUserCount(roomID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if users, ok := h.rooms[roomID]; ok {
		return len(users)
	}
	return 0
}
