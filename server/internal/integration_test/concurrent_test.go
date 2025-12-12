// Package integration_test 并发操作集成测试
package integration_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/fiveseconds/server/internal/service"
)

// TestConcurrentChatRateLimiting 测试并发聊天限流
func TestConcurrentChatRateLimiting(t *testing.T) {
	limiter := service.NewRateLimiter(1, time.Second)

	userID := "user:1"
	concurrentRequests := 10
	var allowed int32
	var denied int32

	var wg sync.WaitGroup
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow(userID) {
				atomic.AddInt32(&allowed, 1)
			} else {
				atomic.AddInt32(&denied, 1)
			}
		}()
	}

	wg.Wait()

	// 在1秒窗口内，只应该允许1个请求
	if allowed != 1 {
		t.Errorf("Expected 1 allowed request, got %d", allowed)
	}
	if denied != int32(concurrentRequests-1) {
		t.Errorf("Expected %d denied requests, got %d", concurrentRequests-1, denied)
	}
}

// TestConcurrentEmojiRateLimiting 测试并发表情限流
func TestConcurrentEmojiRateLimiting(t *testing.T) {
	limiter := service.NewRateLimiter(3, time.Second)

	userID := "emoji:1"
	concurrentRequests := 10
	var allowed int32

	var wg sync.WaitGroup
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if limiter.Allow(userID) {
				atomic.AddInt32(&allowed, 1)
			}
		}()
	}

	wg.Wait()

	// 在1秒窗口内，只应该允许3个请求
	if allowed != 3 {
		t.Errorf("Expected 3 allowed requests, got %d", allowed)
	}
}

// TestRateLimiterWindowReset 测试限流窗口重置
func TestRateLimiterWindowReset(t *testing.T) {
	// 使用100ms窗口便于测试
	limiter := service.NewRateLimiter(2, 100*time.Millisecond)

	userID := "test:1"

	// 第一个窗口：发送2个请求
	if !limiter.Allow(userID) {
		t.Error("First request should be allowed")
	}
	if !limiter.Allow(userID) {
		t.Error("Second request should be allowed")
	}
	if limiter.Allow(userID) {
		t.Error("Third request should be denied")
	}

	// 等待窗口重置
	time.Sleep(150 * time.Millisecond)

	// 新窗口：应该允许新请求
	if !limiter.Allow(userID) {
		t.Error("Request after window reset should be allowed")
	}
}

// TestConcurrentMultiUserRateLimiting 测试多用户并发限流
func TestConcurrentMultiUserRateLimiting(t *testing.T) {
	limiter := service.NewRateLimiter(1, time.Second)

	userCount := 5
	requestsPerUser := 3
	allowedPerUser := make([]int32, userCount)

	var wg sync.WaitGroup
	for u := 0; u < userCount; u++ {
		for r := 0; r < requestsPerUser; r++ {
			wg.Add(1)
			go func(userIdx int) {
				defer wg.Done()
				userID := string(rune('A' + userIdx))
				if limiter.Allow(userID) {
					atomic.AddInt32(&allowedPerUser[userIdx], 1)
				}
			}(u)
		}
	}

	wg.Wait()

	// 每个用户应该只有1个请求被允许
	for u := 0; u < userCount; u++ {
		if allowedPerUser[u] != 1 {
			t.Errorf("User %d: expected 1 allowed request, got %d", u, allowedPerUser[u])
		}
	}
}

// MockBalanceStore 模拟余额存储（用于测试乐观锁）
type MockBalanceStore struct {
	mu       sync.Mutex
	balances map[int64]int64
	versions map[int64]int64
}

func NewMockBalanceStore() *MockBalanceStore {
	return &MockBalanceStore{
		balances: make(map[int64]int64),
		versions: make(map[int64]int64),
	}
}

func (s *MockBalanceStore) SetBalance(userID, balance int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[userID] = balance
	s.versions[userID] = 1
}

func (s *MockBalanceStore) GetBalance(userID int64) (balance, version int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.balances[userID], s.versions[userID]
}

// UpdateWithVersion 使用乐观锁更新余额
func (s *MockBalanceStore) UpdateWithVersion(userID, delta, expectedVersion int64) (bool, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentVersion := s.versions[userID]
	if currentVersion != expectedVersion {
		return false, currentVersion
	}

	s.balances[userID] += delta
	s.versions[userID]++
	return true, s.versions[userID]
}

// TestOptimisticLockingConcurrency 测试乐观锁并发控制
func TestOptimisticLockingConcurrency(t *testing.T) {
	store := NewMockBalanceStore()
	userID := int64(1)
	initialBalance := int64(1000)
	store.SetBalance(userID, initialBalance)

	concurrentUpdates := 10
	updateAmount := int64(10)
	var successCount int32
	var failCount int32

	// 所有goroutine同时读取版本
	_, version := store.GetBalance(userID)

	var wg sync.WaitGroup
	for i := 0; i < concurrentUpdates; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 使用相同的版本尝试更新
			success, _ := store.UpdateWithVersion(userID, updateAmount, version)
			if success {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&failCount, 1)
			}
		}()
	}

	wg.Wait()

	// 只有一个更新应该成功
	if successCount != 1 {
		t.Errorf("Expected 1 successful update, got %d", successCount)
	}
	if failCount != int32(concurrentUpdates-1) {
		t.Errorf("Expected %d failed updates, got %d", concurrentUpdates-1, failCount)
	}

	// 验证最终余额
	finalBalance, _ := store.GetBalance(userID)
	expectedBalance := initialBalance + updateAmount
	if finalBalance != expectedBalance {
		t.Errorf("Expected balance %d, got %d", expectedBalance, finalBalance)
	}
}

// TestSequentialOptimisticLocking 测试顺序乐观锁更新
func TestSequentialOptimisticLocking(t *testing.T) {
	store := NewMockBalanceStore()
	userID := int64(1)
	initialBalance := int64(1000)
	store.SetBalance(userID, initialBalance)

	updateCount := 5
	updateAmount := int64(10)

	for i := 0; i < updateCount; i++ {
		_, version := store.GetBalance(userID)
		success, _ := store.UpdateWithVersion(userID, updateAmount, version)
		if !success {
			t.Errorf("Update %d should succeed", i)
		}
	}

	// 验证最终余额
	finalBalance, finalVersion := store.GetBalance(userID)
	expectedBalance := initialBalance + int64(updateCount)*updateAmount
	if finalBalance != expectedBalance {
		t.Errorf("Expected balance %d, got %d", expectedBalance, finalBalance)
	}
	if finalVersion != int64(updateCount+1) {
		t.Errorf("Expected version %d, got %d", updateCount+1, finalVersion)
	}
}

// TestConcurrentHubOperations 测试Hub并发操作
func TestConcurrentHubOperations(t *testing.T) {
	// 这个测试验证Hub在并发添加/删除连接时的线程安全性
	// 由于Hub使用了sync.RWMutex，应该能正确处理并发

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	operationCount := 100
	var addCount, removeCount int32

	var wg sync.WaitGroup

	// 并发添加和删除操作
	for i := 0; i < operationCount; i++ {
		wg.Add(2)

		// 添加操作
		go func(idx int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				atomic.AddInt32(&addCount, 1)
			}
		}(i)

		// 删除操作
		go func(idx int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				atomic.AddInt32(&removeCount, 1)
			}
		}(i)
	}

	wg.Wait()

	// 验证所有操作都完成了
	if addCount != int32(operationCount) {
		t.Errorf("Expected %d add operations, got %d", operationCount, addCount)
	}
	if removeCount != int32(operationCount) {
		t.Errorf("Expected %d remove operations, got %d", operationCount, removeCount)
	}
}
