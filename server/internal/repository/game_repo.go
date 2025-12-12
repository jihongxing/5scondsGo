package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/fiveseconds/server/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

type GameRepo struct{}

func NewGameRepo() *GameRepo {
	return &GameRepo{}
}

// CreateRound 创建游戏回合
func (r *GameRepo) CreateRound(ctx context.Context, round *model.GameRound) error {
	return r.CreateRoundTx(ctx, nil, round)
}

// CreateRoundTx 创建游戏回合(支持事务)
func (r *GameRepo) CreateRoundTx(ctx context.Context, tx pgx.Tx, round *model.GameRound) error {
	sql := `INSERT INTO game_rounds (room_id, round_number, participant_ids, skipped_ids, bet_amount, pool_amount, commit_hash, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`
	exec := GetExecutor(tx)
	return exec.QueryRow(ctx, sql,
		round.RoomID, round.RoundNumber, round.ParticipantIDs, round.SkippedIDs,
		round.BetAmount, round.PoolAmount, round.CommitHash, round.Status,
	).Scan(&round.ID, &round.CreatedAt)
}

// GetRoundByID 根据ID获取回合
func (r *GameRepo) GetRoundByID(ctx context.Context, id int64) (*model.GameRound, error) {
	sql := `SELECT id, room_id, round_number, participant_ids, skipped_ids, winner_ids,
		bet_amount, pool_amount, prize_per_winner, owner_earning, platform_earning, residual_amount,
		commit_hash, reveal_seed, status, failure_reason, created_at, settled_at
		FROM game_rounds WHERE id = $1`
	round := &model.GameRound{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&round.ID, &round.RoomID, &round.RoundNumber, &round.ParticipantIDs, &round.SkippedIDs, &round.WinnerIDs,
		&round.BetAmount, &round.PoolAmount, &round.PrizePerWinner, &round.OwnerEarning, &round.PlatformEarning, &round.ResidualAmount,
		&round.CommitHash, &round.RevealSeed, &round.Status, &round.FailureReason, &round.CreatedAt, &round.SettledAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return round, err
}

// GetLastRoundNumber 获取房间最后一个回合号
func (r *GameRepo) GetLastRoundNumber(ctx context.Context, roomID int64) (int, error) {
	sql := `SELECT COALESCE(MAX(round_number), 0) FROM game_rounds WHERE room_id = $1`
	var num int
	err := DB.QueryRow(ctx, sql, roomID).Scan(&num)
	return num, err
}

// SettleRound 结算回合
func (r *GameRepo) SettleRound(ctx context.Context, round *model.GameRound) error {
	return r.SettleRoundTx(ctx, nil, round)
}

// SettleRoundTx 结算回合(支持事务)
func (r *GameRepo) SettleRoundTx(ctx context.Context, tx pgx.Tx, round *model.GameRound) error {
	sql := `UPDATE game_rounds SET
		winner_ids = $1, prize_per_winner = $2, owner_earning = $3, platform_earning = $4, residual_amount = $5,
		reveal_seed = $6, status = $7, settled_at = NOW()
		WHERE id = $8`
	exec := GetExecutor(tx)
	tag, err := exec.Exec(ctx, sql,
		round.WinnerIDs, round.PrizePerWinner, round.OwnerEarning, round.PlatformEarning, round.ResidualAmount,
		round.RevealSeed, round.Status, round.ID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// FailRound 标记回合失败
func (r *GameRepo) FailRound(ctx context.Context, roundID int64, reason string) error {
	return r.FailRoundTx(ctx, nil, roundID, reason)
}

// FailRoundTx 标记回合失败(支持事务)
func (r *GameRepo) FailRoundTx(ctx context.Context, tx pgx.Tx, roundID int64, reason string) error {
	sql := `UPDATE game_rounds SET status = $1, failure_reason = $2, settled_at = NOW() WHERE id = $3`
	exec := GetExecutor(tx)
	_, err := exec.Exec(ctx, sql, model.RoundStatusFailed, reason, roundID)
	return err
}

// UpdateRoundStatus 更新回合状态
func (r *GameRepo) UpdateRoundStatus(ctx context.Context, roundID int64, status model.RoundStatus) error {
	return r.UpdateRoundStatusTx(ctx, nil, roundID, status)
}

// UpdateRoundStatusTx 更新回合状态(支持事务)
func (r *GameRepo) UpdateRoundStatusTx(ctx context.Context, tx pgx.Tx, roundID int64, status model.RoundStatus) error {
	sql := `UPDATE game_rounds SET status = $1 WHERE id = $2`
	exec := GetExecutor(tx)
	_, err := exec.Exec(ctx, sql, status, roundID)
	return err
}

// GetPendingRound 获取房间中未结算的回合（状态为 betting 或 playing）
func (r *GameRepo) GetPendingRound(ctx context.Context, roomID int64) (*model.GameRound, error) {
	sql := `SELECT id, room_id, round_number, participant_ids, skipped_ids, winner_ids,
		bet_amount, pool_amount, prize_per_winner, owner_earning, platform_earning, residual_amount,
		commit_hash, reveal_seed, status, failure_reason, created_at, settled_at
		FROM game_rounds 
		WHERE room_id = $1 AND status IN ('betting', 'playing')
		ORDER BY created_at DESC LIMIT 1`
	round := &model.GameRound{}
	err := DB.QueryRow(ctx, sql, roomID).Scan(
		&round.ID, &round.RoomID, &round.RoundNumber, &round.ParticipantIDs, &round.SkippedIDs, &round.WinnerIDs,
		&round.BetAmount, &round.PoolAmount, &round.PrizePerWinner, &round.OwnerEarning, &round.PlatformEarning, &round.ResidualAmount,
		&round.CommitHash, &round.RevealSeed, &round.Status, &round.FailureReason, &round.CreatedAt, &round.SettledAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // 没有未结算的回合
	}
	return round, err
}

// ListRounds 分页获取回合列表
func (r *GameRepo) ListRounds(ctx context.Context, roomID int64, page, pageSize int) ([]*model.GameRound, int64, error) {
	countSQL := `SELECT COUNT(*) FROM game_rounds WHERE room_id = $1`
	var total int64
	if err := DB.QueryRow(ctx, countSQL, roomID).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL := `SELECT id, room_id, round_number, participant_ids, skipped_ids, winner_ids,
		bet_amount, pool_amount, prize_per_winner, owner_earning, platform_earning, residual_amount,
		commit_hash, reveal_seed, status, failure_reason, created_at, settled_at
		FROM game_rounds WHERE room_id = $1 ORDER BY round_number DESC LIMIT $2 OFFSET $3`

	rows, err := DB.Query(ctx, listSQL, roomID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rounds []*model.GameRound
	for rows.Next() {
		round := &model.GameRound{}
		if err := rows.Scan(
			&round.ID, &round.RoomID, &round.RoundNumber, &round.ParticipantIDs, &round.SkippedIDs, &round.WinnerIDs,
			&round.BetAmount, &round.PoolAmount, &round.PrizePerWinner, &round.OwnerEarning, &round.PlatformEarning, &round.ResidualAmount,
			&round.CommitHash, &round.RevealSeed, &round.Status, &round.FailureReason, &round.CreatedAt, &round.SettledAt,
		); err != nil {
			return nil, 0, err
		}
		rounds = append(rounds, round)
	}
	return rounds, total, nil
}

// ===== Transaction 相关 =====

type TransactionRepo struct{}

func NewTransactionRepo() *TransactionRepo {
	return &TransactionRepo{}
}

// Create 创建交易记录
func (r *TransactionRepo) Create(ctx context.Context, tx *model.BalanceTransaction) error {
	return r.CreateTx(ctx, nil, tx)
}

// CreateTx 创建交易记录(支持事务)
func (r *TransactionRepo) CreateTx(ctx context.Context, dbTx pgx.Tx, tx *model.BalanceTransaction) error {
	sql := `INSERT INTO balance_transactions (user_id, room_id, round_id, tx_type, amount, balance_before, balance_after, balance_field, remark)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`
	balanceField := "balance"
	if tx.BalanceField != "" {
		balanceField = tx.BalanceField
	}
	exec := GetExecutor(dbTx)
	return exec.QueryRow(ctx, sql,
		tx.UserID, tx.RoomID, tx.RoundID, tx.Type, tx.Amount, tx.BalanceBefore, tx.BalanceAfter, balanceField, tx.Remark,
	).Scan(&tx.ID, &tx.CreatedAt)
}

// BatchCreateTx 批量创建交易记录（单条 SQL）
func (r *TransactionRepo) BatchCreateTx(ctx context.Context, dbTx pgx.Tx, txs []*model.BalanceTransaction) error {
	if len(txs) == 0 {
		return nil
	}

	// 构建批量 INSERT SQL
	valueStrings := make([]string, 0, len(txs))
	args := make([]interface{}, 0, len(txs)*9)
	argIdx := 1

	for _, tx := range txs {
		balanceField := "balance"
		if tx.BalanceField != "" {
			balanceField = tx.BalanceField
		}
		valueStrings = append(valueStrings, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			argIdx, argIdx+1, argIdx+2, argIdx+3, argIdx+4, argIdx+5, argIdx+6, argIdx+7, argIdx+8,
		))
		args = append(args, tx.UserID, tx.RoomID, tx.RoundID, tx.Type, tx.Amount, tx.BalanceBefore, tx.BalanceAfter, balanceField, tx.Remark)
		argIdx += 9
	}

	sql := fmt.Sprintf(`INSERT INTO balance_transactions 
		(user_id, room_id, round_id, tx_type, amount, balance_before, balance_after, balance_field, remark)
		VALUES %s`, strings.Join(valueStrings, ", "))

	exec := GetExecutor(dbTx)
	_, err := exec.Exec(ctx, sql, args...)
	return err
}

// List 分页获取交易记录
func (r *TransactionRepo) List(ctx context.Context, query *model.TransactionListQuery) ([]*model.BalanceTransaction, int64, error) {
	countSQL := `SELECT COUNT(*) FROM balance_transactions WHERE 1=1`
	listSQL := `SELECT id, user_id, room_id, round_id, tx_type, amount, balance_before, balance_after, remark, created_at
		FROM balance_transactions WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if query.UserID != nil {
		countSQL += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND user_id = $%d`, argIdx)
		args = append(args, *query.UserID)
		argIdx++
	}

	if query.RoomID != nil {
		countSQL += fmt.Sprintf(` AND room_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND room_id = $%d`, argIdx)
		args = append(args, *query.RoomID)
		argIdx++
	}

	if query.Type != nil {
		countSQL += fmt.Sprintf(` AND tx_type = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND tx_type = $%d`, argIdx)
		args = append(args, *query.Type)
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

	var txs []*model.BalanceTransaction
	for rows.Next() {
		tx := &model.BalanceTransaction{}
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.RoomID, &tx.RoundID, &tx.Type, &tx.Amount, &tx.BalanceBefore, &tx.BalanceAfter, &tx.Remark, &tx.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		txs = append(txs, tx)
	}
	return txs, total, nil
}

// ===== FundRequest 相关 =====

type FundRequestRepo struct{}

func NewFundRequestRepo() *FundRequestRepo {
	return &FundRequestRepo{}
}

// Create 创建资金申请
func (r *FundRequestRepo) Create(ctx context.Context, req *model.FundRequest) error {
	sql := `INSERT INTO fund_requests (user_id, request_type, amount, remark)
		VALUES ($1, $2, $3, $4)
		RETURNING id, status, created_at`
	return DB.QueryRow(ctx, sql, req.UserID, req.Type, req.Amount, req.Remark).Scan(&req.ID, &req.Status, &req.CreatedAt)
}

// GetByID 根据ID获取申请
func (r *FundRequestRepo) GetByID(ctx context.Context, id int64) (*model.FundRequest, error) {
	sql := `SELECT id, user_id, request_type, amount, status, remark, operator_id, updated_at, created_at
		FROM fund_requests WHERE id = $1`
	req := &model.FundRequest{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&req.ID, &req.UserID, &req.Type, &req.Amount, &req.Status, &req.Remark, &req.ProcessedBy, &req.ProcessedAt, &req.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return req, err
}

// Process 处理申请
func (r *FundRequestRepo) Process(ctx context.Context, id int64, status model.FundRequestStatus, processedBy int64, remark string) error {
	sql := `UPDATE fund_requests SET status = $1, operator_id = $2, remark = COALESCE(remark, '') || $3
		WHERE id = $4 AND status = 'pending'`
	tag, err := DB.Exec(ctx, sql, status, processedBy, remark, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("request not found or already processed")
	}
	return nil
}

// List 分页获取申请列表
func (r *FundRequestRepo) List(ctx context.Context, query *model.FundRequestListQuery) ([]*model.FundRequest, int64, error) {
	countSQL := `SELECT COUNT(*) FROM fund_requests f WHERE 1=1`
	listSQL := `SELECT f.id, f.user_id, COALESCE(u.username, '') as username, f.request_type, f.amount, f.status, f.remark, f.operator_id, f.updated_at, f.created_at
		FROM fund_requests f LEFT JOIN users u ON f.user_id = u.id WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if query.UserID != nil {
		countSQL += fmt.Sprintf(` AND f.user_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND f.user_id = $%d`, argIdx)
		args = append(args, *query.UserID)
		argIdx++
	}

	// 查询某个 owner 下级玩家的申请
	if query.InvitedBy != nil {
		countSQL += fmt.Sprintf(` AND f.user_id IN (SELECT id FROM users WHERE invited_by = $%d)`, argIdx)
		listSQL += fmt.Sprintf(` AND f.user_id IN (SELECT id FROM users WHERE invited_by = $%d)`, argIdx)
		args = append(args, *query.InvitedBy)
		argIdx++
	}

	if query.Type != nil {
		countSQL += fmt.Sprintf(` AND f.request_type = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND f.request_type = $%d`, argIdx)
		args = append(args, *query.Type)
		argIdx++
	}

	if query.Status != nil {
		countSQL += fmt.Sprintf(` AND f.status = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND f.status = $%d`, argIdx)
		args = append(args, *query.Status)
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL += fmt.Sprintf(` ORDER BY f.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var reqs []*model.FundRequest
	for rows.Next() {
		req := &model.FundRequest{}
		if err := rows.Scan(
			&req.ID, &req.UserID, &req.Username, &req.Type, &req.Amount, &req.Status, &req.Remark, &req.ProcessedBy, &req.ProcessedAt, &req.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		reqs = append(reqs, req)
	}
	return reqs, total, nil
}

// ===== PlatformAccount 相关 =====

type PlatformRepo struct{}

func NewPlatformRepo() *PlatformRepo {
	return &PlatformRepo{}
}

// GetAccount 获取平台账户
func (r *PlatformRepo) GetAccount(ctx context.Context) (*model.PlatformAccount, error) {
	sql := `SELECT id, platform_balance, updated_at FROM platform_account WHERE id = 1`
	acc := &model.PlatformAccount{}
	err := DB.QueryRow(ctx, sql).Scan(&acc.ID, &acc.PlatformBalance, &acc.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return acc, err
}

// UpdateBalance 更新平台余额
func (r *PlatformRepo) UpdateBalance(ctx context.Context, delta decimal.Decimal) error {
	return r.UpdateBalanceTx(ctx, nil, delta)
}

// UpdateBalanceTx 更新平台余额(支持事务)
func (r *PlatformRepo) UpdateBalanceTx(ctx context.Context, tx pgx.Tx, delta decimal.Decimal) error {
	sql := `UPDATE platform_account SET platform_balance = platform_balance + $1, updated_at = NOW() WHERE id = 1`
	exec := GetExecutor(tx)
	_, err := exec.Exec(ctx, sql, delta)
	return err
}

// CheckConservation 检查资金守恒
// 资金守恒公式: 系统内资金总和 = 房主净充值额
// 系统内资金 = 玩家余额 + 房主可用余额 + 房主佣金 + 平台余额
func (r *PlatformRepo) CheckConservation(ctx context.Context) (*model.ConservationCheck, error) {
	result := &model.ConservationCheck{}

	// 1. 获取所有玩家余额（可用+冻结）
	err := DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(balance), 0), 
		COALESCE(SUM(frozen_balance), 0) 
		FROM users WHERE role = 'player'`).Scan(&result.TotalPlayerBalance, &result.TotalPlayerFrozen)
	if err != nil {
		return nil, err
	}

	// 2. 获取所有房主余额信息
	err = DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(balance), 0),
		COALESCE(SUM(owner_room_balance), 0),
		COALESCE(SUM(owner_margin_balance), 0),
		COALESCE(SUM(owner_custody_quota), 0)
		FROM users WHERE role = 'owner'`).Scan(
		&result.TotalOwnerBalance,
		&result.TotalOwnerCommission,
		&result.TotalMargin,
		&result.TotalCustodyQuota,
	)
	if err != nil {
		return nil, err
	}

	// 3. 获取平台余额
	err = DB.QueryRow(ctx, `SELECT COALESCE(platform_balance, 0) FROM platform_account WHERE id = 1`).Scan(&result.PlatformBalance)
	if err != nil {
		return nil, err
	}

	// 4. 从fund_requests表计算外部资金进出（只有房主才能和外部有资金往来）
	// owner_deposit: 房主充值（外部 -> 房主余额）
	// margin_deposit: 保证金充值（外部 -> 房主保证金）
	// owner_withdraw: 房主提现（房主余额 -> 外部）
	err = DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(CASE WHEN request_type = 'owner_deposit' THEN amount ELSE 0 END), 0) AS owner_deposit,
		COALESCE(SUM(CASE WHEN request_type = 'margin_deposit' THEN amount ELSE 0 END), 0) AS margin_deposit,
		COALESCE(SUM(CASE WHEN request_type = 'owner_withdraw' THEN amount ELSE 0 END), 0) AS owner_withdraw
		FROM fund_requests 
		WHERE status = 'approved'`).Scan(&result.TotalOwnerDeposit, &result.TotalMargin, &result.TotalOwnerWithdraw)
	if err != nil {
		// 如果查询失败，不影响主流程，设为0
		result.TotalOwnerDeposit = decimal.Zero
		result.TotalOwnerWithdraw = decimal.Zero
	}
	// 外部注入总额 = 房主充值 + 保证金充值
	result.TotalOwnerDeposit = result.TotalOwnerDeposit.Add(result.TotalMargin)

	// 5. 计算系统内资金总和
	// 系统内资金 = 玩家余额(可用+冻结) + 房主可用余额 + 房主佣金收益 + 房主保证金 + 平台余额
	result.SystemTotalFunds = result.TotalPlayerBalance.
		Add(result.TotalPlayerFrozen).
		Add(result.TotalOwnerBalance).
		Add(result.TotalOwnerCommission).
		Add(result.TotalMargin).
		Add(result.PlatformBalance)

	// 6. 计算预期总额（房主净充值 = 充值 - 提现）
	result.ExpectedTotal = result.TotalOwnerDeposit.Sub(result.TotalOwnerWithdraw)

	// 7. 计算差额
	// 注意：由于保证金是固定的担保资金，不参与日常流转，这里单独考虑
	// 差额 = 系统内资金 - 预期总额
	// 如果差额为0，说明资金守恒
	result.Difference = result.SystemTotalFunds.Sub(result.ExpectedTotal)

	// 允许小额误差（由于精度问题，允许0.01的误差）
	tolerance := decimal.NewFromFloat(0.01)
	result.IsBalanced = result.Difference.Abs().LessThanOrEqual(tolerance)

	return result, nil
}

// CheckConservationByOwner 按房主维度检查资金守恒
func (r *PlatformRepo) CheckConservationByOwner(ctx context.Context, ownerID int64) (*model.ConservationCheck, error) {
	result := &model.ConservationCheck{}

	// 1. 获取该房主名下所有玩家余额
	err := DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(balance), 0), 
		COALESCE(SUM(frozen_balance), 0) 
		FROM users WHERE role = 'player' AND invited_by = $1`, ownerID).Scan(
		&result.TotalPlayerBalance, &result.TotalPlayerFrozen)
	if err != nil {
		return nil, err
	}

	// 2. 获取房主自身余额信息
	err = DB.QueryRow(ctx, `SELECT 
		balance,
		owner_room_balance,
		owner_margin_balance,
		owner_custody_quota
		FROM users WHERE id = $1`, ownerID).Scan(
		&result.TotalOwnerBalance,
		&result.TotalOwnerCommission,
		&result.TotalMargin,
		&result.TotalCustodyQuota,
	)
	if err != nil {
		return nil, err
	}

	// 3. 计算该房主体系内的资金总和
	// 房主体系内资金 = 玩家余额 + 房主可用余额 + 房主佣金
	result.SystemTotalFunds = result.TotalPlayerBalance.
		Add(result.TotalPlayerFrozen).
		Add(result.TotalOwnerBalance).
		Add(result.TotalOwnerCommission)

	// 4. 从交易记录计算该房主的累计充值和提现
	err = DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(CASE WHEN tx_type = 'deposit' AND amount > 0 THEN amount ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN tx_type = 'withdraw' AND amount < 0 THEN ABS(amount) ELSE 0 END), 0)
		FROM balance_transactions
		WHERE user_id = $1`, ownerID).Scan(&result.TotalOwnerDeposit, &result.TotalOwnerWithdraw)
	if err != nil {
		result.TotalOwnerDeposit = decimal.Zero
		result.TotalOwnerWithdraw = decimal.Zero
	}

	// 5. 计算预期总额
	result.ExpectedTotal = result.TotalOwnerDeposit.Sub(result.TotalOwnerWithdraw)

	// 6. 计算差额
	result.Difference = result.SystemTotalFunds.Sub(result.ExpectedTotal)

	tolerance := decimal.NewFromFloat(0.01)
	result.IsBalanced = result.Difference.Abs().LessThanOrEqual(tolerance)

	return result, nil
}

// GetReconciliationReport 获取详细的资金对账报告
func (r *PlatformRepo) GetReconciliationReport(ctx context.Context) (*model.FundReconciliationReport, error) {
	report := &model.FundReconciliationReport{}

	// 1. 获取外部资金注入（从fund_requests表，只有房主才能和外部有资金往来）
	err := DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(CASE WHEN request_type = 'owner_deposit' THEN amount ELSE 0 END), 0) AS owner_deposit,
		COALESCE(SUM(CASE WHEN request_type = 'margin_deposit' THEN amount ELSE 0 END), 0) AS margin_deposit,
		COALESCE(SUM(CASE WHEN request_type = 'owner_withdraw' THEN amount ELSE 0 END), 0) AS owner_withdraw
		FROM fund_requests 
		WHERE status = 'approved'`).Scan(
		&report.ExternalFunds.OwnerDeposit,
		&report.ExternalFunds.MarginDeposit,
		&report.ExternalFunds.OwnerWithdraw,
	)
	if err != nil {
		return nil, err
	}
	report.ExternalFunds.NetInflow = report.ExternalFunds.OwnerDeposit.
		Add(report.ExternalFunds.MarginDeposit).
		Sub(report.ExternalFunds.OwnerWithdraw)

	// 2. 获取玩家余额
	err = DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(balance), 0), 
		COALESCE(SUM(frozen_balance), 0) 
		FROM users WHERE role = 'player'`).Scan(
		&report.SystemFunds.PlayerBalance,
		&report.SystemFunds.PlayerFrozen,
	)
	if err != nil {
		return nil, err
	}

	// 3. 获取房主余额
	err = DB.QueryRow(ctx, `SELECT 
		COALESCE(SUM(balance), 0),
		COALESCE(SUM(owner_room_balance), 0),
		COALESCE(SUM(owner_margin_balance), 0)
		FROM users WHERE role = 'owner'`).Scan(
		&report.SystemFunds.OwnerBalance,
		&report.SystemFunds.OwnerCommission,
		&report.SystemFunds.OwnerMargin,
	)
	if err != nil {
		return nil, err
	}

	// 4. 获取平台余额
	err = DB.QueryRow(ctx, `SELECT COALESCE(platform_balance, 0) FROM platform_account WHERE id = 1`).Scan(
		&report.SystemFunds.PlatformBalance,
	)
	if err != nil {
		report.SystemFunds.PlatformBalance = decimal.Zero
	}

	// 5. 计算系统内资金总和
	report.SystemFunds.Total = report.SystemFunds.PlayerBalance.
		Add(report.SystemFunds.PlayerFrozen).
		Add(report.SystemFunds.OwnerBalance).
		Add(report.SystemFunds.OwnerCommission).
		Add(report.SystemFunds.OwnerMargin).
		Add(report.SystemFunds.PlatformBalance)

	// 6. 对账结果
	report.Reconciliation.ExpectedTotal = report.ExternalFunds.NetInflow
	report.Reconciliation.ActualTotal = report.SystemFunds.Total
	report.Reconciliation.Difference = report.Reconciliation.ActualTotal.Sub(report.Reconciliation.ExpectedTotal)

	tolerance := decimal.NewFromFloat(0.01)
	report.Reconciliation.IsBalanced = report.Reconciliation.Difference.Abs().LessThanOrEqual(tolerance)

	// 7. 差异分析
	// 检查是否有未通过fund_requests记录的保证金
	report.Analysis.UnrecordedMargin = report.SystemFunds.OwnerMargin.Sub(report.ExternalFunds.MarginDeposit)
	
	if !report.Reconciliation.IsBalanced {
		if report.Analysis.UnrecordedMargin.GreaterThan(decimal.Zero) {
			report.Analysis.Explanation = fmt.Sprintf(
				"差异主要来自未通过fund_requests记录的保证金: %.2f。这部分可能是数据库初始化时直接设置的。",
				report.Analysis.UnrecordedMargin.InexactFloat64(),
			)
		} else {
			report.Analysis.Explanation = "存在资金差异，需要进一步排查。"
		}
	} else {
		report.Analysis.Explanation = "资金守恒，无异常。"
	}

	return report, nil
}

// GetUserGameHistory 获取用户游戏历史
func (r *GameRepo) GetUserGameHistory(ctx context.Context, query *model.GameHistoryQuery) ([]*model.GameHistoryItem, int64, error) {
	// 构建查询条件
	countSQL := `SELECT COUNT(*) FROM game_rounds gr
		JOIN rooms rm ON gr.room_id = rm.id
		WHERE $1 = ANY(gr.participant_ids) OR $1 = ANY(gr.skipped_ids)`
	
	listSQL := `SELECT gr.id, gr.room_id, COALESCE(rm.name, 'Room ' || rm.code) as room_name, 
		gr.round_number, gr.bet_amount, gr.winner_ids, gr.prize_per_winner, gr.created_at
		FROM game_rounds gr
		JOIN rooms rm ON gr.room_id = rm.id
		WHERE ($1 = ANY(gr.participant_ids) OR $1 = ANY(gr.skipped_ids))`

	args := []interface{}{query.UserID}
	argIdx := 2

	if query.RoomID != nil {
		countSQL += fmt.Sprintf(` AND gr.room_id = $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND gr.room_id = $%d`, argIdx)
		args = append(args, *query.RoomID)
		argIdx++
	}

	if query.StartDate != nil {
		countSQL += fmt.Sprintf(` AND gr.created_at >= $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND gr.created_at >= $%d`, argIdx)
		args = append(args, *query.StartDate)
		argIdx++
	}

	if query.EndDate != nil {
		countSQL += fmt.Sprintf(` AND gr.created_at <= $%d`, argIdx)
		listSQL += fmt.Sprintf(` AND gr.created_at <= $%d`, argIdx)
		args = append(args, *query.EndDate)
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	pageSize := query.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}

	listSQL += fmt.Sprintf(` ORDER BY gr.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, pageSize, (page-1)*pageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []*model.GameHistoryItem
	for rows.Next() {
		var item model.GameHistoryItem
		var winnerIDs []int64
		var prizePerWinner *decimal.Decimal

		if err := rows.Scan(&item.ID, &item.RoomID, &item.RoomName, &item.RoundNumber,
			&item.BetAmount, &winnerIDs, &prizePerWinner, &item.CreatedAt); err != nil {
			return nil, 0, err
		}

		// 判断结果
		isWinner := false
		for _, wid := range winnerIDs {
			if wid == query.UserID {
				isWinner = true
				break
			}
		}

		if isWinner {
			item.Result = "win"
			if prizePerWinner != nil {
				item.PrizeAmount = *prizePerWinner
			}
		} else {
			// 检查是否跳过
			item.Result = "lose"
			item.PrizeAmount = decimal.Zero
		}

		items = append(items, &item)
	}

	return items, total, nil
}

// GetUserGameStats 获取用户游戏统计
func (r *GameRepo) GetUserGameStats(ctx context.Context, userID int64) (*model.GameStats, error) {
	sql := `SELECT 
		COUNT(*) FILTER (WHERE $1 = ANY(participant_ids)) as total_participated,
		COUNT(*) FILTER (WHERE $1 = ANY(winner_ids)) as total_wins,
		COUNT(*) FILTER (WHERE $1 = ANY(skipped_ids)) as total_skipped,
		COALESCE(SUM(bet_amount) FILTER (WHERE $1 = ANY(participant_ids)), 0) as total_wagered,
		COALESCE(SUM(prize_per_winner) FILTER (WHERE $1 = ANY(winner_ids)), 0) as total_won
		FROM game_rounds
		WHERE status = 'settled' AND ($1 = ANY(participant_ids) OR $1 = ANY(skipped_ids))`

	var stats model.GameStats
	var totalParticipated int
	var totalWon decimal.Decimal

	err := DB.QueryRow(ctx, sql, userID).Scan(
		&totalParticipated, &stats.TotalWins, &stats.TotalSkipped,
		&stats.TotalWagered, &totalWon,
	)
	if err != nil {
		return nil, err
	}

	stats.TotalRounds = totalParticipated + stats.TotalSkipped
	stats.TotalLosses = totalParticipated - stats.TotalWins
	stats.TotalWon = totalWon
	stats.NetProfit = totalWon.Sub(stats.TotalWagered)

	if totalParticipated > 0 {
		stats.WinRate = float64(stats.TotalWins) / float64(totalParticipated) * 100
	}

	return &stats, nil
}

// GetRoundDetail 获取回合详情
func (r *GameRepo) GetRoundDetail(ctx context.Context, roundID int64) (*model.RoundDetail, error) {
	sql := `SELECT gr.id, gr.room_id, COALESCE(rm.name, 'Room ' || rm.code) as room_name,
		gr.round_number, gr.bet_amount, gr.pool_amount, COALESCE(gr.prize_per_winner, 0),
		COALESCE(gr.commit_hash, ''), COALESCE(gr.reveal_seed, ''), gr.status,
		gr.participant_ids, gr.winner_ids, gr.created_at, gr.settled_at
		FROM game_rounds gr
		JOIN rooms rm ON gr.room_id = rm.id
		WHERE gr.id = $1`

	var detail model.RoundDetail
	var participantIDs, winnerIDs []int64

	err := DB.QueryRow(ctx, sql, roundID).Scan(
		&detail.ID, &detail.RoomID, &detail.RoomName, &detail.RoundNumber,
		&detail.BetAmount, &detail.PoolAmount, &detail.PrizePerWinner,
		&detail.CommitHash, &detail.RevealSeed, &detail.Status,
		&participantIDs, &winnerIDs, &detail.CreatedAt, &detail.SettledAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// 获取参与者用户名
	if len(participantIDs) > 0 {
		userSQL := `SELECT id, username FROM users WHERE id = ANY($1)`
		rows, err := DB.Query(ctx, userSQL, participantIDs)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		userMap := make(map[int64]string)
		for rows.Next() {
			var id int64
			var username string
			if err := rows.Scan(&id, &username); err != nil {
				return nil, err
			}
			userMap[id] = username
		}

		winnerSet := make(map[int64]bool)
		for _, wid := range winnerIDs {
			winnerSet[wid] = true
		}

		for _, pid := range participantIDs {
			detail.Participants = append(detail.Participants, model.Participant{
				UserID:   pid,
				Username: userMap[pid],
				IsWinner: winnerSet[pid],
			})
		}

		for _, wid := range winnerIDs {
			detail.Winners = append(detail.Winners, model.Winner{
				UserID:   wid,
				Username: userMap[wid],
			})
		}
	}

	return &detail, nil
}
