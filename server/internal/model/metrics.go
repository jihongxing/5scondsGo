package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	ID               int64           `json:"id" db:"id"`
	OnlinePlayers    int             `json:"online_players" db:"online_players"`
	ActiveRooms      int             `json:"active_rooms" db:"active_rooms"`
	GamesPerMinute   float64         `json:"games_per_minute" db:"games_per_minute"`
	APILatencyP95    float64         `json:"api_latency_p95" db:"api_latency_p95"`
	WSLatencyP95     float64         `json:"ws_latency_p95" db:"ws_latency_p95"`
	DBLatencyP95     float64         `json:"db_latency_p95" db:"db_latency_p95"`
	DailyActiveUsers int             `json:"daily_active_users" db:"daily_active_users"`
	DailyVolume      decimal.Decimal `json:"daily_volume" db:"daily_volume"`
	PlatformRevenue  decimal.Decimal `json:"platform_revenue" db:"platform_revenue"`
	CreatedAt        time.Time       `json:"created_at" db:"created_at"`
}

// RealtimeMetrics 实时指标
type RealtimeMetrics struct {
	OnlinePlayers    int             `json:"online_players"`
	ActiveRooms      int             `json:"active_rooms"`
	GamesPerMinute   float64         `json:"games_per_minute"`
	APILatencyP95    float64         `json:"api_latency_p95"`
	WSLatencyP95     float64         `json:"ws_latency_p95"`
	DBLatencyP95     float64         `json:"db_latency_p95"`
	DailyActiveUsers int             `json:"daily_active_users"`
	DailyVolume      decimal.Decimal `json:"daily_volume"`
	PlatformRevenue  decimal.Decimal `json:"platform_revenue"`
	Timestamp        int64           `json:"timestamp"` // Unix毫秒
}

// MetricsHistoryQuery 指标历史查询
type MetricsHistoryQuery struct {
	TimeRange string `form:"time_range"` // 1h, 24h, 7d, 30d
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=1000"`
}

// MetricsThreshold 指标阈值
type MetricsThreshold struct {
	APILatencyP95Max float64 `json:"api_latency_p95_max"` // 默认 500ms
	WSLatencyP95Max  float64 `json:"ws_latency_p95_max"`  // 默认 200ms
	DBLatencyP95Max  float64 `json:"db_latency_p95_max"`  // 默认 100ms
}

// DefaultMetricsThreshold 默认阈值
var DefaultMetricsThreshold = MetricsThreshold{
	APILatencyP95Max: 500,
	WSLatencyP95Max:  200,
	DBLatencyP95Max:  100,
}

// WSMetricsUpdate WebSocket 指标更新
type WSMetricsUpdate struct {
	Metrics   *RealtimeMetrics `json:"metrics"`
	Alerts    []MetricAlert    `json:"alerts,omitempty"`
	Timestamp int64            `json:"timestamp"`
}

// MetricAlert 指标告警
type MetricAlert struct {
	MetricName string  `json:"metric_name"`
	Value      float64 `json:"value"`
	Threshold  float64 `json:"threshold"`
	Message    string  `json:"message"`
}
