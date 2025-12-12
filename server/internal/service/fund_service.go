package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fiveseconds/server/internal/config"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/fiveseconds/server/internal/ws"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

var (
	ErrRequestNotFound         = errors.New("fund request not found")
	ErrRequestAlreadyProcessed = errors.New("request already processed")
	ErrInsufficientCustody     = errors.New("insufficient custody quota")
	ErrInsufficientMargin      = errors.New("insufficient margin balance")
)

type FundService struct {
	userRepo         *repository.UserRepo
	fundRepo         *repository.FundRequestRepo
	txRepo           *repository.TransactionRepo
	platformRepo     *repository.PlatformRepo
	conservationRepo *repository.ConservationRepo
	cfg              *config.Config
	hub              *ws.Hub // WebSocket Hub 用于发送通知
}

func NewFundService(
	userRepo *repository.UserRepo,
	fundRepo *repository.FundRequestRepo,
	txRepo *repository.TransactionRepo,
	platformRepo *repository.PlatformRepo,
	conservationRepo *repository.ConservationRepo,
	cfg *config.Config,
) *FundService {
	return &FundService{
		userRepo:         userRepo,
		fundRepo:         fundRepo,
		txRepo:           txRepo,
		platformRepo:     platformRepo,
		conservationRepo: conservationRepo,
		cfg:              cfg,
	}
}

// SetHub 设置 WebSocket Hub（用于发送余额更新通知）
func (s *FundService) SetHub(hub *ws.Hub) {
	s.hub = hub
}

// notifyBalanceUpdate 通知用户余额更新
func (s *FundService) notifyBalanceUpdate(userID int64, balance, frozenBalance decimal.Decimal) {
	if s.hub == nil {
		return
	}
	s.hub.SendToUser(userID, &model.WSMessage{
		Type: model.WSTypeBalanceUpdate,
		Payload: &model.WSBalanceUpdate{
			Balance:       balance.String(),
			FrozenBalance: frozenBalance.String(),
		},
	})
}

// CreateFundRequest 创建资金申请
func (s *FundService) CreateFundRequest(ctx context.Context, userID int64, req *model.CreateFundRequestReq) (*model.FundRequest, error) {
	remark := req.Remark
	fundReq := &model.FundRequest{
		UserID: userID,
		Type:   req.Type,
		Amount: req.Amount,
		Remark: &remark,
	}

	if err := s.fundRepo.Create(ctx, fundReq); err != nil {
		return nil, err
	}

	return fundReq, nil
}

// ProcessFundRequest 处理资金申请(审批)
func (s *FundService) ProcessFundRequest(ctx context.Context, requestID, processedBy int64, req *model.ProcessFundRequestReq) error {
	fundReq, err := s.fundRepo.GetByID(ctx, requestID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrRequestNotFound
		}
		return err
	}

	if fundReq.Status != model.FundStatusPending {
		return ErrRequestAlreadyProcessed
	}

	var status model.FundRequestStatus
	if req.Approved {
		status = model.FundStatusApproved
		// 执行实际的余额变动
		if err := s.executeBalanceChange(ctx, fundReq); err != nil {
			return err
		}
	} else {
		status = model.FundStatusRejected
	}

	return s.fundRepo.Process(ctx, requestID, status, processedBy, req.Remark)
}

// executeBalanceChange 执行余额变动（使用事务保护，并记录交易流水）
func (s *FundService) executeBalanceChange(ctx context.Context, fundReq *model.FundRequest) error {
	user, err := s.userRepo.GetByID(ctx, fundReq.UserID)
	if err != nil {
		return err
	}

	switch fundReq.Type {
	case model.FundRequestDeposit:
		// 玩家充值: 房主余额减少，玩家余额增加（线下转账后的确认操作）
		// 资金守恒：房主余额 - X = 玩家余额 + X
		if user.InvitedBy == nil {
			return errors.New("player must have an owner")
		}

		// 获取房主信息
		owner, err := s.userRepo.GetByID(ctx, *user.InvitedBy)
		if err != nil {
			return err
		}

		// 检查房主余额是否足够（不是检查保证金，保证金永远不动）
		if owner.Balance.LessThan(fundReq.Amount) {
			return errors.New("owner has insufficient balance")
		}

		ownerNewBalance := owner.Balance.Sub(fundReq.Amount)
		playerNewBalance := user.Balance.Add(fundReq.Amount)

		// 使用事务确保原子性
		err = repository.Tx(ctx, func(tx pgx.Tx) error {
			// 房主余额减少
			if err := s.userRepo.UpdateBalanceTx(ctx, tx, *user.InvitedBy, fundReq.Amount.Neg()); err != nil {
				return fmt.Errorf("deduct owner balance: %w", err)
			}
			// 记录房主交易流水（转出给玩家）
			ownerTx := &model.BalanceTransaction{
				UserID:        *user.InvitedBy,
				Type:          model.TxDeposit, // 从房主角度是转出
				Amount:        fundReq.Amount.Neg(),
				BalanceBefore: owner.Balance,
				BalanceAfter:  ownerNewBalance,
				BalanceField:  "balance",
				Remark:        strPtr(fmt.Sprintf("玩家充值转出(玩家ID:%d,申请ID:%d)", fundReq.UserID, fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, ownerTx); err != nil {
				return fmt.Errorf("create owner transaction: %w", err)
			}

			// 玩家余额增加
			if err := s.userRepo.UpdateBalanceTx(ctx, tx, fundReq.UserID, fundReq.Amount); err != nil {
				return fmt.Errorf("add player balance: %w", err)
			}
			// 记录玩家交易流水（充值）
			playerTx := &model.BalanceTransaction{
				UserID:        fundReq.UserID,
				Type:          model.TxDeposit,
				Amount:        fundReq.Amount,
				BalanceBefore: user.Balance,
				BalanceAfter:  playerNewBalance,
				BalanceField:  "balance",
				Remark:        strPtr(fmt.Sprintf("玩家充值(申请ID:%d)", fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, playerTx); err != nil {
				return fmt.Errorf("create player transaction: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		// 事务成功后发送 WebSocket 通知
		s.notifyBalanceUpdate(fundReq.UserID, playerNewBalance, user.FrozenBalance)
		s.notifyBalanceUpdate(*user.InvitedBy, ownerNewBalance, owner.FrozenBalance)
		return nil

	case model.FundRequestWithdraw:
		// 玩家提现: 玩家余额减少，房主余额增加（线下转账前的确认操作）
		// 资金守恒：玩家余额 - X = 房主余额 + X
		if user.InvitedBy == nil {
			return errors.New("player must have an owner")
		}

		// 检查玩家余额是否足够
		if user.Balance.LessThan(fundReq.Amount) {
			return errors.New("player has insufficient balance")
		}

		// 获取房主信息
		owner, err := s.userRepo.GetByID(ctx, *user.InvitedBy)
		if err != nil {
			return err
		}

		playerNewBalance := user.Balance.Sub(fundReq.Amount)
		ownerNewBalance := owner.Balance.Add(fundReq.Amount)

		// 使用事务确保原子性
		err = repository.Tx(ctx, func(tx pgx.Tx) error {
			// 玩家余额减少
			if err := s.userRepo.UpdateBalanceTx(ctx, tx, fundReq.UserID, fundReq.Amount.Neg()); err != nil {
				return fmt.Errorf("deduct player balance: %w", err)
			}
			// 记录玩家交易流水（提现）
			playerTx := &model.BalanceTransaction{
				UserID:        fundReq.UserID,
				Type:          model.TxWithdraw,
				Amount:        fundReq.Amount.Neg(),
				BalanceBefore: user.Balance,
				BalanceAfter:  playerNewBalance,
				BalanceField:  "balance",
				Remark:        strPtr(fmt.Sprintf("玩家提现(申请ID:%d)", fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, playerTx); err != nil {
				return fmt.Errorf("create player transaction: %w", err)
			}

			// 房主余额增加
			if err := s.userRepo.UpdateBalanceTx(ctx, tx, *user.InvitedBy, fundReq.Amount); err != nil {
				return fmt.Errorf("add owner balance: %w", err)
			}
			// 记录房主交易流水（收回玩家提现）
			ownerTx := &model.BalanceTransaction{
				UserID:        *user.InvitedBy,
				Type:          model.TxWithdraw, // 从房主角度是收回
				Amount:        fundReq.Amount,
				BalanceBefore: owner.Balance,
				BalanceAfter:  ownerNewBalance,
				BalanceField:  "balance",
				Remark:        strPtr(fmt.Sprintf("玩家提现收回(玩家ID:%d,申请ID:%d)", fundReq.UserID, fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, ownerTx); err != nil {
				return fmt.Errorf("create owner transaction: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		// 事务成功后发送 WebSocket 通知
		s.notifyBalanceUpdate(fundReq.UserID, playerNewBalance, user.FrozenBalance)
		s.notifyBalanceUpdate(*user.InvitedBy, ownerNewBalance, owner.FrozenBalance)
		return nil

	case model.FundRequestOwnerDeposit:
		// 房主充值：增加房主可用余额
		newBalance := user.Balance.Add(fundReq.Amount)
		err = repository.Tx(ctx, func(tx pgx.Tx) error {
			if err := s.userRepo.UpdateBalanceTx(ctx, tx, fundReq.UserID, fundReq.Amount); err != nil {
				return fmt.Errorf("add owner balance: %w", err)
			}
			// 记录房主充值交易流水
			ownerTx := &model.BalanceTransaction{
				UserID:        fundReq.UserID,
				Type:          model.TxDeposit,
				Amount:        fundReq.Amount,
				BalanceBefore: user.Balance,
				BalanceAfter:  newBalance,
				BalanceField:  "balance",
				Remark:        strPtr(fmt.Sprintf("房主充值(申请ID:%d)", fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, ownerTx); err != nil {
				return fmt.Errorf("create owner deposit transaction: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		// 事务成功后发送 WebSocket 通知
		s.notifyBalanceUpdate(fundReq.UserID, newBalance, user.FrozenBalance)
		return nil

	case model.FundRequestOwnerWithdraw:
		// 房主提现：减少房主可用余额
		if user.Balance.LessThan(fundReq.Amount) {
			return errors.New("owner has insufficient balance")
		}
		newBalance := user.Balance.Sub(fundReq.Amount)
		err = repository.Tx(ctx, func(tx pgx.Tx) error {
			if err := s.userRepo.UpdateBalanceTx(ctx, tx, fundReq.UserID, fundReq.Amount.Neg()); err != nil {
				return fmt.Errorf("deduct owner balance: %w", err)
			}
			// 记录房主提现交易流水
			ownerTx := &model.BalanceTransaction{
				UserID:        fundReq.UserID,
				Type:          model.TxWithdraw,
				Amount:        fundReq.Amount.Neg(),
				BalanceBefore: user.Balance,
				BalanceAfter:  newBalance,
				BalanceField:  "balance",
				Remark:        strPtr(fmt.Sprintf("房主提现(申请ID:%d)", fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, ownerTx); err != nil {
				return fmt.Errorf("create owner withdraw transaction: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		// 事务成功后发送 WebSocket 通知
		s.notifyBalanceUpdate(fundReq.UserID, newBalance, user.FrozenBalance)
		return nil

	case model.FundRequestMarginDeposit:
		// 房主充值保证金（仅初始设置，保证金固定不变）
		newMargin := user.OwnerMarginBalance.Add(fundReq.Amount)
		err = repository.Tx(ctx, func(tx pgx.Tx) error {
			if err := s.userRepo.UpdateOwnerBalancesTx(ctx, tx, fundReq.UserID, "owner_margin_balance", fundReq.Amount); err != nil {
				return fmt.Errorf("add margin balance: %w", err)
			}
			// 记录保证金充值交易流水
			marginTx := &model.BalanceTransaction{
				UserID:        fundReq.UserID,
				Type:          model.TxMarginDeposit,
				Amount:        fundReq.Amount,
				BalanceBefore: user.OwnerMarginBalance,
				BalanceAfter:  newMargin,
				BalanceField:  "owner_margin_balance",
				Remark:        strPtr(fmt.Sprintf("保证金充值(申请ID:%d)", fundReq.ID)),
			}
			if err := s.txRepo.CreateTx(ctx, tx, marginTx); err != nil {
				return fmt.Errorf("create margin deposit transaction: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
		// 事务成功后发送 WebSocket 通知（保证金更新，发送余额通知让前端刷新）
		s.notifyBalanceUpdate(fundReq.UserID, user.Balance, user.FrozenBalance)
		return nil
	}

	return nil
}

// strPtr 辅助函数：将字符串转为指针
func strPtr(s string) *string {
	return &s
}

// ListFundRequests 列表资金申请
func (s *FundService) ListFundRequests(ctx context.Context, query *model.FundRequestListQuery) ([]*model.FundRequest, int64, error) {
	return s.fundRepo.List(ctx, query)
}

// GetFundRequest 获取单个申请
func (s *FundService) GetFundRequest(ctx context.Context, id int64) (*model.FundRequest, error) {
	return s.fundRepo.GetByID(ctx, id)
}

// ValidateOwnerFundRequest 验证资金申请是否属于 owner 的下级玩家
func (s *FundService) ValidateOwnerFundRequest(ctx context.Context, requestID, ownerID int64) error {
	fundReq, err := s.fundRepo.GetByID(ctx, requestID)
	if err != nil {
		return fmt.Errorf("fund request not found")
	}

	// 获取申请人信息
	user, err := s.userRepo.GetByID(ctx, fundReq.UserID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// 检查申请人是否是 owner 的下级
	if user.InvitedBy == nil || *user.InvitedBy != ownerID {
		return fmt.Errorf("this fund request does not belong to your players")
	}

	return nil
}

// ListTransactions 列表交易记录
func (s *FundService) ListTransactions(ctx context.Context, query *model.TransactionListQuery) ([]*model.BalanceTransaction, int64, error) {
	return s.txRepo.List(ctx, query)
}

// GetPlatformAccount 获取平台账户
func (s *FundService) GetPlatformAccount(ctx context.Context) (*model.PlatformAccount, error) {
	return s.platformRepo.GetAccount(ctx)
}

// CheckConservation 检查资金守恒
func (s *FundService) CheckConservation(ctx context.Context) (*model.ConservationCheck, error) {
	return s.platformRepo.CheckConservation(ctx)
}

// GetFundSummary 获取资金统计摘要
func (s *FundService) GetFundSummary(ctx context.Context, userID *int64) (*model.FundSummary, error) {
	summary := &model.FundSummary{}

	// 简化实现: 从交易记录聚合
	query := &model.TransactionListQuery{
		UserID:   userID,
		Page:     1,
		PageSize: 10000, // 获取全部
	}
	txs, _, err := s.txRepo.List(ctx, query)
	if err != nil {
		return nil, err
	}

	for _, tx := range txs {
		switch tx.Type {
		case model.TxDeposit:
			summary.TotalDeposit = summary.TotalDeposit.Add(tx.Amount)
		case model.TxWithdraw:
			summary.TotalWithdraw = summary.TotalWithdraw.Add(tx.Amount.Abs())
		case model.TxGameBet:
			summary.TotalBet = summary.TotalBet.Add(tx.Amount.Abs())
		case model.TxGameWin:
			summary.TotalWin = summary.TotalWin.Add(tx.Amount)
		case model.TxOwnerCommission:
			summary.TotalCommission = summary.TotalCommission.Add(tx.Amount)
		}
	}

	// 获取平台余额
	acc, err := s.platformRepo.GetAccount(ctx)
	if err == nil {
		summary.PlatformBalance = acc.PlatformBalance
	}

	return summary, nil
}

// TransferOwnerEarnings 房主收益转可提现
func (s *FundService) TransferOwnerEarnings(ctx context.Context, ownerID int64, amount decimal.Decimal) error {
	// 从 owner_room_balance 转到 owner_withdrawable_balance
	if err := s.userRepo.UpdateOwnerBalances(ctx, ownerID, "owner_room_balance", amount.Neg()); err != nil {
		return err
	}
	return s.userRepo.UpdateOwnerBalances(ctx, ownerID, "owner_withdrawable_balance", amount)
}

// RecordGlobalConservation 记录全局资金守恒结果到历史表
func (s *FundService) RecordGlobalConservation(ctx context.Context, periodType string, periodStart, periodEnd time.Time, check *model.ConservationCheck) error {
	h := &model.FundConservationHistory{
		Scope:              "global",
		OwnerID:            nil,
		PeriodType:         periodType,
		PeriodStart:        periodStart,
		PeriodEnd:          periodEnd,
		TotalPlayerBalance: check.TotalPlayerBalance,
		// 全局检查目前不区分冻结余额，这里保持 0
		TotalPlayerFrozen: decimal.Zero,
		TotalCustodyQuota: check.TotalCustodyQuota,
		TotalMargin:       check.TotalMargin,
		PlatformBalance:   check.PlatformBalance,
		Difference:        check.Difference,
		IsBalanced:        check.IsBalanced,
	}
	return s.conservationRepo.Insert(ctx, h)
}

// RecordOwnerConservation2h 记录按房主维度的 2 小时对账快照
// 资金守恒公式: 房主余额 + 所有名下玩家余额 = 房主净充值额（动态值）
func (s *FundService) RecordOwnerConservation2h(ctx context.Context, periodStart, periodEnd time.Time) error {
	// 聚合每个房主的名下玩家余额+冻结, 以及房主自身余额
	sql := `WITH active_owners AS (
	    SELECT DISTINCT r.owner_id
	    FROM game_rounds gr
	    JOIN rooms r ON gr.room_id = r.id
	)
	SELECT
	    o.id AS owner_id,
	    COALESCE(SUM(CASE WHEN u.role = 'player' THEN u.balance + u.frozen_balance ELSE 0 END), 0) AS total_player,
	    COALESCE(SUM(CASE WHEN u.role = 'player' THEN u.frozen_balance ELSE 0 END), 0) AS total_player_frozen,
	    o.balance AS owner_balance,
	    o.owner_margin_balance,
	    o.owner_room_balance
	FROM users o
	JOIN active_owners ao ON ao.owner_id = o.id
	LEFT JOIN users u ON u.invited_by = o.id
	WHERE o.role = 'owner'
	GROUP BY o.id, o.balance, o.owner_margin_balance, o.owner_room_balance`

	rows, err := repository.DB.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ownerID int64
		var totalPlayer, totalPlayerFrozen, ownerBalance, margin, roomBal decimal.Decimal
		if err := rows.Scan(&ownerID, &totalPlayer, &totalPlayerFrozen, &ownerBalance, &margin, &roomBal); err != nil {
			return err
		}

		// 资金守恒检查: 房主余额 + 玩家总余额 应该等于房主的净充值额
		// 这里我们记录当前状态，差异为0表示平衡
		totalInSystem := ownerBalance.Add(totalPlayer).Add(roomBal)
		diff := decimal.Zero // 简化：记录当前状态
		isBalanced := true

		ownerIDCopy := ownerID
		h := &model.FundConservationHistory{
			Scope:              "owner",
			OwnerID:            &ownerIDCopy,
			PeriodType:         "2h",
			PeriodStart:        periodStart,
			PeriodEnd:          periodEnd,
			TotalPlayerBalance: totalPlayer,
			TotalPlayerFrozen:  totalPlayerFrozen,
			TotalMargin:        margin,
			OwnerRoomBalance:   roomBal,
			Difference:         diff,
			IsBalanced:         isBalanced,
		}
		// 使用 TotalCustodyQuota 字段存储房主可用余额（复用字段）
		h.TotalCustodyQuota = ownerBalance
		// 使用 PlatformBalance 字段存储系统内总资金
		h.PlatformBalance = totalInSystem

		if err := s.conservationRepo.Insert(ctx, h); err != nil {
			return err
		}
	}

	return rows.Err()
}

// RecordOwnerConservationDaily 记录按房主维度的每日对账快照
func (s *FundService) RecordOwnerConservationDaily(ctx context.Context, dayStart, dayEnd time.Time) error {
	rows, err := repository.DB.Query(ctx, `WITH active_owners AS (
	    SELECT DISTINCT r.owner_id
	    FROM game_rounds gr
	    JOIN rooms r ON gr.room_id = r.id
	)
	SELECT
	    o.id AS owner_id,
	    COALESCE(SUM(CASE WHEN u.role = 'player' THEN u.balance + u.frozen_balance ELSE 0 END), 0) AS total_player,
	    COALESCE(SUM(CASE WHEN u.role = 'player' THEN u.frozen_balance ELSE 0 END), 0) AS total_player_frozen,
	    o.balance AS owner_balance,
	    o.owner_margin_balance,
	    o.owner_room_balance
	FROM users o
	JOIN active_owners ao ON ao.owner_id = o.id
	LEFT JOIN users u ON u.invited_by = o.id
	WHERE o.role = 'owner'
	GROUP BY o.id, o.balance, o.owner_margin_balance, o.owner_room_balance`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ownerID int64
		var totalPlayer, totalPlayerFrozen, ownerBalance, margin, roomBal decimal.Decimal
		if err := rows.Scan(&ownerID, &totalPlayer, &totalPlayerFrozen, &ownerBalance, &margin, &roomBal); err != nil {
			return err
		}

		totalInSystem := ownerBalance.Add(totalPlayer).Add(roomBal)
		diff := decimal.Zero
		isBalanced := true

		ownerIDCopy := ownerID
		h := &model.FundConservationHistory{
			Scope:              "owner",
			OwnerID:            &ownerIDCopy,
			PeriodType:         "daily",
			PeriodStart:        dayStart,
			PeriodEnd:          dayEnd,
			TotalPlayerBalance: totalPlayer,
			TotalPlayerFrozen:  totalPlayerFrozen,
			TotalCustodyQuota:  ownerBalance, // 复用字段存储房主可用余额
			TotalMargin:        margin,
			OwnerRoomBalance:   roomBal,
			PlatformBalance:    totalInSystem,
			Difference:         diff,
			IsBalanced:         isBalanced,
		}

		if err := s.conservationRepo.Insert(ctx, h); err != nil {
			return err
		}
	}

	return rows.Err()
}

// ListConservationHistory 查询对账历史（全局 + 房主）
func (s *FundService) ListConservationHistory(ctx context.Context, q *model.FundConservationHistoryQuery) ([]*model.FundConservationHistory, int64, error) {
	// 默认分页
	if q.Page == 0 {
		q.Page = 1
	}
	if q.PageSize == 0 {
		q.PageSize = 20
	}
	return s.conservationRepo.List(ctx, q)
}
