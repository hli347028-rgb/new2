package biz

import (
	"context"
	"time"

	"backend/internal/conf"

	"github.com/shopspring/decimal"
)

// Runtime AIX config (hot-updated from settings)
var (
	StaticRate           = float64(conf.DefaultStaticRate)
	ExitMultiplier       = float64(conf.DefaultExitMultiplier)
	DirectRate           = float64(conf.DefaultDirectRate)
	MgmtThresholds       = conf.DefaultMgmtThresholds()
	MgmtRates            = conf.DefaultMgmtRates()
	MgmtCountsTowardExit = true
	AixPriceInitial      = float64(conf.DefaultAixPrice)
)

// ApplyAixConfig hot-updates AIX business parameters.
func ApplyAixConfig(snap *conf.SystemConfigSnapshot) {
	if snap == nil {
		return
	}
	conf.NormalizeBusinessDefaults(snap)
	StaticRate = snap.StaticRate
	ExitMultiplier = snap.ExitMultiplier
	DirectRate = snap.DirectRate
	MgmtThresholds = append([]float64(nil), snap.MgmtThresholds...)
	MgmtRates = append([]float64(nil), snap.MgmtRates...)
	MgmtCountsTowardExit = snap.MgmtCountsTowardExit
	if snap.AixPriceInitial > 0 {
		AixPriceInitial = snap.AixPriceInitial
	}
}

// StakingOrder alias used by settlement (maps to Order fields)
type StakingOrder struct {
	ID          int64
	UserID      int64
	Principal   string
	ExitCap     string
	EarnedTotal string
	Status      string
	FundSource  string
	CreatedAt   time.Time
}

// ReleaseRecord mapped from reward_logs static_aix for legacy list APIs
type ReleaseRecord struct {
	ID             int64
	UserID         int64
	OrderID        int64
	SettlementDate string
	Rate           string
	Amount         string
	ReleaseDay     int32
	CreatedAt      time.Time
}

// ReferralReward mapped from reward_logs dynamic_usdt
type ReferralReward struct {
	ID                int64
	BeneficiaryUserID int64
	SourceUserID      int64
	SourceOrderID     int64
	Generation        int32
	BaseAmount        string
	Rate              string
	RewardAmount      string
	SettlementDate    string
	CreatedAt         time.Time
}

// EcoReward stub type for disabled eco API
type EcoReward struct {
	ID             int64
	UserID         int64
	SettlementDate string
	CommunityLevel string
	CommunityStake string
	BaseAmount     string
	BaseRate       string
	BaseReward     string
	EqualReward    string
	TotalReward    string
	CreatedAt      time.Time
}

// SettlementBatch AIX settlement batch
type SettlementBatch struct {
	ID             int64
	SettlementDate string
	AixPrice       string
	Status         string
	StaticCount    int32
	StaticAmount   string
	MgmtCount      int32
	MgmtAmount     string
	StartedAt      time.Time
	FinishedAt     *time.Time
	ErrorMsg       string
	CreatedTime    time.Time
}

// CalcExitCap principal * exit_multiplier
func CalcExitCap(principal decimal.Decimal) decimal.Decimal {
	mul := decimal.NewFromFloat(ExitMultiplier)
	if !mul.IsPositive() {
		mul = decimal.NewFromInt(4)
	}
	return principal.Mul(mul)
}

// MgmtLevelByPerf returns W level 0-10 from small-area performance.
func MgmtLevelByPerf(perf decimal.Decimal) int32 {
	level := int32(0)
	for i := len(MgmtThresholds) - 1; i >= 0; i-- {
		th := decimal.NewFromFloat(MgmtThresholds[i])
		if perf.GreaterThanOrEqual(th) {
			return int32(i + 1)
		}
	}
	return level
}

// MgmtRateForLevel returns rate for W level (1-10); 0 for W0.
func MgmtRateForLevel(level int32) float64 {
	if level < 1 || int(level) > len(MgmtRates) {
		return 0
	}
	return MgmtRates[level-1]
}

// StakingRepo settlement & reward persistence (AIX)
type StakingRepo interface {
	ListActiveOrders(ctx context.Context) ([]*StakingOrder, error)
	ListActiveOrdersByUser(ctx context.Context, userID int64) ([]*StakingOrder, error)
	UpdateOrderEarned(ctx context.Context, orderID int64, earnedTotal, status string, exitedTime *time.Time) error

	CreateRewardLog(ctx context.Context, log *RewardLog) error
	ListRewardLogsByUser(ctx context.Context, userID int64) ([]*RewardLog, error)
	ListStaticRewardsAsRelease(ctx context.Context, userID int64) ([]*ReleaseRecord, error)
	ListDynamicRewardsAsReferral(ctx context.Context, userID int64) ([]*ReferralReward, error)
	HasStaticReward(ctx context.Context, orderID int64, date string) (bool, error)

	GetAixPrice(ctx context.Context, date string) (string, error)
	UpsertAixPrice(ctx context.Context, date, price, remark string) error

	HasCompletedSettlement(ctx context.Context, date string) (bool, error)
	CreateSettlementBatch(ctx context.Context, batch *SettlementBatch) error
	FinishSettlementBatch(ctx context.Context, id int64, status string, staticCount int32, staticAmount string, mgmtCount int32, mgmtAmount string, errMsg string) error
	ListSettlementBatches(ctx context.Context, offset, limit int) ([]*SettlementBatch, int64, error)
	SumStaticByDate(ctx context.Context, date string) (string, error)

	// Legacy no-op stubs for old call sites
	ListOrdersByUser(ctx context.Context, userID int64) ([]*StakingOrder, error)
	UpdateOrderAfterRelease(ctx context.Context, orderID int64, releasedAmount, status string, releaseDay, rateIndex int32, rateGoingUp bool) error
	CreateReleaseRecord(ctx context.Context, record *ReleaseRecord) error
	ListReleaseRecordsByUser(ctx context.Context, userID int64) ([]*ReleaseRecord, error)
	CreateReferralReward(ctx context.Context, reward *ReferralReward) error
	ListReferralRewardsByUser(ctx context.Context, userID int64) ([]*ReferralReward, error)
	SumReferralByOrderDate(ctx context.Context, orderID int64, settlementDate string) (string, error)
	CreateEcoReward(ctx context.Context, reward *EcoReward) error
	ListEcoRewardsByUser(ctx context.Context, userID int64) ([]*EcoReward, error)
	CountEcoRewardsByDate(ctx context.Context, date string) (int64, error)
	HasEcoReward(ctx context.Context, userID int64, date string) (bool, error)
	ListReleaseSettlementDates(ctx context.Context) ([]string, error)
	GetLatestSettlementBatch(ctx context.Context, date string) (*SettlementBatch, error)
	SumReleaseByUserDate(ctx context.Context, userID int64, date string) (string, error)
	SumReleaseByUserDateSince(ctx context.Context, userID int64, date string, since time.Time) (string, error)
	SumReleaseByDate(ctx context.Context, date string) (string, error)
	SumReleaseForBatch(ctx context.Context, date string, startedAt time.Time, finishedAt *time.Time) (string, error)
	ListUserIDsWithReleaseOnDate(ctx context.Context, date string) ([]int64, error)
	ListUserIDsWithReleaseOnDateSince(ctx context.Context, date string, since time.Time) ([]int64, error)
	SumSettledByUser(ctx context.Context, userID int64) (string, error)
}

// Legacy helpers kept so admin config compile paths don't break
var (
	ReferralRates  = []float64{conf.DefaultDirectRate}
	MaxReferralGen int32 = 1
)

func ApplyReferralConfig(maxGen int32, rates []float64) {
	if maxGen > 0 {
		MaxReferralGen = maxGen
	}
	if len(rates) > 0 {
		ReferralRates = append([]float64(nil), rates...)
	}
}

func ReferralRateForGen(gen int32) float64 {
	if gen == 1 {
		return DirectRate
	}
	return 0
}

func CalcExitMultiplier(principal decimal.Decimal) decimal.Decimal {
	return decimal.NewFromFloat(ExitMultiplier)
}
