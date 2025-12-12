package service

import (
	"context"
	"strconv"
	"time"

	"github.com/fiveseconds/server/pkg/logger"
	"github.com/fiveseconds/server/pkg/metrics"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

// RoomActivityLogger logs room-related activities
type RoomActivityLogger struct {
	log *logger.Logger
}

// NewRoomActivityLogger creates a new room activity logger
func NewRoomActivityLogger(log *logger.Logger) *RoomActivityLogger {
	return &RoomActivityLogger{
		log: log.With(zap.String("component", "room_activity")),
	}
}

// LogRoomCreated logs when a room is created
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

// LogPlayerJoined logs when a player joins a room
func (r *RoomActivityLogger) LogPlayerJoined(ctx context.Context, roomID int64, userID int64, username string) {
	r.log.WithContext(ctx).Info("player joined room",
		zap.Int64("room_id", roomID),
		zap.Int64("user_id", userID),
		zap.String("username", username),
	)
	metrics.RecordRoomEvent("player_joined")
}


// LogPlayerLeft logs when a player leaves a room
func (r *RoomActivityLogger) LogPlayerLeft(ctx context.Context, roomID int64, userID int64, reason string) {
	r.log.WithContext(ctx).Info("player left room",
		zap.Int64("room_id", roomID),
		zap.Int64("user_id", userID),
		zap.String("reason", reason),
	)
	metrics.RecordRoomEvent("player_left")
}

// LogRoundStarted logs when a game round starts
func (r *RoomActivityLogger) LogRoundStarted(ctx context.Context, roomID int64, roundNumber int, participantCount int, poolAmount decimal.Decimal) {
	r.log.WithContext(ctx).Info("game round started",
		zap.Int64("room_id", roomID),
		zap.Int("round_number", roundNumber),
		zap.Int("participant_count", participantCount),
		zap.String("pool_amount", poolAmount.String()),
	)
	metrics.RecordGameRound(strconv.FormatInt(roomID, 10), "started")
}

// LogRoundSettled logs when a game round is settled
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


// LogRoundFailed logs when a game round fails
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

// LogRoomStatusChanged logs when a room status changes
func (r *RoomActivityLogger) LogRoomStatusChanged(ctx context.Context, roomID int64, oldStatus, newStatus, reason string) {
	r.log.WithContext(ctx).Info("room status changed",
		zap.Int64("room_id", roomID),
		zap.String("old_status", oldStatus),
		zap.String("new_status", newStatus),
		zap.String("reason", reason),
	)
	metrics.RecordRoomEvent("status_changed")
}
