package game

import (
	"context"
	"sync"

	"github.com/fiveseconds/server/internal/cache"
	"github.com/fiveseconds/server/internal/repository"

	"go.uber.org/zap"
)

// Manager 游戏房间管理器
type Manager struct {
	mu         sync.RWMutex
	rooms      map[int64]*RoomProcessor
	broadcaster Broadcaster

	userRepo     *repository.UserRepo
	roomRepo     *repository.RoomRepo
	gameRepo     *repository.GameRepo
	txRepo       *repository.TransactionRepo
	platformRepo *repository.PlatformRepo
	balanceCache *cache.BalanceCache
	riskChecker  RiskChecker
	logger       *zap.Logger
}

// NewManager 创建管理器
func NewManager(
	broadcaster Broadcaster,
	userRepo *repository.UserRepo,
	roomRepo *repository.RoomRepo,
	gameRepo *repository.GameRepo,
	txRepo *repository.TransactionRepo,
	platformRepo *repository.PlatformRepo,
	balanceCache *cache.BalanceCache,
	riskChecker RiskChecker,
	logger *zap.Logger,
) *Manager {
	return &Manager{
		rooms:        make(map[int64]*RoomProcessor),
		broadcaster:  broadcaster,
		userRepo:     userRepo,
		roomRepo:     roomRepo,
		gameRepo:     gameRepo,
		txRepo:       txRepo,
		platformRepo: platformRepo,
		balanceCache: balanceCache,
		riskChecker:  riskChecker,
		logger:       logger,
	}
}

// GetOrCreateRoom 获取或创建房间处理器
func (m *Manager) GetOrCreateRoom(ctx context.Context, roomID int64) (*RoomProcessor, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rp, ok := m.rooms[roomID]; ok {
		return rp, nil
	}

	room, err := m.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	rp := NewRoomProcessor(
		room,
		m.broadcaster,
		m.userRepo,
		m.roomRepo,
		m.gameRepo,
		m.txRepo,
		m.platformRepo,
		m.balanceCache,
		m.riskChecker,
		m.logger,
	)

	// 从数据库加载已有玩家（服务器重启后恢复状态）
	// 所有玩家初始状态为离线，等待他们重新连接 WebSocket
	if err := rp.LoadPlayersFromDB(ctx); err != nil {
		m.logger.Warn("Failed to load players from DB", zap.Int64("room_id", roomID), zap.Error(err))
	}

	rp.Start()
	m.rooms[roomID] = rp

	m.logger.Info("Room processor created", zap.Int64("room_id", roomID))
	return rp, nil
}

// GetRoom 获取房间处理器
func (m *Manager) GetRoom(roomID int64) *RoomProcessor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.rooms[roomID]
}

// RemoveRoom 移除房间处理器
func (m *Manager) RemoveRoom(roomID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rp, ok := m.rooms[roomID]; ok {
		rp.Stop()
		delete(m.rooms, roomID)
		m.logger.Info("Room processor removed", zap.Int64("room_id", roomID))
	}
}

// GetRoomCount 获取活跃房间数量
func (m *Manager) GetRoomCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms)
}

// GetPlayerCount 获取在线玩家数量
func (m *Manager) GetPlayerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, rp := range m.rooms {
		for _, p := range rp.State.Players {
			if p.IsOnline {
				count++
			}
		}
	}
	return count
}

// Shutdown 关闭所有房间
func (m *Manager) Shutdown() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for roomID, rp := range m.rooms {
		rp.Stop()
		delete(m.rooms, roomID)
	}
	m.logger.Info("All room processors stopped")
}
