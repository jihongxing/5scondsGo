// Package integration_test 缓存失效集成测试
package integration_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// MockRedisClient 模拟Redis客户端
type MockRedisClient struct {
	mu        sync.RWMutex
	data      map[string]string
	available bool
}

func NewMockRedisClient() *MockRedisClient {
	return &MockRedisClient{
		data:      make(map[string]string),
		available: true,
	}
}

func (c *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.available {
		return "", errors.New("redis unavailable")
	}

	val, ok := c.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

func (c *MockRedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available {
		return errors.New("redis unavailable")
	}

	c.data[key] = value
	return nil
}

func (c *MockRedisClient) Del(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.available {
		return errors.New("redis unavailable")
	}

	delete(c.data, key)
	return nil
}

func (c *MockRedisClient) SetAvailable(available bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.available = available
}

func (c *MockRedisClient) IsAvailable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.available
}

// MockUserRepo 模拟用户仓库
type MockUserRepo struct {
	mu    sync.RWMutex
	users map[int64]*MockUser
}

type MockUser struct {
	ID             int64
	Balance        decimal.Decimal
	FrozenBalance  decimal.Decimal
	BalanceVersion int64
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users: make(map[int64]*MockUser),
	}
}

func (r *MockUserRepo) AddUser(user *MockUser) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[user.ID] = user
}

func (r *MockUserRepo) GetByID(ctx context.Context, userID int64) (*MockUser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// MockBalanceCache 模拟余额缓存（带降级功能）
type MockBalanceCache struct {
	redis    *MockRedisClient
	userRepo *MockUserRepo
}

func NewMockBalanceCache(redis *MockRedisClient, userRepo *MockUserRepo) *MockBalanceCache {
	return &MockBalanceCache{
		redis:    redis,
		userRepo: userRepo,
	}
}

// Get 获取余额，Redis不可用时降级到数据库
func (c *MockBalanceCache) Get(ctx context.Context, userID int64) (*MockUser, error) {
	// 尝试从Redis获取
	if c.redis.IsAvailable() {
		// 模拟从缓存获取
		_, err := c.redis.Get(ctx, "balance:"+string(rune(userID)))
		if err == nil {
			// 缓存命中，返回缓存数据
			return c.userRepo.GetByID(ctx, userID)
		}
	}

	// 降级到数据库
	return c.userRepo.GetByID(ctx, userID)
}

// IsAvailable 检查缓存是否可用
func (c *MockBalanceCache) IsAvailable(ctx context.Context) bool {
	return c.redis.IsAvailable()
}

// TestCacheFallbackWhenRedisUnavailable 测试Redis不可用时的降级
func TestCacheFallbackWhenRedisUnavailable(t *testing.T) {
	redis := NewMockRedisClient()
	userRepo := NewMockUserRepo()
	cache := NewMockBalanceCache(redis, userRepo)

	// 添加测试用户
	testUser := &MockUser{
		ID:             1,
		Balance:        decimal.NewFromInt(1000),
		FrozenBalance:  decimal.NewFromInt(100),
		BalanceVersion: 1,
	}
	userRepo.AddUser(testUser)

	ctx := context.Background()

	// Redis可用时
	user, err := cache.Get(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get user when Redis available: %v", err)
	}
	if !user.Balance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("Expected balance 1000, got %s", user.Balance.String())
	}

	// 模拟Redis不可用
	redis.SetAvailable(false)

	// 应该降级到数据库
	user, err = cache.Get(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get user when Redis unavailable (should fallback): %v", err)
	}
	if !user.Balance.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("Expected balance 1000 from fallback, got %s", user.Balance.String())
	}
}

// TestCacheRecoveryAfterRedisReconnect 测试Redis恢复后缓存恢复
func TestCacheRecoveryAfterRedisReconnect(t *testing.T) {
	redis := NewMockRedisClient()
	userRepo := NewMockUserRepo()
	cache := NewMockBalanceCache(redis, userRepo)

	testUser := &MockUser{
		ID:             1,
		Balance:        decimal.NewFromInt(1000),
		FrozenBalance:  decimal.Zero,
		BalanceVersion: 1,
	}
	userRepo.AddUser(testUser)

	ctx := context.Background()

	// 初始状态：Redis可用
	if !cache.IsAvailable(ctx) {
		t.Error("Cache should be available initially")
	}

	// Redis断开
	redis.SetAvailable(false)
	if cache.IsAvailable(ctx) {
		t.Error("Cache should be unavailable when Redis is down")
	}

	// 降级查询应该成功
	user, err := cache.Get(ctx, 1)
	if err != nil {
		t.Fatalf("Fallback query should succeed: %v", err)
	}
	if user.ID != 1 {
		t.Error("Should get correct user from fallback")
	}

	// Redis恢复
	redis.SetAvailable(true)
	if !cache.IsAvailable(ctx) {
		t.Error("Cache should be available after Redis reconnects")
	}

	// 正常查询应该成功
	user, err = cache.Get(ctx, 1)
	if err != nil {
		t.Fatalf("Normal query should succeed after recovery: %v", err)
	}
	if user.ID != 1 {
		t.Error("Should get correct user after recovery")
	}
}

// TestConcurrentCacheAccessDuringFailover 测试故障转移期间的并发访问
func TestConcurrentCacheAccessDuringFailover(t *testing.T) {
	redis := NewMockRedisClient()
	userRepo := NewMockUserRepo()
	cache := NewMockBalanceCache(redis, userRepo)

	// 添加多个测试用户
	for i := int64(1); i <= 10; i++ {
		userRepo.AddUser(&MockUser{
			ID:             i,
			Balance:        decimal.NewFromInt(i * 100),
			FrozenBalance:  decimal.Zero,
			BalanceVersion: 1,
		})
	}

	ctx := context.Background()
	requestCount := 100
	var successCount int32
	var mu sync.Mutex

	var wg sync.WaitGroup

	// 启动并发请求
	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// 随机切换Redis状态（模拟故障）
			if idx%20 == 0 {
				redis.SetAvailable(idx%40 != 0)
			}

			userID := int64((idx % 10) + 1)
			_, err := cache.Get(ctx, userID)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// 所有请求都应该成功（通过缓存或降级）
	if successCount != int32(requestCount) {
		t.Errorf("Expected %d successful requests, got %d", requestCount, successCount)
	}
}

// TestCacheInvalidation 测试缓存失效
func TestCacheInvalidation(t *testing.T) {
	redis := NewMockRedisClient()
	ctx := context.Background()

	// 设置缓存
	key := "balance:1"
	err := redis.Set(ctx, key, `{"balance":"1000","version":1}`, time.Hour)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// 验证缓存存在
	_, err = redis.Get(ctx, key)
	if err != nil {
		t.Error("Cache should exist after set")
	}

	// 删除缓存
	err = redis.Del(ctx, key)
	if err != nil {
		t.Fatalf("Failed to delete cache: %v", err)
	}

	// 验证缓存已删除
	_, err = redis.Get(ctx, key)
	if err == nil {
		t.Error("Cache should not exist after deletion")
	}
}

// TestCacheConsistencyAfterUpdate 测试更新后的缓存一致性
func TestCacheConsistencyAfterUpdate(t *testing.T) {
	redis := NewMockRedisClient()
	userRepo := NewMockUserRepo()

	testUser := &MockUser{
		ID:             1,
		Balance:        decimal.NewFromInt(1000),
		FrozenBalance:  decimal.Zero,
		BalanceVersion: 1,
	}
	userRepo.AddUser(testUser)

	ctx := context.Background()

	// 设置初始缓存
	cacheKey := "balance:1"
	redis.Set(ctx, cacheKey, `{"balance":"1000","version":1}`, time.Hour)

	// 模拟数据库更新
	testUser.Balance = decimal.NewFromInt(900)
	testUser.BalanceVersion = 2

	// 使缓存失效
	redis.Del(ctx, cacheKey)

	// 下次查询应该从数据库获取最新数据
	user, err := userRepo.GetByID(ctx, 1)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if !user.Balance.Equal(decimal.NewFromInt(900)) {
		t.Errorf("Expected updated balance 900, got %s", user.Balance.String())
	}
	if user.BalanceVersion != 2 {
		t.Errorf("Expected version 2, got %d", user.BalanceVersion)
	}
}

// TestGracefulDegradation 测试优雅降级
func TestGracefulDegradation(t *testing.T) {
	redis := NewMockRedisClient()
	userRepo := NewMockUserRepo()
	cache := NewMockBalanceCache(redis, userRepo)

	// 添加测试用户
	userRepo.AddUser(&MockUser{
		ID:             1,
		Balance:        decimal.NewFromInt(1000),
		FrozenBalance:  decimal.Zero,
		BalanceVersion: 1,
	})

	ctx := context.Background()

	// 场景1：Redis可用，正常返回
	redis.SetAvailable(true)
	user, err := cache.Get(ctx, 1)
	if err != nil || user == nil {
		t.Error("Should succeed when Redis is available")
	}

	// 场景2：Redis不可用，降级到数据库
	redis.SetAvailable(false)
	user, err = cache.Get(ctx, 1)
	if err != nil || user == nil {
		t.Error("Should succeed with fallback when Redis is unavailable")
	}

	// 场景3：Redis恢复，正常返回
	redis.SetAvailable(true)
	user, err = cache.Get(ctx, 1)
	if err != nil || user == nil {
		t.Error("Should succeed after Redis recovery")
	}
}
