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

var (
	ErrNotFound            = errors.New("record not found")
	ErrVersionConflict     = errors.New("version conflict")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type UserRepo struct{}

func NewUserRepo() *UserRepo {
	return &UserRepo{}
}

// Create 创建用户
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	sql := `INSERT INTO users (username, password_hash, role, invite_code, invited_by, balance, frozen_balance,
		owner_room_balance, owner_margin_balance)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at`
	return DB.QueryRow(ctx, sql,
		user.Username, user.PasswordHash, user.Role, user.InviteCode, user.InvitedBy,
		user.Balance, user.FrozenBalance, user.OwnerRoomBalance, user.OwnerMarginBalance,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByID 根据ID获取用户
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	sql := `SELECT id, username, password_hash, role, invite_code, invited_by, balance, frozen_balance, balance_version,
		owner_room_balance, owner_margin_balance, created_at, updated_at FROM users WHERE id = $1`
	user := &model.User{}
	err := DB.QueryRow(ctx, sql, id).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.InviteCode, &user.InvitedBy,
		&user.Balance, &user.FrozenBalance, &user.BalanceVersion, &user.OwnerRoomBalance, &user.OwnerMarginBalance,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

// GetByUsername 根据用户名获取用户
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	sql := `SELECT id, username, password_hash, role, invite_code, invited_by, balance, frozen_balance, balance_version,
		owner_room_balance, owner_margin_balance, created_at, updated_at FROM users WHERE username = $1`
	user := &model.User{}
	err := DB.QueryRow(ctx, sql, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.InviteCode, &user.InvitedBy,
		&user.Balance, &user.FrozenBalance, &user.BalanceVersion, &user.OwnerRoomBalance, &user.OwnerMarginBalance,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

// GetByInviteCodeAllRoles 根据邀请码获取用户（不分角色）
func (r *UserRepo) GetByInviteCodeAllRoles(ctx context.Context, code string) (*model.User, error) {
	sql := `SELECT id, username, password_hash, role, invite_code, invited_by, balance, frozen_balance, balance_version,
		owner_room_balance, owner_margin_balance, created_at, updated_at FROM users WHERE invite_code = $1`
	user := &model.User{}
	err := DB.QueryRow(ctx, sql, code).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.InviteCode, &user.InvitedBy,
		&user.Balance, &user.FrozenBalance, &user.BalanceVersion, &user.OwnerRoomBalance, &user.OwnerMarginBalance,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

// GetByInviteCode 根据邀请码获取房主
func (r *UserRepo) GetByInviteCode(ctx context.Context, code string) (*model.User, error) {
	sql := `SELECT id, username, password_hash, role, invite_code, invited_by, balance, frozen_balance, balance_version,
		owner_room_balance, owner_margin_balance, created_at, updated_at FROM users WHERE invite_code = $1 AND role = 'owner'`
	user := &model.User{}
	err := DB.QueryRow(ctx, sql, code).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.InviteCode, &user.InvitedBy,
		&user.Balance, &user.FrozenBalance, &user.BalanceVersion, &user.OwnerRoomBalance, &user.OwnerMarginBalance,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return user, err
}

// List 分页列表
func (r *UserRepo) List(ctx context.Context, query *model.UserListQuery) ([]*model.User, int64, error) {
	countSQL := `SELECT COUNT(*) FROM users WHERE 1=1`
	listSQL := `SELECT id, username, password_hash, role, invite_code, invited_by, balance, frozen_balance, balance_version,
		owner_room_balance, owner_margin_balance, created_at, updated_at FROM users WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if query.Role != nil {
		countSQL += ` AND role = $` + string(rune('0'+argIdx))
		listSQL += ` AND role = $` + string(rune('0'+argIdx))
		args = append(args, *query.Role)
		argIdx++
	}

	if query.Search != nil && *query.Search != "" {
		countSQL += ` AND username ILIKE $` + string(rune('0'+argIdx))
		listSQL += ` AND username ILIKE $` + string(rune('0'+argIdx))
		args = append(args, "%"+*query.Search+"%")
		argIdx++
	}

	var total int64
	if err := DB.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listSQL += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+argIdx)) + ` OFFSET $` + string(rune('0'+argIdx+1))
	args = append(args, query.PageSize, (query.Page-1)*query.PageSize)

	rows, err := DB.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		user := &model.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.PasswordHash, &user.Role, &user.InviteCode, &user.InvitedBy,
			&user.Balance, &user.FrozenBalance, &user.BalanceVersion, &user.OwnerRoomBalance, &user.OwnerMarginBalance,
			&user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, user)
	}
	return users, total, nil
}

// ListOwnerPlayers 获取房主名下的玩家统计
func (r *UserRepo) ListOwnerPlayers(ctx context.Context, ownerID int64) ([]*model.PlayerStat, error) {
	sql := `
		SELECT 
			u.id, u.username, u.balance, u.frozen_balance, u.created_at,
			COALESCE(SUM(CASE WHEN tx.tx_type = 'deposit' THEN tx.amount ELSE 0 END), 0) as total_deposit,
			COALESCE(SUM(CASE WHEN tx.tx_type = 'withdraw' THEN ABS(tx.amount) ELSE 0 END), 0) as total_withdraw
		FROM users u
		LEFT JOIN balance_transactions tx ON u.id = tx.user_id
		WHERE u.invited_by = $1 AND u.role = 'player'
		GROUP BY u.id
		ORDER BY u.created_at DESC
	`
	rows, err := DB.Query(ctx, sql, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*model.PlayerStat
	for rows.Next() {
		stat := &model.PlayerStat{}
		if err := rows.Scan(
			&stat.ID, &stat.Username, &stat.Balance, &stat.FrozenBalance, &stat.RegistrationTime,
			&stat.TotalDeposit, &stat.TotalWithdraw,
		); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	return stats, nil
}

// UpdateBalance 更新用户余额(带乐观锁)
func (r *UserRepo) UpdateBalance(ctx context.Context, userID int64, delta decimal.Decimal) error {
	return r.UpdateBalanceTx(ctx, nil, userID, delta)
}

// UpdateBalanceTx 更新用户余额(支持事务)
func (r *UserRepo) UpdateBalanceTx(ctx context.Context, tx pgx.Tx, userID int64, delta decimal.Decimal) error {
	sql := `UPDATE users SET balance = balance + $1, updated_at = NOW() WHERE id = $2 AND balance + $1 >= 0`
	exec := GetExecutor(tx)
	tag, err := exec.Exec(ctx, sql, delta, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient balance or user not found")
	}
	return nil
}

// UpdateFrozenBalance 更新冻结余额
func (r *UserRepo) UpdateFrozenBalance(ctx context.Context, userID int64, delta decimal.Decimal) error {
	sql := `UPDATE users SET frozen_balance = frozen_balance + $1, updated_at = NOW() WHERE id = $2 AND frozen_balance + $1 >= 0`
	tag, err := DB.Exec(ctx, sql, delta, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient frozen balance or user not found")
	}
	return nil
}

// UpdateOwnerBalances 更新房主各类余额
func (r *UserRepo) UpdateOwnerBalances(ctx context.Context, userID int64, field string, delta decimal.Decimal) error {
	return r.UpdateOwnerBalancesTx(ctx, nil, userID, field, delta)
}

// UpdateOwnerBalancesTx 更新房主各类余额(支持事务)
func (r *UserRepo) UpdateOwnerBalancesTx(ctx context.Context, tx pgx.Tx, userID int64, field string, delta decimal.Decimal) error {
	validFields := map[string]bool{
		"owner_room_balance":   true, // 佣金收益
		"owner_margin_balance": true, // 保证金（仅初始设置）
	}
	if !validFields[field] {
		return errors.New("invalid field")
	}
	sql := `UPDATE users SET ` + field + ` = ` + field + ` + $1, updated_at = NOW() WHERE id = $2 AND ` + field + ` + $1 >= 0`
	exec := GetExecutor(tx)
	tag, err := exec.Exec(ctx, sql, delta, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient balance or user not found")
	}
	return nil
}

// GetTotalPlayerBalance 获取某房主下所有玩家的总余额
func (r *UserRepo) GetTotalPlayerBalance(ctx context.Context, ownerID int64) (decimal.Decimal, error) {
	sql := `SELECT COALESCE(SUM(balance + frozen_balance), 0) FROM users WHERE invited_by = $1 AND role = 'player'`
	var total decimal.Decimal
	err := DB.QueryRow(ctx, sql, ownerID).Scan(&total)
	return total, err
}

// UsernameExists 检查用户名是否存在
func (r *UserRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := DB.QueryRow(ctx, sql, username).Scan(&exists)
	return exists, err
}

// InviteCodeExists 检查邀请码是否存在
func (r *UserRepo) InviteCodeExists(ctx context.Context, code string) (bool, error) {
	sql := `SELECT EXISTS(SELECT 1 FROM users WHERE invite_code = $1)`
	var exists bool
	err := DB.QueryRow(ctx, sql, code).Scan(&exists)
	return exists, err
}

// UpdateBalanceWithVersion 使用乐观锁更新余额
// 返回新余额和新版本号
func (r *UserRepo) UpdateBalanceWithVersion(ctx context.Context, userID int64, delta decimal.Decimal, expectedVersion int64) (decimal.Decimal, int64, error) {
	sql := `UPDATE users 
		SET balance = balance + $1, 
		    balance_version = balance_version + 1, 
		    updated_at = NOW() 
		WHERE id = $2 
		  AND balance_version = $3 
		  AND balance + $1 >= 0
		RETURNING balance, balance_version`

	var newBalance decimal.Decimal
	var newVersion int64
	err := DB.QueryRow(ctx, sql, delta, userID, expectedVersion).Scan(&newBalance, &newVersion)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// 需要区分是版本冲突还是余额不足
			// 先检查版本
			var currentVersion int64
			var currentBalance decimal.Decimal
			checkSQL := `SELECT balance, balance_version FROM users WHERE id = $1`
			if checkErr := DB.QueryRow(ctx, checkSQL, userID).Scan(&currentBalance, &currentVersion); checkErr == nil {
				if currentVersion != expectedVersion {
					return decimal.Zero, 0, ErrVersionConflict
				}
				if currentBalance.Add(delta).LessThan(decimal.Zero) {
					return decimal.Zero, 0, ErrInsufficientBalance
				}
			}
			return decimal.Zero, 0, ErrNotFound
		}
		return decimal.Zero, 0, err
	}
	return newBalance, newVersion, nil
}

// GetBalanceVersion 获取用户余额版本号
func (r *UserRepo) GetBalanceVersion(ctx context.Context, userID int64) (int64, error) {
	sql := `SELECT balance_version FROM users WHERE id = $1`
	var version int64
	err := DB.QueryRow(ctx, sql, userID).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return version, err
}

// UpdateDeviceFingerprint 更新用户设备指纹
func (r *UserRepo) UpdateDeviceFingerprint(ctx context.Context, userID int64, fingerprint string) error {
	sql := `UPDATE users SET device_fingerprint = $1, updated_at = NOW() WHERE id = $2`
	_, err := DB.Exec(ctx, sql, fingerprint, userID)
	return err
}

// GetDeviceFingerprint 获取用户设备指纹
func (r *UserRepo) GetDeviceFingerprint(ctx context.Context, userID int64) (string, error) {
	sql := `SELECT COALESCE(device_fingerprint, '') FROM users WHERE id = $1`
	var fingerprint string
	err := DB.QueryRow(ctx, sql, userID).Scan(&fingerprint)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return fingerprint, err
}

// UpdateLanguage 更新用户语言偏好
func (r *UserRepo) UpdateLanguage(ctx context.Context, userID int64, language string) error {
	sql := `UPDATE users SET language = $1, updated_at = NOW() WHERE id = $2`
	_, err := DB.Exec(ctx, sql, language, userID)
	return err
}

// GetLanguage 获取用户语言偏好
func (r *UserRepo) GetLanguage(ctx context.Context, userID int64) (string, error) {
	sql := `SELECT COALESCE(language, 'zh') FROM users WHERE id = $1`
	var language string
	err := DB.QueryRow(ctx, sql, userID).Scan(&language)
	if errors.Is(err, pgx.ErrNoRows) {
		return "zh", ErrNotFound
	}
	return language, err
}

// BatchDeductBalanceResult 批量扣款结果
type BatchDeductBalanceResult struct {
	UserID     int64
	NewBalance decimal.Decimal
}

// BatchDeductBalanceTx 批量扣款（单条 SQL）
// 返回成功扣款的用户ID和新余额
func (r *UserRepo) BatchDeductBalanceTx(ctx context.Context, tx pgx.Tx, userIDs []int64, amount decimal.Decimal) ([]BatchDeductBalanceResult, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	// 使用单条 UPDATE + RETURNING 批量扣款
	// 只有余额足够的用户才会被更新
	sql := `UPDATE users 
		SET balance = balance - $1, 
		    balance_version = balance_version + 1,
		    updated_at = NOW() 
		WHERE id = ANY($2) 
		  AND balance >= $1
		RETURNING id, balance`

	exec := GetExecutor(tx)
	rows, err := exec.Query(ctx, sql, amount, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BatchDeductBalanceResult
	for rows.Next() {
		var r BatchDeductBalanceResult
		if err := rows.Scan(&r.UserID, &r.NewBalance); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// BatchAddBalanceResult 批量加款结果
type BatchAddBalanceResult struct {
	UserID     int64
	NewBalance decimal.Decimal
}

// BatchAddBalanceTx 批量加款（单条 SQL）
// amounts 是一个 map，key 是用户ID，value 是要增加的金额
func (r *UserRepo) BatchAddBalanceTx(ctx context.Context, tx pgx.Tx, amounts map[int64]decimal.Decimal) ([]BatchAddBalanceResult, error) {
	if len(amounts) == 0 {
		return nil, nil
	}

	// 构建批量更新 SQL
	// UPDATE users SET balance = balance + CASE id WHEN 1 THEN 100 WHEN 2 THEN 200 END WHERE id IN (1, 2)
	userIDs := make([]int64, 0, len(amounts))
	caseStmts := make([]string, 0, len(amounts))
	args := []interface{}{}
	argIdx := 1

	for userID, amount := range amounts {
		userIDs = append(userIDs, userID)
		// 显式转换为 NUMERIC 类型，避免类型不匹配
		caseStmts = append(caseStmts, fmt.Sprintf("WHEN $%d THEN $%d::NUMERIC", argIdx, argIdx+1))
		args = append(args, userID, amount.String())
		argIdx += 2
	}

	sql := fmt.Sprintf(`UPDATE users 
		SET balance = balance + CASE id %s END,
		    balance_version = balance_version + 1,
		    updated_at = NOW()
		WHERE id = ANY($%d)
		RETURNING id, balance`, 
		strings.Join(caseStmts, " "), argIdx)
	args = append(args, userIDs)

	exec := GetExecutor(tx)
	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []BatchAddBalanceResult
	for rows.Next() {
		var r BatchAddBalanceResult
		if err := rows.Scan(&r.UserID, &r.NewBalance); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	return results, rows.Err()
}
