// Package metrics provides Prometheus metrics collection.
package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
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

	// WebSocket metrics
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

	// Database metrics
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

	// Game metrics
	gameRoundsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "game_rounds_total",
			Help: "Total number of game rounds",
		},
		[]string{"room_id", "status"},
	)

	// Business metrics
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

	// Room event metrics
	roomEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "room_events_total",
			Help: "Total number of room events",
		},
		[]string{"event_type"},
	)

	// Fund anomaly metrics
	fundAnomalyTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "fund_anomaly_total",
			Help: "Total number of fund anomalies detected",
		},
		[]string{"type"},
	)

	// Reconciliation metrics
	reconciliationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reconciliation_total",
			Help: "Total number of reconciliation runs",
		},
		[]string{"status", "period"},
	)
)

// RecordHTTPRequest records an HTTP request metric
func RecordHTTPRequest(method, path string, status int, duration time.Duration) {
	statusStr := strconv.Itoa(status)
	httpRequestsTotal.WithLabelValues(method, path, statusStr).Inc()
	httpRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
}

// RecordWSConnection records WebSocket connection change
func RecordWSConnection(delta int) {
	wsConnectionsActive.Add(float64(delta))
}

// RecordWSMessage records a WebSocket message
func RecordWSMessage(msgType, direction string) {
	wsMessagesTotal.WithLabelValues(msgType, direction).Inc()
}

// RecordDBQuery records a database query metric
func RecordDBQuery(operation string, duration time.Duration, err error) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
	if err != nil {
		dbQueryErrors.WithLabelValues(operation).Inc()
	}
}

// RecordGameRound records a game round event
func RecordGameRound(roomID, status string) {
	gameRoundsTotal.WithLabelValues(roomID, status).Inc()
}


// SetOnlinePlayers sets the online players gauge
func SetOnlinePlayers(count int) {
	onlinePlayersGauge.Set(float64(count))
}

// SetActiveRooms sets the active rooms gauge
func SetActiveRooms(count int) {
	activeRoomsGauge.Set(float64(count))
}

// RecordRoomEvent records a room event
func RecordRoomEvent(eventType string) {
	roomEventsTotal.WithLabelValues(eventType).Inc()
}

// RecordFundAnomaly records a fund anomaly
func RecordFundAnomaly(anomalyType string) {
	fundAnomalyTotal.WithLabelValues(anomalyType).Inc()
}

// RecordReconciliation records a reconciliation run
func RecordReconciliation(status, period string) {
	reconciliationTotal.WithLabelValues(status, period).Inc()
}
