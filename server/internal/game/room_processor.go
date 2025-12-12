package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fiveseconds/server/internal/cache"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	PhaseDuration        = 5 * time.Second   // 每个阶段5秒
	DefaultMinPlayers    = 2                 // 默认最少参与人数（实际使用 winner_count + 1）
	ActiveTickInterval   = 1 * time.Second   // 活跃阶段 tick 间隔
	WaitingTickInterval  = 3 * time.Second   // 等待阶段 tick 间隔
	OfflineTimeout       = 2 * time.Minute   // 离线超时时间，超过后自动移除玩家
	OfflineCheckInterval = 30 * time.Second  // 离线检查间隔
)

// getMinPlayers 获取最小参与人数（winner_count + 1，至少有一个输家）
func (rp *RoomProcessor) getMinPlayers() int {
	minPlayers := rp.Room.WinnerCount + 1
	if minPlayers < DefaultMinPlayers {
		minPlayers = DefaultMinPlayers
	}
	return minPlayers
}

// RiskChecker 风控检查接口
type RiskChecker interface {
	OnRoundSettled(ctx context.Context, participants []int64, winners []int64)
}

// RoomProcessor 房间游戏处理器
type RoomProcessor struct {
	mu sync.RWMutex

	RoomID      int64
	Room        *model.Room
	State       *model.RoomState
	Broadcaster Broadcaster

	userRepo     *repository.UserRepo
	roomRepo     *repository.RoomRepo
	gameRepo     *repository.GameRepo
	txRepo       *repository.TransactionRepo
	platformRepo *repository.PlatformRepo
	balanceCache *cache.BalanceCache
	riskChecker  RiskChecker
	commitReveal *CommitReveal
	logger       *zap.Logger

	stopCh       chan struct{}
	ticker       *time.Ticker
	phaseTicker  *time.Ticker
	lastTickState *tickState // 上次 tick 的状态快照，用于增量比较
}

// Broadcaster 广播接口
type Broadcaster interface {
	BroadcastToRoom(roomID int64, msg *model.WSMessage)
	SendToUser(userID int64, msg *model.WSMessage)
}

// NewRoomProcessor 创建房间处理器
func NewRoomProcessor(
	room *model.Room,
	broadcaster Broadcaster,
	userRepo *repository.UserRepo,
	roomRepo *repository.RoomRepo,
	gameRepo *repository.GameRepo,
	txRepo *repository.TransactionRepo,
	platformRepo *repository.PlatformRepo,
	balanceCache *cache.BalanceCache,
	riskChecker RiskChecker,
	logger *zap.Logger,
) *RoomProcessor {
	return &RoomProcessor{
		RoomID:       room.ID,
		Room:         room,
		Broadcaster:  broadcaster,
		userRepo:     userRepo,
		roomRepo:     roomRepo,
		gameRepo:     gameRepo,
		txRepo:       txRepo,
		platformRepo: platformRepo,
		balanceCache: balanceCache,
		riskChecker:  riskChecker,
		commitReveal: NewCommitReveal(),
		logger:       logger.With(zap.Int64("room_id", room.ID)),
		State: &model.RoomState{
			Phase:        model.PhaseWaiting,
			PhaseEndTime: time.Now(),
			CurrentRound: 0,
			Players:      make(map[int64]*model.PlayerState),
			Spectators:   make(map[int64]*model.SpectatorState),
			PoolAmount:   decimal.Zero,
		},
		stopCh: make(chan struct{}),
	}
}

// tickState 用于增量比较的状态快照
type tickState struct {
	Phase          model.GamePhase
	PoolAmount     string
	PlayerCount    int
	SpectatorCount int
}

// Start 启动处理器
func (rp *RoomProcessor) Start() {
	rp.ticker = time.NewTicker(100 * time.Millisecond)
	rp.phaseTicker = time.NewTicker(WaitingTickInterval) // 初始为等待阶段间隔
	go rp.loop()
	go rp.phaseTickLoop()
	go rp.offlineCheckLoop() // 离线超时检查
	rp.logger.Info("Room processor started")
}

// Stop 停止处理器
func (rp *RoomProcessor) Stop() {
	close(rp.stopCh)
	if rp.ticker != nil {
		rp.ticker.Stop()
	}
	if rp.phaseTicker != nil {
		rp.phaseTicker.Stop()
	}
	rp.logger.Info("Room processor stopped")
}

// loop 主循环
func (rp *RoomProcessor) loop() {
	defer func() {
		if r := recover(); r != nil {
			rp.logger.Error("Room processor panic recovered", zap.Any("panic", r), zap.Stack("stack"))
		}
	}()
	
	for {
		select {
		case <-rp.stopCh:
			return
		case <-rp.ticker.C:
			rp.safeTick()
		}
	}
}

// safeTick 安全的 tick，捕获 panic
func (rp *RoomProcessor) safeTick() {
	defer func() {
		if r := recover(); r != nil {
			rp.logger.Error("Tick panic recovered", zap.Any("panic", r), zap.Stack("stack"))
			// 尝试恢复到等待状态
			rp.mu.Lock()
			rp.State.Phase = model.PhaseWaiting
			rp.State.PhaseEndTime = time.Now()
			rp.mu.Unlock()
		}
	}()
	rp.tick()
}

// phaseTickLoop 阶段 tick 循环（用于增量状态广播）
func (rp *RoomProcessor) phaseTickLoop() {
	defer func() {
		if r := recover(); r != nil {
			rp.logger.Error("Phase tick loop panic recovered", zap.Any("panic", r), zap.Stack("stack"))
		}
	}()
	
	for {
		select {
		case <-rp.stopCh:
			return
		case <-rp.phaseTicker.C:
			rp.safeSendPhaseTick()
		}
	}
}

// offlineCheckLoop 离线超时检查循环
func (rp *RoomProcessor) offlineCheckLoop() {
	ticker := time.NewTicker(OfflineCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rp.stopCh:
			return
		case <-ticker.C:
			rp.checkOfflinePlayers()
		}
	}
}

// checkOfflinePlayers 检查并移除超时离线的玩家
func (rp *RoomProcessor) checkOfflinePlayers() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	now := time.Now()
	var toRemove []int64

	for userID, p := range rp.State.Players {
		if !p.IsOnline && p.OfflineSince != nil {
			if now.Sub(*p.OfflineSince) > OfflineTimeout {
				toRemove = append(toRemove, userID)
			}
		}
	}

	// 移除超时的玩家
	for _, userID := range toRemove {
		delete(rp.State.Players, userID)

		// 从数据库移除
		if rp.roomRepo != nil {
			ctx := context.Background()
			if err := rp.roomRepo.RemovePlayer(ctx, rp.RoomID, userID); err != nil {
				rp.logger.Warn("Failed to remove offline player from DB", zap.Int64("user_id", userID), zap.Error(err))
			}
		}

		// 广播玩家离开
		rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
			Type: model.WSTypePlayerLeave,
			Payload: &model.WSPlayerLeave{
				UserID: userID,
			},
		})

		rp.logger.Info("Removed offline player due to timeout", zap.Int64("user_id", userID))
	}
}

// safeSendPhaseTick 安全的发送阶段 tick
func (rp *RoomProcessor) safeSendPhaseTick() {
	defer func() {
		if r := recover(); r != nil {
			rp.logger.Error("Send phase tick panic recovered", zap.Any("panic", r), zap.Stack("stack"))
		}
	}()
	rp.sendPhaseTick()
}

// sendPhaseTick 发送状态更新
// 始终包含阶段和奖池信息，确保客户端状态同步
func (rp *RoomProcessor) sendPhaseTick() {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	now := time.Now()
	currentState := &tickState{
		Phase:          rp.State.Phase,
		PoolAmount:     rp.State.PoolAmount.String(),
		PlayerCount:    len(rp.State.Players),
		SpectatorCount: len(rp.State.Spectators),
	}

	// 检查是否有状态变化
	hasChanges := rp.lastTickState == nil ||
		rp.lastTickState.Phase != currentState.Phase ||
		rp.lastTickState.PoolAmount != currentState.PoolAmount ||
		rp.lastTickState.PlayerCount != currentState.PlayerCount ||
		rp.lastTickState.SpectatorCount != currentState.SpectatorCount

	// 如果没有变化且处于等待阶段，跳过广播
	if !hasChanges && rp.State.Phase == model.PhaseWaiting {
		return
	}

	// 始终包含阶段和奖池信息，确保客户端状态同步
	phase := string(currentState.Phase)
	poolAmount := currentState.PoolAmount
	playerCount := currentState.PlayerCount
	spectatorCount := currentState.SpectatorCount

	tick := &model.WSPhaseTick{
		ServerTime:     now.UnixMilli(),
		PhaseEndTime:   rp.State.PhaseEndTime.UnixMilli(),
		TimeRemaining:  rp.State.PhaseEndTime.Sub(now).Milliseconds(),
		Phase:          &phase,
		PoolAmount:     &poolAmount,
		PlayerCount:    &playerCount,
		SpectatorCount: &spectatorCount,
	}

	// 广播
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type:    model.WSTypePhaseTick,
		Payload: tick,
	})

	// 更新上次状态
	rp.lastTickState = currentState
}

// adjustTickInterval 根据阶段调整 tick 间隔
func (rp *RoomProcessor) adjustTickInterval() {
	if rp.phaseTicker == nil {
		return
	}

	var interval time.Duration
	if rp.State.Phase == model.PhaseWaiting {
		interval = WaitingTickInterval
	} else {
		interval = ActiveTickInterval
	}

	rp.phaseTicker.Reset(interval)
}

// tick 每个时钟周期检查状态转换
func (rp *RoomProcessor) tick() {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	now := time.Now()
	if now.Before(rp.State.PhaseEndTime) {
		return
	}

	// 阶段转换
	switch rp.State.Phase {
	case model.PhaseWaiting:
		rp.tryStartCountdown()
	case model.PhaseCountdown:
		rp.enterBetting()
	case model.PhaseBetting:
		rp.enterInGame()
	case model.PhaseInGame:
		rp.enterSettlement()
	case model.PhaseSettlement:
		rp.enterReset()
	case model.PhaseReset:
		rp.enterWaiting()
	}
}

// tryStartCountdown 尝试开始倒计时
func (rp *RoomProcessor) tryStartCountdown() {
	// 检查在线且设置自动准备的玩家数量
	readyCount := 0
	for _, p := range rp.State.Players {
		if p.IsOnline && p.AutoReady {
			readyCount++
		}
	}

	if readyCount >= rp.getMinPlayers() {
		rp.State.Phase = model.PhaseCountdown
		rp.State.PhaseEndTime = time.Now().Add(PhaseDuration)
		rp.State.CurrentRound++
		rp.broadcastPhaseChange()
		rp.logger.Info("Phase changed", zap.String("phase", "countdown"), zap.Int("round", rp.State.CurrentRound))
	}
}

// enterBetting 进入下注阶段(自动扣款)
func (rp *RoomProcessor) enterBetting() {
	ctx := context.Background()

	// 生成 commit
	seed, commitHash, err := rp.commitReveal.GenerateCommit()
	if err != nil {
		rp.logger.Error("Generate commit failed", zap.Error(err))
		rp.enterWaiting()
		return
	}
	rp.State.Seed = seed
	rp.State.CommitHash = commitHash

	// 先检查哪些玩家可以参与（余额足够且在线且准备）
	betAmount := rp.Room.BetAmount
	eligiblePlayers := []int64{}
	skipped := []int64{}
	disqualifiedPlayers := []model.WSPlayerDisqualified{}

	for userID, p := range rp.State.Players {
		if !p.IsOnline {
			skipped = append(skipped, userID)
			continue
		}
		if !p.AutoReady {
			skipped = append(skipped, userID)
			continue
		}
		if p.Balance.LessThan(betAmount) {
			skipped = append(skipped, userID)
			// 标记玩家被取消资格
			p.Disqualified = true
			p.DisqualifyReason = "insufficient_balance"
			disqualifiedPlayers = append(disqualifiedPlayers, model.WSPlayerDisqualified{
				UserID:   userID,
				Username: p.Username,
				Reason:   "insufficient_balance",
			})
			// 发送个人通知给被取消资格的玩家
			rp.Broadcaster.SendToUser(userID, &model.WSMessage{
				Type: model.WSTypePlayerDisqualified,
				Payload: &model.WSPlayerDisqualified{
					UserID:   userID,
					Username: p.Username,
					Reason:   "insufficient_balance",
				},
			})
			// 广播玩家状态更新（让其他玩家看到）
			disqualified := true
			reason := "insufficient_balance"
			rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
				Type: model.WSTypePlayerUpdate,
				Payload: &model.WSPlayerUpdate{
					UserID:           userID,
					Disqualified:     &disqualified,
					DisqualifyReason: &reason,
				},
			})
			rp.logger.Info("Player disqualified due to insufficient balance",
				zap.Int64("user_id", userID),
				zap.String("balance", p.Balance.String()),
				zap.String("bet_amount", betAmount.String()))
			continue
		}
		eligiblePlayers = append(eligiblePlayers, userID)
	}

	// 检查参与人数（必须大于 winner_count，至少有一个输家）
	minPlayers := rp.getMinPlayers()
	if len(eligiblePlayers) < minPlayers {
		// 人数不足，广播回合取消消息（包含被取消资格的玩家列表）
		rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
			Type: model.WSTypeRoundCancelled,
			Payload: &model.WSRoundCancelled{
				Reason:              "not_enough_players",
				DisqualifiedPlayers: disqualifiedPlayers,
				MinPlayersRequired:  minPlayers,
				CurrentPlayers:      len(eligiblePlayers),
			},
		})
		// 重置被取消资格玩家的状态（下一轮可以重新参与）
		rp.resetDisqualifiedPlayers()
		rp.enterWaiting()
		return
	}

	// 使用事务执行批量扣款操作（优化：单条 SQL）
	participants := []int64{}
	poolAmount := decimal.Zero
	playerNewBalances := make(map[int64]decimal.Decimal)
	playerOldBalances := make(map[int64]decimal.Decimal)

	// 记录扣款前的余额
	for _, userID := range eligiblePlayers {
		if p := rp.State.Players[userID]; p != nil {
			playerOldBalances[userID] = p.Balance
		}
	}

	err = repository.Tx(ctx, func(tx pgx.Tx) error {
		// 批量扣款（单条 SQL）
		deductResults, err := rp.userRepo.BatchDeductBalanceTx(ctx, tx, eligiblePlayers, betAmount)
		if err != nil {
			return fmt.Errorf("batch deduct balance: %w", err)
		}

		// 处理扣款结果
		for _, result := range deductResults {
			participants = append(participants, result.UserID)
			playerNewBalances[result.UserID] = result.NewBalance
			poolAmount = poolAmount.Add(betAmount)
		}

		// 检查参与人数（必须大于 winner_count）
		if len(participants) < rp.getMinPlayers() {
			return fmt.Errorf("not enough participants after deduction: need %d, got %d", rp.getMinPlayers(), len(participants))
		}

		// 批量创建交易记录（单条 SQL）
		txRecords := make([]*model.BalanceTransaction, 0, len(participants))
		for _, userID := range participants {
			oldBalance := playerOldBalances[userID]
			newBalance := playerNewBalances[userID]
			txRecords = append(txRecords, &model.BalanceTransaction{
				UserID:        userID,
				RoomID:        &rp.RoomID,
				Type:          model.TxGameBet,
				Amount:        betAmount.Neg(),
				BalanceBefore: oldBalance,
				BalanceAfter:  newBalance,
			})
		}
		if err := rp.txRepo.BatchCreateTx(ctx, tx, txRecords); err != nil {
			return fmt.Errorf("batch create bet transactions: %w", err)
		}

		return nil
	})

	if err != nil {
		rp.logger.Error("Betting transaction failed", zap.Error(err))
		// 事务失败会自动回滚，无需手动退款
		rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
			Type: model.WSTypeRoundFailed,
			Payload: &model.WSRoundFailed{
				Reason:   "betting_failed",
				Refunded: []int64{},
			},
		})
		rp.enterWaiting()
		return
	}

	// 事务成功后更新内存状态
	for userID, newBalance := range playerNewBalances {
		if p := rp.State.Players[userID]; p != nil {
			p.Balance = newBalance
		}
		// 使缓存失效
		if rp.balanceCache != nil {
			if err := rp.balanceCache.Invalidate(ctx, userID); err != nil {
				rp.logger.Warn("Failed to invalidate balance cache", zap.Int64("user_id", userID), zap.Error(err))
			}
		}
	}

	// 更新跳过列表（包括余额不足的）
	for _, userID := range eligiblePlayers {
		if _, ok := playerNewBalances[userID]; !ok {
			skipped = append(skipped, userID)
		}
	}

	rp.State.Participants = participants
	rp.State.SkippedPlayers = skipped
	rp.State.PoolAmount = poolAmount

	// 创建回合记录
	lastNum, _ := rp.gameRepo.GetLastRoundNumber(ctx, rp.RoomID)
	round := &model.GameRound{
		RoomID:         rp.RoomID,
		RoundNumber:    lastNum + 1,
		ParticipantIDs: participants,
		SkippedIDs:     skipped,
		BetAmount:      rp.Room.BetAmount,
		PoolAmount:     poolAmount,
		CommitHash:     &commitHash,
		Status:         model.RoundStatusBetting,
	}
	if err := rp.gameRepo.CreateRound(ctx, round); err != nil {
		rp.logger.Error("Create round failed", zap.Error(err))
		// 创建回合失败，需要退款并返回等待状态
		rp.refundAndWait(ctx, participants)
		return
	}
	rp.State.RoundID = round.ID

	rp.State.Phase = model.PhaseBetting
	rp.State.PhaseEndTime = time.Now().Add(PhaseDuration)
	rp.broadcastPhaseChange()

	// 广播下注完成信息
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeBettingDone,
		Payload: &model.WSBettingDone{
			PoolAmount:   poolAmount.String(),
			Participants: participants,
			Skipped:      skipped,
		},
	})

	rp.logger.Info("Phase changed", zap.String("phase", "betting"),
		zap.Int("participants", len(participants)), zap.String("pool", poolAmount.String()))
}

// enterInGame 进入游戏中阶段
func (rp *RoomProcessor) enterInGame() {
	rp.State.Phase = model.PhaseInGame
	rp.State.PhaseEndTime = time.Now().Add(PhaseDuration)
	rp.broadcastPhaseChange()
	rp.logger.Info("Phase changed", zap.String("phase", "in_game"))
}

// enterSettlement 进入结算阶段
func (rp *RoomProcessor) enterSettlement() {
	ctx := context.Background()

	// 使用 commit-reveal 选择赢家
	winners := rp.commitReveal.SelectWinners(rp.State.Participants, rp.Room.WinnerCount, rp.State.Seed)
	revealSeed := rp.commitReveal.Reveal(rp.State.Seed)

	// 计算抽成
	// 注意：费率已经是小数形式，如 0.03 表示 3%
	poolAmount := rp.State.PoolAmount
	ownerRate := rp.Room.OwnerCommissionRate
	platformRate := rp.Room.PlatformCommissionRate

	ownerEarning := poolAmount.Mul(ownerRate).Round(2)
	platformEarning := poolAmount.Mul(platformRate).Round(2)
	prizePool := poolAmount.Sub(ownerEarning).Sub(platformEarning)

	var prizePerWinner decimal.Decimal
	var residual decimal.Decimal

	if len(winners) > 0 {
		prizePerWinner = prizePool.Div(decimal.NewFromInt(int64(len(winners)))).Round(2)
		residual = prizePool.Sub(prizePerWinner.Mul(decimal.NewFromInt(int64(len(winners)))))
	}

	// 使用事务执行批量结算操作（优化：减少 SQL 次数）
	winnerNames := []string{}
	winnerBalances := make(map[int64]decimal.Decimal) // 记录赢家新余额
	winnerOldBalances := make(map[int64]decimal.Decimal) // 记录赢家旧余额

	// 收集赢家信息
	winnerAmounts := make(map[int64]decimal.Decimal)
	for _, winnerID := range winners {
		if p := rp.State.Players[winnerID]; p != nil {
			winnerAmounts[winnerID] = prizePerWinner
			winnerOldBalances[winnerID] = p.Balance
			winnerNames = append(winnerNames, p.Username)
		}
	}

	err := repository.Tx(ctx, func(tx pgx.Tx) error {
		// 1. 批量发放奖金给赢家（单条 SQL）
		if len(winnerAmounts) > 0 {
			addResults, err := rp.userRepo.BatchAddBalanceTx(ctx, tx, winnerAmounts)
			if err != nil {
				return fmt.Errorf("batch add winner balance: %w", err)
			}
			for _, result := range addResults {
				winnerBalances[result.UserID] = result.NewBalance
			}
		}

		// 2. 批量创建赢家交易记录（单条 SQL）
		if len(winners) > 0 {
			txRecords := make([]*model.BalanceTransaction, 0, len(winners))
			for _, winnerID := range winners {
				oldBalance := winnerOldBalances[winnerID]
				newBalance := winnerBalances[winnerID]
				if newBalance.IsZero() {
					// 如果没有在结果中，使用计算值
					newBalance = oldBalance.Add(prizePerWinner)
					winnerBalances[winnerID] = newBalance
				}
				txRecords = append(txRecords, &model.BalanceTransaction{
					UserID:        winnerID,
					RoomID:        &rp.RoomID,
					RoundID:       &rp.State.RoundID,
					Type:          model.TxGameWin,
					Amount:        prizePerWinner,
					BalanceBefore: oldBalance,
					BalanceAfter:  newBalance,
				})
			}
			if err := rp.txRepo.BatchCreateTx(ctx, tx, txRecords); err != nil {
				return fmt.Errorf("batch create win transactions: %w", err)
			}
		}

		// 3. 房主抽成
		if err := rp.userRepo.UpdateOwnerBalancesTx(ctx, tx, rp.Room.OwnerID, "owner_room_balance", ownerEarning); err != nil {
			return fmt.Errorf("add owner earning: %w", err)
		}

		// 4. 平台抽成（合并残值）
		totalPlatformEarning := platformEarning.Add(residual)
		if err := rp.platformRepo.UpdateBalanceTx(ctx, tx, totalPlatformEarning); err != nil {
			return fmt.Errorf("add platform earning: %w", err)
		}

		// 5. 更新回合记录
		round := &model.GameRound{
			ID:              rp.State.RoundID,
			WinnerIDs:       winners,
			PrizePerWinner:  &prizePerWinner,
			OwnerEarning:    &ownerEarning,
			PlatformEarning: &platformEarning,
			ResidualAmount:  &residual,
			RevealSeed:      &revealSeed,
			Status:          model.RoundStatusSettled,
		}
		if err := rp.gameRepo.SettleRoundTx(ctx, tx, round); err != nil {
			return fmt.Errorf("settle round: %w", err)
		}

		return nil
	})

	if err != nil {
		rp.logger.Error("Settlement transaction failed", zap.Error(err))
		// 结算失败，标记回合失败并退款
		rp.handleSettlementFailure(ctx, "settlement_error")
		return
	}

	// 事务成功后更新内存状态和缓存
	for winnerID, newBalance := range winnerBalances {
		if p := rp.State.Players[winnerID]; p != nil {
			p.Balance = newBalance
		}
		// 使缓存失效，下次读取时从数据库加载
		if rp.balanceCache != nil {
			if err := rp.balanceCache.Invalidate(ctx, winnerID); err != nil {
				rp.logger.Warn("Failed to invalidate winner balance cache", zap.Int64("user_id", winnerID), zap.Error(err))
			}
		}
	}

	rp.State.Phase = model.PhaseSettlement
	rp.State.PhaseEndTime = time.Now().Add(PhaseDuration)
	rp.broadcastPhaseChange()

	// 广播结果
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeRoundResult,
		Payload: &model.WSRoundResult{
			RoundID:        rp.State.RoundID,
			Winners:        winners,
			WinnerNames:    winnerNames,
			PrizePerWinner: prizePerWinner.String(),
			RevealSeed:     revealSeed,
			CommitHash:     rp.State.CommitHash,
		},
	})

	rp.logger.Info("Phase changed", zap.String("phase", "settlement"),
		zap.Int64s("winners", winners), zap.String("prize", prizePerWinner.String()))

	// 异步执行风控检查
	if rp.riskChecker != nil {
		go rp.riskChecker.OnRoundSettled(context.Background(), rp.State.Participants, winners)
	}
}

// handleSettlementFailure 处理结算失败，退款给所有参与者（使用批量操作优化）
func (rp *RoomProcessor) handleSettlementFailure(ctx context.Context, reason string) {
	betAmount := rp.Room.BetAmount
	playerNewBalances := make(map[int64]decimal.Decimal)
	playerOldBalances := make(map[int64]decimal.Decimal)

	// 收集退款信息
	refundAmounts := make(map[int64]decimal.Decimal)
	for _, userID := range rp.State.Participants {
		if p := rp.State.Players[userID]; p != nil {
			refundAmounts[userID] = betAmount
			playerOldBalances[userID] = p.Balance
		}
	}

	// 退款给所有参与者
	err := repository.Tx(ctx, func(tx pgx.Tx) error {
		// 批量退款（单条 SQL）
		if len(refundAmounts) > 0 {
			addResults, err := rp.userRepo.BatchAddBalanceTx(ctx, tx, refundAmounts)
			if err != nil {
				return fmt.Errorf("batch refund: %w", err)
			}
			for _, result := range addResults {
				playerNewBalances[result.UserID] = result.NewBalance
			}

			// 批量创建退款交易记录（单条 SQL）
			txRecords := make([]*model.BalanceTransaction, 0, len(rp.State.Participants))
			for _, userID := range rp.State.Participants {
				oldBalance := playerOldBalances[userID]
				newBalance := playerNewBalances[userID]
				if newBalance.IsZero() {
					newBalance = oldBalance.Add(betAmount)
					playerNewBalances[userID] = newBalance
				}
				txRecords = append(txRecords, &model.BalanceTransaction{
					UserID:        userID,
					RoomID:        &rp.RoomID,
					Type:          model.TxGameRefund,
					Amount:        betAmount,
					BalanceBefore: oldBalance,
					BalanceAfter:  newBalance,
				})
			}
			if err := rp.txRepo.BatchCreateTx(ctx, tx, txRecords); err != nil {
				return fmt.Errorf("batch create refund transactions: %w", err)
			}
		}

		// 标记回合失败
		if err := rp.gameRepo.FailRound(ctx, rp.State.RoundID, reason); err != nil {
			rp.logger.Warn("Failed to mark round as failed", zap.Error(err))
		}

		return nil
	})

	if err != nil {
		rp.logger.Error("Settlement failure refund also failed", zap.Error(err))
	} else {
		// 更新内存状态
		for userID, newBalance := range playerNewBalances {
			if p := rp.State.Players[userID]; p != nil {
				p.Balance = newBalance
			}
		}
	}

	// 广播失败
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeRoundFailed,
		Payload: &model.WSRoundFailed{
			Reason:   reason,
			Refunded: rp.State.Participants,
		},
	})

	rp.enterWaiting()
}

// enterReset 进入重置阶段
func (rp *RoomProcessor) enterReset() {
	rp.State.Phase = model.PhaseReset
	rp.State.PhaseEndTime = time.Now().Add(PhaseDuration)
	rp.broadcastPhaseChange()
	rp.logger.Info("Phase changed", zap.String("phase", "reset"))
}

// enterWaiting 进入等待阶段
func (rp *RoomProcessor) enterWaiting() {
	rp.State.Phase = model.PhaseWaiting
	rp.State.PhaseEndTime = time.Now()
	rp.State.Participants = nil
	rp.State.SkippedPlayers = nil
	rp.State.PoolAmount = decimal.Zero
	rp.State.CommitHash = ""
	rp.State.Seed = nil
	rp.State.RoundID = 0
	// 重置被取消资格玩家的状态
	rp.resetDisqualifiedPlayers()
	rp.broadcastPhaseChange()
	rp.logger.Info("Phase changed", zap.String("phase", "waiting"))
}

// resetDisqualifiedPlayers 重置被取消资格玩家的状态
func (rp *RoomProcessor) resetDisqualifiedPlayers() {
	for _, p := range rp.State.Players {
		if p.Disqualified {
			p.Disqualified = false
			p.DisqualifyReason = ""
			// 广播状态重置
			disqualified := false
			reason := ""
			rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
				Type: model.WSTypePlayerUpdate,
				Payload: &model.WSPlayerUpdate{
					UserID:           p.UserID,
					Disqualified:     &disqualified,
					DisqualifyReason: &reason,
				},
			})
		}
	}
}

// refundAndWait 退款并返回等待（使用批量操作优化）
func (rp *RoomProcessor) refundAndWait(ctx context.Context, participants []int64) {
	if len(participants) == 0 {
		rp.enterWaiting()
		return
	}

	betAmount := rp.Room.BetAmount
	playerNewBalances := make(map[int64]decimal.Decimal)
	playerOldBalances := make(map[int64]decimal.Decimal)

	// 收集退款信息
	refundAmounts := make(map[int64]decimal.Decimal)
	for _, userID := range participants {
		if p := rp.State.Players[userID]; p != nil {
			refundAmounts[userID] = betAmount
			playerOldBalances[userID] = p.Balance
		}
	}

	err := repository.Tx(ctx, func(tx pgx.Tx) error {
		// 批量退款（单条 SQL）
		addResults, err := rp.userRepo.BatchAddBalanceTx(ctx, tx, refundAmounts)
		if err != nil {
			return fmt.Errorf("batch refund: %w", err)
		}
		for _, result := range addResults {
			playerNewBalances[result.UserID] = result.NewBalance
		}

		// 批量创建退款交易记录（单条 SQL）
		txRecords := make([]*model.BalanceTransaction, 0, len(participants))
		for _, userID := range participants {
			oldBalance := playerOldBalances[userID]
			newBalance := playerNewBalances[userID]
			if newBalance.IsZero() {
				newBalance = oldBalance.Add(betAmount)
				playerNewBalances[userID] = newBalance
			}
			txRecords = append(txRecords, &model.BalanceTransaction{
				UserID:        userID,
				RoomID:        &rp.RoomID,
				Type:          model.TxGameRefund,
				Amount:        betAmount,
				BalanceBefore: oldBalance,
				BalanceAfter:  newBalance,
			})
		}
		if err := rp.txRepo.BatchCreateTx(ctx, tx, txRecords); err != nil {
			return fmt.Errorf("batch create refund transactions: %w", err)
		}

		return nil
	})

	if err != nil {
		rp.logger.Error("Refund transaction failed", zap.Error(err))
	} else {
		// 事务成功后更新内存状态
		for userID, newBalance := range playerNewBalances {
			if p := rp.State.Players[userID]; p != nil {
				p.Balance = newBalance
			}
			// 使缓存失效
			if rp.balanceCache != nil {
				if cacheErr := rp.balanceCache.Invalidate(ctx, userID); cacheErr != nil {
					rp.logger.Warn("Failed to invalidate balance cache after refund", zap.Int64("user_id", userID), zap.Error(cacheErr))
				}
			}
		}
	}

	// 广播失败
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeRoundFailed,
		Payload: &model.WSRoundFailed{
			Reason:   "not_enough_players",
			Refunded: participants,
		},
	})

	rp.enterWaiting()
}

// broadcastPhaseChange 广播阶段变化
func (rp *RoomProcessor) broadcastPhaseChange() {
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypePhaseChange,
		Payload: &model.WSPhaseChange{
			Phase:        rp.State.Phase,
			PhaseEndTime: rp.State.PhaseEndTime.UnixMilli(),
			Round:        rp.State.CurrentRound,
		},
	})

	// 调整 tick 间隔
	rp.adjustTickInterval()
}

// ===== 外部调用接口 =====

// AddPlayer 添加玩家
func (rp *RoomProcessor) AddPlayer(user *model.User) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// 加载余额到缓存
	balance := user.Balance
	if rp.balanceCache != nil {
		ctx := context.Background()
		if cached, err := rp.balanceCache.LoadFromDB(ctx, user.ID); err == nil {
			balance = cached.Balance
		}
	}

	// 检查是否是重连（已存在的玩家）
	if existingPlayer, exists := rp.State.Players[user.ID]; exists {
		// 重连：保留原有的 AutoReady 状态，只更新在线状态和余额
		existingPlayer.IsOnline = true
		existingPlayer.Balance = balance
		return
	}

	// 新玩家：尝试从数据库恢复 auto_ready 状态
	autoReady := false
	if rp.roomRepo != nil {
		ctx := context.Background()
		if roomPlayer, err := rp.roomRepo.GetRoomPlayer(ctx, rp.RoomID, user.ID); err == nil && roomPlayer != nil {
			autoReady = roomPlayer.AutoReady
		}
	}

	rp.State.Players[user.ID] = &model.PlayerState{
		UserID:    user.ID,
		Username:  user.Username,
		Balance:   balance,
		AutoReady: autoReady,
		IsOnline:  true,
	}

	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypePlayerJoin,
		Payload: &model.WSPlayerJoin{
			UserID:   user.ID,
			Username: user.Username,
		},
	})
}

// LoadPlayersFromDB 从数据库加载已有玩家（服务器重启后恢复状态）
// 所有玩家初始状态为离线，等待他们重新连接 WebSocket
func (rp *RoomProcessor) LoadPlayersFromDB(ctx context.Context) error {
	if rp.roomRepo == nil {
		return nil
	}

	// 获取房间中的所有玩家
	roomPlayers, err := rp.roomRepo.GetRoomPlayers(ctx, rp.RoomID)
	if err != nil {
		return err
	}

	rp.mu.Lock()
	defer rp.mu.Unlock()

	for _, rp2 := range roomPlayers {
		// 获取用户信息
		user, err := rp.userRepo.GetByID(ctx, rp2.UserID)
		if err != nil {
			rp.logger.Warn("Failed to get user for room player", zap.Int64("user_id", rp2.UserID), zap.Error(err))
			continue
		}

		// 加载余额
		balance := user.Balance
		if rp.balanceCache != nil {
			if cached, err := rp.balanceCache.LoadFromDB(ctx, user.ID); err == nil {
				balance = cached.Balance
			}
		}

		// 添加玩家，初始状态为离线
		rp.State.Players[user.ID] = &model.PlayerState{
			UserID:    user.ID,
			Username:  user.Username,
			Balance:   balance,
			AutoReady: rp2.AutoReady,
			IsOnline:  false, // 初始为离线，等待 WebSocket 连接
		}
	}

	rp.logger.Info("Loaded players from DB", zap.Int("count", len(roomPlayers)))
	return nil
}

// RemovePlayer 移除玩家（主动离开房间时调用）
func (rp *RoomProcessor) RemovePlayer(userID int64) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	delete(rp.State.Players, userID)

	// 从数据库移除玩家记录
	if rp.roomRepo != nil {
		ctx := context.Background()
		if err := rp.roomRepo.RemovePlayer(ctx, rp.RoomID, userID); err != nil {
			rp.logger.Warn("Failed to remove player from DB", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	// 使缓存失效
	if rp.balanceCache != nil {
		ctx := context.Background()
		if err := rp.balanceCache.Invalidate(ctx, userID); err != nil {
			rp.logger.Warn("Failed to invalidate balance cache on player leave", zap.Int64("user_id", userID), zap.Error(err))
		}
	}

	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypePlayerLeave,
		Payload: &model.WSPlayerLeave{
			UserID: userID,
		},
	})

	rp.logger.Info("Player removed from room", zap.Int64("user_id", userID))
}

// SetPlayerOnline 设置玩家在线状态
func (rp *RoomProcessor) SetPlayerOnline(userID int64, online bool) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if p := rp.State.Players[userID]; p != nil {
		p.IsOnline = online
		if online {
			// 上线时清除离线时间
			p.OfflineSince = nil
		} else {
			// 离线时记录离线时间
			now := time.Now()
			p.OfflineSince = &now
		}
		rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
			Type: model.WSTypePlayerUpdate,
			Payload: &model.WSPlayerUpdate{
				UserID:   userID,
				IsOnline: &online,
			},
		})
	}
}

// SetAutoReady 设置自动准备
func (rp *RoomProcessor) SetAutoReady(userID int64, autoReady bool) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if p := rp.State.Players[userID]; p != nil {
		p.AutoReady = autoReady
		rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
			Type: model.WSTypePlayerUpdate,
			Payload: &model.WSPlayerUpdate{
				UserID:    userID,
				AutoReady: &autoReady,
			},
		})
	}
}

// GetRoomState 获取房间状态快照
func (rp *RoomProcessor) GetRoomState() *model.WSRoomState {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	players := make(map[int64]*model.WSPlayerState)
	for id, p := range rp.State.Players {
		players[id] = &model.WSPlayerState{
			UserID:           p.UserID,
			Username:         p.Username,
			Balance:          p.Balance.String(),
			AutoReady:        p.AutoReady,
			IsOnline:         p.IsOnline,
			Disqualified:     p.Disqualified,
			DisqualifyReason: p.DisqualifyReason,
		}
	}

	return &model.WSRoomState{
		RoomID:       rp.RoomID,
		RoomName:     rp.Room.Name,
		BetAmount:    rp.Room.BetAmount.String(),
		WinnerCount:  rp.Room.WinnerCount,
		MaxPlayers:   rp.Room.MaxPlayers,
		Phase:        rp.State.Phase,
		PhaseEndTime: rp.State.PhaseEndTime.UnixMilli(),
		CurrentRound: rp.State.CurrentRound,
		Players:      players,
		PoolAmount:   rp.State.PoolAmount.String(),
	}
}

// UpdatePlayerBalance 更新玩家余额(外部充值/提现后调用)
func (rp *RoomProcessor) UpdatePlayerBalance(userID int64, balance decimal.Decimal) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if p := rp.State.Players[userID]; p != nil {
		p.Balance = balance
		rp.Broadcaster.SendToUser(userID, &model.WSMessage{
			Type: model.WSTypeBalanceUpdate,
			Payload: &model.WSBalanceUpdate{
				Balance:       balance.String(),
				FrozenBalance: "0",
			},
		})
	}
}


// ===== 观战者相关接口 =====

const MaxSpectators = 50 // 最大观战人数

// AddSpectator 添加观战者
func (rp *RoomProcessor) AddSpectator(user *model.User) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// 检查是否已是玩家
	if _, exists := rp.State.Players[user.ID]; exists {
		return ErrAlreadyParticipant
	}

	// 检查是否已是观战者
	if _, exists := rp.State.Spectators[user.ID]; exists {
		return ErrAlreadySpectator
	}

	// 检查观战人数上限
	if len(rp.State.Spectators) >= MaxSpectators {
		return ErrSpectatorLimitReached
	}

	rp.State.Spectators[user.ID] = &model.SpectatorState{
		UserID:   user.ID,
		Username: user.Username,
		JoinedAt: time.Now(),
	}

	// 广播观战者加入
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeSpectatorJoin,
		Payload: &model.WSSpectatorJoin{
			UserID:   user.ID,
			Username: user.Username,
		},
	})

	rp.logger.Info("Spectator joined", zap.Int64("user_id", user.ID), zap.String("username", user.Username))
	return nil
}

// RemoveSpectator 移除观战者
func (rp *RoomProcessor) RemoveSpectator(userID int64) {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	if _, exists := rp.State.Spectators[userID]; !exists {
		return
	}

	delete(rp.State.Spectators, userID)

	// 广播观战者离开
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeSpectatorLeave,
		Payload: &model.WSSpectatorLeave{
			UserID: userID,
		},
	})

	rp.logger.Info("Spectator left", zap.Int64("user_id", userID))
}

// SpectatorToParticipant 观战者切换为参与者
func (rp *RoomProcessor) SpectatorToParticipant(user *model.User) error {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	// 检查是否是观战者
	spectator, exists := rp.State.Spectators[user.ID]
	if !exists {
		return ErrNotSpectator
	}

	// 检查房间是否已满
	if len(rp.State.Players) >= rp.Room.MaxPlayers {
		return ErrRoomFull
	}

	// 从观战者列表移除
	delete(rp.State.Spectators, user.ID)

	// 从数据库/缓存获取最新余额
	balance := user.Balance
	if rp.balanceCache != nil {
		ctx := context.Background()
		if cached, err := rp.balanceCache.LoadFromDB(ctx, user.ID); err == nil {
			balance = cached.Balance
		}
	}

	// 添加到玩家列表
	rp.State.Players[user.ID] = &model.PlayerState{
		UserID:    user.ID,
		Username:  spectator.Username,
		Balance:   balance,
		AutoReady: false,
		IsOnline:  true,
	}

	// 广播观战者切换为参与者
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypeSpectatorSwitch,
		Payload: &model.WSSpectatorSwitch{
			UserID:   user.ID,
			Username: spectator.Username,
		},
	})

	// 同时广播玩家加入
	rp.Broadcaster.BroadcastToRoom(rp.RoomID, &model.WSMessage{
		Type: model.WSTypePlayerJoin,
		Payload: &model.WSPlayerJoin{
			UserID:   user.ID,
			Username: spectator.Username,
		},
	})

	rp.logger.Info("Spectator switched to participant", zap.Int64("user_id", user.ID))
	return nil
}

// IsSpectator 检查用户是否是观战者
func (rp *RoomProcessor) IsSpectator(userID int64) bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	_, exists := rp.State.Spectators[userID]
	return exists
}

// IsParticipant 检查用户是否是参与者
func (rp *RoomProcessor) IsParticipant(userID int64) bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	_, exists := rp.State.Players[userID]
	return exists
}

// GetSpectatorCount 获取观战者数量
func (rp *RoomProcessor) GetSpectatorCount() int {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return len(rp.State.Spectators)
}

// GetRoomStateForUser 获取房间状态快照（包含用户是否为观战者的信息）
func (rp *RoomProcessor) GetRoomStateForUser(userID int64) *model.WSRoomState {
	rp.mu.RLock()
	defer rp.mu.RUnlock()

	players := make(map[int64]*model.WSPlayerState)
	for id, p := range rp.State.Players {
		players[id] = &model.WSPlayerState{
			UserID:    p.UserID,
			Username:  p.Username,
			Balance:   p.Balance.String(),
			AutoReady: p.AutoReady,
			IsOnline:  p.IsOnline,
		}
	}

	spectators := make(map[int64]*model.WSSpectatorState)
	for id, s := range rp.State.Spectators {
		spectators[id] = &model.WSSpectatorState{
			UserID:   s.UserID,
			Username: s.Username,
		}
	}

	_, isSpectator := rp.State.Spectators[userID]

	return &model.WSRoomState{
		RoomID:        rp.RoomID,
		RoomName:      rp.Room.Name,
		BetAmount:     rp.Room.BetAmount.String(),
		WinnerCount:   rp.Room.WinnerCount,
		MaxPlayers:    rp.Room.MaxPlayers,
		MaxSpectators: MaxSpectators,
		Phase:         rp.State.Phase,
		PhaseEndTime:  rp.State.PhaseEndTime.UnixMilli(),
		CurrentRound:  rp.State.CurrentRound,
		Players:       players,
		Spectators:    spectators,
		PoolAmount:    rp.State.PoolAmount.String(),
		IsSpectator:   isSpectator,
	}
}
