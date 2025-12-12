package service

import (
	"context"
	"errors"
	"sync"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
)

var (
	ErrCannotAddSelf = errors.New("cannot add yourself as friend")
)

type FriendService struct {
	repo *repository.FriendRepo
	mu   sync.RWMutex

	// 在线状态缓存 userID -> roomID (0表示在线但不在房间)
	onlineUsers map[int64]int64
}

func NewFriendService(repo *repository.FriendRepo) *FriendService {
	return &FriendService{
		repo:        repo,
		onlineUsers: make(map[int64]int64),
	}
}

// SendFriendRequest 发送好友请求
func (s *FriendService) SendFriendRequest(ctx context.Context, fromUserID, toUserID int64) (*model.FriendRequest, error) {
	if fromUserID == toUserID {
		return nil, ErrCannotAddSelf
	}
	return s.repo.CreateFriendRequest(ctx, fromUserID, toUserID)
}

// AcceptFriendRequest 接受好友请求
func (s *FriendService) AcceptFriendRequest(ctx context.Context, requestID int64, userID int64) error {
	// 获取请求
	req, err := s.repo.GetFriendRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	// 验证是请求的接收者
	if req.ToUserID != userID {
		return errors.New("not authorized to accept this request")
	}

	// 验证状态
	if req.Status != model.FriendRequestPending {
		return errors.New("request is not pending")
	}

	// 创建好友关系
	if err := s.repo.CreateFriendship(ctx, req.FromUserID, req.ToUserID); err != nil {
		return err
	}

	// 更新请求状态
	return s.repo.UpdateFriendRequestStatus(ctx, requestID, model.FriendRequestAccepted)
}

// RejectFriendRequest 拒绝好友请求
func (s *FriendService) RejectFriendRequest(ctx context.Context, requestID int64, userID int64) error {
	// 获取请求
	req, err := s.repo.GetFriendRequestByID(ctx, requestID)
	if err != nil {
		return err
	}

	// 验证是请求的接收者
	if req.ToUserID != userID {
		return errors.New("not authorized to reject this request")
	}

	// 验证状态
	if req.Status != model.FriendRequestPending {
		return errors.New("request is not pending")
	}

	return s.repo.UpdateFriendRequestStatus(ctx, requestID, model.FriendRequestRejected)
}

// RemoveFriend 删除好友
func (s *FriendService) RemoveFriend(ctx context.Context, userID, friendID int64) error {
	return s.repo.RemoveFriendship(ctx, userID, friendID)
}

// GetFriendList 获取好友列表（包含在线状态）
func (s *FriendService) GetFriendList(ctx context.Context, userID int64) ([]*model.FriendInfo, error) {
	friends, err := s.repo.GetFriendList(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 填充在线状态
	s.mu.RLock()
	for _, f := range friends {
		if roomID, ok := s.onlineUsers[f.ID]; ok {
			f.IsOnline = true
			if roomID > 0 {
				f.CurrentRoom = &roomID
			}
		}
	}
	s.mu.RUnlock()

	return friends, nil
}

// GetPendingRequests 获取待处理的好友请求
func (s *FriendService) GetPendingRequests(ctx context.Context, userID int64) ([]*model.FriendRequest, error) {
	return s.repo.GetPendingRequestsForUser(ctx, userID)
}

// AreFriends 检查是否是好友
func (s *FriendService) AreFriends(ctx context.Context, userID1, userID2 int64) (bool, error) {
	return s.repo.AreFriends(ctx, userID1, userID2)
}

// SetUserOnline 设置用户在线状态
func (s *FriendService) SetUserOnline(userID int64, roomID int64) {
	s.mu.Lock()
	s.onlineUsers[userID] = roomID
	s.mu.Unlock()
}

// SetUserOffline 设置用户离线
func (s *FriendService) SetUserOffline(userID int64) {
	s.mu.Lock()
	delete(s.onlineUsers, userID)
	s.mu.Unlock()
}

// IsUserOnline 检查用户是否在线
func (s *FriendService) IsUserOnline(userID int64) bool {
	s.mu.RLock()
	_, ok := s.onlineUsers[userID]
	s.mu.RUnlock()
	return ok
}

// GetUserRoom 获取用户当前所在房间
func (s *FriendService) GetUserRoom(userID int64) (int64, bool) {
	s.mu.RLock()
	roomID, ok := s.onlineUsers[userID]
	s.mu.RUnlock()
	if !ok || roomID == 0 {
		return 0, false
	}
	return roomID, true
}

// GetOnlineFriendIDs 获取在线好友ID列表
func (s *FriendService) GetOnlineFriendIDs(ctx context.Context, userID int64) ([]int64, error) {
	friendIDs, err := s.repo.GetFriendIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	var onlineIDs []int64
	s.mu.RLock()
	for _, id := range friendIDs {
		if _, ok := s.onlineUsers[id]; ok {
			onlineIDs = append(onlineIDs, id)
		}
	}
	s.mu.RUnlock()

	return onlineIDs, nil
}

// GetFriendIDs 获取好友ID列表
func (s *FriendService) GetFriendIDs(ctx context.Context, userID int64) ([]int64, error) {
	return s.repo.GetFriendIDs(ctx, userID)
}
