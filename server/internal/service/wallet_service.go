package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

// WalletService 钱包服务
type WalletService struct {
	userRepo *repository.UserRepo
	txRepo   *repository.TransactionRepo
}

// NewWalletService 创建钱包服务
func NewWalletService(userRepo *repository.UserRepo, txRepo *repository.TransactionRepo) *WalletService {
	return &WalletService{
		userRepo: userRepo,
		txRepo:   txRepo,
	}
}

// WalletInfo 钱包信息
type WalletInfo struct {
	AvailableBalance decimal.Decimal `json:"available_balance"` // 可用余额
	FrozenBalance    decimal.Decimal `json:"frozen_balance"`    // 冻结余额（游戏中）
	TotalBalance     decimal.Decimal `json:"total_balance"`     // 总余额
	// 房主专属字段
	IsOwner              bool            `json:"is_owner"`
	OwnerMarginBalance   decimal.Decimal `json:"owner_margin_balance,omitempty"`   // 保证金（固定不变）
	OwnerRoomBalance     decimal.Decimal `json:"owner_room_balance,omitempty"`     // 佣金收益
	OwnerTotalCommission decimal.Decimal `json:"owner_total_commission,omitempty"` // 累计佣金收益
}

// GetWallet 获取钱包信息
func (s *WalletService) GetWallet(ctx context.Context, userID int64) (*WalletInfo, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	info := &WalletInfo{
		AvailableBalance: user.Balance,
		FrozenBalance:    user.FrozenBalance,
		TotalBalance:     user.Balance.Add(user.FrozenBalance),
		IsOwner:          user.IsOwner(),
	}

	// 房主专属字段
	if user.IsOwner() {
		info.OwnerMarginBalance = user.OwnerMarginBalance
		info.OwnerRoomBalance = user.OwnerRoomBalance

		// 计算累计佣金收益
		totalCommission, err := s.getTotalOwnerCommission(ctx, userID)
		if err == nil {
			info.OwnerTotalCommission = totalCommission
		}
	}

	return info, nil
}

// getTotalOwnerCommission 获取房主累计佣金收益
func (s *WalletService) getTotalOwnerCommission(ctx context.Context, userID int64) (decimal.Decimal, error) {
	txType := model.TxOwnerCommission
	query := &model.TransactionListQuery{
		UserID:   &userID,
		Type:     &txType,
		Page:     1,
		PageSize: 100000,
	}
	txs, _, err := s.txRepo.List(ctx, query)
	if err != nil {
		return decimal.Zero, err
	}

	total := decimal.Zero
	for _, tx := range txs {
		total = total.Add(tx.Amount)
	}
	return total, nil
}

// TransferEarningsToBalance 房主佣金收益转可用余额（使用事务保护）
func (s *WalletService) TransferEarningsToBalance(ctx context.Context, userID int64, amount decimal.Decimal) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.IsOwner() {
		return errors.New("only owner can transfer earnings")
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("amount must be positive")
	}

	if user.OwnerRoomBalance.LessThan(amount) {
		return errors.New("insufficient commission balance")
	}

	// 使用事务确保原子性：从 owner_room_balance 转到 balance（可用余额）
	return repository.Tx(ctx, func(tx pgx.Tx) error {
		// 1. 从佣金余额扣除
		newCommissionBalance := user.OwnerRoomBalance.Sub(amount)
		if err := s.userRepo.UpdateOwnerBalancesTx(ctx, tx, userID, "owner_room_balance", amount.Neg()); err != nil {
			return fmt.Errorf("deduct commission balance: %w", err)
		}
		// 记录佣金扣除交易
		commissionTx := &model.BalanceTransaction{
			UserID:        userID,
			Type:          model.TxEarningsTransfer,
			Amount:        amount.Neg(),
			BalanceBefore: user.OwnerRoomBalance,
			BalanceAfter:  newCommissionBalance,
			BalanceField:  "owner_room_balance",
			Remark:        strPtr("佣金转可用余额(扣除)"),
		}
		if err := s.txRepo.CreateTx(ctx, tx, commissionTx); err != nil {
			return fmt.Errorf("create commission deduct transaction: %w", err)
		}

		// 2. 增加可用余额
		newBalance := user.Balance.Add(amount)
		if err := s.userRepo.UpdateBalanceTx(ctx, tx, userID, amount); err != nil {
			return fmt.Errorf("add available balance: %w", err)
		}
		// 记录可用余额增加交易
		balanceTx := &model.BalanceTransaction{
			UserID:        userID,
			Type:          model.TxEarningsTransfer,
			Amount:        amount,
			BalanceBefore: user.Balance,
			BalanceAfter:  newBalance,
			BalanceField:  "balance",
			Remark:        strPtr("佣金转可用余额(增加)"),
		}
		if err := s.txRepo.CreateTx(ctx, tx, balanceTx); err != nil {
			return fmt.Errorf("create balance add transaction: %w", err)
		}

		return nil
	})
}

// TransactionRecord 交易记录
type TransactionRecord struct {
	ID          int64           `json:"id"`
	Type        string          `json:"type"`
	TypeDisplay string          `json:"type_display"`
	Amount      decimal.Decimal `json:"amount"`
	Balance     decimal.Decimal `json:"balance"`
	Remark      string          `json:"remark,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// GetTransactions 获取交易历史
func (s *WalletService) GetTransactions(ctx context.Context, userID int64, page, pageSize int) ([]*TransactionRecord, int64, error) {
	query := &model.TransactionListQuery{
		UserID:   &userID,
		Page:     page,
		PageSize: pageSize,
	}

	txs, total, err := s.txRepo.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	records := make([]*TransactionRecord, len(txs))
	for i, tx := range txs {
		records[i] = &TransactionRecord{
			ID:          tx.ID,
			Type:        string(tx.Type),
			TypeDisplay: getTransactionTypeDisplay(tx.Type),
			Amount:      tx.Amount,
			Balance:     tx.BalanceAfter,
			Remark:      getTransactionRemark(tx),
			CreatedAt:   tx.CreatedAt,
		}
	}

	return records, total, nil
}

// EarningsSummary 收益统计
type EarningsSummary struct {
	TotalWinnings decimal.Decimal `json:"total_winnings"`
	TotalLosses   decimal.Decimal `json:"total_losses"`
	NetProfit     decimal.Decimal `json:"net_profit"`
	TotalRounds   int             `json:"total_rounds"`
	WinRate       float64         `json:"win_rate"`
	
	// 按时间段统计
	TodayProfit   decimal.Decimal `json:"today_profit"`
	WeekProfit    decimal.Decimal `json:"week_profit"`
	MonthProfit   decimal.Decimal `json:"month_profit"`
}

// GetEarnings 获取收益统计
func (s *WalletService) GetEarnings(ctx context.Context, userID int64) (*EarningsSummary, error) {
	summary := &EarningsSummary{}

	// 获取所有游戏相关交易
	query := &model.TransactionListQuery{
		UserID:   &userID,
		Page:     1,
		PageSize: 100000, // 获取全部
	}

	txs, _, err := s.txRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := todayStart.AddDate(0, 0, -int(now.Weekday()))
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	for _, tx := range txs {
		switch tx.Type {
		case model.TxGameWin:
			summary.TotalWinnings = summary.TotalWinnings.Add(tx.Amount)
			summary.TotalRounds++
			
			if tx.CreatedAt.After(todayStart) {
				summary.TodayProfit = summary.TodayProfit.Add(tx.Amount)
			}
			if tx.CreatedAt.After(weekStart) {
				summary.WeekProfit = summary.WeekProfit.Add(tx.Amount)
			}
			if tx.CreatedAt.After(monthStart) {
				summary.MonthProfit = summary.MonthProfit.Add(tx.Amount)
			}
			
		case model.TxGameBet:
			summary.TotalLosses = summary.TotalLosses.Add(tx.Amount.Abs())
			
			if tx.CreatedAt.After(todayStart) {
				summary.TodayProfit = summary.TodayProfit.Sub(tx.Amount.Abs())
			}
			if tx.CreatedAt.After(weekStart) {
				summary.WeekProfit = summary.WeekProfit.Sub(tx.Amount.Abs())
			}
			if tx.CreatedAt.After(monthStart) {
				summary.MonthProfit = summary.MonthProfit.Sub(tx.Amount.Abs())
			}
		}
	}

	summary.NetProfit = summary.TotalWinnings.Sub(summary.TotalLosses)
	
	// 计算胜率（基于交易记录中的获胜次数和总下注次数）
	winCount := 0
	betCount := 0
	for _, tx := range txs {
		if tx.Type == model.TxGameWin {
			winCount++
		}
		if tx.Type == model.TxGameBet {
			betCount++
		}
	}
	if betCount > 0 {
		summary.WinRate = float64(winCount) / float64(betCount) * 100
	}
	summary.TotalRounds = betCount

	return summary, nil
}

// getTransactionTypeDisplay 获取交易类型显示名称
func getTransactionTypeDisplay(txType model.TransactionType) string {
	switch txType {
	case model.TxDeposit:
		return "充值"
	case model.TxWithdraw:
		return "提现"
	case model.TxMarginDeposit:
		return "保证金充值"
	case model.TxGameBet:
		return "游戏下注"
	case model.TxGameWin:
		return "游戏获胜"
	case model.TxGameRefund:
		return "游戏退款"
	case model.TxOwnerCommission:
		return "房主佣金"
	case model.TxPlatformShare:
		return "平台抽成"
	case model.TxEarningsTransfer:
		return "佣金转余额"
	case model.TxFreeze:
		return "冻结"
	case model.TxUnfreeze:
		return "解冻"
	default:
		return string(txType)
	}
}

// getTransactionRemark 获取交易备注
func getTransactionRemark(tx *model.BalanceTransaction) string {
	if tx.RoomID != nil {
		return "房间 #" + string(rune(*tx.RoomID))
	}
	return ""
}
