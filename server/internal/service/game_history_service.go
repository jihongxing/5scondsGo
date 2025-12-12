package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"

	"go.uber.org/zap"
)

// 游戏历史相关错误
var (
	ErrRoundNotFound = errors.New("round not found")
	ErrInvalidSeed   = errors.New("invalid reveal seed")
)

// GameHistoryService 游戏历史服务
type GameHistoryService struct {
	gameRepo *repository.GameRepo
	logger   *zap.Logger
}

// NewGameHistoryService 创建游戏历史服务
func NewGameHistoryService(gameRepo *repository.GameRepo, logger *zap.Logger) *GameHistoryService {
	return &GameHistoryService{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// GetHistory 获取游戏历史
func (s *GameHistoryService) GetHistory(ctx context.Context, query *model.GameHistoryQuery) ([]*model.GameHistoryItem, int64, error) {
	return s.gameRepo.GetUserGameHistory(ctx, query)
}

// GetStats 获取游戏统计
func (s *GameHistoryService) GetStats(ctx context.Context, userID int64) (*model.GameStats, error) {
	return s.gameRepo.GetUserGameStats(ctx, userID)
}

// GetRoundDetail 获取回合详情
func (s *GameHistoryService) GetRoundDetail(ctx context.Context, roundID int64) (*model.RoundDetail, error) {
	return s.gameRepo.GetRoundDetail(ctx, roundID)
}

// GetReplayData 获取回放数据
func (s *GameHistoryService) GetReplayData(ctx context.Context, roundID int64) (*model.ReplayData, error) {
	detail, err := s.gameRepo.GetRoundDetail(ctx, roundID)
	if err != nil {
		return nil, err
	}

	replay := &model.ReplayData{
		RoundID:        detail.ID,
		RoomName:       detail.RoomName,
		RoundNumber:    detail.RoundNumber,
		BetAmount:      detail.BetAmount,
		PoolAmount:     detail.PoolAmount,
		PrizePerWinner: detail.PrizePerWinner,
		CommitHash:     detail.CommitHash,
		RevealSeed:     detail.RevealSeed,
		Participants:   detail.Participants,
		Winners:        detail.Winners,
		CreatedAt:      detail.CreatedAt,
		SettledAt:      detail.SettledAt,
	}

	// 构建阶段数据（模拟）
	replay.Phases = []model.PhaseData{
		{Phase: model.PhaseCountdown, Duration: 5},
		{Phase: model.PhaseBetting, Duration: 5},
		{Phase: model.PhaseInGame, Duration: 5},
		{Phase: model.PhaseSettlement, Duration: 5},
	}

	return replay, nil
}


// VerifyRound 验证回合结果
func (s *GameHistoryService) VerifyRound(ctx context.Context, roundID int64) (*VerificationResult, error) {
	detail, err := s.gameRepo.GetRoundDetail(ctx, roundID)
	if err != nil {
		return nil, err
	}

	if detail.RevealSeed == "" {
		return nil, ErrInvalidSeed
	}

	// 验证 commit hash
	hash := sha256.Sum256([]byte(detail.RevealSeed))
	computedHash := hex.EncodeToString(hash[:])

	hashMatch := computedHash == detail.CommitHash

	// 重新计算赢家
	participantIDs := make([]int64, 0, len(detail.Participants))
	for _, p := range detail.Participants {
		participantIDs = append(participantIDs, p.UserID)
	}

	// 排序以确保一致性
	sort.Slice(participantIDs, func(i, j int) bool {
		return participantIDs[i] < participantIDs[j]
	})

	computedWinners := computeWinners(detail.RevealSeed, participantIDs, len(detail.Winners))

	// 比较赢家
	actualWinnerIDs := make([]int64, 0, len(detail.Winners))
	for _, w := range detail.Winners {
		actualWinnerIDs = append(actualWinnerIDs, w.UserID)
	}
	sort.Slice(actualWinnerIDs, func(i, j int) bool {
		return actualWinnerIDs[i] < actualWinnerIDs[j]
	})

	winnersMatch := compareInt64Slices(computedWinners, actualWinnerIDs)

	return &VerificationResult{
		RoundID:         roundID,
		CommitHash:      detail.CommitHash,
		RevealSeed:      detail.RevealSeed,
		ComputedHash:    computedHash,
		HashMatch:       hashMatch,
		ActualWinners:   actualWinnerIDs,
		ComputedWinners: computedWinners,
		WinnersMatch:    winnersMatch,
		IsValid:         hashMatch && winnersMatch,
	}, nil
}

// VerificationResult 验证结果
type VerificationResult struct {
	RoundID         int64   `json:"round_id"`
	CommitHash      string  `json:"commit_hash"`
	RevealSeed      string  `json:"reveal_seed"`
	ComputedHash    string  `json:"computed_hash"`
	HashMatch       bool    `json:"hash_match"`
	ActualWinners   []int64 `json:"actual_winners"`
	ComputedWinners []int64 `json:"computed_winners"`
	WinnersMatch    bool    `json:"winners_match"`
	IsValid         bool    `json:"is_valid"`
}

// computeWinners 使用种子计算赢家
func computeWinners(seed string, participantIDs []int64, winnerCount int) []int64 {
	if len(participantIDs) == 0 || winnerCount <= 0 {
		return nil
	}

	if winnerCount > len(participantIDs) {
		winnerCount = len(participantIDs)
	}

	// 使用种子生成随机数
	hash := sha256.Sum256([]byte(seed))
	
	// Fisher-Yates shuffle 的确定性版本
	shuffled := make([]int64, len(participantIDs))
	copy(shuffled, participantIDs)

	for i := len(shuffled) - 1; i > 0; i-- {
		// 使用 hash 的不同字节来生成索引
		j := int(hash[i%32]) % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	winners := shuffled[:winnerCount]
	sort.Slice(winners, func(i, j int) bool {
		return winners[i] < winners[j]
	})

	return winners
}

// compareInt64Slices 比较两个 int64 切片
func compareInt64Slices(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
