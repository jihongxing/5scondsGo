package repository

import (
	"context"
	"time"

	"github.com/fiveseconds/server/internal/model"
	"github.com/shopspring/decimal"
)

// MetricsRepo 指标仓库
type MetricsRepo struct{}

// NewMetricsRepo 创建指标仓库
func NewMetricsRepo() *MetricsRepo {
	return &MetricsRepo{}
}

// SaveSnapshot 保存指标快照
func (r *MetricsRepo) SaveSnapshot(ctx context.Context, snapshot *model.MetricsSnapshot) error {
	sql := `INSERT INTO metrics_snapshots (
		online_players, active_rooms, games_per_minute,
		api_latency_p95, ws_latency_p95, db_latency_p95,
		daily_active_users, daily_volume, platform_revenue
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, created_at`

	return DB.QueryRow(ctx, sql,
		snapshot.OnlinePlayers, snapshot.ActiveRooms, snapshot.GamesPerMinute,
		snapshot.APILatencyP95, snapshot.WSLatencyP95, snapshot.DBLatencyP95,
		snapshot.DailyActiveUsers, snapshot.DailyVolume, snapshot.PlatformRevenue,
	).Scan(&snapshot.ID, &snapshot.CreatedAt)
}

// GetHistory 获取历史指标
func (r *MetricsRepo) GetHistory(ctx context.Context, from, to time.Time, limit int) ([]*model.MetricsSnapshot, error) {
	sql := `SELECT id, online_players, active_rooms, games_per_minute,
		api_latency_p95, ws_latency_p95, db_latency_p95,
		daily_active_users, daily_volume, platform_revenue, created_at
		FROM metrics_snapshots
		WHERE created_at BETWEEN $1 AND $2
		ORDER BY created_at DESC
		LIMIT $3`

	rows, err := DB.Query(ctx, sql, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []*model.MetricsSnapshot
	for rows.Next() {
		s := &model.MetricsSnapshot{}
		if err := rows.Scan(
			&s.ID, &s.OnlinePlayers, &s.ActiveRooms, &s.GamesPerMinute,
			&s.APILatencyP95, &s.WSLatencyP95, &s.DBLatencyP95,
			&s.DailyActiveUsers, &s.DailyVolume, &s.PlatformRevenue, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, s)
	}
	return snapshots, nil
}

// GetOnlinePlayersCount 获取在线玩家数
// 查询当前在房间内且在线的玩家数量
func (r *MetricsRepo) GetOnlinePlayersCount(ctx context.Context) (int, error) {
	sql := `SELECT COUNT(*) FROM room_players WHERE left_at IS NULL AND is_online = true`
	var count int
	err := DB.QueryRow(ctx, sql).Scan(&count)
	return count, err
}

// GetActiveRoomsCount 获取活跃房间数
func (r *MetricsRepo) GetActiveRoomsCount(ctx context.Context) (int, error) {
	sql := `SELECT COUNT(*) FROM rooms WHERE status = 'active'`
	var count int
	err := DB.QueryRow(ctx, sql).Scan(&count)
	return count, err
}

// GetGamesPerMinute 获取每分钟游戏数
func (r *MetricsRepo) GetGamesPerMinute(ctx context.Context) (float64, error) {
	// 统计过去5分钟的游戏数，然后除以5
	sql := `SELECT COUNT(*) FROM game_rounds 
		WHERE created_at > NOW() - INTERVAL '5 minutes'`
	var count int
	if err := DB.QueryRow(ctx, sql).Scan(&count); err != nil {
		return 0, err
	}
	return float64(count) / 5.0, nil
}

// GetDailyActiveUsers 获取日活用户数
func (r *MetricsRepo) GetDailyActiveUsers(ctx context.Context) (int, error) {
	sql := `SELECT COUNT(DISTINCT user_id) FROM room_players 
		WHERE joined_at > NOW() - INTERVAL '24 hours'`
	var count int
	err := DB.QueryRow(ctx, sql).Scan(&count)
	return count, err
}

// GetDailyVolume 获取日交易量
func (r *MetricsRepo) GetDailyVolume(ctx context.Context) (decimal.Decimal, error) {
	sql := `SELECT COALESCE(SUM(ABS(amount)), 0) FROM balance_transactions 
		WHERE created_at > NOW() - INTERVAL '24 hours'
		AND tx_type IN ('game_bet', 'game_win', 'deposit', 'withdraw')`
	var volume decimal.Decimal
	err := DB.QueryRow(ctx, sql).Scan(&volume)
	return volume, err
}

// GetPlatformRevenue 获取平台收入（今日）
// 从 game_rounds 表汇总已结算回合的平台抽成
func (r *MetricsRepo) GetPlatformRevenue(ctx context.Context) (decimal.Decimal, error) {
	sql := `SELECT COALESCE(SUM(platform_earning), 0) + COALESCE(SUM(residual_amount), 0)
		FROM game_rounds 
		WHERE status = 'settled' 
		AND settled_at > NOW() - INTERVAL '24 hours'`
	var revenue decimal.Decimal
	err := DB.QueryRow(ctx, sql).Scan(&revenue)
	return revenue, err
}

// CleanupOldSnapshots 清理旧快照（保留30天）
func (r *MetricsRepo) CleanupOldSnapshots(ctx context.Context) (int64, error) {
	sql := `DELETE FROM metrics_snapshots WHERE created_at < NOW() - INTERVAL '30 days'`
	tag, err := DB.Exec(ctx, sql)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
