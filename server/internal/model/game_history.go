package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// GameHistoryItem 游戏历史记录项
type GameHistoryItem struct {
	ID             int64           `json:"id"`
	RoomID         int64           `json:"room_id"`
	RoomName       string          `json:"room_name"`
	RoundNumber    int             `json:"round_number"`
	BetAmount      decimal.Decimal `json:"bet_amount"`
	Result         string          `json:"result"` // win/lose/skipped
	PrizeAmount    decimal.Decimal `json:"prize_amount"`
	CreatedAt      time.Time       `json:"created_at"`
}

// GameHistoryQuery 游戏历史查询参数
type GameHistoryQuery struct {
	UserID    int64      `form:"-"`
	RoomID    *int64     `form:"room_id"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	Page      int        `form:"page"`
	PageSize  int        `form:"page_size"`
}

// GameStats 游戏统计
type GameStats struct {
	TotalRounds  int             `json:"total_rounds"`
	TotalWins    int             `json:"total_wins"`
	TotalLosses  int             `json:"total_losses"`
	TotalSkipped int             `json:"total_skipped"`
	WinRate      float64         `json:"win_rate"`
	TotalWagered decimal.Decimal `json:"total_wagered"`
	TotalWon     decimal.Decimal `json:"total_won"`
	NetProfit    decimal.Decimal `json:"net_profit"`
}

// RoundDetail 回合详情
type RoundDetail struct {
	ID              int64           `json:"id"`
	RoomID          int64           `json:"room_id"`
	RoomName        string          `json:"room_name"`
	RoundNumber     int             `json:"round_number"`
	BetAmount       decimal.Decimal `json:"bet_amount"`
	PoolAmount      decimal.Decimal `json:"pool_amount"`
	PrizePerWinner  decimal.Decimal `json:"prize_per_winner"`
	CommitHash      string          `json:"commit_hash"`
	RevealSeed      string          `json:"reveal_seed"`
	Status          RoundStatus     `json:"status"`
	Participants    []Participant   `json:"participants"`
	Winners         []Winner        `json:"winners"`
	CreatedAt       time.Time       `json:"created_at"`
	SettledAt       *time.Time      `json:"settled_at"`
}

// Participant 参与者信息
type Participant struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	IsWinner bool   `json:"is_winner"`
}

// Winner 赢家信息
type Winner struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// ReplayData 回放数据
type ReplayData struct {
	RoundID        int64           `json:"round_id"`
	RoomName       string          `json:"room_name"`
	RoundNumber    int             `json:"round_number"`
	BetAmount      decimal.Decimal `json:"bet_amount"`
	PoolAmount     decimal.Decimal `json:"pool_amount"`
	PrizePerWinner decimal.Decimal `json:"prize_per_winner"`
	CommitHash     string          `json:"commit_hash"`
	RevealSeed     string          `json:"reveal_seed"`
	Participants   []Participant   `json:"participants"`
	Winners        []Winner        `json:"winners"`
	Phases         []PhaseData     `json:"phases"`
	CreatedAt      time.Time       `json:"created_at"`
	SettledAt      *time.Time      `json:"settled_at"`
}

// PhaseData 阶段数据
type PhaseData struct {
	Phase     GamePhase `json:"phase"`
	Duration  int       `json:"duration"` // 秒
	Timestamp time.Time `json:"timestamp"`
}
