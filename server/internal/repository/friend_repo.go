package repository

import (
	"context"
	"errors"

	"github.com/fiveseconds/server/internal/model"
	"github.com/jackc/pgx/v5"
)

var (
	ErrFriendRequestExists = errors.New("friend request already exists")
	ErrAlreadyFriends      = errors.New("already friends")
	ErrFriendLimitReached  = errors.New("friend limit reached")
	ErrFriendRequestNotFound = errors.New("friend request not found")
	ErrNotFriends          = errors.New("not friends")
)

const MaxFriends = 200

type FriendRepo struct{}

func NewFriendRepo() *FriendRepo {
	return &FriendRepo{}
}

// CreateFriendRequest 创建好友请求
func (r *FriendRepo) CreateFriendRequest(ctx context.Context, fromUserID, toUserID int64) (*model.FriendRequest, error) {
	// 检查是否已经是好友
	isFriend, err := r.AreFriends(ctx, fromUserID, toUserID)
	if err != nil {
		return nil, err
	}
	if isFriend {
		return nil, ErrAlreadyFriends
	}

	// 检查是否已有待处理的请求
	existing, err := r.GetPendingRequest(ctx, fromUserID, toUserID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrFriendRequestExists
	}

	// 检查好友数量限制
	count, err := r.GetFriendCount(ctx, toUserID)
	if err != nil {
		return nil, err
	}
	if count >= MaxFriends {
		return nil, ErrFriendLimitReached
	}

	sql := `INSERT INTO friend_requests (from_user_id, to_user_id, status)
		VALUES ($1, $2, 'pending')
		RETURNING id, from_user_id, to_user_id, status, created_at, updated_at`

	req := &model.FriendRequest{}
	err = DB.QueryRow(ctx, sql, fromUserID, toUserID).Scan(
		&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	return req, err
}

// GetPendingRequest 获取待处理的好友请求
func (r *FriendRepo) GetPendingRequest(ctx context.Context, fromUserID, toUserID int64) (*model.FriendRequest, error) {
	sql := `SELECT id, from_user_id, to_user_id, status, created_at, updated_at
		FROM friend_requests
		WHERE from_user_id = $1 AND to_user_id = $2 AND status = 'pending'`

	req := &model.FriendRequest{}
	err := DB.QueryRow(ctx, sql, fromUserID, toUserID).Scan(
		&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return req, err
}


// GetFriendRequestByID 根据ID获取好友请求
func (r *FriendRepo) GetFriendRequestByID(ctx context.Context, id int64) (*model.FriendRequest, error) {
	sql := `SELECT id, from_user_id, to_user_id, status, created_at, updated_at
		FROM friend_requests WHERE id = $1`

	req := &model.FriendRequest{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrFriendRequestNotFound
	}
	return req, err
}

// UpdateFriendRequestStatus 更新好友请求状态
func (r *FriendRepo) UpdateFriendRequestStatus(ctx context.Context, id int64, status model.FriendRequestStatus) error {
	sql := `UPDATE friend_requests SET status = $1, updated_at = NOW() WHERE id = $2`
	tag, err := DB.Exec(ctx, sql, status, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrFriendRequestNotFound
	}
	return nil
}

// GetPendingRequestsForUser 获取用户收到的待处理好友请求
func (r *FriendRepo) GetPendingRequestsForUser(ctx context.Context, userID int64) ([]*model.FriendRequest, error) {
	sql := `SELECT fr.id, fr.from_user_id, fr.to_user_id, fr.status, fr.created_at, fr.updated_at,
			u.id, u.username, u.role
		FROM friend_requests fr
		JOIN users u ON fr.from_user_id = u.id
		WHERE fr.to_user_id = $1 AND fr.status = 'pending'
		ORDER BY fr.created_at DESC`

	rows, err := DB.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*model.FriendRequest
	for rows.Next() {
		req := &model.FriendRequest{}
		fromUser := &model.UserPublicInfo{}
		if err := rows.Scan(
			&req.ID, &req.FromUserID, &req.ToUserID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
			&fromUser.ID, &fromUser.Username, &fromUser.Role,
		); err != nil {
			return nil, err
		}
		req.FromUser = fromUser
		requests = append(requests, req)
	}
	return requests, nil
}

// CreateFriendship 创建好友关系（双向）
func (r *FriendRepo) CreateFriendship(ctx context.Context, userID1, userID2 int64) error {
	// 检查好友数量限制
	count1, err := r.GetFriendCount(ctx, userID1)
	if err != nil {
		return err
	}
	count2, err := r.GetFriendCount(ctx, userID2)
	if err != nil {
		return err
	}
	if count1 >= MaxFriends || count2 >= MaxFriends {
		return ErrFriendLimitReached
	}

	// 创建双向好友关系
	sql := `INSERT INTO friends (user_id, friend_id) VALUES ($1, $2), ($2, $1)
		ON CONFLICT (user_id, friend_id) DO NOTHING`
	_, err = DB.Exec(ctx, sql, userID1, userID2)
	return err
}

// RemoveFriendship 删除好友关系（双向）
func (r *FriendRepo) RemoveFriendship(ctx context.Context, userID1, userID2 int64) error {
	sql := `DELETE FROM friends WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`
	tag, err := DB.Exec(ctx, sql, userID1, userID2)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFriends
	}
	return nil
}

// AreFriends 检查两个用户是否是好友
func (r *FriendRepo) AreFriends(ctx context.Context, userID1, userID2 int64) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM friends WHERE user_id = $1 AND friend_id = $2)`
	var exists bool
	err := DB.QueryRow(ctx, sql, userID1, userID2).Scan(&exists)
	return exists, err
}

// GetFriendCount 获取好友数量
func (r *FriendRepo) GetFriendCount(ctx context.Context, userID int64) (int, error) {
	sql := `SELECT COUNT(*) FROM friends WHERE user_id = $1`
	var count int
	err := DB.QueryRow(ctx, sql, userID).Scan(&count)
	return count, err
}

// GetFriendIDs 获取好友ID列表
func (r *FriendRepo) GetFriendIDs(ctx context.Context, userID int64) ([]int64, error) {
	sql := `SELECT friend_id FROM friends WHERE user_id = $1`
	rows, err := DB.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// GetFriendList 获取好友列表（包含用户信息）
func (r *FriendRepo) GetFriendList(ctx context.Context, userID int64) ([]*model.FriendInfo, error) {
	sql := `SELECT u.id, u.username, u.role
		FROM friends f
		JOIN users u ON f.friend_id = u.id
		WHERE f.user_id = $1
		ORDER BY u.username`

	rows, err := DB.Query(ctx, sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var friends []*model.FriendInfo
	for rows.Next() {
		f := &model.FriendInfo{}
		if err := rows.Scan(&f.ID, &f.Username, &f.Role); err != nil {
			return nil, err
		}
		friends = append(friends, f)
	}
	return friends, nil
}
