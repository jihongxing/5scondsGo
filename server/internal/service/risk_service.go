package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RiskControlService 风控服务
type RiskControlService struct {
	riskRepo     *repository.RiskRepo
	alertManager *AlertManager
	config       model.RiskConfig
	logger       *zap.Logger
}

// NewRiskControlService 创建风控服务
func NewRiskControlService(
	riskRepo *repository.RiskRepo,
	alertManager *AlertManager,
	logger *zap.Logger,
) *RiskControlService {
	return &RiskControlService{
		riskRepo:     riskRepo,
		alertManager: alertManager,
		config:       model.DefaultRiskConfig,
		logger:       logger.With(zap.String("service", "risk_control")),
	}
}

// CheckConsecutiveWins 检查连续获胜
func (s *RiskControlService) CheckConsecutiveWins(ctx context.Context, userID int64, isWinner bool) error {
	if isWinner {
		// 获取当前连续获胜次数
		wins, err := s.riskRepo.GetUserConsecutiveWins(ctx, userID)
		if err != nil {
			s.logger.Error("Failed to get consecutive wins", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}

		wins++
		if err := s.riskRepo.UpdateUserConsecutiveWins(ctx, userID, wins); err != nil {
			s.logger.Error("Failed to update consecutive wins", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}

		// 检查是否超过阈值
		if wins > s.config.ConsecutiveWinThreshold {
			// 检查是否已有待处理的标记
			hasPending, err := s.riskRepo.HasPendingFlag(ctx, userID, model.RiskFlagConsecutiveWins)
			if err != nil {
				s.logger.Error("Failed to check pending flag", zap.Error(err))
				return err
			}

			if !hasPending {
				// 创建风控标记
				details := &model.RiskFlagDetails{
					ConsecutiveWins: wins,
				}
				if err := s.createRiskFlag(ctx, userID, model.RiskFlagConsecutiveWins, details); err != nil {
					return err
				}
				s.logger.Warn("Consecutive wins threshold exceeded",
					zap.Int64("user_id", userID),
					zap.Int("consecutive_wins", wins))
			}
		}
	} else {
		// 输了，重置连续获胜次数
		if err := s.riskRepo.ResetUserConsecutiveWins(ctx, userID); err != nil {
			s.logger.Error("Failed to reset consecutive wins", zap.Int64("user_id", userID), zap.Error(err))
			return err
		}
	}

	return nil
}

// CheckWinRate 检查胜率
func (s *RiskControlService) CheckWinRate(ctx context.Context, userID int64) error {
	winRate, totalRounds, err := s.riskRepo.GetUserWinRate(ctx, userID, s.config.WinRateMinRounds)
	if err != nil {
		s.logger.Error("Failed to get win rate", zap.Int64("user_id", userID), zap.Error(err))
		return err
	}

	// 只有达到最小回合数才检查
	if totalRounds < s.config.WinRateMinRounds {
		return nil
	}

	if winRate > s.config.WinRateThreshold {
		// 检查是否已有待处理的标记
		hasPending, err := s.riskRepo.HasPendingFlag(ctx, userID, model.RiskFlagHighWinRate)
		if err != nil {
			s.logger.Error("Failed to check pending flag", zap.Error(err))
			return err
		}

		if !hasPending {
			details := &model.RiskFlagDetails{
				WinRate:     winRate,
				TotalRounds: totalRounds,
			}
			if err := s.createRiskFlag(ctx, userID, model.RiskFlagHighWinRate, details); err != nil {
				return err
			}
			s.logger.Warn("High win rate detected",
				zap.Int64("user_id", userID),
				zap.Float64("win_rate", winRate),
				zap.Int("total_rounds", totalRounds))
		}
	}

	return nil
}


// CheckDeviceFingerprint 检查设备指纹（多账户检测）
func (s *RiskControlService) CheckDeviceFingerprint(ctx context.Context, userID int64, fingerprint string) error {
	if fingerprint == "" {
		return nil
	}

	userIDs, err := s.riskRepo.GetUsersByDeviceFingerprint(ctx, fingerprint)
	if err != nil {
		s.logger.Error("Failed to get users by fingerprint", zap.Error(err))
		return err
	}

	// 如果有多个用户使用相同设备
	if len(userIDs) > 1 {
		// 检查是否已有待处理的标记
		hasPending, err := s.riskRepo.HasPendingFlag(ctx, userID, model.RiskFlagMultiAccount)
		if err != nil {
			s.logger.Error("Failed to check pending flag", zap.Error(err))
			return err
		}

		if !hasPending {
			details := &model.RiskFlagDetails{
				DeviceFingerprint: fingerprint,
				RelatedUserIDs:    userIDs,
			}
			if err := s.createRiskFlag(ctx, userID, model.RiskFlagMultiAccount, details); err != nil {
				return err
			}
			s.logger.Warn("Multi-account detected",
				zap.Int64("user_id", userID),
				zap.String("fingerprint", fingerprint),
				zap.Int64s("related_users", userIDs))
		}
	}

	return nil
}

// CheckLargeTransaction 检查大额交易
func (s *RiskControlService) CheckLargeTransaction(ctx context.Context, userID int64, amount decimal.Decimal) error {
	if amount.Abs().LessThan(s.config.LargeTransactionAmount) {
		return nil
	}

	details := &model.RiskFlagDetails{
		TransactionAmount: amount,
	}
	if err := s.createRiskFlag(ctx, userID, model.RiskFlagLargeTransaction, details); err != nil {
		return err
	}

	// 同时触发告警
	if s.alertManager != nil {
		s.alertManager.TriggerLargeTransactionAlert(ctx, userID, amount)
	}

	s.logger.Warn("Large transaction detected",
		zap.Int64("user_id", userID),
		zap.String("amount", amount.String()))

	return nil
}

// CheckDailyVolume 检查日交易量
func (s *RiskControlService) CheckDailyVolume(ctx context.Context, userID int64) error {
	volumeStr, err := s.riskRepo.GetUserDailyVolume(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get daily volume", zap.Int64("user_id", userID), zap.Error(err))
		return err
	}

	volume, err := decimal.NewFromString(volumeStr)
	if err != nil {
		return err
	}

	if volume.GreaterThan(s.config.DailyVolumeThreshold) {
		// 触发告警
		if s.alertManager != nil {
			s.alertManager.TriggerDailyVolumeAlert(ctx, userID, volume)
		}
		s.logger.Warn("Daily volume threshold exceeded",
			zap.Int64("user_id", userID),
			zap.String("volume", volume.String()))
	}

	return nil
}

// createRiskFlag 创建风控标记
func (s *RiskControlService) createRiskFlag(ctx context.Context, userID int64, flagType model.RiskFlagType, details *model.RiskFlagDetails) error {
	flag := &model.RiskFlag{
		UserID:   userID,
		FlagType: flagType,
		Status:   model.RiskFlagStatusPending,
	}

	if err := s.riskRepo.CreateFlagWithDetails(ctx, flag, details); err != nil {
		s.logger.Error("Failed to create risk flag", zap.Error(err))
		return err
	}

	// 触发告警
	if s.alertManager != nil {
		s.alertManager.TriggerRiskFlagAlert(ctx, flag)
	}

	return nil
}

// ReviewFlag 审核风控标记
func (s *RiskControlService) ReviewFlag(ctx context.Context, flagID int64, action string, reviewedBy int64) error {
	var status model.RiskFlagStatus
	switch action {
	case "confirm":
		status = model.RiskFlagStatusConfirmed
	case "dismiss":
		status = model.RiskFlagStatusDismissed
	default:
		return fmt.Errorf("invalid action: %s", action)
	}

	return s.riskRepo.ReviewFlag(ctx, flagID, status, reviewedBy)
}

// ListFlags 列表风控标记
func (s *RiskControlService) ListFlags(ctx context.Context, query *model.RiskFlagListQuery) ([]*model.RiskFlag, int64, error) {
	return s.riskRepo.ListFlags(ctx, query)
}

// GetFlag 获取风控标记
func (s *RiskControlService) GetFlag(ctx context.Context, id int64) (*model.RiskFlag, error) {
	return s.riskRepo.GetFlagByID(ctx, id)
}

// OnRoundSettled 回合结算后的风控检查
func (s *RiskControlService) OnRoundSettled(ctx context.Context, participants []int64, winners []int64) {
	winnerSet := make(map[int64]bool)
	for _, w := range winners {
		winnerSet[w] = true
	}

	for _, userID := range participants {
		isWinner := winnerSet[userID]

		// 检查连续获胜
		if err := s.CheckConsecutiveWins(ctx, userID, isWinner); err != nil {
			s.logger.Error("Failed to check consecutive wins", zap.Int64("user_id", userID), zap.Error(err))
		}

		// 检查胜率（只对赢家检查，减少查询）
		if isWinner {
			if err := s.CheckWinRate(ctx, userID); err != nil {
				s.logger.Error("Failed to check win rate", zap.Int64("user_id", userID), zap.Error(err))
			}
		}
	}
}

// GetFlagDetails 解析风控标记详情
func (s *RiskControlService) GetFlagDetails(flag *model.RiskFlag) (*model.RiskFlagDetails, error) {
	var details model.RiskFlagDetails
	if err := json.Unmarshal([]byte(flag.Details), &details); err != nil {
		return nil, err
	}
	return &details, nil
}
