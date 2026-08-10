package v1

type GetBalanceRequest struct {
	Token string `json:"token"`
}

type GetBalanceReply struct {
	Address         string `json:"address"`
	Balance         string `json:"balance"`
	ReleasedBalance string `json:"released_balance"`
	ClaimableAmount string `json:"claimable_amount"`
	PendingAmount   string `json:"pending_amount"`
	ClaimedAmount   string `json:"claimed_amount"`
	TotalNodes      int32  `json:"total_nodes"`
	ServerTime      int64  `json:"server_time"`
	NextReleaseAt   int64  `json:"next_release_at"`
	UnexitedAmount  string `json:"unexited_amount"`
}

type CreateRechargeRequest struct {
	Token  string `json:"token"`
	Amount string `json:"amount"`
}

type CreateRechargeReply struct {
	RechargeID       int64    `json:"recharge_id"`
	Amount           string   `json:"amount"`
	DepositAddress   string   `json:"deposit_address"`
	UsdtContract     string   `json:"usdt_contract"`
	TokenSymbol      string   `json:"token_symbol"`
	Message          string   `json:"message"`
	ExpireAt         int64    `json:"expire_at"`
	DevMode          bool     `json:"dev_mode"`
	DepositAddresses []string `json:"deposit_addresses"`
	SplitAmounts     []string `json:"split_amounts"`
}

type ConfirmRechargeRequest struct {
	Token      string   `json:"token"`
	RechargeID int64    `json:"recharge_id"`
	TxHash     string   `json:"tx_hash"`
	Signature  string   `json:"signature"`
	TxHashes   []string `json:"tx_hashes"`
}

type ConfirmRechargeReply struct {
	Balance string `json:"balance"`
	Amount  string `json:"amount"`
}

type ListRechargesRequest struct {
	Token string `json:"token"`
}

type RechargeRecord struct {
	ID          int64  `json:"id"`
	Amount      string `json:"amount"`
	TxHash      string `json:"tx_hash"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	ConfirmedAt int64  `json:"confirmed_at"`
}

type ListRechargesReply struct {
	Recharges []*RechargeRecord `json:"recharges"`
}

type CreateWithdrawRequest struct {
	Token      string `json:"token"`
	Amount     string `json:"amount"`
	ToAddress  string `json:"to_address"`
	Signature  string `json:"signature"`
	WithdrawAt int64  `json:"withdraw_at"`
}

type CreateWithdrawReply struct {
	WithdrawID      int64  `json:"withdraw_id"`
	Amount          string `json:"amount"`
	Fee             string `json:"fee"`
	NetAmount       string `json:"net_amount"`
	ToAddress       string `json:"to_address"`
	Balance         string `json:"balance"`
	Status          string `json:"status"`
	ReleasedBalance string `json:"released_balance"`
}

type ClaimToAccountRequest struct {
	Token  string `json:"token"`
	Amount string `json:"amount"`
}

type ClaimToAccountReply struct {
	Balance         string `json:"balance"`
	ReleasedBalance string `json:"released_balance"`
	Amount          string `json:"amount"`
}

type ListWithdrawalsRequest struct {
	Token string `json:"token"`
}

type WithdrawalRecord struct {
	ID        int64  `json:"id"`
	Amount    string `json:"amount"`
	Fee       string `json:"fee"`
	NetAmount string `json:"net_amount"`
	ToAddress string `json:"to_address"`
	Status    string `json:"status"`
	TxHash    string `json:"tx_hash"`
	CreatedAt int64  `json:"created_at"`
}

type ListWithdrawalsReply struct {
	Withdrawals []*WithdrawalRecord `json:"withdrawals"`
}

type ListProductsRequest struct{}

type Product struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Stock       int32  `json:"stock"`
}

type ListProductsReply struct {
	Products []*Product `json:"products"`
}

type SubscribeRequest struct {
	Token     string `json:"token"`
	ProductID int64  `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Amount    string `json:"amount"`
}

type SubscribeReply struct {
	OrderID     int64  `json:"order_id"`
	ProductName string `json:"product_name"`
	TotalAmount string `json:"total_amount"`
	Balance     string `json:"balance"`
}

type ListOrdersRequest struct {
	Token string `json:"token"`
}

type Order struct {
	ID             int64  `json:"id"`
	ProductID      int64  `json:"product_id"`
	ProductName    string `json:"product_name"`
	Quantity       int32  `json:"quantity"`
	TotalAmount    string `json:"total_amount"`
	CreatedAt      int64  `json:"created_at"`
	Status         string `json:"status"`
	ExitMultiplier string `json:"exit_multiplier"`
	ExitTarget     string `json:"exit_target"`
	ReleasedAmount string `json:"released_amount"`
	ReleaseDay     int32  `json:"release_day"`
	CycleDay       int32  `json:"cycle_day"`
}

type ListOrdersReply struct {
	Orders []*Order `json:"orders"`
}

type ListReleaseRecordsRequest struct {
	Token string `json:"token"`
}

type ReleaseRecord struct {
	ID                  int64  `json:"id"`
	OrderID             int64  `json:"order_id"`
	SettlementDate      string `json:"settlement_date"`
	Rate                string `json:"rate"`
	Amount              string `json:"amount"`
	ReleaseDay          int32  `json:"release_day"`
	CreatedAt           int64  `json:"created_at"`
	OrderIndex          int32  `json:"order_index"`
	ReferralDistributed string `json:"referral_distributed"`
	ExitMultiplier      string `json:"exit_multiplier"`
	Money               string `json:"money"`
}

type ListReleaseRecordsReply struct {
	Records []*ReleaseRecord `json:"records"`
}

type ListReferralRewardsRequest struct {
	Token string `json:"token"`
}

type ReferralReward struct {
	ID             int64  `json:"id"`
	SourceUserID   int64  `json:"source_user_id"`
	SourceAddress  string `json:"source_address"`
	SourceOrderID  int64  `json:"source_order_id"`
	Generation     int32  `json:"generation"`
	BaseAmount     string `json:"base_amount"`
	Rate           string `json:"rate"`
	RewardAmount   string `json:"reward_amount"`
	SettlementDate string `json:"settlement_date"`
	CreatedAt      int64  `json:"created_at"`
	OrderIndex     int32  `json:"order_index"`
}

type ListReferralRewardsReply struct {
	Rewards []*ReferralReward `json:"rewards"`
}

type ListEcoRewardsRequest struct {
	Token string `json:"token"`
}

type EcoReward struct {
	ID             int64  `json:"id"`
	SettlementDate string `json:"settlement_date"`
	CommunityLevel string `json:"community_level"`
	CommunityStake string `json:"community_stake"`
	BaseAmount     string `json:"base_amount"`
	BaseRate       string `json:"base_rate"`
	BaseReward     string `json:"base_reward"`
	EqualReward    string `json:"equal_reward"`
	TotalReward    string `json:"total_reward"`
	CreatedAt      int64  `json:"created_at"`
}

type ListEcoRewardsReply struct {
	Rewards []*EcoReward `json:"rewards"`
}

type ListClaimRecordsRequest struct {
	Token string `json:"token"`
}

type ClaimRecord struct {
	ID        int64  `json:"id"`
	Amount    string `json:"amount"`
	CreatedAt int64  `json:"created_at"`
}

type ListClaimRecordsReply struct {
	Claims []*ClaimRecord `json:"claims"`
}
