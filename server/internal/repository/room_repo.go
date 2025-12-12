package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/fiveseconds/server/internal/model"

	"github.com/jackc/pgx/v5"
)

type RoomRepo struct{}

func NewRoomRepo() *RoomRepo {
	return &RoomRepo{}
}

// Create 创建房间
func (r *RoomRepo) Create(ctx context.Context, room *model.Room) error {
	sql := `INSERT INTO rooms (owner_id, name, code, bet_amount, winner_count, max_players,
		owner_commission, platform_commission, status, password)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`
	return DB.QueryRow(ctx, sql,
		room.OwnerID, room.Name, room.InviteCode, room.BetAmount, room.WinnerCount, room.MaxPlayers,
		room.OwnerCommissionRate, room.PlatformCommissionRate, room.Status, room.Password,
	).Scan(&room.ID, &room.CreatedAt, &room.UpdatedAt)
}

// GetByID 根据ID获取房间
func (r *RoomRepo) GetByID(ctx context.Context, id int64) (*model.Room, error) {
	sql := `SELECT id, owner_id, name, code, bet_amount, winner_count, max_players,
		owner_commission, platform_commission, status, password, created_at, updated_at
		FROM rooms WHERE id = $1`
	room := &model.Room{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&room.ID, &room.OwnerID, &room.Name, &room.InviteCode, &room.BetAmount, &room.WinnerCount, &room.MaxPlayers,
		&room.OwnerCommissionRate, &room.PlatformCommissionRate, &room.Status, &room.Password, &room.CreatedAt, &room.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return room, err
}

// GetByInviteCode 根据邀请码获取房间
func (r *RoomRepo) GetByInviteCode(ctx context.Context, code string) (*model.Room, error) {
	sql := `SELECT id, owner_id, name, code, bet_amount, winner_count, max_players,
		owner_commission, platform_commission, status, created_at, updated_at
		FROM rooms WHERE code = $1`
	room := &model.Room{}
	err := DB.QueryRow(ctx, sql, code).Scan(
		&room.ID, &room.OwnerID, &room.Name, &room.InviteCode, &room.BetAmount, &room.WinnerCount, &room.MaxPlayers,
		&room.OwnerCommissionRate, &room.PlatformCommissionRate, &room.Status, &room.CreatedAt, &room.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return room, err
}

// ListByOwner 获取房主的房间列表
func (r *RoomRepo) ListByOwner(ctx context.Context, ownerID int64) ([]*model.Room, error) {
	sql := `SELECT id, owner_id, name, code, bet_amount, winner_count, max_players,
		owner_commission, platform_commission, status, created_at, updated_at
		FROM rooms WHERE owner_id = $1 ORDER BY created_at DESC`
	rows, err := DB.Query(ctx, sql, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		room := &model.Room{}
		if err := rows.Scan(
			&room.ID, &room.OwnerID, &room.Name, &room.InviteCode, &room.BetAmount, &room.WinnerCount, &room.MaxPlayers,
			&room.OwnerCommissionRate, &room.PlatformCommissionRate, &room.Status, &room.CreatedAt, &room.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// List 分页列表
func (r *RoomRepo) List(ctx context.Context, query *model.RoomListQuery) ([]*model.Room, int64, error) {
	countSQL := `SELECT COUNT(*) FROM rooms WHERE 1=1`
	listSQL := `SELECT r.id, r.owner_id, r.name, r.code, r.bet_amount, r.winner_count, r.max_players,
		r.owner_commission, r.platform_commission, r.status, r.password, r.created_at, r.updated_at,
		COALESCE(u.username, '') as owner_name
		FROM rooms r LEFT JOIN users u ON r.owner_id = u.id WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if query.OwnerID != nil {
		countSQL += fmt.Sprintf(` AND owner_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND r.owner_id = $%d`, argIdx)
		args = append(args, *query.OwnerID)
		argIdx++
	}

	if query.InvitedBy != nil {
		countSQL += fmt.Sprintf(` AND owner_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND r.owner_id = $%d`, argIdx)
		args = append(args, *query.InvitedBy)
		argIdx++
	}

	if query.Status != nil {
		countSQL += fmt.Sprintf(` AND status = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND r.status = $%d`, argIdx)
		args = append(args, *query.Status)
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL += fmt.Sprintf(` ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		room := &model.Room{}
		var ownerName string
		if err := rows.Scan(
			&room.ID, &room.OwnerID, &room.Name, &room.InviteCode, &room.BetAmount, &room.WinnerCount, &room.MaxPlayers,
			&room.OwnerCommissionRate, &room.PlatformCommissionRate, &room.Status, &room.Password, &room.CreatedAt, &room.UpdatedAt,
			&ownerName,
		); err != nil {
			return nil, 0, err
		}
		room.OwnerName = ownerName
		rooms = append(rooms, room)
	}
	return rooms, total, nil
}

// UpdateStatus 更新房间状态
func (r *RoomRepo) UpdateStatus(ctx context.Context, roomID int64, status model.RoomStatus) error {
	sql := `UPDATE rooms SET status = $1, updated_at = NOW() WHERE id = $2`
	tag, err := DB.Exec(ctx, sql, status, roomID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Update 更新房间配置
func (r *RoomRepo) Update(ctx context.Context, room *model.Room) error {
	sql := `UPDATE rooms SET name = $1, bet_amount = $2, winner_count = $3, max_players = $4,
		owner_commission = $5, platform_commission = $6, updated_at = NOW()
		WHERE id = $7`
	tag, err := DB.Exec(ctx, sql,
		room.Name, room.BetAmount, room.WinnerCount, room.MaxPlayers,
		room.OwnerCommissionRate, room.PlatformCommissionRate, room.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// InviteCodeExists 检查房间邀请码是否存在
func (r *RoomRepo) InviteCodeExists(ctx context.Context, code string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM rooms WHERE code = $1)`
	var exists bool
	err := DB.QueryRow(ctx, sql, code).Scan(&exists)
	return exists, err
}

// ===== RoomPlayer 相关 =====

// AddPlayer 添加玩家到房间
func (r *RoomRepo) AddPlayer(ctx context.Context, rp *model.RoomPlayer) error {
	sql := `INSERT INTO room_players (room_id, user_id, auto_ready)
		VALUES ($1, $2, $3)
		ON CONFLICT (room_id, user_id) DO UPDATE SET left_at = NULL, auto_ready = $3
		RETURNING joined_at`
	return DB.QueryRow(ctx, sql, rp.RoomID, rp.UserID, rp.AutoReady).Scan(&rp.JoinedAt)
}

// RemovePlayer 从房间移除玩家
func (r *RoomRepo) RemovePlayer(ctx context.Context, roomID, userID int64) error {
	sql := `UPDATE room_players SET left_at = NOW() WHERE room_id = $1 AND user_id = $2 AND left_at IS NULL`
	_, err := DB.Exec(ctx, sql, roomID, userID)
	return err
}

// GetRoomPlayers 获取房间内的玩家
func (r *RoomRepo) GetRoomPlayers(ctx context.Context, roomID int64) ([]*model.RoomPlayer, error) {
	sql := `SELECT rp.room_id, rp.user_id, rp.auto_ready, rp.joined_at, rp.left_at, u.username
		FROM room_players rp
		JOIN users u ON u.id = rp.user_id
		WHERE rp.room_id = $1 AND rp.left_at IS NULL`
	rows, err := DB.Query(ctx, sql, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []*model.RoomPlayer
	for rows.Next() {
		p := &model.RoomPlayer{}
		var username string
		if err := rows.Scan(&p.RoomID, &p.UserID, &p.AutoReady, &p.JoinedAt, &p.LeftAt, &username); err != nil {
			return nil, err
		}
		players = append(players, p)
	}
	return players, nil
}

// GetPlayerRoom 获取玩家当前所在房间
func (r *RoomRepo) GetPlayerRoom(ctx context.Context, userID int64) (*model.RoomPlayer, error) {
	sql := `SELECT room_id, user_id, auto_ready, joined_at, left_at
		FROM room_players WHERE user_id = $1 AND left_at IS NULL`
	rp := &model.RoomPlayer{}
	err := DB.QueryRow(ctx, sql, userID).Scan(&rp.RoomID, &rp.UserID, &rp.AutoReady, &rp.JoinedAt, &rp.LeftAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rp, err
}

// GetRoomPlayer 获取指定房间中的指定玩家
func (r *RoomRepo) GetRoomPlayer(ctx context.Context, roomID, userID int64) (*model.RoomPlayer, error) {
	sql := `SELECT room_id, user_id, auto_ready, joined_at, left_at
		FROM room_players WHERE room_id = $1 AND user_id = $2 AND left_at IS NULL`
	rp := &model.RoomPlayer{}
	err := DB.QueryRow(ctx, sql, roomID, userID).Scan(&rp.RoomID, &rp.UserID, &rp.AutoReady, &rp.JoinedAt, &rp.LeftAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return rp, err
}

// UpdatePlayerAutoReady 更新玩家自动准备状态
func (r *RoomRepo) UpdatePlayerAutoReady(ctx context.Context, roomID, userID int64, autoReady bool) error {
	sql := `UPDATE room_players SET auto_ready = $1 WHERE room_id = $2 AND user_id = $3 AND left_at IS NULL`
	_, err := DB.Exec(ctx, sql, autoReady, roomID, userID)
	return err
}

// CountRoomPlayers 统计房间玩家数
func (r *RoomRepo) CountRoomPlayers(ctx context.Context, roomID int64) (int, error) {
	sql := `SELECT COUNT(*) FROM room_players WHERE room_id = $1 AND left_at IS NULL`
	var count int
	err := DB.QueryRow(ctx, sql, roomID).Scan(&count)
	return count, err
}
