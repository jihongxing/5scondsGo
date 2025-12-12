package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/fiveseconds/server/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

const (
	// BalanceCachePrefix Redis 键前缀
	BalanceCachePrefix = "balance_cache:"
	// BalanceCacheTTL 缓存过期时间
	BalanceCacheTTL = 30 * time.Minute
)

var (
	ErrCacheMiss           = errors.New("cache miss")
	ErrVersionConflict     = errors.New("version conflict")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// CachedBalance 缓存的余额信息
type CachedBalance struct {
	Balance       decimal.Decimal `json:"balance"`
	FrozenBalance decimal.Decimal `json:"frozen_balance"`
	Version       int64           `json:"version"`
	CachedAt      time.Time       `json:"cached_at"`
}

// BalanceCache 余额缓存组件
type BalanceCache struct {
	redis    *redis.Client
	userRepo *repository.UserRepo
	logger   *zap.Logger
}

// NewBalanceCache 创建余额缓存实例
func NewBalanceCache(redisClient *redis.Client, userRepo *repository.UserRepo, logger *zap.Logger) *BalanceCache {
	return &BalanceCache{
		redis:    redisClient,
		userRepo: userRepo,
		logger:   logger.With(zap.String("component", "balance_cache")),
	}
}

// cacheKey 生成缓存键
func (c *BalanceCache) cacheKey(userID int64) string {
	return fmt.Sprintf("%s%d", BalanceCachePrefix, userID)
}

// Get 从缓存获取余额，如果缓存未命中则从数据库加载
func (c *BalanceCache) Get(ctx context.Context, userID int64) (*CachedBalance, error) {
	key := c.cacheKey(userID)

	// 尝试从 Redis 获取
	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var cached CachedBalance
		if err := json.Unmarshal(data, &cached); err == nil {
			return &cached, nil
		}
		c.logger.Warn("Failed to unmarshal cached balance", zap.Error(err))
	} else if !errors.Is(err, redis.Nil) {
		c.logger.Warn("Redis get error", zap.Error(err))
	}

	// 缓存未命中，从数据库加载
	return c.LoadFromDB(ctx, userID)
}

// LoadFromDB 从数据库加载余额并缓存
func (c *BalanceCache) LoadFromDB(ctx context.Context, userID int64) (*CachedBalance, error) {
	user, err := c.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user from db: %w", err)
	}

	cached := &CachedBalance{
		Balance:       user.Balance,
		FrozenBalance: user.FrozenBalance,
		Version:       user.BalanceVersion,
		CachedAt:      time.Now(),
	}

	// 写入缓存
	if err := c.Set(ctx, userID, cached); err != nil {
		c.logger.Warn("Failed to set cache after db load", zap.Error(err))
	}

	return cached, nil
}

// Set 设置缓存
func (c *BalanceCache) Set(ctx context.Context, userID int64, balance *CachedBalance) error {
	key := c.cacheKey(userID)

	data, err := json.Marshal(balance)
	if err != nil {
		return fmt.Errorf("marshal balance: %w", err)
	}

	if err := c.redis.Set(ctx, key, data, BalanceCacheTTL).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	return nil
}

// Invalidate 使缓存失效
func (c *BalanceCache) Invalidate(ctx context.Context, userID int64) error {
	key := c.cacheKey(userID)
	if err := c.redis.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}


// UpdateWithVersion 使用乐观锁更新余额
// 只有当 expectedVersion 与数据库中的版本匹配时才会更新
func (c *BalanceCache) UpdateWithVersion(ctx context.Context, userID int64, delta decimal.Decimal, expectedVersion int64) (*CachedBalance, error) {
	// 使用数据库乐观锁更新
	newBalance, newVersion, err := c.userRepo.UpdateBalanceWithVersion(ctx, userID, delta, expectedVersion)
	if err != nil {
		if errors.Is(err, repository.ErrVersionConflict) {
			return nil, ErrVersionConflict
		}
		if errors.Is(err, repository.ErrInsufficientBalance) {
			return nil, ErrInsufficientBalance
		}
		return nil, fmt.Errorf("update balance with version: %w", err)
	}

	// 更新缓存
	cached := &CachedBalance{
		Balance:       newBalance,
		FrozenBalance: decimal.Zero, // 需要从数据库获取完整信息
		Version:       newVersion,
		CachedAt:      time.Now(),
	}

	// 重新从数据库加载完整信息
	fullCached, err := c.LoadFromDB(ctx, userID)
	if err != nil {
		c.logger.Warn("Failed to reload after update", zap.Error(err))
		return cached, nil
	}

	return fullCached, nil
}

const (
	// MaxRetries 乐观锁重试次数
	MaxRetries = 3
	// RetryDelay 重试间隔
	RetryDelay = 10 * time.Millisecond
)

// DeductBalance 扣除余额（用于下注，带重试机制）
func (c *BalanceCache) DeductBalance(ctx context.Context, userID int64, amount decimal.Decimal) (*CachedBalance, error) {
	var lastErr error
	
	for i := 0; i < MaxRetries; i++ {
		// 获取当前缓存
		cached, err := c.Get(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("get cached balance: %w", err)
		}

		// 检查余额是否足够
		if cached.Balance.LessThan(amount) {
			return nil, ErrInsufficientBalance
		}

		// 使用乐观锁更新
		result, err := c.UpdateWithVersion(ctx, userID, amount.Neg(), cached.Version)
		if err == nil {
			return result, nil
		}
		
		if errors.Is(err, ErrVersionConflict) {
			lastErr = err
			// 版本冲突，等待后重试
			if i < MaxRetries-1 {
				time.Sleep(RetryDelay * time.Duration(i+1))
				// 使缓存失效，强制从数据库重新加载
				_ = c.Invalidate(ctx, userID)
			}
			continue
		}
		
		// 其他错误直接返回
		return nil, err
	}
	
	return nil, fmt.Errorf("deduct balance failed after %d retries: %w", MaxRetries, lastErr)
}

// AddBalance 增加余额（用于奖金发放，带重试机制）
func (c *BalanceCache) AddBalance(ctx context.Context, userID int64, amount decimal.Decimal) (*CachedBalance, error) {
	var lastErr error
	
	for i := 0; i < MaxRetries; i++ {
		// 获取当前缓存
		cached, err := c.Get(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("get cached balance: %w", err)
		}

		// 使用乐观锁更新
		result, err := c.UpdateWithVersion(ctx, userID, amount, cached.Version)
		if err == nil {
			return result, nil
		}
		
		if errors.Is(err, ErrVersionConflict) {
			lastErr = err
			// 版本冲突，等待后重试
			if i < MaxRetries-1 {
				time.Sleep(RetryDelay * time.Duration(i+1))
				// 使缓存失效，强制从数据库重新加载
				_ = c.Invalidate(ctx, userID)
			}
			continue
		}
		
		// 其他错误直接返回
		return nil, err
	}
	
	return nil, fmt.Errorf("add balance failed after %d retries: %w", MaxRetries, lastErr)
}

// RefreshCache 刷新缓存（从数据库重新加载）
func (c *BalanceCache) RefreshCache(ctx context.Context, userID int64) (*CachedBalance, error) {
	// 先使缓存失效
	if err := c.Invalidate(ctx, userID); err != nil {
		c.logger.Warn("Failed to invalidate cache before refresh", zap.Error(err))
	}

	// 从数据库重新加载
	return c.LoadFromDB(ctx, userID)
}

// GetMultiple 批量获取多个用户的余额
func (c *BalanceCache) GetMultiple(ctx context.Context, userIDs []int64) (map[int64]*CachedBalance, error) {
	result := make(map[int64]*CachedBalance)

	// 构建所有键
	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = c.cacheKey(id)
	}

	// 批量获取
	values, err := c.redis.MGet(ctx, keys...).Result()
	if err != nil {
		c.logger.Warn("Redis mget error, falling back to db", zap.Error(err))
		// 降级到数据库
		for _, id := range userIDs {
			cached, err := c.LoadFromDB(ctx, id)
			if err == nil {
				result[id] = cached
			}
		}
		return result, nil
	}

	// 处理结果
	missedIDs := []int64{}
	for i, val := range values {
		userID := userIDs[i]
		if val == nil {
			missedIDs = append(missedIDs, userID)
			continue
		}

		var cached CachedBalance
		if str, ok := val.(string); ok {
			if err := json.Unmarshal([]byte(str), &cached); err == nil {
				result[userID] = &cached
				continue
			}
		}
		missedIDs = append(missedIDs, userID)
	}

	// 从数据库加载缺失的
	for _, id := range missedIDs {
		cached, err := c.LoadFromDB(ctx, id)
		if err == nil {
			result[id] = cached
		}
	}

	return result, nil
}

// IsAvailable 检查缓存服务是否可用
func (c *BalanceCache) IsAvailable(ctx context.Context) bool {
	if c.redis == nil {
		return false
	}
	return c.redis.Ping(ctx).Err() == nil
}
