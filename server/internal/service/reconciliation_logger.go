package service

import (
	"context"
	"time"

	"github.com/fiveseconds/server/pkg/logger"
	"github.com/fiveseconds/server/pkg/metrics"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// ReconciliationLogger logs reconciliation activities
type ReconciliationLogger struct {
	log          *logger.Logger
	alertManager *AlertManager
}

// NewReconciliationLogger creates a new reconciliation logger
func NewReconciliationLogger(log *logger.Logger, alertManager *AlertManager) *ReconciliationLogger {
	return &ReconciliationLogger{
		log:          log.With(zap.String("component", "reconciliation")),
		alertManager: alertManager,
	}
}

// ReconciliationResult holds the result of a reconciliation run
type ReconciliationResult struct {
	PeriodType         string          // "2h" or "daily"
	PeriodStart        time.Time
	PeriodEnd          time.Time
	TotalPlayerBalance decimal.Decimal
	TotalCustodyQuota  decimal.Decimal
	TotalMargin        decimal.Decimal
	PlatformBalance    decimal.Decimal
	Difference         decimal.Decimal
	IsBalanced         bool
	Duration           time.Duration
}

// LogReconciliationStarted logs when reconciliation starts
func (r *ReconciliationLogger) LogReconciliationStarted(ctx context.Context, periodType string, periodStart, periodEnd time.Time) {
	r.log.WithContext(ctx).Info("reconciliation started",
		zap.String("period_type", periodType),
		zap.Time("period_start", periodStart),
		zap.Time("period_end", periodEnd),
	)
}


// LogReconciliationResult logs the result of a reconciliation run
func (r *ReconciliationLogger) LogReconciliationResult(ctx context.Context, result *ReconciliationResult) {
	fields := []zap.Field{
		zap.String("period_type", result.PeriodType),
		zap.Time("period_start", result.PeriodStart),
		zap.Time("period_end", result.PeriodEnd),
		zap.String("total_player_balance", result.TotalPlayerBalance.String()),
		zap.String("total_custody_quota", result.TotalCustodyQuota.String()),
		zap.String("total_margin", result.TotalMargin.String()),
		zap.String("platform_balance", result.PlatformBalance.String()),
		zap.String("difference", result.Difference.String()),
		zap.Bool("is_balanced", result.IsBalanced),
		zap.Duration("duration", result.Duration),
	}

	if result.IsBalanced {
		r.log.WithContext(ctx).Info("reconciliation completed successfully", fields...)
		metrics.RecordReconciliation("success", result.PeriodType)
	} else {
		r.log.WithContext(ctx).Error("CRITICAL: reconciliation detected imbalance", fields...)
		metrics.RecordReconciliation("imbalance", result.PeriodType)

		// Trigger alert if alert manager is available
		if r.alertManager != nil {
			r.alertManager.TriggerConservationFailedAlert(ctx, result.Difference)
		}
	}
}


// OwnerReconciliationResult holds the result of an owner reconciliation
type OwnerReconciliationResult struct {
	OwnerID            int64
	OwnerUsername      string
	PlayerCount        int
	TotalPlayerBalance decimal.Decimal
	CustodyQuota       decimal.Decimal
	MarginBalance      decimal.Decimal
	IsBalanced         bool
	Difference         decimal.Decimal
}

// LogOwnerReconciliation logs the result of an owner reconciliation
func (r *ReconciliationLogger) LogOwnerReconciliation(ctx context.Context, result *OwnerReconciliationResult) {
	fields := []zap.Field{
		zap.Int64("owner_id", result.OwnerID),
		zap.String("owner_username", result.OwnerUsername),
		zap.Int("player_count", result.PlayerCount),
		zap.String("total_player_balance", result.TotalPlayerBalance.String()),
		zap.String("custody_quota", result.CustodyQuota.String()),
		zap.String("margin_balance", result.MarginBalance.String()),
		zap.Bool("is_balanced", result.IsBalanced),
		zap.String("difference", result.Difference.String()),
	}

	if result.IsBalanced {
		r.log.WithContext(ctx).Info("owner reconciliation passed", fields...)
	} else {
		r.log.WithContext(ctx).Warn("owner reconciliation detected discrepancy", fields...)
	}
}

// LogReconciliationError logs a reconciliation error
func (r *ReconciliationLogger) LogReconciliationError(ctx context.Context, periodType string, err error, retryCount int) {
	r.log.WithContext(ctx).Error("reconciliation failed",
		zap.String("period_type", periodType),
		zap.Error(err),
		zap.Int("retry_count", retryCount),
	)
	metrics.RecordReconciliation("error", periodType)
}
