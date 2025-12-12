package repository

import (
	"context"
	"fmt"

	"github.com/fiveseconds/server/internal/model"
)

// ConservationRepo 对账历史仓库
// 负责读写 fund_conservation_history 表

type ConservationRepo struct{}

func NewConservationRepo() *ConservationRepo {
	return &ConservationRepo{}
}

// Insert 插入一条对账历史记录
func (r *ConservationRepo) Insert(ctx context.Context, h *model.FundConservationHistory) error {
	sql := `INSERT INTO fund_conservation_history (
        scope, owner_id, period_type, period_start, period_end,
        total_player_balance, total_player_frozen, total_custody_quota, total_margin,
        owner_room_balance, owner_withdrawable_balance, owner_frozen_balance, platform_balance,
        difference, is_balanced
    ) VALUES (
        $1, $2, $3, $4, $5,
        $6, $7, $8, $9,
        $10, $11, $12, $13,
        $14, $15
    ) RETURNING id, created_at`

	return DB.QueryRow(ctx, sql,
		h.Scope, h.OwnerID, h.PeriodType, h.PeriodStart, h.PeriodEnd,
		h.TotalPlayerBalance, h.TotalPlayerFrozen, h.TotalCustodyQuota, h.TotalMargin,
		h.OwnerRoomBalance, h.OwnerWithdrawableBalance, h.OwnerFrozenBalance, h.PlatformBalance,
		h.Difference, h.IsBalanced,
	).Scan(&h.ID, &h.CreatedAt)
}

// FundConservationHistoryQuery 查询条件
// 注意：时间范围按 created_at 过滤，便于简单实现

func (r *ConservationRepo) List(ctx context.Context, q *model.FundConservationHistoryQuery) ([]*model.FundConservationHistory, int64, error) {
	countSQL := `SELECT COUNT(*) FROM fund_conservation_history WHERE 1=1`
	listSQL := `SELECT id, scope, owner_id, period_type, period_start, period_end,
        total_player_balance, total_player_frozen, total_custody_quota, total_margin,
        owner_room_balance, owner_withdrawable_balance, owner_frozen_balance, platform_balance,
        difference, is_balanced, created_at
        FROM fund_conservation_history WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if q.Scope != nil {
		countSQL += fmt.Sprintf(" AND scope = $%d", argIdx)
		listSQL += fmt.Sprintf(" AND scope = $%d", argIdx)
		args = append(args, *q.Scope)
		argIdx++
	}

	if q.OwnerID != nil {
		countSQL += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		listSQL += fmt.Sprintf(" AND owner_id = $%d", argIdx)
		args = append(args, *q.OwnerID)
		argIdx++
	}

	if q.PeriodType != nil {
		countSQL += fmt.Sprintf(" AND period_type = $%d", argIdx)
		listSQL += fmt.Sprintf(" AND period_type = $%d", argIdx)
		args = append(args, *q.PeriodType)
		argIdx++
	}

	if q.FromCreatedAt != nil {
		countSQL += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		listSQL += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *q.FromCreatedAt)
		argIdx++
	}

	if q.ToCreatedAt != nil {
		countSQL += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		listSQL += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *q.ToCreatedAt)
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, q.PageSize, (q.Page-1)*q.PageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.FundConservationHistory
	for rows.Next() {
		h := &model.FundConservationHistory{}
		if err := rows.Scan(
			&h.ID, &h.Scope, &h.OwnerID, &h.PeriodType, &h.PeriodStart, &h.PeriodEnd,
			&h.TotalPlayerBalance, &h.TotalPlayerFrozen, &h.TotalCustodyQuota, &h.TotalMargin,
			&h.OwnerRoomBalance, &h.OwnerWithdrawableBalance, &h.OwnerFrozenBalance, &h.PlatformBalance,
			&h.Difference, &h.IsBalanced, &h.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, h)
	}

	return items, total, nil
}
