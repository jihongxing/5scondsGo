package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// GamePhase 游戏阶段
type GamePhase string

const (
	PhaseWaiting    GamePhase = "waiting"
	PhaseCountdown  GamePhase = "countdown"
	PhaseBetting    GamePhase = "betting"
	PhaseInGame     GamePhase = "in_game"
	PhaseSettlement GamePhase = "settlement"
	PhaseReset      GamePhase = "reset"
)

// RoundStatus 回合状态
type RoundStatus string

const (
	RoundStatusBetting RoundStatus = "betting"
	RoundStatusPlaying RoundStatus = "playing"
	RoundStatusSettled RoundStatus = "settled"
	RoundStatusFailed  RoundStatus = "failed"
)

// GameRound 游戏回合
type GameRound struct {
	ID          int64  `json:"id" db:"id"`
	RoomID      int64  `json:"room_id" db:"room_id"`
	RoundNumber int    `json:"round_number" db:"round_number"`

	// 参与信息
	ParticipantIDs []int64 `json:"participant_ids" db:"participant_ids"`
	SkippedIDs     []int64 `json:"skipped_ids" db:"skipped_ids"`
	WinnerIDs      []int64 `json:"winner_ids" db:"winner_ids"`

	// 金额
	BetAmount       decimal.Decimal  `json:"bet_amount" db:"bet_amount"`
	PoolAmount      decimal.Decimal  `json:"pool_amount" db:"pool_amount"`
	PrizePerWinner  *decimal.Decimal `json:"prize_per_winner,omitempty" db:"prize_per_winner"`
	OwnerEarning    *decimal.Decimal `json:"owner_earning,omitempty" db:"owner_earning"`
	PlatformEarning *decimal.Decimal `json:"platform_earning,omitempty" db:"platform_earning"`
	ResidualAmount  *decimal.Decimal `json:"residual_amount,omitempty" db:"residual_amount"`

	// Commit-Reveal 随机
	CommitHash *string `json:"commit_hash,omitempty" db:"commit_hash"`
	RevealSeed *string `json:"reveal_seed,omitempty" db:"reveal_seed"`

	// 状态
	Status        RoundStatus `json:"status" db:"status"`
	FailureReason *string     `json:"failure_reason,omitempty" db:"failure_reason"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	SettledAt *time.Time `json:"settled_at,omitempty" db:"settled_at"`
}

// RoomState 房间内存状态(用于 RoomProcessor)
type RoomState struct {
	Phase         GamePhase              `json:"phase"`
	PhaseEndTime  time.Time              `json:"phase_end_time"`
	CurrentRound  int                    `json:"current_round"`
	Players       map[int64]*PlayerState `json:"players"`
	Spectators    map[int64]*SpectatorState `json:"spectators"` // 观战者
	
	// 当前回合数据
	RoundID        int64             `json:"round_id,omitempty"`
	Participants   []int64           `json:"participants,omitempty"`
	SkippedPlayers []int64           `json:"skipped_players,omitempty"`
	PoolAmount     decimal.Decimal   `json:"pool_amount"`
	CommitHash     string            `json:"commit_hash,omitempty"`
	Seed           []byte            `json:"-"` // 内存中保存,不序列化
}

// PlayerState 玩家内存状态
type PlayerState struct {
	UserID            int64           `json:"user_id"`
	Username          string          `json:"username"`
	Balance           decimal.Decimal `json:"balance"`
	AutoReady         bool            `json:"auto_ready"`
	IsOnline          bool            `json:"is_online"`
	OfflineSince      *time.Time      `json:"-"` // 离线开始时间，用于超时清理
	Disqualified      bool            `json:"disqualified"`       // 是否被取消资格（余额不足等）
	DisqualifyReason  string          `json:"disqualify_reason"`  // 取消资格原因
}

// PhaseInfo 阶段信息(用于广播)
type PhaseInfo struct {
	Phase      GamePhase `json:"phase"`
	TimeLeft   int       `json:"time_left"`
	Round      int       `json:"round"`
	PoolAmount string    `json:"pool_amount,omitempty"`
}

// RoundResult 回合结果(用于广播)
type RoundResult struct {
	RoundID        int64    `json:"round_id"`
	Winners        []int64  `json:"winners"`
	WinnerNames    []string `json:"winner_names"`
	PrizePerWinner string   `json:"prize_per_winner"`
	RevealSeed     string   `json:"reveal_seed"`
	CommitHash     string   `json:"commit_hash"`
}

// RoundFailed 回合失败(用于广播)
type RoundFailed struct {
	Reason   string  `json:"reason"`
	Refunded []int64 `json:"refunded"`
}

// BettingCompleted 下注完成(用于广播)
type BettingCompleted struct {
	PoolAmount   string  `json:"pool_amount"`
	Participants []int64 `json:"participants"`
	Skipped      []int64 `json:"skipped"`
}
