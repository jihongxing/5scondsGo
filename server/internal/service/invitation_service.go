package service

import (
	"context"
	"errors"
	"time"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
)

var (
	ErrCannotInviteSelf    = errors.New("cannot invite yourself")
	ErrInvitationExpired   = errors.New("invitation expired")
	ErrNotInvitationTarget = errors.New("not the invitation target")
)

// InvitationBroadcaster 邀请广播接口
type InvitationBroadcaster interface {
	SendToUser(userID int64, msg *model.WSMessage)
}

type InvitationService struct {
	repo         *repository.InvitationRepo
	friendRepo   *repository.FriendRepo
	roomRepo     *repository.RoomRepo
	userRepo     *repository.UserRepo
	broadcaster  InvitationBroadcaster
}

func NewInvitationService(repo *repository.InvitationRepo, friendRepo *repository.FriendRepo, roomRepo *repository.RoomRepo, userRepo *repository.UserRepo, broadcaster InvitationBroadcaster) *InvitationService {
	return &InvitationService{
		repo:        repo,
		friendRepo:  friendRepo,
		roomRepo:    roomRepo,
		userRepo:    userRepo,
		broadcaster: broadcaster,
	}
}

// SendInvitation 发送房间邀请
func (s *InvitationService) SendInvitation(ctx context.Context, roomID, fromUserID, toUserID int64) (*model.RoomInvitation, error) {
	if fromUserID == toUserID {
		return nil, ErrCannotInviteSelf
	}

	// 验证是好友关系（可选，根据需求）
	// isFriend, err := s.friendRepo.AreFriends(ctx, fromUserID, toUserID)
	// if err != nil {
	// 	return nil, err
	// }
	// if !isFriend {
	// 	return nil, errors.New("can only invite friends")
	// }

	inv, err := s.repo.CreateInvitation(ctx, roomID, fromUserID, toUserID)
	if err != nil {
		return nil, err
	}

	// 广播邀请通知给目标用户
	s.broadcastInvitation(ctx, inv, fromUserID)

	return inv, nil
}

// broadcastInvitation 广播邀请通知
func (s *InvitationService) broadcastInvitation(ctx context.Context, inv *model.RoomInvitation, fromUserID int64) {
	if s.broadcaster == nil {
		return
	}

	// 获取房间信息
	room, err := s.roomRepo.GetByID(ctx, inv.RoomID)
	if err != nil {
		return
	}

	// 获取发送者信息
	fromUser, err := s.userRepo.GetByID(ctx, fromUserID)
	if err != nil {
		return
	}

	// 获取房间当前玩家数
	playerCount, _ := s.roomRepo.CountRoomPlayers(ctx, inv.RoomID)

	s.broadcaster.SendToUser(inv.ToUserID, &model.WSMessage{
		Type: model.WSTypeRoomInvitation,
		Payload: &model.WSRoomInvitation{
			InvitationID: inv.ID,
			RoomID:       inv.RoomID,
			RoomName:     room.Name,
			BetAmount:    room.BetAmount.String(),
			PlayerCount:  playerCount,
			FromUserID:   fromUserID,
			FromUsername: fromUser.Username,
		},
	})
}

// AcceptInvitation 接受邀请
func (s *InvitationService) AcceptInvitation(ctx context.Context, invitationID, userID int64) (*model.RoomInvitation, error) {
	inv, err := s.repo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}

	// 验证是邀请的接收者
	if inv.ToUserID != userID {
		return nil, ErrNotInvitationTarget
	}

	// 验证状态
	if inv.Status != model.InvitationPending {
		return nil, errors.New("invitation is not pending")
	}

	// 验证是否过期
	if time.Now().After(inv.ExpiresAt) {
		_ = s.repo.UpdateInvitationStatus(ctx, invitationID, model.InvitationExpired)
		return nil, ErrInvitationExpired
	}

	// 更新状态
	if err := s.repo.UpdateInvitationStatus(ctx, invitationID, model.InvitationAccepted); err != nil {
		return nil, err
	}

	inv.Status = model.InvitationAccepted

	// 自动添加好友关系（邀请人和被邀请人互为好友）
	if s.friendRepo != nil {
		// 忽略错误（可能已经是好友或达到好友上限）
		_ = s.friendRepo.CreateFriendship(ctx, inv.FromUserID, inv.ToUserID)
	}

	// 广播邀请响应给发送者
	s.broadcastInviteResponse(ctx, inv, userID, true)

	return inv, nil
}

// DeclineInvitation 拒绝邀请
func (s *InvitationService) DeclineInvitation(ctx context.Context, invitationID, userID int64) error {
	inv, err := s.repo.GetInvitationByID(ctx, invitationID)
	if err != nil {
		return err
	}

	// 验证是邀请的接收者
	if inv.ToUserID != userID {
		return ErrNotInvitationTarget
	}

	// 验证状态
	if inv.Status != model.InvitationPending {
		return errors.New("invitation is not pending")
	}

	if err := s.repo.UpdateInvitationStatus(ctx, invitationID, model.InvitationDeclined); err != nil {
		return err
	}

	// 广播邀请响应给发送者
	s.broadcastInviteResponse(ctx, inv, userID, false)

	return nil
}

// broadcastInviteResponse 广播邀请响应
func (s *InvitationService) broadcastInviteResponse(ctx context.Context, inv *model.RoomInvitation, fromUserID int64, accepted bool) {
	if s.broadcaster == nil {
		return
	}

	// 获取响应者信息
	fromUser, err := s.userRepo.GetByID(ctx, fromUserID)
	if err != nil {
		return
	}

	s.broadcaster.SendToUser(inv.FromUserID, &model.WSMessage{
		Type: model.WSTypeInviteResponse,
		Payload: &model.WSInviteResponse{
			InvitationID: inv.ID,
			Accepted:     accepted,
			FromUserID:   fromUserID,
			FromUsername: fromUser.Username,
		},
	})
}

// GetPendingInvitations 获取待处理的邀请
func (s *InvitationService) GetPendingInvitations(ctx context.Context, userID int64) ([]*model.RoomInvitation, error) {
	return s.repo.GetPendingInvitationsForUser(ctx, userID)
}

// CreateInviteLink 创建邀请链接
func (s *InvitationService) CreateInviteLink(ctx context.Context, roomID, userID int64, maxUses *int) (*model.InviteLink, error) {
	return s.repo.CreateInviteLink(ctx, roomID, userID, maxUses)
}

// UseInviteLink 使用邀请链接
func (s *InvitationService) UseInviteLink(ctx context.Context, code string, userID int64) (*model.InviteLink, error) {
	link, err := s.repo.GetInviteLinkByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	// 验证是否过期
	if time.Now().After(link.ExpiresAt) {
		return nil, repository.ErrInviteLinkExpired
	}

	// 验证使用次数
	if link.MaxUses != nil && link.UseCount >= *link.MaxUses {
		return nil, repository.ErrInviteLinkMaxUses
	}

	// 增加使用次数
	if err := s.repo.IncrementInviteLinkUseCount(ctx, link.ID); err != nil {
		return nil, err
	}

	// 自动添加好友关系（链接创建者和使用者互为好友）
	if s.friendRepo != nil && link.CreatedBy != userID {
		// 忽略错误（可能已经是好友或达到好友上限）
		_ = s.friendRepo.CreateFriendship(ctx, link.CreatedBy, userID)
	}

	link.UseCount++
	return link, nil
}

// GetInvitationByID 根据ID获取邀请
func (s *InvitationService) GetInvitationByID(ctx context.Context, id int64) (*model.RoomInvitation, error) {
	return s.repo.GetInvitationByID(ctx, id)
}
