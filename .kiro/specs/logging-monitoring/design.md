# Design Document: Logging and Monitoring System

## Overview

本设计文档描述 5SecondsGo 游戏平台的完整日志和监控系统实现。系统基于现有的 Zap 日志库和 Prometheus 指标端点，扩展实现结构化日志、分布式追踪、Loki 集成、Grafana 仪表盘和告警规则。

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        5SecondsGo Server                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Handler   │  │   Service   │  │ Repository  │              │
│  │  (HTTP/WS)  │  │   Layer     │  │   Layer     │              │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘              │
│         │                │                │                      │
│         ▼                ▼                ▼                      │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Observability Layer                         │    │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐            │    │
│  │  │  Logger   │  │  Metrics  │  │  Tracer   │            │    │
│  │  │  (Zap)    │  │(Prometheus)│  │(Trace ID) │            │    │
│  │  └─────┬─────┘  └─────┬─────┘  └─────┬─────┘            │    │
│  └────────┼──────────────┼──────────────┼──────────────────┘    │
│           │              │              │                        │
└───────────┼──────────────┼──────────────┼────────────────────────┘
            │              │              │
            ▼              ▼              ▼
     ┌──────────┐   ┌──────────┐   ┌──────────┐
     │ Promtail │   │Prometheus│   │  Stdout  │
     └────┬─────┘   └────┬─────┘   └──────────┘
          │              │
          ▼              ▼
     ┌──────────┐   ┌──────────┐
     │   Loki   │   │ Grafana  │◄─── Alertmanager
     └────┬─────┘   └──────────┘
          │              ▲
          └──────────────┘
```

## Components and Interfaces

### 1. Trace Context Manager

```go
// pkg/trace/context.go

package trace

import (
    "context"
    "github.com/google/uuid"
)

type contextKey string

const (
    TraceIDKey   contextKey = "trace_id"
    SessionIDKey contextKey = "session_id"
    UserIDKey    contextKey = "user_id"
)

// TraceContext 追踪上下文
type TraceContext struct {
    TraceID   string
    SessionID string
    UserID    int64
}

// NewTraceID 生成新的 Trace ID
func NewTraceID() string {
    return uuid.New().String()
}

// WithTraceID 将 Trace ID 注入 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 从 context 获取 Trace ID
func GetTraceID(ctx context.Context) string {
    if v := ctx.Value(TraceIDKey); v != nil {
        return v.(string)
    }
    return ""
}

// WithTraceContext 注入完整追踪上下文
func WithTraceContext(ctx context.Context, tc *TraceContext) context.Context {
    ctx = context.WithValue(ctx, TraceIDKey, tc.TraceID)
    ctx = context.WithValue(ctx, SessionIDKey, tc.SessionID)
    if tc.UserID > 0 {
        ctx = context.WithValue(ctx, UserIDKey, tc.UserID)
    }
    return ctx
}
```

### 2. Structured Logger Wrapper

```go
// pkg/logger/logger.go

package logger

import (
    "context"
    "sync/atomic"
    
    "github.com/fiveseconds/server/pkg/trace"
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

// Logger 结构化日志包装器
type Logger struct {
    base  *zap.Logger
    level *atomic.Value // 存储当前日志级别
}

// Config 日志配置
type Config struct {
    Level       string `yaml:"level"`       // debug/info/warn/error
    Format      string `yaml:"format"`      // json/console
    ServiceName string `yaml:"service_name"`
    Environment string `yaml:"environment"`
}

// New 创建新的 Logger
func New(cfg *Config) (*Logger, error) {
    level := parseLevel(cfg.Level)
    
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "timestamp",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        FunctionKey:    zapcore.OmitKey,
        MessageKey:     "message",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.MillisDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
    
    config := zap.Config{
        Level:            zap.NewAtomicLevelAt(level),
        Development:      cfg.Environment != "production",
        Encoding:         cfg.Format,
        EncoderConfig:    encoderConfig,
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
        InitialFields: map[string]interface{}{
            "service": cfg.ServiceName,
            "env":     cfg.Environment,
        },
    }
    
    base, err := config.Build(zap.AddCallerSkip(1))
    if err != nil {
        return nil, err
    }
    
    levelVal := &atomic.Value{}
    levelVal.Store(level)
    
    return &Logger{base: base, level: levelVal}, nil
}

// WithContext 从 context 提取追踪信息并添加到日志字段
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
    fields := []zap.Field{}
    
    if traceID := trace.GetTraceID(ctx); traceID != "" {
        fields = append(fields, zap.String("trace_id", traceID))
    }
    if sessionID := ctx.Value(trace.SessionIDKey); sessionID != nil {
        fields = append(fields, zap.String("session_id", sessionID.(string)))
    }
    if userID := ctx.Value(trace.UserIDKey); userID != nil {
        fields = append(fields, zap.Int64("user_id", userID.(int64)))
    }
    
    return l.base.With(fields...)
}

// SetLevel 动态设置日志级别
func (l *Logger) SetLevel(levelStr string) error {
    level := parseLevel(levelStr)
    l.level.Store(level)
    // 重建 logger 以应用新级别
    return nil
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() string {
    level := l.level.Load().(zapcore.Level)
    return level.String()
}
```

### 3. HTTP Logging Middleware

```go
// internal/middleware/logging.go

package middleware

import (
    "time"
    
    "github.com/fiveseconds/server/pkg/logger"
    "github.com/fiveseconds/server/pkg/metrics"
    "github.com/fiveseconds/server/pkg/trace"
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

const (
    TraceIDHeader = "X-Trace-ID"
)

// RequestLogging 请求日志中间件
func RequestLogging(log *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 获取或生成 Trace ID
        traceID := c.GetHeader(TraceIDHeader)
        if traceID == "" {
            traceID = trace.NewTraceID()
        }
        
        // 注入 context
        ctx := trace.WithTraceID(c.Request.Context(), traceID)
        c.Request = c.Request.WithContext(ctx)
        
        // 设置响应头
        c.Header(TraceIDHeader, traceID)
        
        // 处理请求
        c.Next()
        
        // 计算延迟
        latency := time.Since(start)
        status := c.Writer.Status()
        
        // 记录指标
        metrics.RecordHTTPRequest(c.Request.Method, c.FullPath(), status, latency)
        
        // 记录日志
        fields := []zap.Field{
            zap.String("trace_id", traceID),
            zap.String("method", c.Request.Method),
            zap.String("path", c.Request.URL.Path),
            zap.Int("status", status),
            zap.Duration("latency", latency),
            zap.String("client_ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        }
        
        if status >= 500 {
            log.WithContext(ctx).Error("request completed with error", fields...)
        } else if status >= 400 {
            log.WithContext(ctx).Warn("request completed with client error", fields...)
        } else {
            log.WithContext(ctx).Info("request completed", fields...)
        }
    }
}
```

### 4. Database Query Logger

```go
// internal/repository/db_logger.go

package repository

import (
    "context"
    "time"
    
    "github.com/fiveseconds/server/pkg/logger"
    "github.com/fiveseconds/server/pkg/metrics"
    "github.com/fiveseconds/server/pkg/trace"
    "go.uber.org/zap"
)

const SlowQueryThreshold = 100 * time.Millisecond

// DBLogger 数据库查询日志器
type DBLogger struct {
    log *logger.Logger
}

// NewDBLogger 创建数据库日志器
func NewDBLogger(log *logger.Logger) *DBLogger {
    return &DBLogger{log: log}
}

// LogQuery 记录查询
func (d *DBLogger) LogQuery(ctx context.Context, operation string, query string, duration time.Duration, err error) {
    // 记录指标
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
```

### 5. Prometheus Metrics Collector

```go
// pkg/metrics/metrics.go

package metrics

import (
    "time"
    
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP 请求指标
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )
    
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
        },
        []string{"method", "path"},
    )
    
    // WebSocket 指标
    wsConnectionsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "websocket_connections_active",
            Help: "Number of active WebSocket connections",
        },
    )
    
    wsMessagesTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "websocket_messages_total",
            Help: "Total number of WebSocket messages",
        },
        []string{"type", "direction"},
    )
    
    // 数据库指标
    dbQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration in seconds",
            Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
        },
        []string{"operation"},
    )
    
    dbQueryErrors = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_query_errors_total",
            Help: "Total number of database query errors",
        },
        []string{"operation"},
    )
    
    // 业务指标
    gameRoundsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "game_rounds_total",
            Help: "Total number of game rounds",
        },
        []string{"room_id", "status"},
    )
    
    onlinePlayersGauge = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "online_players",
            Help: "Number of online players",
        },
    )
    
    activeRoomsGauge = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "active_rooms",
            Help: "Number of active rooms",
        },
    )
)

// RecordHTTPRequest 记录 HTTP 请求
func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
    statusStr := strconv.Itoa(status)
    httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
    httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordWSConnection WebSocket 连接变化
func RecordWSConnection(delta int) {
    wsConnectionsActive.Add(float64(delta))
}

// RecordWSMessage 记录 WebSocket 消息
func RecordWSMessage(msgType, direction string) {
    wsMessagesTotal.WithLabelValues(msgType, direction).Inc()
}

// RecordDBQuery 记录数据库查询
func RecordDBQuery(operation string, duration time.Duration, err error) {
    dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
    if err != nil {
        dbQueryErrors.WithLabelValues(operation).Inc()
    }
}

// RecordGameRound 记录游戏回合
func RecordGameRound(roomID, status string) {
    gameRoundsTotal.WithLabelValues(roomID, status).Inc()
}

// SetOnlinePlayers 设置在线玩家数
func SetOnlinePlayers(count int) {
    onlinePlayersGauge.Set(float64(count))
}

// SetActiveRooms 设置活跃房间数
func SetActiveRooms(count int) {
    activeRoomsGauge.Set(float64(count))
}
```

### 6. Log Level API Handler

```go
// internal/handler/admin_log_handler.go

package handler

import (
    "net/http"
    
    "github.com/fiveseconds/server/pkg/logger"
    "github.com/gin-gonic/gin"
)

// LogLevelHandler 日志级别处理器
type LogLevelHandler struct {
    logger *logger.Logger
}

// NewLogLevelHandler 创建日志级别处理器
func NewLogLevelHandler(log *logger.Logger) *LogLevelHandler {
    return &LogLevelHandler{logger: log}
}

// GetLogLevel 获取当前日志级别
func (h *LogLevelHandler) GetLogLevel(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "level": h.logger.GetLevel(),
    })
}

// SetLogLevel 设置日志级别
func (h *LogLevelHandler) SetLogLevel(c *gin.Context) {
    var req struct {
        Level string `json:"level" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    validLevels := map[string]bool{
        "debug": true, "info": true, "warn": true, "error": true,
    }
    
    if !validLevels[req.Level] {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "invalid log level, must be one of: debug, info, warn, error",
        })
        return
    }
    
    oldLevel := h.logger.GetLevel()
    if err := h.logger.SetLevel(req.Level); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    h.logger.WithContext(c.Request.Context()).Info("log level changed",
        zap.String("old_level", oldLevel),
        zap.String("new_level", req.Level),
    )
    
    c.JSON(http.StatusOK, gin.H{
        "old_level": oldLevel,
        "new_level": req.Level,
    })
}
```

### 7. Room Activity Logger

```go
// internal/service/room_logger.go

package service

import (
    "context"
    
    "github.com/fiveseconds/server/pkg/logger"
    "github.com/fiveseconds/server/pkg/metrics"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// RoomActivityLogger 房间活动日志器
type RoomActivityLogger struct {
    log *logger.Logger
}

// NewRoomActivityLogger 创建房间活动日志器
func NewRoomActivityLogger(log *logger.Logger) *RoomActivityLogger {
    return &RoomActivityLogger{log: log.With(zap.String("component", "room_activity"))}
}

// LogRoomCreated 记录房间创建
func (r *RoomActivityLogger) LogRoomCreated(ctx context.Context, roomID int64, ownerID int64, code string, betAmount decimal.Decimal, config map[string]interface{}) {
    r.log.WithContext(ctx).Info("room created",
        zap.Int64("room_id", roomID),
        zap.Int64("owner_id", ownerID),
        zap.String("room_code", code),
        zap.String("bet_amount", betAmount.String()),
        zap.Any("config", config),
    )
    metrics.RecordRoomEvent("created")
}

// LogPlayerJoined 记录玩家加入
func (r *RoomActivityLogger) LogPlayerJoined(ctx context.Context, roomID int64, userID int64, username string) {
    r.log.WithContext(ctx).Info("player joined room",
        zap.Int64("room_id", roomID),
        zap.Int64("user_id", userID),
        zap.String("username", username),
    )
    metrics.RecordRoomEvent("player_joined")
}

// LogPlayerLeft 记录玩家离开
func (r *RoomActivityLogger) LogPlayerLeft(ctx context.Context, roomID int64, userID int64, reason string) {
    r.log.WithContext(ctx).Info("player left room",
        zap.Int64("room_id", roomID),
        zap.Int64("user_id", userID),
        zap.String("reason", reason),
    )
    metrics.RecordRoomEvent("player_left")
}

// LogRoundStarted 记录回合开始
func (r *RoomActivityLogger) LogRoundStarted(ctx context.Context, roomID int64, roundNumber int, participantCount int, poolAmount decimal.Decimal) {
    r.log.WithContext(ctx).Info("game round started",
        zap.Int64("room_id", roomID),
        zap.Int("round_number", roundNumber),
        zap.Int("participant_count", participantCount),
        zap.String("pool_amount", poolAmount.String()),
    )
    metrics.RecordGameRound(strconv.FormatInt(roomID, 10), "started")
}

// LogRoundSettled 记录回合结算
func (r *RoomActivityLogger) LogRoundSettled(ctx context.Context, roomID int64, roundNumber int, winnerIDs []int64, prizePerWinner decimal.Decimal, ownerEarning decimal.Decimal, platformEarning decimal.Decimal, settlementTime time.Duration) {
    r.log.WithContext(ctx).Info("game round settled",
        zap.Int64("room_id", roomID),
        zap.Int("round_number", roundNumber),
        zap.Int64s("winner_ids", winnerIDs),
        zap.String("prize_per_winner", prizePerWinner.String()),
        zap.String("owner_earning", ownerEarning.String()),
        zap.String("platform_earning", platformEarning.String()),
        zap.Duration("settlement_time", settlementTime),
    )
    metrics.RecordGameRound(strconv.FormatInt(roomID, 10), "settled")
}

// LogRoundFailed 记录回合失败
func (r *RoomActivityLogger) LogRoundFailed(ctx context.Context, roomID int64, roundNumber int, reason string, refundedUserIDs []int64, refundAmount decimal.Decimal) {
    r.log.WithContext(ctx).Warn("game round failed",
        zap.Int64("room_id", roomID),
        zap.Int("round_number", roundNumber),
        zap.String("failure_reason", reason),
        zap.Int64s("refunded_user_ids", refundedUserIDs),
        zap.String("refund_amount", refundAmount.String()),
    )
    metrics.RecordGameRound(strconv.FormatInt(roomID, 10), "failed")
}

// LogRoomStatusChanged 记录房间状态变更
func (r *RoomActivityLogger) LogRoomStatusChanged(ctx context.Context, roomID int64, oldStatus, newStatus, reason string) {
    r.log.WithContext(ctx).Info("room status changed",
        zap.Int64("room_id", roomID),
        zap.String("old_status", oldStatus),
        zap.String("new_status", newStatus),
        zap.String("reason", reason),
    )
    metrics.RecordRoomEvent("status_changed")
}
```

### 8. Fund Anomaly Logger

```go
// internal/service/fund_anomaly_logger.go

package service

import (
    "context"
    
    "github.com/fiveseconds/server/pkg/logger"
    "github.com/fiveseconds/server/pkg/metrics"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

const (
    LargeTransactionThreshold = 10000
    ConsecutiveWinThreshold   = 10
    HighWinRateThreshold      = 0.80
    HighWinRateMinRounds      = 50
)

// FundAnomalyLogger 资金异常日志器
type FundAnomalyLogger struct {
    log          *logger.Logger
    alertManager *AlertManager
}

// NewFundAnomalyLogger 创建资金异常日志器
func NewFundAnomalyLogger(log *logger.Logger, alertManager *AlertManager) *FundAnomalyLogger {
    return &FundAnomalyLogger{
        log:          log.With(zap.String("component", "fund_anomaly")),
        alertManager: alertManager,
    }
}

// LogNegativeBalance 记录负余额
func (f *FundAnomalyLogger) LogNegativeBalance(ctx context.Context, userID int64, balance decimal.Decimal, lastTxID int64, lastTxType string) {
    f.log.WithContext(ctx).Error("CRITICAL: negative balance detected",
        zap.Int64("user_id", userID),
        zap.String("balance", balance.String()),
        zap.Int64("last_tx_id", lastTxID),
        zap.String("last_tx_type", lastTxType),
    )
    metrics.RecordFundAnomaly("negative_balance")
    
    // 触发告警
    f.alertManager.CreateAlert(ctx, &Alert{
        Type:     AlertTypeNegativeBalance,
        Severity: SeverityCritical,
        Title:    "Negative Balance Detected",
        Details: map[string]interface{}{
            "user_id":      userID,
            "balance":      balance.String(),
            "last_tx_id":   lastTxID,
            "last_tx_type": lastTxType,
        },
    })
}

// LogLargeTransaction 记录大额交易
func (f *FundAnomalyLogger) LogLargeTransaction(ctx context.Context, userID int64, amount decimal.Decimal, txType string, txID int64) {
    f.log.WithContext(ctx).Warn("large transaction detected",
        zap.Int64("user_id", userID),
        zap.String("amount", amount.String()),
        zap.String("tx_type", txType),
        zap.Int64("tx_id", txID),
    )
    metrics.RecordFundAnomaly("large_transaction")
}

// LogConsecutiveWins 记录连续获胜
func (f *FundAnomalyLogger) LogConsecutiveWins(ctx context.Context, userID int64, winStreak int, roomID int64) {
    f.log.WithContext(ctx).Warn("consecutive wins detected",
        zap.Int64("user_id", userID),
        zap.Int("win_streak", winStreak),
        zap.Int64("room_id", roomID),
    )
    metrics.RecordFundAnomaly("consecutive_wins")
}

// LogHighWinRate 记录高胜率
func (f *FundAnomalyLogger) LogHighWinRate(ctx context.Context, userID int64, winRate float64, totalRounds int, roomID int64) {
    f.log.WithContext(ctx).Warn("high win rate detected",
        zap.Int64("user_id", userID),
        zap.Float64("win_rate", winRate),
        zap.Int("total_rounds", totalRounds),
        zap.Int64("room_id", roomID),
    )
    metrics.RecordFundAnomaly("high_win_rate")
}

// LogDuplicateDeviceFingerprint 记录重复设备指纹
func (f *FundAnomalyLogger) LogDuplicateDeviceFingerprint(ctx context.Context, userIDs []int64, fingerprint string) {
    f.log.WithContext(ctx).Warn("duplicate device fingerprint detected",
        zap.Int64s("user_ids", userIDs),
        zap.String("fingerprint", fingerprint),
    )
    metrics.RecordFundAnomaly("duplicate_fingerprint")
}

// LogInsufficientCustodyQuota 记录托管额度不足
func (f *FundAnomalyLogger) LogInsufficientCustodyQuota(ctx context.Context, ownerID int64, quota decimal.Decimal, requestedAmount decimal.Decimal) {
    f.log.WithContext(ctx).Warn("insufficient custody quota for withdrawal",
        zap.Int64("owner_id", ownerID),
        zap.String("custody_quota", quota.String()),
        zap.String("requested_amount", requestedAmount.String()),
    )
    metrics.RecordFundAnomaly("insufficient_custody")
}
```

### 9. Reconciliation Logger

```go
// internal/service/reconciliation_logger.go

package service

import (
    "context"
    "time"
    
    "github.com/fiveseconds/server/pkg/logger"
    "github.com/fiveseconds/server/pkg/metrics"
    "github.com/shopspring/decimal"
    "go.uber.org/zap"
)

// ReconciliationLogger 对账日志器
type ReconciliationLogger struct {
    log          *logger.Logger
    alertManager *AlertManager
}

// NewReconciliationLogger 创建对账日志器
func NewReconciliationLogger(log *logger.Logger, alertManager *AlertManager) *ReconciliationLogger {
    return &ReconciliationLogger{
        log:          log.With(zap.String("component", "reconciliation")),
        alertManager: alertManager,
    }
}

// ReconciliationResult 对账结果
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

// LogReconciliationResult 记录对账结果
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
        
        // 触发告警
        r.alertManager.CreateAlert(ctx, &Alert{
            Type:     AlertTypeFundImbalance,
            Severity: SeverityCritical,
            Title:    "Fund Conservation Imbalance Detected",
            Details: map[string]interface{}{
                "period_type":          result.PeriodType,
                "difference":           result.Difference.String(),
                "total_player_balance": result.TotalPlayerBalance.String(),
                "platform_balance":     result.PlatformBalance.String(),
            },
        })
    }
}

// OwnerReconciliationResult 房主对账结果
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

// LogOwnerReconciliation 记录房主对账结果
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

// LogReconciliationError 记录对账错误
func (r *ReconciliationLogger) LogReconciliationError(ctx context.Context, periodType string, err error, retryCount int) {
    r.log.WithContext(ctx).Error("reconciliation failed",
        zap.String("period_type", periodType),
        zap.Error(err),
        zap.Int("retry_count", retryCount),
    )
    metrics.RecordReconciliation("error", periodType)
}

// LogReconciliationStarted 记录对账开始
func (r *ReconciliationLogger) LogReconciliationStarted(ctx context.Context, periodType string, periodStart, periodEnd time.Time) {
    r.log.WithContext(ctx).Info("reconciliation started",
        zap.String("period_type", periodType),
        zap.Time("period_start", periodStart),
        zap.Time("period_end", periodEnd),
    )
}
```

## Data Models

### Log Entry Format (Loki Compatible)

```json
{
    "timestamp": "2025-12-09T10:30:00.000Z",
    "level": "info",
    "message": "request completed",
    "service": "fiveseconds",
    "env": "production",
    "trace_id": "550e8400-e29b-41d4-a716-446655440000",
    "method": "POST",
    "path": "/api/rooms/ABC123/join",
    "status": 200,
    "latency": 45.23,
    "client_ip": "192.168.1.100",
    "user_id": 12345
}
```

### Metrics Snapshot Model

```go
type MetricsSnapshot struct {
    ID               int64           `json:"id"`
    Timestamp        time.Time       `json:"timestamp"`
    OnlinePlayers    int             `json:"online_players"`
    ActiveRooms      int             `json:"active_rooms"`
    GamesPerMinute   float64         `json:"games_per_minute"`
    APILatencyP95    float64         `json:"api_latency_p95"`
    WSLatencyP95     float64         `json:"ws_latency_p95"`
    DBLatencyP95     float64         `json:"db_latency_p95"`
    DailyActiveUsers int             `json:"daily_active_users"`
    DailyVolume      decimal.Decimal `json:"daily_volume"`
    PlatformRevenue  decimal.Decimal `json:"platform_revenue"`
    ErrorRate        float64         `json:"error_rate"`
}
```



## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Trace ID Uniqueness
*For any* set of HTTP requests processed by the system, all generated trace IDs SHALL be unique (no duplicates).
**Validates: Requirements 1.1**

### Property 2: Log Entry Structure Completeness
*For any* log entry produced during request processing, the entry SHALL contain: timestamp, level, message, service name, environment, and trace_id fields in valid JSON format.
**Validates: Requirements 1.2, 3.1, 3.3**

### Property 3: Request Log Field Completeness
*For any* completed HTTP request, the log entry SHALL contain method, path, status code, latency (as positive duration), and trace_id.
**Validates: Requirements 1.3**

### Property 4: Slow Query Detection Threshold
*For any* database query with execution time greater than 100ms, the system SHALL log it as a slow query; for queries under 100ms, no slow query log SHALL be produced.
**Validates: Requirements 2.2**

### Property 5: P95 Latency Calculation Correctness
*For any* sequence of recorded latency values, the P95 calculation SHALL return a value such that 95% of the recorded values are less than or equal to it.
**Validates: Requirements 2.3**

### Property 6: HTTP Metrics Recording
*For any* completed HTTP request, the Prometheus histogram SHALL be updated with the correct method and path labels, and the recorded duration SHALL match the actual request latency.
**Validates: Requirements 4.1**

### Property 7: WebSocket Connection Gauge Accuracy
*For any* sequence of WebSocket connection and disconnection events, the active connections gauge SHALL equal the number of currently open connections.
**Validates: Requirements 4.2**

### Property 8: Database Query Metrics Recording
*For any* database query execution, the Prometheus histogram SHALL be updated with the operation label and the recorded duration SHALL be positive.
**Validates: Requirements 4.4**

### Property 9: Log Level Change Validity
*For any* valid log level (debug, info, warn, error), calling SetLevel SHALL change the current level; for any invalid level string, the current level SHALL remain unchanged and an error SHALL be returned.
**Validates: Requirements 7.1, 7.3**

### Property 10: Trace ID Propagation
*For any* HTTP request with an X-Trace-ID header, the system SHALL use that trace ID; for requests without the header, a new unique trace ID SHALL be generated. In both cases, the trace ID SHALL be preserved in spawned goroutines.
**Validates: Requirements 8.1, 8.3**

### Property 11: WebSocket Session ID Persistence
*For any* WebSocket connection, all messages processed on that connection SHALL have the same session ID.
**Validates: Requirements 8.4**

### Property 12: Room Activity Log Completeness
*For any* room lifecycle event (create, join, leave, round start, round settle, round fail, status change), the log entry SHALL contain the room ID and all event-specific required fields.
**Validates: Requirements 9.1, 9.2, 9.3, 9.4, 9.5, 9.6**

### Property 13: Fund Anomaly Detection Thresholds
*For any* transaction with amount greater than 10000, a large transaction warning SHALL be logged; for any player with balance less than 0, a critical negative balance alert SHALL be logged.
**Validates: Requirements 10.1, 10.2**

### Property 14: Consecutive Win Detection
*For any* player with more than 10 consecutive wins, a warning log SHALL be produced with the win streak count and room ID.
**Validates: Requirements 10.3**

### Property 15: Reconciliation Log Completeness
*For any* reconciliation execution (2-hour or daily), the log SHALL contain period type, start time, end time, all balance totals, difference amount, and balanced status.
**Validates: Requirements 11.1, 11.3, 11.4**

### Property 16: Reconciliation Imbalance Alert
*For any* reconciliation that detects an imbalance (difference != 0), a critical alert SHALL be created with the discrepancy details.
**Validates: Requirements 11.2**

## Error Handling

### Log Write Failures
- If stdout write fails, buffer logs in memory (max 1000 entries)
- Retry writes with exponential backoff
- Drop oldest logs if buffer is full

### Metrics Collection Failures
- If Prometheus registry fails, log error and continue
- Metrics collection should not block request processing
- Use default values (0) for failed metric reads

### Trace Context Propagation Failures
- If context is nil, generate new trace ID
- If trace ID extraction fails, generate new trace ID
- Log warning for context propagation failures

## Testing Strategy

### Unit Testing
- Test trace ID generation uniqueness
- Test log entry JSON structure
- Test P95 calculation algorithm
- Test log level validation
- Test slow query threshold detection

### Property-Based Testing
Using `github.com/leanovate/gopter`:

1. **Trace ID Uniqueness Property Test**
   - Generate random number of requests
   - Verify all trace IDs are unique

2. **Log Structure Property Test**
   - Generate random log entries
   - Verify JSON structure contains required fields

3. **P95 Calculation Property Test**
   - Generate random latency sequences
   - Verify P95 value satisfies the 95th percentile definition

4. **Log Level Change Property Test**
   - Generate random valid/invalid level strings
   - Verify correct behavior for each case

5. **WebSocket Gauge Property Test**
   - Generate random connect/disconnect sequences
   - Verify gauge equals net connections

6. **Room Activity Log Property Test**
   - Generate random room events
   - Verify log entries contain all required fields

7. **Fund Anomaly Detection Property Test**
   - Generate random transactions with various amounts
   - Verify large transactions (>10000) trigger warnings
   - Verify negative balances trigger critical alerts

8. **Reconciliation Log Property Test**
   - Generate random reconciliation results
   - Verify log entries contain all required fields
   - Verify imbalances trigger alerts

### Integration Testing
- Test end-to-end request logging with trace ID
- Test Prometheus metrics endpoint
- Test log level API endpoint
- Test database query logging
- Test room activity logging during game flow
- Test fund anomaly detection with real transactions
- Test reconciliation logging with mock data

## Configuration Files

### Promtail Configuration (promtail.yaml)

```yaml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: fiveseconds
    static_configs:
      - targets:
          - localhost
        labels:
          job: fiveseconds
          __path__: /var/log/fiveseconds/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            trace_id: trace_id
            service: service
      - labels:
          level:
          service:
```

### Prometheus Alerting Rules (alerts.yaml)

```yaml
groups:
  - name: fiveseconds
    rules:
      - alert: HighAPILatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High API latency detected"
          description: "API P95 latency is above 500ms for 5 minutes"

      - alert: CriticalAPILatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Critical API latency"
          description: "API P95 latency is above 1s for 2 minutes"

      - alert: HighErrorRate
        expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
          description: "Error rate is above 1% for 5 minutes"

      - alert: CriticalErrorRate
        expr: sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m])) > 0.05
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "Critical error rate"
          description: "Error rate is above 5% for 2 minutes"

      - alert: HighDBConnectionUsage
        expr: pg_stat_activity_count / pg_settings_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High database connection usage"
          description: "Database connection pool utilization is above 80%"
```

### Grafana Dashboard JSON

Dashboard JSON files will be provided in `deploy/grafana/dashboards/`:
- `system-overview.json` - Request rate, error rate, latency percentiles
- `business-metrics.json` - Online players, active rooms, daily volume
- `infrastructure.json` - CPU, memory, DB connections, Redis connections

