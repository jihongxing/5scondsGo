package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/fiveseconds/server/internal/model"
	"github.com/jackc/pgx/v5"
)

var (
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInviteLinkExpired  = errors.New("invite link expired")
	ErrInviteLinkInvalid  = errors.New("invalid invite link")
	ErrInviteLinkMaxUses  = errors.New("invite link max uses reached")
)

const InvitationExpireHours = 1 // 邀请1小时过期
const InviteLinkExpireHours = 24 // 邀请链接24小时过期

type InvitationRepo struct{}

func NewInvitationRepo() *InvitationRepo {
	return &InvitationRepo{}
}

// CreateInvitation 创建房间邀请
func (r *InvitationRepo) CreateInvitation(ctx context.Context, roomID, fromUserID, toUserID int64) (*model.RoomInvitation, error) {
	expiresAt := time.Now().Add(time.Hour * InvitationExpireHours)
	
	sql := `INSERT INTO room_invitations (room_id, from_user_id, to_user_id, status, expires_at)
		VALUES ($1, $2, $3, 'pending', $4)
		RETURNING id, room_id, from_user_id, to_user_id, status, created_at, expires_at`

	inv := &model.RoomInvitation{}
	err := DB.QueryRow(ctx, sql, roomID, fromUserID, toUserID, expiresAt).Scan(
		&inv.ID, &inv.RoomID, &inv.FromUserID, &inv.ToUserID, &inv.Status, &inv.CreatedAt, &inv.ExpiresAt,
	)
	return inv, err
}

// GetInvitationByID 根据ID获取邀请
func (r *InvitationRepo) GetInvitationByID(ctx context.Context, id int64) (*model.RoomInvitation, error) {
	sql := `SELECT ri.id, ri.room_id, ri.from_user_id, ri.to_user_id, ri.status, ri.created_at, ri.expires_at,
			COALESCE(rm.name, 'Room ' || rm.code) as room_name, rm.bet_amount, u.username as from_username
		FROM room_invitations ri
		JOIN rooms rm ON ri.room_id = rm.id
		JOIN users u ON ri.from_user_id = u.id
		WHERE ri.id = $1`

	inv := &model.RoomInvitation{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&inv.ID, &inv.RoomID, &inv.FromUserID, &inv.ToUserID, &inv.Status, &inv.CreatedAt, &inv.ExpiresAt,
		&inv.RoomName, &inv.BetAmount, &inv.FromUsername,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvitationNotFound
	}
	return inv, err
}

// GetPendingInvitationsForUser 获取用户收到的待处理邀请
func (r *InvitationRepo) GetPendingInvitationsForUser(ctx context.Context, userID int64) ([]*model.RoomInvitation, error) {
	sql := `SELECT ri.id, ri.room_id, ri.from_user_id, ri.to_user_id, ri.status, ri.created_at, ri.expires_at,
			COALESCE(rm.name, 'Room ' || rm.code) as room_name, rm.bet_amount, u.username as from_username
		FROM room_invitations ri
		JOIN rooms rm ON ri.room_id = rm.id
		JOIN users u ON ri.from_user_id = u.id
		WHERE ri.to_user_id = $1 AND ri.status = 'pending' AND ri.expires_at > NOW()
		ORDER BY ri.created_at DESC`

	rows, err := DB.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*model.RoomInvitation
	for rows.Next() {
		inv := &model.RoomInvitation{}
		if err := rows.Scan(
			&inv.ID, &inv.RoomID, &inv.FromUserID, &inv.ToUserID, &inv.Status, &inv.CreatedAt, &inv.ExpiresAt,
			&inv.RoomName, &inv.BetAmount, &inv.FromUsername,
		); err != nil {
			return nil, err
		}
		invitations = append(invitations, inv)
	}
	return invitations, nil
}

// UpdateInvitationStatus 更新邀请状态
func (r *InvitationRepo) UpdateInvitationStatus(ctx context.Context, id int64, status model.InvitationStatus) error {
	sql := `UPDATE room_invitations SET status = $1 WHERE id = $2`
	tag, err := DB.Exec(ctx, sql, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInvitationNotFound
	}
	return nil
}

// CreateInviteLink 创建邀请链接
func (r *InvitationRepo) CreateInviteLink(ctx context.Context, roomID, createdBy int64, maxUses *int) (*model.InviteLink, error) {
	code, err := generateInviteCode()
	if err != nil {
		return nil, err
	}
	
	expiresAt := time.Now().Add(time.Hour * InviteLinkExpireHours)
	
	sql := `INSERT INTO invite_links (room_id, code, created_by, expires_at, max_uses)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, room_id, code, created_by, created_at, expires_at, use_count, max_uses`

	link := &model.InviteLink{}
	err = DB.QueryRow(ctx, sql, roomID, code, createdBy, expiresAt, maxUses).Scan(
		&link.ID, &link.RoomID, &link.Code, &link.CreatedBy, &link.CreatedAt, &link.ExpiresAt, &link.UseCount, &link.MaxUses,
	)
	return link, err
}

// GetInviteLinkByCode 根据code获取邀请链接
func (r *InvitationRepo) GetInviteLinkByCode(ctx context.Context, code string) (*model.InviteLink, error) {
	sql := `SELECT id, room_id, code, created_by, created_at, expires_at, use_count, max_uses
		FROM invite_links WHERE code = $1`

	link := &model.InviteLink{}
	err := DB.QueryRow(ctx, sql, code).Scan(
		&link.ID, &link.RoomID, &link.Code, &link.CreatedBy, &link.CreatedAt, &link.ExpiresAt, &link.UseCount, &link.MaxUses,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInviteLinkInvalid
	}
	return link, err
}

// IncrementInviteLinkUseCount 增加邀请链接使用次数
func (r *InvitationRepo) IncrementInviteLinkUseCount(ctx context.Context, id int64) error {
	sql := `UPDATE invite_links SET use_count = use_count + 1 WHERE id = $1`
	_, err := DB.Exec(ctx, sql, id)
	return err
}

// generateInviteCode 生成邀请码
func generateInviteCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
