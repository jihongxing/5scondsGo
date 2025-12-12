package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fiveseconds/server/internal/model"
	"github.com/jackc/pgx/v5"
)

// RiskRepo 风控仓库
type RiskRepo struct{}

// NewRiskRepo 创建风控仓库
func NewRiskRepo() *RiskRepo {
	return &RiskRepo{}
}

// CreateFlag 创建风控标记
func (r *RiskRepo) CreateFlag(ctx context.Context, flag *model.RiskFlag) error {
	sql := `INSERT INTO risk_flags (user_id, flag_type, details, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`
	return DB.QueryRow(ctx, sql,
		flag.UserID, flag.FlagType, flag.Details, flag.Status,
	).Scan(&flag.ID, &flag.CreatedAt)
}

// GetFlagByID 根据ID获取风控标记
func (r *RiskRepo) GetFlagByID(ctx context.Context, id int64) (*model.RiskFlag, error) {
	sql := `SELECT id, user_id, flag_type, details, status, reviewed_by, reviewed_at, created_at
		FROM risk_flags WHERE id = $1`
	flag := &model.RiskFlag{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&flag.ID, &flag.UserID, &flag.FlagType, &flag.Details,
		&flag.Status, &flag.ReviewedBy, &flag.ReviewedAt, &flag.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return flag, err
}

// ListFlags 列表风控标记
func (r *RiskRepo) ListFlags(ctx context.Context, query *model.RiskFlagListQuery) ([]*model.RiskFlag, int64, error) {
	countSQL := `SELECT COUNT(*) FROM risk_flags WHERE 1=1`
	listSQL := `SELECT id, user_id, flag_type, details, status, reviewed_by, reviewed_at, created_at
		FROM risk_flags WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if query.UserID != nil {
		countSQL += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		args = append(args, *query.UserID)
		argIdx++
	}
	if query.FlagType != nil {
		countSQL += fmt.Sprintf(` AND flag_type = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND flag_type = $%d`, argIdx)
		args = append(args, *query.FlagType)
		argIdx++
	}
	if query.Status != nil {
		countSQL += fmt.Sprintf(` AND status = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, *query.Status)
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var flags []*model.RiskFlag
	for rows.Next() {
		flag := &model.RiskFlag{}
		if err := rows.Scan(
			&flag.ID, &flag.UserID, &flag.FlagType, &flag.Details,
			&flag.Status, &flag.ReviewedBy, &flag.ReviewedAt, &flag.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		flags = append(flags, flag)
	}
	return flags, total, nil
}

// ReviewFlag 审核风控标记
func (r *RiskRepo) ReviewFlag(ctx context.Context, id int64, status model.RiskFlagStatus, reviewedBy int64) error {
	sql := `UPDATE risk_flags SET status = $1, reviewed_by = $2, reviewed_at = NOW()
		WHERE id = $3 AND status = 'pending'`
	tag, err := DB.Exec(ctx, sql, status, reviewedBy, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("flag not found or already reviewed")
	}
	return nil
}


// HasPendingFlag 检查用户是否有待处理的同类型风控标记
func (r *RiskRepo) HasPendingFlag(ctx context.Context, userID int64, flagType model.RiskFlagType) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM risk_flags WHERE user_id = $1 AND flag_type = $2 AND status = 'pending')`
	var exists bool
	err := DB.QueryRow(ctx, sql, userID, flagType).Scan(&exists)
	return exists, err
}

// GetUserConsecutiveWins 获取用户连续获胜次数
func (r *RiskRepo) GetUserConsecutiveWins(ctx context.Context, userID int64) (int, error) {
	sql := `SELECT COALESCE(consecutive_wins, 0) FROM users WHERE id = $1`
	var wins int
	err := DB.QueryRow(ctx, sql, userID).Scan(&wins)
	return wins, err
}

// UpdateUserConsecutiveWins 更新用户连续获胜次数
func (r *RiskRepo) UpdateUserConsecutiveWins(ctx context.Context, userID int64, wins int) error {
	sql := `UPDATE users SET consecutive_wins = $1, last_win_at = NOW() WHERE id = $2`
	_, err := DB.Exec(ctx, sql, wins, userID)
	return err
}

// ResetUserConsecutiveWins 重置用户连续获胜次数
func (r *RiskRepo) ResetUserConsecutiveWins(ctx context.Context, userID int64) error {
	sql := `UPDATE users SET consecutive_wins = 0 WHERE id = $1`
	_, err := DB.Exec(ctx, sql, userID)
	return err
}

// GetUserWinRate 获取用户胜率（最近N回合）
func (r *RiskRepo) GetUserWinRate(ctx context.Context, userID int64, rounds int) (float64, int, error) {
	sql := `
		WITH recent_rounds AS (
			SELECT gr.id, gr.winner_ids
			FROM game_rounds gr
			JOIN rooms r ON gr.room_id = r.id
			WHERE $1 = ANY(gr.participant_ids)
			  AND gr.status = 'settled'
			ORDER BY gr.created_at DESC
			LIMIT $2
		)
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE $1 = ANY(winner_ids)) as wins
		FROM recent_rounds
	`
	var total, wins int
	err := DB.QueryRow(ctx, sql, userID, rounds).Scan(&total, &wins)
	if err != nil {
		return 0, 0, err
	}
	if total == 0 {
		return 0, 0, nil
	}
	return float64(wins) / float64(total), total, nil
}

// GetUsersByDeviceFingerprint 获取使用相同设备指纹的用户
func (r *RiskRepo) GetUsersByDeviceFingerprint(ctx context.Context, fingerprint string) ([]int64, error) {
	sql := `SELECT id FROM users WHERE device_fingerprint = $1`
	rows, err := DB.Query(ctx, sql, fingerprint)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, id)
	}
	return userIDs, nil
}

// GetUserDailyVolume 获取用户当日交易量
func (r *RiskRepo) GetUserDailyVolume(ctx context.Context, userID int64) (string, error) {
	sql := `
		SELECT COALESCE(SUM(ABS(amount)), 0)
		FROM balance_transactions
		WHERE user_id = $1
		  AND created_at >= CURRENT_DATE
	`
	var volume string
	err := DB.QueryRow(ctx, sql, userID).Scan(&volume)
	return volume, err
}

// CreateFlagWithDetails 创建带详情的风控标记
func (r *RiskRepo) CreateFlagWithDetails(ctx context.Context, flag *model.RiskFlag, details *model.RiskFlagDetails) error {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return err
	}
	flag.Details = string(detailsJSON)
	return r.CreateFlag(ctx, flag)
}
