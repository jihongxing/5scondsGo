package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// TransactionType 交易类型
type TransactionType string

const (
	TxDeposit           TransactionType = "deposit"            // 充值
	TxWithdraw          TransactionType = "withdraw"           // 提现
	TxMarginDeposit     TransactionType = "margin_deposit"     // 保证金充值（仅初始设置）
	TxGameBet           TransactionType = "game_bet"           // 游戏下注
	TxGameWin           TransactionType = "game_win"           // 游戏获胜
	TxGameRefund        TransactionType = "game_refund"        // 游戏退款
	TxOwnerCommission   TransactionType = "owner_commission"   // 房主佣金
	TxPlatformShare     TransactionType = "platform_share"     // 平台抽成
	TxFreeze            TransactionType = "freeze"             // 冻结
	TxUnfreeze          TransactionType = "unfreeze"           // 解冻
	TxEarningsTransfer  TransactionType = "earnings_transfer"  // 佣金转可用余额
)

// BalanceTransaction 余额交易记录
type BalanceTransaction struct {
	ID            int64           `json:"id" db:"id"`
	UserID        int64           `json:"user_id" db:"user_id"`
	RoomID        *int64          `json:"room_id,omitempty" db:"room_id"`
	RoundID       *int64          `json:"round_id,omitempty" db:"round_id"`
	Type          TransactionType `json:"type" db:"tx_type"`
	Amount        decimal.Decimal `json:"amount" db:"amount"`
	BalanceBefore decimal.Decimal `json:"balance_before" db:"balance_before"`
	BalanceAfter  decimal.Decimal `json:"balance_after" db:"balance_after"`
	BalanceField  string          `json:"balance_field" db:"balance_field"`
	Remark        *string         `json:"remark,omitempty" db:"remark"`
	CreatedAt     time.Time       `json:"created_at" db:"created_at"`
}

// FundRequestType 资金申请类型
type FundRequestType string

const (
	FundRequestDeposit       FundRequestType = "deposit"        // 玩家充值（房主确认后：房主余额-，玩家余额+）
	FundRequestWithdraw      FundRequestType = "withdraw"       // 玩家提现（房主确认后：玩家余额-，房主余额+）
	FundRequestOwnerDeposit  FundRequestType = "owner_deposit"  // 房主充值（增加房主可用余额）
	FundRequestOwnerWithdraw FundRequestType = "owner_withdraw" // 房主提现（减少房主可用余额）
	FundRequestMarginDeposit FundRequestType = "margin_deposit" // 保证金充值（仅初始设置，固定不变）
)

// FundRequestStatus 资金申请状态
type FundRequestStatus string

const (
	FundStatusPending  FundRequestStatus = "pending"
	FundStatusApproved FundRequestStatus = "approved"
	FundStatusRejected FundRequestStatus = "rejected"
)

// FundRequest 资金申请
type FundRequest struct {
	ID          int64             `json:"id" db:"id"`
	UserID      int64             `json:"user_id" db:"user_id"`
	Username    string            `json:"username,omitempty" db:"-"`
	Type        FundRequestType   `json:"type" db:"request_type"`
	Amount      decimal.Decimal   `json:"amount" db:"amount"`
	Status      FundRequestStatus `json:"status" db:"status"`
	Remark      *string           `json:"remark,omitempty" db:"remark"`
	ProcessedBy *int64            `json:"processed_by,omitempty" db:"operator_id"`
	ProcessedAt *time.Time        `json:"processed_at,omitempty" db:"updated_at"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
}

// PlatformAccount 平台账户
type PlatformAccount struct {
	ID              int64           `json:"id" db:"id"`
	PlatformBalance decimal.Decimal `json:"platform_balance" db:"platform_balance"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

// ===== 请求/响应类型 =====

// CreateFundRequestReq 创建资金申请请求
type CreateFundRequestReq struct {
	Type   FundRequestType `json:"type" binding:"required"`
	Amount decimal.Decimal `json:"amount" binding:"required"`
	Remark string          `json:"remark"`
}

// ProcessFundRequestReq 处理资金申请请求
type ProcessFundRequestReq struct {
	Approved bool   `json:"approved"`
	Remark   string `json:"remark"`
}

// FundRequestListQuery 资金申请列表查询
type FundRequestListQuery struct {
	UserID    *int64             `form:"user_id"`
	InvitedBy *int64             `form:"invited_by"` // 查询某个 owner 下级玩家的申请
	Type      *FundRequestType   `form:"type"`
	Status    *FundRequestStatus `form:"status"`
	Page      int                `form:"page" binding:"min=1"`
	PageSize  int                `form:"page_size" binding:"min=1,max=100"`
}

// FundRequestListResp 资金申请列表响应
type FundRequestListResp struct {
	Total int64          `json:"total"`
	Items []*FundRequest `json:"items"`
}

// TransactionListQuery 交易记录查询
type TransactionListQuery struct {
	UserID   *int64           `form:"user_id"`
	RoomID   *int64           `form:"room_id"`
	Type     *TransactionType `form:"type"`
	Page     int              `form:"page" binding:"min=1"`
	PageSize int              `form:"page_size" binding:"min=1,max=100"`
}

// TransactionListResp 交易记录响应
type TransactionListResp struct {
	Total int64                 `json:"total"`
	Items []*BalanceTransaction `json:"items"`
}

// FundSummary 资金统计摘要
type FundSummary struct {
	TotalDeposit    decimal.Decimal `json:"total_deposit"`
	TotalWithdraw   decimal.Decimal `json:"total_withdraw"`
	TotalBet        decimal.Decimal `json:"total_bet"`
	TotalWin        decimal.Decimal `json:"total_win"`
	TotalCommission decimal.Decimal `json:"total_commission"`
	PlatformBalance decimal.Decimal `json:"platform_balance"`
}

// ConservationCheck 资金守恒检查结果
// 资金守恒公式: 玩家总余额 + 房主佣金收益 + 平台余额 = 房主净充值额（系统总资金入口）
// 简化公式: 系统内资金总和 = 玩家余额 + 房主可用余额 + 房主佣金 + 平台余额
type ConservationCheck struct {
	IsBalanced bool `json:"is_balanced"`

	// 玩家侧
	TotalPlayerBalance decimal.Decimal `json:"total_player_balance"` // 玩家可用余额总和
	TotalPlayerFrozen  decimal.Decimal `json:"total_player_frozen"`  // 玩家冻结余额总和

	// 房主侧
	TotalOwnerBalance    decimal.Decimal `json:"total_owner_balance"`     // 房主可用余额总和
	TotalOwnerCommission decimal.Decimal `json:"total_owner_commission"`  // 房主佣金收益总和
	TotalMargin          decimal.Decimal `json:"total_margin"`            // 房主保证金总和
	TotalCustodyQuota    decimal.Decimal `json:"total_custody_quota"`     // 房主托管额度总和（历史兼容）

	// 平台侧
	PlatformBalance decimal.Decimal `json:"platform_balance"` // 平台账户余额

	// 汇总
	SystemTotalFunds   decimal.Decimal `json:"system_total_funds"`   // 系统内资金总和
	TotalOwnerDeposit  decimal.Decimal `json:"total_owner_deposit"`  // 房主累计充值
	TotalOwnerWithdraw decimal.Decimal `json:"total_owner_withdraw"` // 房主累计提现
	ExpectedTotal      decimal.Decimal `json:"expected_total"`       // 预期总额（净充值）
	Difference         decimal.Decimal `json:"difference"`           // 差额
}

// FundReconciliationReport 资金对账报告（详细版）
type FundReconciliationReport struct {
	// 外部资金注入（只有房主才能和外部有资金往来）
	ExternalFunds struct {
		OwnerDeposit  decimal.Decimal `json:"owner_deposit"`  // 房主充值总额（已批准）
		MarginDeposit decimal.Decimal `json:"margin_deposit"` // 保证金充值总额（已批准）
		OwnerWithdraw decimal.Decimal `json:"owner_withdraw"` // 房主提现总额（已批准）
		NetInflow     decimal.Decimal `json:"net_inflow"`     // 净流入 = 充值 - 提现
	} `json:"external_funds"`

	// 系统内资金分布
	SystemFunds struct {
		PlayerBalance    decimal.Decimal `json:"player_balance"`    // 玩家可用余额
		PlayerFrozen     decimal.Decimal `json:"player_frozen"`     // 玩家冻结余额
		OwnerBalance     decimal.Decimal `json:"owner_balance"`     // 房主可用余额
		OwnerCommission  decimal.Decimal `json:"owner_commission"`  // 房主佣金收益
		OwnerMargin      decimal.Decimal `json:"owner_margin"`      // 房主保证金
		PlatformBalance  decimal.Decimal `json:"platform_balance"`  // 平台余额
		Total            decimal.Decimal `json:"total"`             // 系统内资金总和
	} `json:"system_funds"`

	// 对账结果
	Reconciliation struct {
		ExpectedTotal decimal.Decimal `json:"expected_total"` // 预期总额（净流入）
		ActualTotal   decimal.Decimal `json:"actual_total"`   // 实际总额（系统内资金）
		Difference    decimal.Decimal `json:"difference"`     // 差异
		IsBalanced    bool            `json:"is_balanced"`    // 是否平衡
	} `json:"reconciliation"`

	// 差异分析
	Analysis struct {
		UnrecordedMargin decimal.Decimal `json:"unrecorded_margin"` // 未记录的保证金（数据库直接设置）
		Explanation      string          `json:"explanation"`       // 差异说明
	} `json:"analysis"`
}

// FundConservationHistory 对账历史记录（全局 + 房主维度）
type FundConservationHistory struct {
	ID                       int64           `json:"id" db:"id"`
	Scope                    string          `json:"scope" db:"scope"`                 // global/owner
	OwnerID                  *int64          `json:"owner_id,omitempty" db:"owner_id"` // scope=owner 时有效
	PeriodType               string          `json:"period_type" db:"period_type"`     // 2h/daily
	PeriodStart              time.Time       `json:"period_start" db:"period_start"`
	PeriodEnd                time.Time       `json:"period_end" db:"period_end"`
	TotalPlayerBalance       decimal.Decimal `json:"total_player_balance" db:"total_player_balance"`
	TotalPlayerFrozen        decimal.Decimal `json:"total_player_frozen" db:"total_player_frozen"`
	TotalCustodyQuota        decimal.Decimal `json:"total_custody_quota" db:"total_custody_quota"`
	TotalMargin              decimal.Decimal `json:"total_margin" db:"total_margin"`
	OwnerRoomBalance         decimal.Decimal `json:"owner_room_balance" db:"owner_room_balance"`
	OwnerWithdrawableBalance decimal.Decimal `json:"owner_withdrawable_balance" db:"owner_withdrawable_balance"`
	OwnerFrozenBalance       decimal.Decimal `json:"owner_frozen_balance" db:"owner_frozen_balance"`
	PlatformBalance          decimal.Decimal `json:"platform_balance" db:"platform_balance"`
	Difference               decimal.Decimal `json:"difference" db:"difference"`
	IsBalanced               bool            `json:"is_balanced" db:"is_balanced"`
	CreatedAt                time.Time       `json:"created_at" db:"created_at"`
}

// FundConservationHistoryQuery 对账历史查询参数
type FundConservationHistoryQuery struct {
	Scope         *string    `form:"scope"` // global / owner
	OwnerID       *int64     `form:"owner_id"`
	PeriodType    *string    `form:"period_type"` // 2h / daily
	FromCreatedAt *time.Time `form:"from_created_at" time_format:"2006-01-02T15:04:05Z07:00"`
	ToCreatedAt   *time.Time `form:"to_created_at" time_format:"2006-01-02T15:04:05Z07:00"`
	Page          int        `form:"page" binding:"min=1"`
	PageSize      int        `form:"page_size" binding:"min=1,max=100"`
}
