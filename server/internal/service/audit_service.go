package service

import (
	"context"
	"time"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/fiveseconds/server/pkg/logger"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// AuditService 资金审计服务
type AuditService struct {
	fundService      *FundService
	platformRepo     *repository.PlatformRepo
	conservationRepo *repository.ConservationRepo
	alertManager     *AlertManager
	logger           *logger.Logger
}

// NewAuditService 创建审计服务
func NewAuditService(
	fundService *FundService,
	platformRepo *repository.PlatformRepo,
	conservationRepo *repository.ConservationRepo,
	alertManager *AlertManager,
	log *logger.Logger,
) *AuditService {
	return &AuditService{
		fundService:      fundService,
		platformRepo:     platformRepo,
		conservationRepo: conservationRepo,
		alertManager:     alertManager,
		logger:           log.With(zap.String("service", "audit")),
	}
}

// RunGlobalConservationCheck 执行全局资金守恒检查
func (s *AuditService) RunGlobalConservationCheck(ctx context.Context) (*model.ConservationCheck, error) {
	s.logger.WithContext(ctx).Info("Starting global conservation check")

	check, err := s.platformRepo.CheckConservation(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Error("Global conservation check failed", zap.Error(err))
		return nil, err
	}

	// 记录检查结果
	s.logger.WithContext(ctx).Info("Global conservation check completed",
		zap.Bool("is_balanced", check.IsBalanced),
		zap.String("system_total", check.SystemTotalFunds.String()),
		zap.String("expected_total", check.ExpectedTotal.String()),
		zap.String("difference", check.Difference.String()),
	)

	// 如果不平衡，触发告警
	if !check.IsBalanced {
		s.logger.WithContext(ctx).Error("CRITICAL: Global conservation check FAILED",
			zap.String("difference", check.Difference.String()),
		)
		if s.alertManager != nil {
			s.alertManager.TriggerConservationFailedAlert(ctx, check.Difference)
		}
	}

	return check, nil
}

// RunPeriodicAudit 执行定期审计（2小时/每日）
func (s *AuditService) RunPeriodicAudit(ctx context.Context, periodType string) error {
	now := time.Now()
	var periodStart, periodEnd time.Time

	switch periodType {
	case "2h":
		periodEnd = now
		periodStart = now.Add(-2 * time.Hour)
	case "daily":
		periodEnd = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		periodStart = periodEnd.AddDate(0, 0, -1)
	default:
		periodType = "2h"
		periodEnd = now
		periodStart = now.Add(-2 * time.Hour)
	}

	s.logger.WithContext(ctx).Info("Starting periodic audit",
		zap.String("period_type", periodType),
		zap.Time("period_start", periodStart),
		zap.Time("period_end", periodEnd),
	)

	// 1. 执行全局对账
	check, err := s.RunGlobalConservationCheck(ctx)
	if err != nil {
		return err
	}

	// 2. 记录全局对账历史
	if err := s.fundService.RecordGlobalConservation(ctx, periodType, periodStart, periodEnd, check); err != nil {
		s.logger.WithContext(ctx).Error("Failed to record global conservation", zap.Error(err))
	}

	// 3. 执行按房主维度的对账
	if periodType == "2h" {
		if err := s.fundService.RecordOwnerConservation2h(ctx, periodStart, periodEnd); err != nil {
			s.logger.WithContext(ctx).Error("Failed to record owner conservation 2h", zap.Error(err))
		}
	} else if periodType == "daily" {
		if err := s.fundService.RecordOwnerConservationDaily(ctx, periodStart, periodEnd); err != nil {
			s.logger.WithContext(ctx).Error("Failed to record owner conservation daily", zap.Error(err))
		}
	}

	s.logger.WithContext(ctx).Info("Periodic audit completed",
		zap.String("period_type", periodType),
		zap.Bool("is_balanced", check.IsBalanced),
	)

	return nil
}

// AuditResult 审计结果
type AuditResult struct {
	Timestamp          time.Time                 `json:"timestamp"`
	GlobalCheck        *model.ConservationCheck  `json:"global_check"`
	OwnerChecks        []*OwnerAuditResult       `json:"owner_checks,omitempty"`
	TransactionSummary *TransactionAuditSummary  `json:"transaction_summary"`
	Anomalies          []string                  `json:"anomalies,omitempty"`
}

// OwnerAuditResult 房主审计结果
type OwnerAuditResult struct {
	OwnerID            int64                    `json:"owner_id"`
	OwnerUsername      string                   `json:"owner_username"`
	Check              *model.ConservationCheck `json:"check"`
	PlayerCount        int                      `json:"player_count"`
	TotalPlayerBalance decimal.Decimal          `json:"total_player_balance"`
}

// TransactionAuditSummary 交易审计摘要
type TransactionAuditSummary struct {
	TotalTransactions   int64           `json:"total_transactions"`
	TotalDeposits       decimal.Decimal `json:"total_deposits"`
	TotalWithdrawals    decimal.Decimal `json:"total_withdrawals"`
	TotalBets           decimal.Decimal `json:"total_bets"`
	TotalWinnings       decimal.Decimal `json:"total_winnings"`
	TotalCommissions    decimal.Decimal `json:"total_commissions"`
	TotalPlatformShare  decimal.Decimal `json:"total_platform_share"`
	NetFlow             decimal.Decimal `json:"net_flow"` // 净流入
}

// RunFullAudit 执行完整审计
func (s *AuditService) RunFullAudit(ctx context.Context) (*AuditResult, error) {
	result := &AuditResult{
		Timestamp: time.Now(),
		Anomalies: []string{},
	}

	// 1. 全局对账
	globalCheck, err := s.RunGlobalConservationCheck(ctx)
	if err != nil {
		return nil, err
	}
	result.GlobalCheck = globalCheck

	if !globalCheck.IsBalanced {
		result.Anomalies = append(result.Anomalies,
			"全局资金不平衡，差额: "+globalCheck.Difference.String())
	}

	// 2. 交易摘要
	summary, err := s.getTransactionSummary(ctx)
	if err != nil {
		s.logger.WithContext(ctx).Warn("Failed to get transaction summary", zap.Error(err))
	} else {
		result.TransactionSummary = summary
	}

	// 3. 检查异常情况
	anomalies := s.checkAnomalies(ctx, globalCheck, summary)
	result.Anomalies = append(result.Anomalies, anomalies...)

	return result, nil
}

// getTransactionSummary 获取交易摘要
func (s *AuditService) getTransactionSummary(ctx context.Context) (*TransactionAuditSummary, error) {
	summary := &TransactionAuditSummary{}

	// 从数据库聚合交易数据
	sql := `SELECT 
		COUNT(*) as total,
		COALESCE(SUM(CASE WHEN tx_type = 'deposit' AND amount > 0 THEN amount ELSE 0 END), 0) as deposits,
		COALESCE(SUM(CASE WHEN tx_type = 'withdraw' AND amount < 0 THEN ABS(amount) ELSE 0 END), 0) as withdrawals,
		COALESCE(SUM(CASE WHEN tx_type = 'game_bet' THEN ABS(amount) ELSE 0 END), 0) as bets,
		COALESCE(SUM(CASE WHEN tx_type = 'game_win' THEN amount ELSE 0 END), 0) as winnings,
		COALESCE(SUM(CASE WHEN tx_type = 'owner_commission' THEN amount ELSE 0 END), 0) as commissions,
		COALESCE(SUM(CASE WHEN tx_type = 'platform_share' THEN amount ELSE 0 END), 0) as platform_share
		FROM balance_transactions`

	err := repository.DB.QueryRow(ctx, sql).Scan(
		&summary.TotalTransactions,
		&summary.TotalDeposits,
		&summary.TotalWithdrawals,
		&summary.TotalBets,
		&summary.TotalWinnings,
		&summary.TotalCommissions,
		&summary.TotalPlatformShare,
	)
	if err != nil {
		return nil, err
	}

	// 计算净流入
	summary.NetFlow = summary.TotalDeposits.Sub(summary.TotalWithdrawals)

	return summary, nil
}

// checkAnomalies 检查异常情况
func (s *AuditService) checkAnomalies(ctx context.Context, check *model.ConservationCheck, summary *TransactionAuditSummary) []string {
	anomalies := []string{}

	// 1. 检查负余额用户
	var negativeCount int
	err := repository.DB.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE balance < 0 OR frozen_balance < 0`).Scan(&negativeCount)
	if err == nil && negativeCount > 0 {
		anomalies = append(anomalies, "存在负余额用户: "+string(rune(negativeCount))+"个")
	}

	// 2. 检查交易记录与余额是否匹配
	if summary != nil {
		// 游戏下注应该等于游戏获胜+佣金+平台抽成（在完美情况下）
		// 实际上由于退款等情况，这个等式不一定成立，但差额不应该太大
		gameIncome := summary.TotalWinnings.Add(summary.TotalCommissions).Add(summary.TotalPlatformShare)
		gameDiff := summary.TotalBets.Sub(gameIncome)
		// 允许一定误差（退款等情况）
		if gameDiff.Abs().GreaterThan(decimal.NewFromInt(100)) {
			anomalies = append(anomalies, "游戏资金流向异常，差额: "+gameDiff.String())
		}
	}

	return anomalies
}

// GetAuditHistory 获取审计历史
func (s *AuditService) GetAuditHistory(ctx context.Context, query *model.FundConservationHistoryQuery) ([]*model.FundConservationHistory, int64, error) {
	return s.conservationRepo.List(ctx, query)
}
