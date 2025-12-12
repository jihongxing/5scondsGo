package service

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/fiveseconds/server/internal/cache"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

var _ = decimal.Zero // 避免未使用的导入警告

const (
	// Redis keys for metrics
	metricsRealtimeKey    = "metrics:realtime"
	metricsAPILatencyKey  = "metrics:api_latency"
	metricsWSLatencyKey   = "metrics:ws_latency"
	metricsDBLatencyKey   = "metrics:db_latency"
	metricsGameCountKey   = "metrics:game_count"
	
	// Latency sample window (5 minutes)
	latencySampleWindow = 5 * time.Minute
)

// MetricsBroadcaster 指标广播接口
type MetricsBroadcaster interface {
	BroadcastToAdmins(msg *model.WSMessage)
}

// MonitoringService 监控服务
type MonitoringService struct {
	metricsRepo *repository.MetricsRepo
	broadcaster MetricsBroadcaster
	logger      *zap.Logger
	
	// 内存中的延迟样本
	apiLatencies []float64
	wsLatencies  []float64
	dbLatencies  []float64
	latencyMu    sync.RWMutex
	
	// 停止信号
	stopCh chan struct{}
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(
	metricsRepo *repository.MetricsRepo,
	broadcaster MetricsBroadcaster,
	logger *zap.Logger,
) *MonitoringService {
	return &MonitoringService{
		metricsRepo:  metricsRepo,
		broadcaster:  broadcaster,
		logger:       logger.With(zap.String("service", "monitoring")),
		apiLatencies: make([]float64, 0, 1000),
		wsLatencies:  make([]float64, 0, 1000),
		dbLatencies:  make([]float64, 0, 1000),
		stopCh:       make(chan struct{}),
	}
}

// Start 启动监控服务
func (s *MonitoringService) Start() {
	// 每分钟保存快照
	go s.snapshotLoop()
	// 每10秒广播指标
	go s.broadcastLoop()
	
	s.logger.Info("Monitoring service started")
}

// Stop 停止监控服务
func (s *MonitoringService) Stop() {
	close(s.stopCh)
	s.logger.Info("Monitoring service stopped")
}

// snapshotLoop 快照循环
func (s *MonitoringService) snapshotLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := s.saveSnapshot(ctx); err != nil {
				s.logger.Error("Failed to save metrics snapshot", zap.Error(err))
			}
			cancel()
		case <-s.stopCh:
			return
		}
	}
}

// broadcastLoop 广播循环
func (s *MonitoringService) broadcastLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			metrics, err := s.GetRealtimeMetrics(ctx)
			cancel()
			
			if err != nil {
				s.logger.Error("Failed to get realtime metrics", zap.Error(err))
				continue
			}
			
			s.broadcastMetrics(metrics)
		case <-s.stopCh:
			return
		}
	}
}

// saveSnapshot 保存快照
func (s *MonitoringService) saveSnapshot(ctx context.Context) error {
	metrics, err := s.GetRealtimeMetrics(ctx)
	if err != nil {
		return err
	}
	
	snapshot := &model.MetricsSnapshot{
		OnlinePlayers:    metrics.OnlinePlayers,
		ActiveRooms:      metrics.ActiveRooms,
		GamesPerMinute:   metrics.GamesPerMinute,
		APILatencyP95:    metrics.APILatencyP95,
		WSLatencyP95:     metrics.WSLatencyP95,
		DBLatencyP95:     metrics.DBLatencyP95,
		DailyActiveUsers: metrics.DailyActiveUsers,
		DailyVolume:      metrics.DailyVolume,
		PlatformRevenue:  metrics.PlatformRevenue,
	}
	
	if err := s.metricsRepo.SaveSnapshot(ctx, snapshot); err != nil {
		return err
	}
	
	s.logger.Debug("Metrics snapshot saved", zap.Int64("id", snapshot.ID))
	return nil
}

// broadcastMetrics 广播指标
func (s *MonitoringService) broadcastMetrics(metrics *model.RealtimeMetrics) {
	if s.broadcaster == nil {
		return
	}
	
	// 检查阈值告警
	alerts := s.checkThresholds(metrics)
	
	s.broadcaster.BroadcastToAdmins(&model.WSMessage{
		Type: model.WSTypeMetricsUpdate,
		Payload: &model.WSMetricsUpdate{
			Metrics:   metrics,
			Alerts:    alerts,
			Timestamp: time.Now().UnixMilli(),
		},
	})
}

// checkThresholds 检查阈值
func (s *MonitoringService) checkThresholds(metrics *model.RealtimeMetrics) []model.MetricAlert {
	var alerts []model.MetricAlert
	threshold := model.DefaultMetricsThreshold
	
	if metrics.APILatencyP95 > threshold.APILatencyP95Max {
		alerts = append(alerts, model.MetricAlert{
			MetricName: "api_latency_p95",
			Value:      metrics.APILatencyP95,
			Threshold:  threshold.APILatencyP95Max,
			Message:    "API延迟P95超过阈值",
		})
	}
	
	if metrics.WSLatencyP95 > threshold.WSLatencyP95Max {
		alerts = append(alerts, model.MetricAlert{
			MetricName: "ws_latency_p95",
			Value:      metrics.WSLatencyP95,
			Threshold:  threshold.WSLatencyP95Max,
			Message:    "WebSocket延迟P95超过阈值",
		})
	}
	
	if metrics.DBLatencyP95 > threshold.DBLatencyP95Max {
		alerts = append(alerts, model.MetricAlert{
			MetricName: "db_latency_p95",
			Value:      metrics.DBLatencyP95,
			Threshold:  threshold.DBLatencyP95Max,
			Message:    "数据库延迟P95超过阈值",
		})
	}
	
	return alerts
}

// GetRealtimeMetrics 获取实时指标
func (s *MonitoringService) GetRealtimeMetrics(ctx context.Context) (*model.RealtimeMetrics, error) {
	// 尝试从 Redis 缓存获取
	cached, err := s.getFromCache(ctx)
	if err == nil && cached != nil {
		return cached, nil
	}
	
	// 从数据库获取
	metrics := &model.RealtimeMetrics{
		Timestamp: time.Now().UnixMilli(),
	}
	
	// 并行获取各项指标
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error
	
	wg.Add(6)
	
	go func() {
		defer wg.Done()
		count, err := s.metricsRepo.GetOnlinePlayersCount(ctx)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metrics.OnlinePlayers = count
		}
		mu.Unlock()
	}()
	
	go func() {
		defer wg.Done()
		count, err := s.metricsRepo.GetActiveRoomsCount(ctx)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metrics.ActiveRooms = count
		}
		mu.Unlock()
	}()
	
	go func() {
		defer wg.Done()
		gpm, err := s.metricsRepo.GetGamesPerMinute(ctx)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metrics.GamesPerMinute = gpm
		}
		mu.Unlock()
	}()
	
	go func() {
		defer wg.Done()
		dau, err := s.metricsRepo.GetDailyActiveUsers(ctx)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metrics.DailyActiveUsers = dau
		}
		mu.Unlock()
	}()
	
	go func() {
		defer wg.Done()
		volume, err := s.metricsRepo.GetDailyVolume(ctx)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metrics.DailyVolume = volume
		}
		mu.Unlock()
	}()
	
	go func() {
		defer wg.Done()
		revenue, err := s.metricsRepo.GetPlatformRevenue(ctx)
		mu.Lock()
		if err != nil {
			errs = append(errs, err)
		} else {
			metrics.PlatformRevenue = revenue
		}
		mu.Unlock()
	}()
	
	wg.Wait()
	
	// 获取延迟指标
	metrics.APILatencyP95 = s.getAPILatencyP95()
	metrics.WSLatencyP95 = s.getWSLatencyP95()
	metrics.DBLatencyP95 = s.getDBLatencyP95()
	
	// 缓存结果
	s.saveToCache(ctx, metrics)
	
	if len(errs) > 0 {
		s.logger.Warn("Some metrics failed to fetch", zap.Int("error_count", len(errs)))
	}
	
	return metrics, nil
}

// GetHistoricalMetrics 获取历史指标
func (s *MonitoringService) GetHistoricalMetrics(ctx context.Context, query *model.MetricsHistoryQuery) ([]*model.MetricsSnapshot, error) {
	var from time.Time
	to := time.Now()
	limit := 1000
	
	switch query.TimeRange {
	case "1h":
		from = to.Add(-1 * time.Hour)
		limit = 60 // 每分钟一条
	case "24h":
		from = to.Add(-24 * time.Hour)
		limit = 288 // 每5分钟一条
	case "7d":
		from = to.Add(-7 * 24 * time.Hour)
		limit = 336 // 每30分钟一条
	case "30d":
		from = to.Add(-30 * 24 * time.Hour)
		limit = 720 // 每小时一条
	default:
		from = to.Add(-1 * time.Hour)
		limit = 60
	}
	
	return s.metricsRepo.GetHistory(ctx, from, to, limit)
}

// RecordAPILatency 记录API延迟
func (s *MonitoringService) RecordAPILatency(latencyMs float64) {
	s.latencyMu.Lock()
	defer s.latencyMu.Unlock()
	
	s.apiLatencies = append(s.apiLatencies, latencyMs)
	if len(s.apiLatencies) > 1000 {
		s.apiLatencies = s.apiLatencies[len(s.apiLatencies)-1000:]
	}
	
	// 同时写入 Redis
	s.recordLatencyToRedis(metricsAPILatencyKey, latencyMs)
}

// RecordWSLatency 记录WebSocket延迟
func (s *MonitoringService) RecordWSLatency(latencyMs float64) {
	s.latencyMu.Lock()
	defer s.latencyMu.Unlock()
	
	s.wsLatencies = append(s.wsLatencies, latencyMs)
	if len(s.wsLatencies) > 1000 {
		s.wsLatencies = s.wsLatencies[len(s.wsLatencies)-1000:]
	}
	
	s.recordLatencyToRedis(metricsWSLatencyKey, latencyMs)
}

// RecordDBLatency 记录数据库延迟
func (s *MonitoringService) RecordDBLatency(latencyMs float64) {
	s.latencyMu.Lock()
	defer s.latencyMu.Unlock()
	
	s.dbLatencies = append(s.dbLatencies, latencyMs)
	if len(s.dbLatencies) > 1000 {
		s.dbLatencies = s.dbLatencies[len(s.dbLatencies)-1000:]
	}
	
	s.recordLatencyToRedis(metricsDBLatencyKey, latencyMs)
}

// recordLatencyToRedis 记录延迟到Redis
func (s *MonitoringService) recordLatencyToRedis(key string, latencyMs float64) {
	if cache.RedisClient == nil {
		return
	}
	
	ctx := context.Background()
	score := float64(time.Now().UnixNano())
	member := strconv.FormatFloat(latencyMs, 'f', 2, 64) + ":" + strconv.FormatInt(time.Now().UnixNano(), 10)
	
	cache.RedisClient.ZAdd(ctx, key, redis.Z{Score: score, Member: member})
	// 清理5分钟前的数据
	cutoff := float64(time.Now().Add(-latencySampleWindow).UnixNano())
	cache.RedisClient.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatFloat(cutoff, 'f', 0, 64))
}

// getAPILatencyP95 获取API延迟P95
func (s *MonitoringService) getAPILatencyP95() float64 {
	return s.calculateP95(s.apiLatencies)
}

// getWSLatencyP95 获取WebSocket延迟P95
func (s *MonitoringService) getWSLatencyP95() float64 {
	return s.calculateP95(s.wsLatencies)
}

// getDBLatencyP95 获取数据库延迟P95
func (s *MonitoringService) getDBLatencyP95() float64 {
	return s.calculateP95(s.dbLatencies)
}

// calculateP95 计算P95
func (s *MonitoringService) calculateP95(latencies []float64) float64 {
	s.latencyMu.RLock()
	defer s.latencyMu.RUnlock()
	
	if len(latencies) == 0 {
		return 0
	}
	
	// 复制并排序
	sorted := make([]float64, len(latencies))
	copy(sorted, latencies)
	sort.Float64s(sorted)
	
	// P95 索引
	idx := int(float64(len(sorted)) * 0.95)
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	
	return sorted[idx]
}

// getFromCache 从缓存获取
func (s *MonitoringService) getFromCache(ctx context.Context) (*model.RealtimeMetrics, error) {
	if cache.RedisClient == nil {
		return nil, nil
	}
	
	data, err := cache.RedisClient.Get(ctx, metricsRealtimeKey).Bytes()
	if err != nil {
		return nil, err
	}
	
	var metrics model.RealtimeMetrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	
	// 检查是否过期（10秒）
	if time.Now().UnixMilli()-metrics.Timestamp > 10000 {
		return nil, nil
	}
	
	return &metrics, nil
}

// saveToCache 保存到缓存
func (s *MonitoringService) saveToCache(ctx context.Context, metrics *model.RealtimeMetrics) {
	if cache.RedisClient == nil {
		return
	}
	
	data, err := json.Marshal(metrics)
	if err != nil {
		return
	}
	
	cache.RedisClient.Set(ctx, metricsRealtimeKey, data, 10*time.Second)
}

// IncrementGameCount 增加游戏计数
func (s *MonitoringService) IncrementGameCount(ctx context.Context) {
	if cache.RedisClient == nil {
		return
	}
	
	key := metricsGameCountKey + ":" + time.Now().Format("200601021504")
	cache.RedisClient.Incr(ctx, key)
	cache.RedisClient.Expire(ctx, key, 10*time.Minute)
}
