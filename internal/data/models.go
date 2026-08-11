package data

import (
	"time"

	"github.com/shopspring/decimal"
)

const decimalType = "decimal(36,18)"

type UserPO struct {
	ID              int64           `gorm:"primaryKey;autoIncrement"`
	Address         string          `gorm:"uniqueIndex;size:42;not null"`
	InviterID       *int64          `gorm:"index"`
	InviteCode      string          `gorm:"uniqueIndex;size:64;not null"`
	UsdtRecharge    decimal.Decimal `gorm:"column:usdt_recharge;type:decimal(36,18);default:0;not null"`
	UsdtReward      decimal.Decimal `gorm:"column:usdt_reward;type:decimal(36,18);default:0;not null"`
	AixBalance      decimal.Decimal `gorm:"column:aix_balance;type:decimal(36,18);default:0;not null"`       // AIX 代币数（静态换算入账）
	StaticUsdtTotal decimal.Decimal `gorm:"column:static_usdt_total;type:decimal(36,18);default:0;not null"` // 静态总收益（USDT 金本位累计）
	MgmtLevel       int32           `gorm:"column:mgmt_level;default:0;not null"`
	MgmtLevelLocked bool            `gorm:"column:mgmt_level_locked;default:false;not null"`
	LargeAreaPerf   decimal.Decimal `gorm:"column:large_area_perf;type:decimal(36,18);default:0;not null"`
	SmallAreaPerf   decimal.Decimal `gorm:"column:small_area_perf;type:decimal(36,18);default:0;not null"`
	TeamPerf        decimal.Decimal `gorm:"column:team_perf;type:decimal(36,18);default:0;not null"`
	Status          int32           `gorm:"default:1;not null"`
	Role            string          `gorm:"size:16;default:user;not null"` // app admin helper, not in business DDL
	CreatedTime     time.Time       `gorm:"column:created_time;autoCreateTime"`
	UpdatedTime     time.Time       `gorm:"column:updated_time;autoUpdateTime"`
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

// WithdrawalPO 仅支持提现 AIX 代币（合约信息未就绪时链上打款留空）
type WithdrawalPO struct {
	ID          int64           `gorm:"primaryKey;autoIncrement"`
	UserID      int64           `gorm:"index;not null"`
	Asset       string          `gorm:"size:16;default:AIX;not null"`
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

type AixPricePO struct {
	ID            int64           `gorm:"primaryKey;autoIncrement"`
	Price         decimal.Decimal `gorm:"type:decimal(36,18);not null"`
	EffectiveDate string          `gorm:"column:effective_date;type:date;uniqueIndex;not null"`
	Remark        string          `gorm:"size:255"`
	CreatedTime   time.Time       `gorm:"column:created_time;autoCreateTime"`
}

func (AixPricePO) TableName() string { return "aix_prices" }

type SettlementBatchPO struct {
	ID             int64           `gorm:"primaryKey;autoIncrement"`
	SettlementDate string          `gorm:"column:settlement_date;type:date;uniqueIndex;not null"`
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
