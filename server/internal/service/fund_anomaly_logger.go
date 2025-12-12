package service

import (
	"context"

	"github.com/fiveseconds/server/pkg/logger"
	"github.com/fiveseconds/server/pkg/metrics"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	// FundAnomalyLargeTransactionThreshold is the threshold for large transaction warnings
	FundAnomalyLargeTransactionThreshold = 10000
	// FundAnomalyConsecutiveWinThreshold is the threshold for consecutive win warnings
	FundAnomalyConsecutiveWinThreshold = 10
	// FundAnomalyHighWinRateThreshold is the threshold for high win rate warnings (80%)
	FundAnomalyHighWinRateThreshold = 0.80
	// FundAnomalyHighWinRateMinRounds is the minimum rounds for high win rate detection
	FundAnomalyHighWinRateMinRounds = 50
)

// FundAnomalyLogger logs fund-related anomalies
type FundAnomalyLogger struct {
	log          *logger.Logger
	alertManager *AlertManager
}

// NewFundAnomalyLogger creates a new fund anomaly logger
func NewFundAnomalyLogger(log *logger.Logger, alertManager *AlertManager) *FundAnomalyLogger {
	return &FundAnomalyLogger{
		log:          log.With(zap.String("component", "fund_anomaly")),
		alertManager: alertManager,
	}
}

// LogNegativeBalance logs a critical alert for negative balance
func (f *FundAnomalyLogger) LogNegativeBalance(ctx context.Context, userID int64, balance decimal.Decimal, lastTxID int64, lastTxType string) {
	f.log.WithContext(ctx).Error("CRITICAL: negative balance detected",
		zap.Int64("user_id", userID),
		zap.String("balance", balance.String()),
		zap.Int64("last_tx_id", lastTxID),
		zap.String("last_tx_type", lastTxType),
	)
	metrics.RecordFundAnomaly("negative_balance")

	// Trigger alert if alert manager is available
	if f.alertManager != nil {
		f.alertManager.TriggerNegativeBalanceAlert(ctx, userID, balance)
	}
}

// LogLargeTransaction logs a warning for large transactions
func (f *FundAnomalyLogger) LogLargeTransaction(ctx context.Context, userID int64, amount decimal.Decimal, txType string, txID int64) {
	f.log.WithContext(ctx).Warn("large transaction detected",
		zap.Int64("user_id", userID),
		zap.String("amount", amount.String()),
		zap.String("tx_type", txType),
		zap.Int64("tx_id", txID),
	)
	metrics.RecordFundAnomaly("large_transaction")
}

// LogConsecutiveWins logs a warning for consecutive wins
func (f *FundAnomalyLogger) LogConsecutiveWins(ctx context.Context, userID int64, winStreak int, roomID int64) {
	f.log.WithContext(ctx).Warn("consecutive wins detected",
		zap.Int64("user_id", userID),
		zap.Int("win_streak", winStreak),
		zap.Int64("room_id", roomID),
	)
	metrics.RecordFundAnomaly("consecutive_wins")
}


// LogHighWinRate logs a warning for high win rate
func (f *FundAnomalyLogger) LogHighWinRate(ctx context.Context, userID int64, winRate float64, totalRounds int, roomID int64) {
	f.log.WithContext(ctx).Warn("high win rate detected",
		zap.Int64("user_id", userID),
		zap.Float64("win_rate", winRate),
		zap.Int("total_rounds", totalRounds),
		zap.Int64("room_id", roomID),
	)
	metrics.RecordFundAnomaly("high_win_rate")
}

// LogDuplicateDeviceFingerprint logs a warning for duplicate device fingerprints
func (f *FundAnomalyLogger) LogDuplicateDeviceFingerprint(ctx context.Context, userIDs []int64, fingerprint string) {
	f.log.WithContext(ctx).Warn("duplicate device fingerprint detected",
		zap.Int64s("user_ids", userIDs),
		zap.String("fingerprint", fingerprint),
	)
	metrics.RecordFundAnomaly("duplicate_fingerprint")
}

// LogInsufficientCustodyQuota logs a warning for insufficient custody quota
func (f *FundAnomalyLogger) LogInsufficientCustodyQuota(ctx context.Context, ownerID int64, quota decimal.Decimal, requestedAmount decimal.Decimal) {
	f.log.WithContext(ctx).Warn("insufficient custody quota for withdrawal",
		zap.Int64("owner_id", ownerID),
		zap.String("custody_quota", quota.String()),
		zap.String("requested_amount", requestedAmount.String()),
	)
	metrics.RecordFundAnomaly("insufficient_custody")
}

// ShouldLogLargeTransaction checks if a transaction amount exceeds the threshold
func ShouldLogLargeTransaction(amount decimal.Decimal) bool {
	return amount.GreaterThan(decimal.NewFromInt(FundAnomalyLargeTransactionThreshold))
}

// ShouldLogConsecutiveWins checks if win streak exceeds the threshold
func ShouldLogConsecutiveWins(winStreak int) bool {
	return winStreak > FundAnomalyConsecutiveWinThreshold
}

// ShouldLogHighWinRate checks if win rate exceeds the threshold
func ShouldLogHighWinRate(winRate float64, totalRounds int) bool {
	return totalRounds >= FundAnomalyHighWinRateMinRounds && winRate > FundAnomalyHighWinRateThreshold
}
