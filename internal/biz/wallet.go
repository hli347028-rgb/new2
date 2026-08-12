package biz

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

const (
	RechargeStatusPending   = "pending"
	RechargeStatusConfirmed = "confirmed"
	RechargeStatusRejected  = "rejected"

	PayFromRecharge = "recharge"
	PayFromReward   = "reward"

	OrderStatusActive = "active"
	OrderStatusExited = "exited"

	// Legacy aliases
	OrderStatusCompleted    = "exited"
	WithdrawStatusPending   = "pending"
	WithdrawStatusCompleted = "completed"
	WithdrawStatusFailed    = "failed"

	TokenUSDT = "USDT"
	TokenAIX  = "AIX"
	TokenWIN  = "WIN"

	RewardTypeStaticAix       = "static_aix"
	RewardTypeDynamicUsdt     = "dynamic_usdt"
	RewardTypeMgmt            = "mgmt"
	RewardTypeMgmtPoolRelease = "mgmt_pool_release"
	RewardTypeExitAccel       = "exit_accel"
	RewardTypeTransferIn      = "transfer_in"
	RewardTypeTransferOut     = "transfer_out"
)

// GetWinPrice 返回 WIN 代币价格（USDT/枚）。
// TODO: 后续接入链上预言机或外部价格源，当前暂返回管理后台配置价。
func GetWinPrice() float64 {
	return WinPrice
}

// Recharge represents a USDT recharge order.
type Recharge struct {
	ID            int64
	UserID        int64
	Address       string // from_address / user address for display
	Amount        string
	Message       string
	TxHash        string
	FromAddress   string
	ToAddress     string
	Status        string
	ExpireAt      time.Time
	CreatedAt     time.Time
	ConfirmedAt   *time.Time
	CreatedTime   time.Time
	ConfirmedTime *time.Time
}

// Transfer internal transfer record.
type Transfer struct {
	ID                int64
	FromUserID        int64
	ToUserID          int64
	Asset             string
	Amount            string
	PayFrom           string
	FromRechargeDebit string
	FromRewardDebit   string
	ToCreditReward    string
	ToCreditAix       string
	Remark            string
	CreatedTime       time.Time
}

type SelfTransferRecord struct {
	ID          int64
	Asset       string
	Amount      string
	FromWallet  string
	ToWallet    string
	CreatedTime time.Time
}

type LinealTransferRecord struct {
	ID                  int64
	Direction           string
	Relationship        string
	CounterpartyUserID  int64
	CounterpartyAddress string
	Asset               string
	Amount              string
	FromWallet          string
	ToWallet            string
	CreatedTime         time.Time
}

// RewardLog AIX reward ledger row.
type RewardLog struct {
	ID             int64
	UserID         int64
	FromUserID     *int64
	OrderID        *int64
	BatchID        *int64
	Type           string
	Asset          string
	Amount         string
	BaseAmount     string
	Rate           string
	ExitApplied    string
	Meta           string
	SettlementDate string
	CreatedTime    time.Time
}

type MgmtRewardSummary struct {
	Released string
	Pending  string
	Total    string
}

type MgmtReward struct {
	ID             int64
	UserID         int64
	FromUserID     int64
	SourceOrderID  int64
	BaseAmount     string
	Rate           string
	TotalAmount    string
	ReleasedAmount string
	PendingAmount  string
	CreatedTime    time.Time
}

// ClaimRecord legacy stub
type ClaimRecord struct {
	ID        int64
	UserID    int64
	Amount    string
	CreatedAt time.Time
}

// OrderReleaseSummary compatibility for GetBalance mapping
type OrderReleaseSummary struct {
	ExitTotal     string
	ReleasedTotal string
	PendingTotal  string
	UnexitedTotal string
	TotalNodes    int32
}

// Withdrawal legacy stub
type Withdrawal struct {
	ID        int64
	UserID    int64
	Address   string
	ToAddress string
	Amount    string
	Fee       string
	NetAmount string
	Status    string
	TxHash    string
	Asset     string
	CreatedAt time.Time
}

// ExchangeRecord AIX → WIN 兑换记录
type ExchangeRecord struct {
	ID            int64
	UserID        int64
	UserAddress   string
	FromAsset     string
	FromAmount    string
	ToAsset       string
	ToAmount      string
	ExchangePrice string
	Status        string
	Remark        string
	CreatedTime   time.Time
}

// Product legacy stub
type Product struct {
	ID          int64
	Name        string
	Description string
	Price       string
	Stock       int32
	Status      int32
}

// Order represents an AIX subscribe order.
type Order struct {
	ID           int64
	UserID       int64
	Principal    string
	ExitCap      string
	EarnedTotal  string
	DirectBase   string
	FromRecharge string
	FromReward   string
	FundSource   string
	Status       string
	ExitedTime   *time.Time
	CreatedTime  time.Time
	UpdatedTime  time.Time

	// Legacy field aliases for proto/admin mapping
	ProductID      int64
	ProductName    string
	Quantity       int32
	TotalAmount    string
	ExitMultiplier string
	ExitTarget     string
	ReleasedAmount string
	ReleaseDay     int32
	CycleDay       int32
	RateGoingUp    bool
	CreatedAt      time.Time
}

// SyncCompatFields fills legacy order aliases.
func (o *Order) SyncCompatFields() {
	if o == nil {
		return
	}
	o.TotalAmount = o.Principal
	o.ExitTarget = o.ExitCap
	o.ReleasedAmount = o.EarnedTotal
	o.ProductName = o.FundSource
	o.ExitMultiplier = "4"
	o.Quantity = 1
	o.CreatedAt = o.CreatedTime
}

// WalletRepo handles wallet persistence.
type WalletRepo interface {
	CreateRecharge(ctx context.Context, recharge *Recharge) (*Recharge, error)
	FindRecharge(ctx context.Context, id int64) (*Recharge, error)
	FindRechargeByTxHash(ctx context.Context, txHash string) (*Recharge, error)
	ConfirmRecharge(ctx context.Context, id int64, txHash string) error
	ConfirmRechargeCredit(ctx context.Context, id int64, txHash string) (newRechargeBalance string, err error)
	DeletePendingRecharge(ctx context.Context, id int64) error
	AutoCreditChainRecharge(ctx context.Context, txHash, fromAddress, toAddress, amount string, blockNumber uint64) (credited bool, err error)
	ListRechargesByUser(ctx context.Context, userID int64) ([]*Recharge, error)

	Subscribe(ctx context.Context, userID int64, amount, payFrom string, exitMul, directRate float64) (*Order, string, error)
	ListOrdersByUser(ctx context.Context, userID int64) ([]*Order, error)
	ListAllOrders(ctx context.Context) ([]*AdminOrderDetail, error)
	FindOrder(ctx context.Context, id int64) (*Order, error)
	RemainingExitCapacity(ctx context.Context, userID int64) (string, error)

	CreateTransfer(ctx context.Context, t *Transfer) (*Transfer, error)
	ListTransfersByUser(ctx context.Context, userID int64) ([]*Transfer, error)
	// MoveRechargeToReward 同用户：充值钱包 USDT → 奖励钱包（不触发直推）
	MoveRechargeToReward(ctx context.Context, userID int64, amount string) (rechargeBal, rewardBal string, err error)

	CreateRewardLog(ctx context.Context, log *RewardLog) error
	ListRewardLogsByUser(ctx context.Context, userID int64) ([]*RewardLog, error)
	GetMgmtRewardSummary(ctx context.Context, userID int64) (*MgmtRewardSummary, error)
	ListMgmtRewardsByUser(ctx context.Context, userID int64) ([]*MgmtReward, error)

	GetAixPrice(ctx context.Context, date string) (string, error)
	UpsertAixPrice(ctx context.Context, date, price, remark string) error

	// ExchangeAixToWin AIX → WIN 兑换：扣 AixBalance，加 WinBalance，记录 ExchangeRecord
	ExchangeAixToWin(ctx context.Context, userID int64, aixAmount string) (*ExchangeRecord, string, string, error)
	ListExchangeRecordsByUser(ctx context.Context, userID int64) ([]*ExchangeRecord, error)
	ListAllExchangeRecords(ctx context.Context) ([]*ExchangeRecord, error)

	// CreateWinWithdrawal WIN 代币提现（AIX 当前禁止提现）
	CreateWinWithdrawal(ctx context.Context, userID int64, amount, toAddress string) (*Withdrawal, string, error)
	ListWithdrawalsByUser(ctx context.Context, userID int64) ([]*Withdrawal, error)
	// Legacy stubs
	CreateWithdrawal(ctx context.Context, userID int64, amount, fee, netAmount, toAddress string) (*Withdrawal, string, error)
	CreateClaimRecord(ctx context.Context, record *ClaimRecord) error
	ListClaimRecordsByUser(ctx context.Context, userID int64) ([]*ClaimRecord, error)
	SumClaimedByUser(ctx context.Context, userID int64) (string, error)
	SumWithdrawnByUser(ctx context.Context, userID int64) (string, error)
	ListAllWithdrawals(ctx context.Context) ([]*Withdrawal, error)
	ApproveWithdrawal(ctx context.Context, id int64) error
	ListProducts(ctx context.Context) ([]*Product, error)
	FindProduct(ctx context.Context, id int64) (*Product, error)
	FindProductByPrice(ctx context.Context, price string) (*Product, error)
	SubscribeProduct(ctx context.Context, userID, productID int64, quantity int32, totalAmount string) (*Order, string, error)
	SubscribeByAmount(ctx context.Context, userID int64, totalAmount, productName string) (*Order, string, error)
	ListAllProducts(ctx context.Context) ([]*Product, error)
	CreateProduct(ctx context.Context, product *Product) (*Product, error)
	AdminUpdateProduct(ctx context.Context, product *Product) (*Product, error)
	AdminUpdateOrder(ctx context.Context, update *AdminOrderUpdate) (*Order, error)
}

// AdminOrderDetail 管理员可见的订单详情
type AdminOrderDetail struct {
	Order       *Order
	UserAddress string
}

// AdminOrderUpdate 管理员修改订单字段
type AdminOrderUpdate struct {
	OrderID        int64
	Quantity       int32
	TotalAmount    string
	Status         string
	ExitMultiplier string
	ExitTarget     string
	ReleasedAmount string
	ReleaseDay     int32
	CycleDay       int32
}

func ParseAmount(amount string) (decimal.Decimal, error) {
	return decimal.NewFromString(amount)
}
