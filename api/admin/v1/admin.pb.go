package v1

type ListUsersRequest struct {
	Token string `json:"token"`
}

type AdminUser struct {
	ID                int64  `json:"id"`
	Address           string `json:"address"`
	Balance           string `json:"balance"`
	ReleasedBalance   string `json:"released_balance"`
	Role              string `json:"role"`
	InviterID         *int64 `json:"inviter_id"`
	InviterAddress    string `json:"inviter_address"`
	InviteeCount      int32  `json:"invitee_count"`
	CommunityLevel    string `json:"community_level"`
	CommunityStake    string `json:"community_stake"`
	TeamStake         string `json:"team_stake"`
	LargeAreaPerf     string `json:"large_area_perf"`
	ShareProfitTotal  string `json:"share_profit_total"`
	EcoRewardTotal    string `json:"eco_reward_total"`
	WinBalance        string `json:"win_balance"`
	PendingMgmtReward string `json:"pending_mgmt_reward"`
	WithdrawReset     bool   `json:"withdraw_reset"`
	CreatedAt         int64  `json:"created_at"`
}

type ListUsersReply struct {
	Users []*AdminUser `json:"users"`
}

type UpdateUserRequest struct {
	Token             string `json:"token"`
	UserID            int64  `json:"user_id"`
	Balance           string `json:"balance"`
	ReleasedBalance   string `json:"released_balance"`
	Role              string `json:"role"`
	CommunityLevel    string `json:"community_level"`
	CommunityStake    string `json:"community_stake"`
	TeamStake         string `json:"team_stake"`
	ShareProfitTotal  string `json:"share_profit_total"`
	EcoRewardTotal    string `json:"eco_reward_total"`
	WinBalance        string `json:"win_balance"`
	PendingMgmtReward string `json:"pending_mgmt_reward"`
	InviterID         *int64 `json:"inviter_id"`
	WithdrawReset     *bool  `json:"withdraw_reset"`
}

type UpdateUserReply struct {
	User *AdminUser `json:"user"`
}

type GetConfigRequest struct {
	Token string `json:"token"`
}

type SystemConfig struct {
	JwtSecret       string   `json:"jwt_secret"`
	ChallengeTTL    string   `json:"challenge_ttl"`
	AdminAddresses  []string `json:"admin_addresses"`
	DepositAddress  string   `json:"deposit_address"`
	UsdtContract    string   `json:"usdt_contract"`
	UsdtDecimals    int32    `json:"usdt_decimals"`
	RPCURL          string   `json:"rpc_url"`
	MinSubscribe    string   `json:"min_subscribe"`
	WithdrawFeeRate float64  `json:"withdraw_fee_rate"`
	WinPrice        float64  `json:"win_price"`
}

type GetConfigReply struct {
	Config *SystemConfig `json:"config"`
}

type UpdateConfigRequest struct {
	Token           string   `json:"token"`
	JwtSecret       string   `json:"jwt_secret"`
	ChallengeTTL    string   `json:"challenge_ttl"`
	AdminAddresses  []string `json:"admin_addresses"`
	DepositAddress  string   `json:"deposit_address"`
	UsdtContract    string   `json:"usdt_contract"`
	UsdtDecimals    int32    `json:"usdt_decimals"`
	RPCURL          string   `json:"rpc_url"`
	MinSubscribe    string   `json:"min_subscribe"`
	WithdrawFeeRate float64  `json:"withdraw_fee_rate"`
	WinPrice        float64  `json:"win_price"`
}

type UpdateConfigReply struct {
	Config *SystemConfig `json:"config"`
}

type UpdateProductRequest struct {
	Token       string `json:"token"`
	ProductID   int64  `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Stock       int32  `json:"stock"`
	Status      int32  `json:"status"`
}

type AdminProduct struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Stock       int32  `json:"stock"`
	Status      int32  `json:"status"`
}

type ListProductsRequest struct {
	Token string `json:"token"`
}

type ListProductsReply struct {
	Products []*AdminProduct `json:"products"`
}

type UpdateProductReply struct {
	Product *AdminProduct `json:"product"`
}

type TriggerSettlementRequest struct {
	Token          string `json:"token"`
	SettlementDate string `json:"settlement_date"`
}

type TriggerSettlementReply struct {
	Message string `json:"message"`
}

type ListOrdersRequest struct {
	Token string `json:"token"`
}

type AdminOrder struct {
	ID             int64  `json:"id"`
	UserID         int64  `json:"user_id"`
	UserAddress    string `json:"user_address"`
	ProductID      int64  `json:"product_id"`
	ProductName    string `json:"product_name"`
	Quantity       int32  `json:"quantity"`
	TotalAmount    string `json:"total_amount"`
	Status         string `json:"status"`
	ExitMultiplier string `json:"exit_multiplier"`
	ExitTarget     string `json:"exit_target"`
	ReleasedAmount string `json:"released_amount"`
	ReleaseDay     int32  `json:"release_day"`
	CycleDay       int32  `json:"cycle_day"`
	CreatedAt      int64  `json:"created_at"`
}

type ListOrdersReply struct {
	Orders []*AdminOrder `json:"orders"`
}

type UpdateOrderRequest struct {
	Token          string `json:"token"`
	OrderID        int64  `json:"order_id"`
	Quantity       int32  `json:"quantity"`
	TotalAmount    string `json:"total_amount"`
	Status         string `json:"status"`
	ExitMultiplier string `json:"exit_multiplier"`
	ExitTarget     string `json:"exit_target"`
	ReleasedAmount string `json:"released_amount"`
	ReleaseDay     int32  `json:"release_day"`
	CycleDay       int32  `json:"cycle_day"`
}

type UpdateOrderReply struct {
	Order *AdminOrder `json:"order"`
}
