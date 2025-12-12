package service

import (
	"context"
	"errors"

	"github.com/fiveseconds/server/internal/game"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"

	"github.com/shopspring/decimal"
)

var (
	ErrRoomFull         = errors.New("room is full")
	ErrRoomNotActive    = errors.New("room is not active")
	ErrNotRoomOwner     = errors.New("not room owner")
	ErrAlreadyInRoom    = errors.New("already in room")
	ErrNotInRoom        = errors.New("not in room")
	ErrInvalidBetAmount = errors.New("invalid bet amount")
	ErrInvalidPassword  = errors.New("invalid room password")
)

const (
	// MinPlayers 最小参与人数
	MinPlayers = 2
)

// 有效的下注金额
var ValidBetAmounts = []decimal.Decimal{
	decimal.NewFromInt(5),
	decimal.NewFromInt(10),
	decimal.NewFromInt(20),
	decimal.NewFromInt(50),
	decimal.NewFromInt(100),
	decimal.NewFromInt(200),
}

type RoomService struct {
	roomRepo *repository.RoomRepo
	userRepo *repository.UserRepo
	manager  *game.Manager
}

func NewRoomService(roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, manager *game.Manager) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
		manager:  manager,
	}
}

// MinMarginBalanceForRoom 创建房间所需的最低保证金
var MinMarginBalanceForRoom = decimal.NewFromInt(2000)

// ErrInsufficientMarginForRoom 保证金不足错误
var ErrInsufficientMarginForRoom = errors.New("insufficient margin balance, minimum 2000 required to create room")

// CreateRoom 创建房间
func (s *RoomService) CreateRoom(ctx context.Context, ownerID int64, req *model.CreateRoomReq) (*model.Room, error) {
	// 验证房主保证金余额
	owner, err := s.userRepo.GetByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if owner.OwnerMarginBalance.LessThan(MinMarginBalanceForRoom) {
		return nil, ErrInsufficientMarginForRoom
	}

	betAmount := req.GetBetAmountDecimal()
	ownerCommissionRate := req.GetOwnerCommissionRateDecimal()
	platformCommissionRate := req.GetPlatformCommissionRateDecimal()

	// 验证下注金额
	if betAmount.IsZero() {
		return nil, errors.New("invalid bet amount format")
	}
	if !isValidBetAmount(betAmount) {
		return nil, ErrInvalidBetAmount
	}

	// 验证赢家数量必须小于最大玩家数
	if req.WinnerCount >= req.MaxPlayers {
		return nil, errors.New("winner count must be less than max players")
	}

	// 验证赢家数量必须小于最小参与人数（至少2人）
	if req.WinnerCount >= MinPlayers {
		// 如果赢家数量 >= 最小参与人数，则无法正常游戏
		// 但这个检查可能过于严格，因为实际参与人数可能更多
		// 保守起见，只检查 WinnerCount < MaxPlayers
	}

	// 验证房主佣金率（0-8%，即 0.00-0.08）
	maxOwnerRate, _ := decimal.NewFromString("0.08")
	if ownerCommissionRate.LessThan(decimal.Zero) || ownerCommissionRate.GreaterThan(maxOwnerRate) {
		return nil, errors.New("owner commission rate must be between 0% and 8%")
	}

	// 如果前端没有传平台佣金率，使用默认值 2%
	if platformCommissionRate.IsZero() {
		platformCommissionRate, _ = decimal.NewFromString("0.02")
	}

	// 验证总抽成比例（不能超过 10%，即 0.10）
	maxTotalRate, _ := decimal.NewFromString("0.10")
	totalRate := ownerCommissionRate.Add(platformCommissionRate)
	if totalRate.GreaterThan(maxTotalRate) {
		return nil, errors.New("total commission rate cannot exceed 10%")
	}

	// 生成房间邀请码
	var inviteCode string
	for {
		code, err := game.GenerateInviteCode()
		if err != nil {
			return nil, err
		}
		exists, err := s.roomRepo.InviteCodeExists(ctx, code)
		if err != nil {
			return nil, err
		}
		if !exists {
			inviteCode = code
			break
		}
	}

	room := &model.Room{
		OwnerID:                ownerID,
		Name:                   req.Name,
		InviteCode:             inviteCode,
		BetAmount:              betAmount,
		WinnerCount:            req.WinnerCount,
		MaxPlayers:             req.MaxPlayers,
		OwnerCommissionRate:    ownerCommissionRate,
		PlatformCommissionRate: platformCommissionRate,
		Status:                 model.RoomStatusActive,
	}
	if req.Password != "" {
		room.Password = &req.Password
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom 获取房间
func (s *RoomService) GetRoom(ctx context.Context, roomID int64) (*model.Room, error) {
	return s.roomRepo.GetByID(ctx, roomID)
}

// GetRoomByInviteCode 根据邀请码获取房间
func (s *RoomService) GetRoomByInviteCode(ctx context.Context, code string) (*model.Room, error) {
	return s.roomRepo.GetByInviteCode(ctx, code)
}

// ListRooms 列表房间（附带当前玩家数）
func (s *RoomService) ListRooms(ctx context.Context, query *model.RoomListQuery) ([]*model.RoomListItem, int64, error) {
	// 获取当前用户信息，根据 invited_by 过滤
	// 注意：Handler 必须在调用前把 UserID 放到 Context 或者直接传 UserID 进来
	// 但这里 query 只有 UserID (不, query 甚至没有 UserID)
	// 这是一个设计问题：Handler 应该把 Context User 的信息查出来，或者 query 应该包含 Filter 条件。
	// 我们假设 Handler 负责把 "InvitedBy" 填入 query (如果是 Player)
	// 但 ListRooms 签名是 query。
	// 让我们在 handler 里处理。Handler 调用 authService 拿到 invitedBy，然后填入 query。
	// 这里只负责调用 Repo。

	rooms, total, err := s.roomRepo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	items := make([]*model.RoomListItem, 0, len(rooms))
	for _, room := range rooms {
		count, err := s.roomRepo.CountRoomPlayers(ctx, room.ID)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, &model.RoomListItem{
			Room:           room,
			CurrentPlayers: count,
			HasPassword:    room.Password != nil && *room.Password != "",
		})
	}

	return items, total, nil
}

// ListOwnerRooms 获取房主的房间
func (s *RoomService) ListOwnerRooms(ctx context.Context, ownerID int64) ([]*model.Room, error) {
	return s.roomRepo.ListByOwner(ctx, ownerID)
}

// UpdateRoom 更新房间配置
func (s *RoomService) UpdateRoom(ctx context.Context, roomID, ownerID int64, req *model.UpdateRoomReq) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.OwnerID != ownerID {
		return ErrNotRoomOwner
	}

	if req.Name != "" {
		room.Name = req.Name
	}
	if !req.BetAmount.IsZero() {
		if !isValidBetAmount(req.BetAmount) {
			return ErrInvalidBetAmount
		}
		room.BetAmount = req.BetAmount
	}
	if req.WinnerCount > 0 {
		room.WinnerCount = req.WinnerCount
	}
	if req.MaxPlayers > 0 {
		room.MaxPlayers = req.MaxPlayers
	}

	return s.roomRepo.Update(ctx, room)
}

// UpdateRoomStatus 更新房间状态（房主调用）
func (s *RoomService) UpdateRoomStatus(ctx context.Context, roomID, ownerID int64, status model.RoomStatus) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.OwnerID != ownerID {
		return ErrNotRoomOwner
	}

	return s.roomRepo.UpdateStatus(ctx, roomID, status)
}

// AdminUpdateRoomStatus 管理员更新房间状态（不校验 owner）
func (s *RoomService) AdminUpdateRoomStatus(ctx context.Context, roomID int64, status model.RoomStatus) error {
	return s.roomRepo.UpdateStatus(ctx, roomID, status)
}

// JoinRoom 加入房间
func (s *RoomService) JoinRoom(ctx context.Context, userID, roomID int64, password string) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.Status != model.RoomStatusActive {
		return ErrRoomNotActive
	}

	// 验证密码
	if room.Password != nil && *room.Password != "" {
		if *room.Password != password {
			return ErrInvalidPassword
		}
	}

	// 检查是否已在其他房间
	currentRoom, err := s.roomRepo.GetPlayerRoom(ctx, userID)
	if err != nil {
		return err
	}
	if currentRoom != nil {
		if currentRoom.RoomID == roomID {
			// 已经在目标房间，直接返回成功
			return nil
		}
		// 在其他房间，先离开（RemovePlayer 会处理内存、数据库清理和广播）
		if processor := s.manager.GetRoom(currentRoom.RoomID); processor != nil {
			processor.RemovePlayer(userID)
		} else {
			// 如果游戏引擎中没有这个房间，直接从数据库删除
			if err := s.roomRepo.RemovePlayer(ctx, currentRoom.RoomID, userID); err != nil {
				return err
			}
		}
	}

	// 检查房间是否已满
	count, err := s.roomRepo.CountRoomPlayers(ctx, roomID)
	if err != nil {
		return err
	}
	if count >= room.MaxPlayers {
		return ErrRoomFull
	}

	// 加入房间
	rp := &model.RoomPlayer{
		RoomID:    roomID,
		UserID:    userID,
		AutoReady: false,
	}
	if err := s.roomRepo.AddPlayer(ctx, rp); err != nil {
		return err
	}

	// 通知游戏引擎
	user, _ := s.userRepo.GetByID(ctx, userID)
	if processor, err := s.manager.GetOrCreateRoom(ctx, roomID); err == nil && user != nil {
		processor.AddPlayer(user)
	}

	return nil
}

// LeaveRoom 离开房间
func (s *RoomService) LeaveRoom(ctx context.Context, userID int64) error {
	currentRoom, err := s.roomRepo.GetPlayerRoom(ctx, userID)
	if err != nil {
		return err
	}
	if currentRoom == nil {
		return ErrNotInRoom
	}

	// 通知游戏引擎（RemovePlayer 会处理内存、数据库清理和广播）
	if processor := s.manager.GetRoom(currentRoom.RoomID); processor != nil {
		processor.RemovePlayer(userID)
	} else {
		// 如果游戏引擎中没有这个房间，直接从数据库删除
		if err := s.roomRepo.RemovePlayer(ctx, currentRoom.RoomID, userID); err != nil {
			return err
		}
	}

	return nil
}

// SetAutoReady 设置自动准备
func (s *RoomService) SetAutoReady(ctx context.Context, userID int64, autoReady bool) error {
	currentRoom, err := s.roomRepo.GetPlayerRoom(ctx, userID)
	if err != nil {
		return err
	}
	if currentRoom == nil {
		return ErrNotInRoom
	}

	if err := s.roomRepo.UpdatePlayerAutoReady(ctx, currentRoom.RoomID, userID, autoReady); err != nil {
		return err
	}

	// 通知游戏引擎
	if processor := s.manager.GetRoom(currentRoom.RoomID); processor != nil {
		processor.SetAutoReady(userID, autoReady)
	}

	return nil
}

// GetRoomPlayers 获取房间玩家列表
func (s *RoomService) GetRoomPlayers(ctx context.Context, roomID int64) ([]*model.RoomPlayer, error) {
	return s.roomRepo.GetRoomPlayers(ctx, roomID)
}

// GetPlayerRoom 获取玩家当前房间
func (s *RoomService) GetPlayerRoom(ctx context.Context, userID int64) (*model.RoomPlayer, error) {
	return s.roomRepo.GetPlayerRoom(ctx, userID)
}

func isValidBetAmount(amount decimal.Decimal) bool {
	for _, valid := range ValidBetAmounts {
		if amount.Equal(valid) {
			return true
		}
	}
	return false
}

// GetUserByID 获取用户信息
func (s *RoomService) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

// JoinAsSpectator 以观战者身份加入房间
func (s *RoomService) JoinAsSpectator(ctx context.Context, userID, roomID int64, password string) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.Status != model.RoomStatusActive {
		return ErrRoomNotActive
	}

	// 验证密码
	if room.Password != nil && *room.Password != "" {
		if *room.Password != password {
			return ErrInvalidPassword
		}
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 通过游戏引擎添加观战者
	processor, err := s.manager.GetOrCreateRoom(ctx, roomID)
	if err != nil {
		return err
	}

	return processor.AddSpectator(user)
}

// SwitchToParticipant 观战者切换为参与者
func (s *RoomService) SwitchToParticipant(ctx context.Context, userID, roomID int64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	if room.Status != model.RoomStatusActive {
		return ErrRoomNotActive
	}

	// 获取用户信息
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// 通过游戏引擎切换
	processor := s.manager.GetRoom(roomID)
	if processor == nil {
		return errors.New("room processor not found")
	}

	return processor.SpectatorToParticipant(user)
}
