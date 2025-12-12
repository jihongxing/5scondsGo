package repository

import (
	"context"
	"time"

	"github.com/fiveseconds/server/pkg/logger"
	"github.com/fiveseconds/server/pkg/metrics"
	"github.com/fiveseconds/server/pkg/trace"
	"go.uber.org/zap"
)

// SlowQueryThreshold defines the threshold for slow query detection
const SlowQueryThreshold = 100 * time.Millisecond

// DBLogger provides database query logging functionality
type DBLogger struct {
	log *logger.Logger
}

// NewDBLogger creates a new database logger
func NewDBLogger(log *logger.Logger) *DBLogger {
	return &DBLogger{log: log}
}

// LogQuery logs a database query with timing and error information
func (d *DBLogger) LogQuery(ctx context.Context, operation string, query string, duration time.Duration, err error) {
	// Record metrics
	metrics.RecordDBQuery(operation, duration, err)

	traceID := trace.GetTraceID(ctx)

	fields := []zap.Field{
		zap.String("trace_id", traceID),
		zap.String("operation", operation),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		d.log.WithContext(ctx).Error("database query failed", fields...)
		return
	}

	if duration > SlowQueryThreshold {
		fields = append(fields, zap.String("query", query))
		d.log.WithContext(ctx).Warn("slow query detected", fields...)
	} else {
		d.log.WithContext(ctx).Debug("database query executed", fields...)
	}
}


// QueryTimer helps time database queries
type QueryTimer struct {
	ctx       context.Context
	operation string
	query     string
	start     time.Time
	logger    *DBLogger
}

// StartQuery starts timing a query
func (d *DBLogger) StartQuery(ctx context.Context, operation, query string) *QueryTimer {
	return &QueryTimer{
		ctx:       ctx,
		operation: operation,
		query:     query,
		start:     time.Now(),
		logger:    d,
	}
}

// End ends the query timing and logs the result
func (t *QueryTimer) End(err error) {
	duration := time.Since(t.start)
	t.logger.LogQuery(t.ctx, t.operation, t.query, duration, err)
}
