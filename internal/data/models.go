package data

import (
	"time"

	"github.com/shopspring/decimal"
)

const decimalType = "decimal(36,18)"

type UserPO struct {
	ID                int64           `gorm:"primaryKey;autoIncrement"`
	Address           string          `gorm:"uniqueIndex;size:42;not null"`
	InviterID         *int64          `gorm:"index"`
	InviteCode        string          `gorm:"uniqueIndex;size:64;not null"`
	UsdtRecharge      decimal.Decimal `gorm:"column:usdt_recharge;type:decimal(36,18);default:0;not null"`
	UsdtReward        decimal.Decimal `gorm:"column:usdt_reward;type:decimal(36,18);default:0;not null"`
	AixBalance        decimal.Decimal `gorm:"column:aix_balance;type:decimal(36,18);default:0;not null"`         // AIX 代币数（静态换算入账）
	WinBalance        decimal.Decimal `gorm:"column:win_balance;type:decimal(36,18);default:0;not null"`         // WIN 代币数
	PendingMgmtReward decimal.Decimal `gorm:"column:pending_mgmt_reward;type:decimal(36,18);default:0;not null"` // 兼容旧列；业务以 OverflowReward 为准
	OverflowReward    decimal.Decimal `gorm:"column:overflow_reward;type:decimal(36,18);default:0;not null"`     // 溢出奖励（订单全部出局后剩余直推/管理奖）
	StaticUsdtTotal   decimal.Decimal `gorm:"column:static_usdt_total;type:decimal(36,18);default:0;not null"`   // 静态总收益（USDT 金本位累计）
	MgmtLevel         int32           `gorm:"column:mgmt_level;default:0;not null"`
	MgmtLevelLocked   bool            `gorm:"column:mgmt_level_locked;default:false;not null"`
	LargeAreaPerf     decimal.Decimal `gorm:"column:large_area_perf;type:decimal(36,18);default:0;not null"`
	SmallAreaPerf     decimal.Decimal `gorm:"column:small_area_perf;type:decimal(36,18);default:0;not null"`
	TeamPerf          decimal.Decimal `gorm:"column:team_perf;type:decimal(36,18);default:0;not null"`
	Status            int32           `gorm:"default:1;not null"`
	Role              string          `gorm:"size:16;default:user;not null"` // app admin helper, not in business DDL
	CreatedTime       time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime       time.Time       `gorm:"column:updated_time;autoUpdateTime"`
}

func (UserPO) TableName() string { return "users" }

type OrderPO struct {
	ID           int64           `gorm:"primaryKey;autoIncrement"`
	UserID       int64           `gorm:"index;not null"`
	Principal    decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	ExitCap      decimal.Decimal `gorm:"column:exit_cap;type:decimal(36,18);not null"`
	EarnedTotal  decimal.Decimal `gorm:"column:earned_total;type:decimal(36,18);default:0;not null"`
	DirectBase   decimal.Decimal `gorm:"column:direct_base;type:decimal(36,18);default:0;not null"`
	FromRecharge decimal.Decimal `gorm:"column:from_recharge;type:decimal(36,18);default:0;not null"`
	FromReward   decimal.Decimal `gorm:"column:from_reward;type:decimal(36,18);default:0;not null"`
	FromWin      decimal.Decimal `gorm:"column:from_win;type:decimal(36,18);default:0;not null"` // WIN 扣款数量（按认购时 win_price 折算）
	FundSource   string          `gorm:"column:fund_source;size:16;not null"`
	Status       string          `gorm:"size:16;default:active;not null"`
	ExitedTime   *time.Time      `gorm:"column:exited_time"`
	CreatedTime  time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime  time.Time       `gorm:"column:updated_time;autoUpdateTime"`
}

func (OrderPO) TableName() string { return "orders" }

type RechargePO struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	UserID        int64           `gorm:"index;not null"`
	Asset         string          `gorm:"size:16;default:USDT;not null;index"` // USDT | WIN
	Amount        decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	TxHash        string          `gorm:"size:66;uniqueIndex;not null"`
	FromAddress   string          `gorm:"column:from_address;size:42"`
	ToAddress     string          `gorm:"column:to_address;size:42"`
	Status        string          `gorm:"size:16;default:pending;not null"`
	Message       string          `gorm:"type:text"` // signing message for confirm flow
	ExpireAt      *time.Time      `gorm:"column:expire_at"`
	ConfirmedTime *time.Time      `gorm:"column:confirmed_time"`
	CreatedTime   time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime   time.Time       `gorm:"column:updated_time;autoUpdateTime"`
}

func (RechargePO) TableName() string { return "recharges" }

type TransferPO struct {
	ID                int64           `gorm:"primaryKey;autoIncrement"`
	FromUserID        int64           `gorm:"column:from_user_id;index;not null"`
	ToUserID          int64           `gorm:"column:to_user_id;index;not null"`
	Asset             string          `gorm:"size:16;not null"`
	Amount            decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	PayFrom           string          `gorm:"column:pay_from;size:16"`
	FromRechargeDebit decimal.Decimal `gorm:"column:from_recharge_debit;type:decimal(36,18);default:0;not null"`
	FromRewardDebit   decimal.Decimal `gorm:"column:from_reward_debit;type:decimal(36,18);default:0;not null"`
	ToCreditReward    decimal.Decimal `gorm:"column:to_credit_reward;type:decimal(36,18);default:0;not null"`
	ToCreditAix       decimal.Decimal `gorm:"column:to_credit_aix;type:decimal(36,18);default:0;not null"`
	Remark            string          `gorm:"size:255"`
	CreatedTime       time.Time       `gorm:"column:created_time;autoCreateTime"`
}

func (TransferPO) TableName() string { return "transfers" }

// WithdrawalPO 支持 AIX 与 WIN 代币提现（AIX 当前禁止提现，仅 WIN 可提现）
type WithdrawalPO struct {
	ID          int64           `gorm:"primaryKey;autoIncrement"`
	UserID      int64           `gorm:"index;not null"`
	Asset       string          `gorm:"size:16;default:AIX;not null"` // AIX 或 WIN
	Amount      decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	Fee         decimal.Decimal `gorm:"type:decimal(36,18);default:0;not null"`
	PayAmount   decimal.Decimal `gorm:"column:pay_amount;type:decimal(36,18);not null"`
	ToAddress   string          `gorm:"column:to_address;size:42;not null"`
	TxHash      string          `gorm:"column:tx_hash;size:66"`
	Status      string          `gorm:"size:16;default:pending;not null"`
	Remark      string          `gorm:"size:255"`
	CreatedTime time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime time.Time       `gorm:"column:updated_time;autoUpdateTime"`
}

func (WithdrawalPO) TableName() string { return "withdrawals" }

// ExchangeRecordPO AIX → WIN 兑换记录
type ExchangeRecordPO struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	UserID        int64           `gorm:"index;not null"`
	FromAsset     string          `gorm:"column:from_asset;size:16;not null"` // 固定 AIX
	FromAmount    decimal.Decimal `gorm:"column:from_amount;type:decimal(36,18);not null"`
	ToAsset       string          `gorm:"column:to_asset;size:16;not null"` // 固定 WIN
	ToAmount      decimal.Decimal `gorm:"column:to_amount;type:decimal(36,18);not null"`
	FeeAmount     decimal.Decimal `gorm:"column:fee_amount;type:decimal(36,18);not null;default:0"`
	ExchangePrice decimal.Decimal `gorm:"column:exchange_price;type:decimal(36,18);not null"` // 兑换时的 WIN 价格（USDT/枚）
	FeeRate       decimal.Decimal `gorm:"column:fee_rate;type:decimal(12,6);not null;default:0"` // 兑换时的手续费率
	Status        string          `gorm:"size:16;default:completed;not null"`                 // completed
	Remark        string          `gorm:"size:255"`
	CreatedTime   time.Time       `gorm:"column:created_time;autoCreateTime"`
}

func (ExchangeRecordPO) TableName() string { return "exchange_records" }

type RewardLogPO struct {
	ID             int64            `gorm:"primaryKey;autoIncrement"`
	UserID         int64            `gorm:"index;not null"`
	FromUserID     *int64           `gorm:"column:from_user_id"`
	OrderID        *int64           `gorm:"column:order_id;index"`
	BatchID        *int64           `gorm:"column:batch_id;index"`
	Type           string           `gorm:"size:32;not null"`
	Asset          string           `gorm:"size:16;not null"`
	Amount         decimal.Decimal  `gorm:"type:decimal(36,18);not null"`
	BaseAmount     *decimal.Decimal `gorm:"column:base_amount;type:decimal(36,18)"`
	Rate           *decimal.Decimal `gorm:"type:decimal(36,18)"`
	ExitApplied    decimal.Decimal  `gorm:"column:exit_applied;type:decimal(36,18);default:0;not null"`
	Meta           *string          `gorm:"type:json"`
	SettlementDate *string          `gorm:"column:settlement_date;type:date;index"`
	CreatedTime    time.Time        `gorm:"column:created_time;autoCreateTime"`
}

func (RewardLogPO) TableName() string { return "reward_logs" }

// MgmtRewardPO stores the full one-time differential entitlement generated by
// a downline subscription. ReleasedAmount is credited as the beneficiary's
// own subscription principal creates management-reward capacity; the
// remainder stays pending without being lost.
type MgmtRewardPO struct {
	ID             int64           `gorm:"primaryKey;autoIncrement"`
	UserID         int64           `gorm:"column:user_id;index;not null;uniqueIndex:uk_mgmt_source"`
	FromUserID     int64           `gorm:"column:from_user_id;index;not null"`
	SourceOrderID  int64           `gorm:"column:source_order_id;index;not null;uniqueIndex:uk_mgmt_source"`
	BaseAmount     decimal.Decimal `gorm:"column:base_amount;type:decimal(36,18);not null"`
	Rate           decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	TotalAmount    decimal.Decimal `gorm:"column:total_amount;type:decimal(36,18);not null"`
	ReleasedAmount decimal.Decimal `gorm:"column:released_amount;type:decimal(36,18);default:0;not null"`
	CreatedTime    time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime    time.Time       `gorm:"column:updated_time;autoUpdateTime"`
}

func (MgmtRewardPO) TableName() string { return "mgmt_rewards" }

type AixPricePO struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	Price         decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	EffectiveDate string          `gorm:"column:effective_date;type:date;uniqueIndex;not null"`
	Remark        string          `gorm:"size:255"`
	CreatedTime   time.Time       `gorm:"column:created_time;autoCreateTime"`
}

func (AixPricePO) TableName() string { return "aix_prices" }

// WinPricePO 当前 WIN 价格（全表仅保留 1 条，固定 ID=1，预言机/后台均覆盖更新）。
type WinPricePO struct {
	ID          int64           `gorm:"primaryKey"`
	Price       decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	Source      string          `gorm:"size:32;default:oracle;not null"`
	UpdatedTime time.Time       `gorm:"column:updated_time;autoUpdateTime"`
	CreatedTime time.Time       `gorm:"column:created_time;autoCreateTime"`
}

const WinPriceRowID int64 = 1

func (WinPricePO) TableName() string { return "win_prices" }

type SettlementBatchPO struct {
	ID             int64           `gorm:"primaryKey;autoIncrement"`
	SettlementDate string          `gorm:"column:settlement_date;type:date;index;not null"`
	AixPrice       decimal.Decimal `gorm:"column:aix_price;type:decimal(36,18);not null"`
	Status         string          `gorm:"size:16;default:running;not null"`
	StaticCount    int32           `gorm:"column:static_count;default:0;not null"`
	StaticAmount   decimal.Decimal `gorm:"column:static_amount;type:decimal(36,18);default:0;not null"`
	MgmtCount      int32           `gorm:"column:mgmt_count;default:0;not null"`
	MgmtAmount     decimal.Decimal `gorm:"column:mgmt_amount;type:decimal(36,18);default:0;not null"`
	StartedTime    *time.Time      `gorm:"column:started_time"`
	FinishedTime   *time.Time      `gorm:"column:finished_time"`
	ErrorMsg       string          `gorm:"column:error_msg;size:512"`
	CreatedTime    time.Time       `gorm:"column:created_time;autoCreateTime"`
}

func (SettlementBatchPO) TableName() string { return "settlement_batches" }

type SettingPO struct {
	ID          int64     `gorm:"primaryKey;autoIncrement"`
	Key         string    `gorm:"uniqueIndex;size:64;column:key;not null"`
	Value       string    `gorm:"type:json;not null"`
	CreatedTime time.Time `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime time.Time `gorm:"column:updated_time;autoUpdateTime"`
}

func (SettingPO) TableName() string { return "settings" }
