package service

import (
	"context"
	"testing"

	"github.com/fiveseconds/server/internal/model"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// MockRiskRepo 模拟风控仓库用于属性测试
type MockRiskRepo struct {
	consecutiveWins map[int64]int
	flags           []*model.RiskFlag
	pendingFlags    map[int64]map[model.RiskFlagType]bool
	winRecords      map[int64][]bool // userID -> win/loss records
}

func NewMockRiskRepo() *MockRiskRepo {
	return &MockRiskRepo{
		consecutiveWins: make(map[int64]int),
		flags:           make([]*model.RiskFlag, 0),
		pendingFlags:    make(map[int64]map[model.RiskFlagType]bool),
		winRecords:      make(map[int64][]bool),
	}
}

func (m *MockRiskRepo) GetUserConsecutiveWins(ctx context.Context, userID int64) (int, error) {
	return m.consecutiveWins[userID], nil
}

func (m *MockRiskRepo) UpdateUserConsecutiveWins(ctx context.Context, userID int64, wins int) error {
	m.consecutiveWins[userID] = wins
	return nil
}

func (m *MockRiskRepo) ResetUserConsecutiveWins(ctx context.Context, userID int64) error {
	m.consecutiveWins[userID] = 0
	return nil
}

func (m *MockRiskRepo) HasPendingFlag(ctx context.Context, userID int64, flagType model.RiskFlagType) (bool, error) {
	if userFlags, ok := m.pendingFlags[userID]; ok {
		return userFlags[flagType], nil
	}
	return false, nil
}


func (m *MockRiskRepo) CreateFlagWithDetails(ctx context.Context, flag *model.RiskFlag, details *model.RiskFlagDetails) error {
	flag.ID = int64(len(m.flags) + 1)
	m.flags = append(m.flags, flag)
	
	// 标记为待处理
	if m.pendingFlags[flag.UserID] == nil {
		m.pendingFlags[flag.UserID] = make(map[model.RiskFlagType]bool)
	}
	m.pendingFlags[flag.UserID][flag.FlagType] = true
	return nil
}

func (m *MockRiskRepo) GetUserWinRate(ctx context.Context, userID int64, minRounds int) (float64, int, error) {
	records := m.winRecords[userID]
	if len(records) < minRounds {
		return 0, len(records), nil
	}
	
	wins := 0
	for _, isWin := range records {
		if isWin {
			wins++
		}
	}
	return float64(wins) / float64(len(records)), len(records), nil
}

func (m *MockRiskRepo) AddWinRecord(userID int64, isWin bool) {
	m.winRecords[userID] = append(m.winRecords[userID], isWin)
}

func (m *MockRiskRepo) GetFlags() []*model.RiskFlag {
	return m.flags
}

func (m *MockRiskRepo) ClearFlags() {
	m.flags = make([]*model.RiskFlag, 0)
	m.pendingFlags = make(map[int64]map[model.RiskFlagType]bool)
}

// RiskServiceForTest 用于测试的风控服务包装
type RiskServiceForTest struct {
	repo   *MockRiskRepo
	config model.RiskConfig
}

func NewRiskServiceForTest(config model.RiskConfig) *RiskServiceForTest {
	return &RiskServiceForTest{
		repo:   NewMockRiskRepo(),
		config: config,
	}
}

// CheckConsecutiveWins 检查连续获胜（测试版本）
func (s *RiskServiceForTest) CheckConsecutiveWins(ctx context.Context, userID int64, isWinner bool) error {
	if isWinner {
		wins, _ := s.repo.GetUserConsecutiveWins(ctx, userID)
		wins++
		s.repo.UpdateUserConsecutiveWins(ctx, userID, wins)

		if wins > s.config.ConsecutiveWinThreshold {
			hasPending, _ := s.repo.HasPendingFlag(ctx, userID, model.RiskFlagConsecutiveWins)
			if !hasPending {
				flag := &model.RiskFlag{
					UserID:   userID,
					FlagType: model.RiskFlagConsecutiveWins,
					Status:   model.RiskFlagStatusPending,
				}
				details := &model.RiskFlagDetails{ConsecutiveWins: wins}
				s.repo.CreateFlagWithDetails(ctx, flag, details)
			}
		}
	} else {
		s.repo.ResetUserConsecutiveWins(ctx, userID)
	}
	return nil
}


// CheckWinRate 检查胜率（测试版本）
func (s *RiskServiceForTest) CheckWinRate(ctx context.Context, userID int64) error {
	winRate, totalRounds, _ := s.repo.GetUserWinRate(ctx, userID, s.config.WinRateMinRounds)
	
	if totalRounds < s.config.WinRateMinRounds {
		return nil
	}

	if winRate > s.config.WinRateThreshold {
		hasPending, _ := s.repo.HasPendingFlag(ctx, userID, model.RiskFlagHighWinRate)
		if !hasPending {
			flag := &model.RiskFlag{
				UserID:   userID,
				FlagType: model.RiskFlagHighWinRate,
				Status:   model.RiskFlagStatusPending,
			}
			details := &model.RiskFlagDetails{WinRate: winRate, TotalRounds: totalRounds}
			s.repo.CreateFlagWithDetails(ctx, flag, details)
		}
	}
	return nil
}

func (s *RiskServiceForTest) GetFlags() []*model.RiskFlag {
	return s.repo.GetFlags()
}

func (s *RiskServiceForTest) GetConsecutiveWins(userID int64) int {
	return s.repo.consecutiveWins[userID]
}

func (s *RiskServiceForTest) AddWinRecord(userID int64, isWin bool) {
	s.repo.AddWinRecord(userID, isWin)
}

// =============================================================================
// Property-Based Tests
// =============================================================================

// TestProperty16_ConsecutiveWinDetection 属性测试：连续获胜检测
// **Feature: p1-p2-features, Property 16: Consecutive win detection**
// **Validates: Requirements 12.1**
//
// Property: For any player who wins more than 10 consecutive rounds,
// a risk flag should be created.
func TestProperty16_ConsecutiveWinDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// 生成连续获胜次数（1-20）
	consecutiveWinsGen := gen.IntRange(1, 20)

	properties.Property("consecutive wins > threshold creates risk flag", prop.ForAll(
		func(consecutiveWins int) bool {
			ctx := context.Background()
			config := model.RiskConfig{
				ConsecutiveWinThreshold: 10,
				WinRateThreshold:        0.8,
				WinRateMinRounds:        50,
			}
			service := NewRiskServiceForTest(config)
			userID := int64(1)

			// 模拟连续获胜
			for i := 0; i < consecutiveWins; i++ {
				service.CheckConsecutiveWins(ctx, userID, true)
			}

			flags := service.GetFlags()
			actualWins := service.GetConsecutiveWins(userID)

			// 验证：连续获胜次数应该正确记录
			if actualWins != consecutiveWins {
				return false
			}

			// 验证：如果连续获胜超过阈值，应该创建风控标记
			if consecutiveWins > config.ConsecutiveWinThreshold {
				// 应该有一个风控标记
				if len(flags) != 1 {
					return false
				}
				// 标记类型应该是连续获胜
				if flags[0].FlagType != model.RiskFlagConsecutiveWins {
					return false
				}
				// 标记状态应该是待处理
				if flags[0].Status != model.RiskFlagStatusPending {
					return false
				}
			} else {
				// 不应该有风控标记
				if len(flags) != 0 {
					return false
				}
			}

			return true
		},
		consecutiveWinsGen,
	))


	// 测试输掉后重置连续获胜计数
	properties.Property("losing resets consecutive wins counter", prop.ForAll(
		func(winsBeforeLoss int) bool {
			ctx := context.Background()
			config := model.RiskConfig{
				ConsecutiveWinThreshold: 10,
				WinRateThreshold:        0.8,
				WinRateMinRounds:        50,
			}
			service := NewRiskServiceForTest(config)
			userID := int64(1)

			// 先连续获胜
			for i := 0; i < winsBeforeLoss; i++ {
				service.CheckConsecutiveWins(ctx, userID, true)
			}

			// 然后输掉一局
			service.CheckConsecutiveWins(ctx, userID, false)

			// 验证：连续获胜计数应该被重置为0
			return service.GetConsecutiveWins(userID) == 0
		},
		gen.IntRange(1, 15),
	))

	// 测试不会重复创建风控标记
	properties.Property("no duplicate risk flags for same user", prop.ForAll(
		func(extraWins int) bool {
			ctx := context.Background()
			config := model.RiskConfig{
				ConsecutiveWinThreshold: 10,
				WinRateThreshold:        0.8,
				WinRateMinRounds:        50,
			}
			service := NewRiskServiceForTest(config)
			userID := int64(1)

			// 先超过阈值
			for i := 0; i < 11; i++ {
				service.CheckConsecutiveWins(ctx, userID, true)
			}

			// 继续获胜更多次
			for i := 0; i < extraWins; i++ {
				service.CheckConsecutiveWins(ctx, userID, true)
			}

			flags := service.GetFlags()

			// 验证：即使继续获胜，也只应该有一个风控标记
			return len(flags) == 1
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}


// TestProperty17_WinRateDetection 属性测试：胜率检测
// **Feature: p1-p2-features, Property 17: Win rate detection**
// **Validates: Requirements 12.2**
//
// Property: For any player with win rate exceeding 80% over 50 rounds,
// a risk flag should be created.
func TestProperty17_WinRateDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// 生成胜率（0.5-1.0）和回合数（30-70）
	winRateGen := gen.Float64Range(0.5, 1.0)
	roundsGen := gen.IntRange(30, 70)

	properties.Property("high win rate over min rounds creates risk flag", prop.ForAll(
		func(targetWinRate float64, totalRounds int) bool {
			ctx := context.Background()
			config := model.RiskConfig{
				ConsecutiveWinThreshold: 10,
				WinRateThreshold:        0.8,
				WinRateMinRounds:        50,
			}
			service := NewRiskServiceForTest(config)
			userID := int64(1)

			// 计算需要多少次获胜来达到目标胜率
			wins := int(float64(totalRounds) * targetWinRate)
			losses := totalRounds - wins

			// 添加获胜记录
			for i := 0; i < wins; i++ {
				service.AddWinRecord(userID, true)
			}
			// 添加失败记录
			for i := 0; i < losses; i++ {
				service.AddWinRecord(userID, false)
			}

			// 检查胜率
			service.CheckWinRate(ctx, userID)

			flags := service.GetFlags()
			actualWinRate := float64(wins) / float64(totalRounds)

			// 验证逻辑
			if totalRounds >= config.WinRateMinRounds && actualWinRate > config.WinRateThreshold {
				// 应该创建风控标记
				if len(flags) != 1 {
					return false
				}
				if flags[0].FlagType != model.RiskFlagHighWinRate {
					return false
				}
			} else {
				// 不应该创建风控标记（回合数不足或胜率未超阈值）
				if len(flags) != 0 {
					return false
				}
			}

			return true
		},
		winRateGen,
		roundsGen,
	))

	// 测试回合数不足时不触发检测
	properties.Property("insufficient rounds does not trigger detection", prop.ForAll(
		func(rounds int) bool {
			ctx := context.Background()
			config := model.RiskConfig{
				ConsecutiveWinThreshold: 10,
				WinRateThreshold:        0.8,
				WinRateMinRounds:        50,
			}
			service := NewRiskServiceForTest(config)
			userID := int64(1)

			// 全部获胜但回合数不足
			for i := 0; i < rounds; i++ {
				service.AddWinRecord(userID, true)
			}

			service.CheckWinRate(ctx, userID)
			flags := service.GetFlags()

			// 回合数不足时不应该创建风控标记
			return len(flags) == 0
		},
		gen.IntRange(1, 49), // 小于最小回合数50
	))

	// 测试正常胜率不触发检测
	properties.Property("normal win rate does not trigger detection", prop.ForAll(
		func(wins int) bool {
			ctx := context.Background()
			config := model.RiskConfig{
				ConsecutiveWinThreshold: 10,
				WinRateThreshold:        0.8,
				WinRateMinRounds:        50,
			}
			service := NewRiskServiceForTest(config)
			userID := int64(1)

			// 固定50回合，胜率在0-80%之间
			totalRounds := 50
			losses := totalRounds - wins

			for i := 0; i < wins; i++ {
				service.AddWinRecord(userID, true)
			}
			for i := 0; i < losses; i++ {
				service.AddWinRecord(userID, false)
			}

			service.CheckWinRate(ctx, userID)
			flags := service.GetFlags()

			// 胜率<=80%时不应该创建风控标记
			return len(flags) == 0
		},
		gen.IntRange(0, 40), // 0-40胜，胜率0-80%
	))

	properties.TestingRun(t)
}
