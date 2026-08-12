package conf

const SettingsKeySystemConfig = "system_config"

// SystemConfigSnapshot 可热更新的系统参数（AIX）
type SystemConfigSnapshot struct {
	JwtSecret        string   `json:"jwt_secret"`
	ChallengeTTL     string   `json:"challenge_ttl"`
	AdminAddresses   []string `json:"admin_addresses"`
	DepositAddress   string   `json:"deposit_address"`
	DepositAddresses []string `json:"deposit_addresses"`
	UsdtContract     string   `json:"usdt_contract"`
	UsdtDecimals     int32    `json:"usdt_decimals"`
	RPCURL           string   `json:"rpc_url"`
	MinSubscribe     string   `json:"min_subscribe"`

	// AIX 业务参数
	StaticRate            float64   `json:"static_rate"`             // 日静态利率（%），默认 0.5
	ExitMultiplier        float64   `json:"exit_multiplier"`         // 出局倍数，默认 4
	DirectRate            float64   `json:"direct_rate"`             // 直推比例，默认 0.5
	MgmtThresholds        []float64 `json:"mgmt_thresholds"`         // W1–W10 小区业绩门槛
	MgmtRates             []float64 `json:"mgmt_rates"`              // W1–W10 管理奖比例
	AixPriceInitial       float64   `json:"aix_price_initial"`       // 初始 AIX 价格
	WinPrice              float64   `json:"win_price"`               // WIN 代币价格（USDT/枚）
	MgmtCountsTowardExit  bool      `json:"mgmt_counts_toward_exit"` // 管理奖是否计入出局
	MgmtCountsTowardExitP *bool     `json:"-"`                       // 内部：区分 JSON 缺省与 false

	// 兼容旧 admin 字段（忽略）
	WithdrawFeeRate float64   `json:"withdraw_fee_rate,omitempty"`
	ReleaseMinRate  float64   `json:"release_min_rate,omitempty"`
	ReleaseMaxRate  float64   `json:"release_max_rate,omitempty"`
	MaxReferralGen  int32     `json:"max_referral_gen,omitempty"`
	ReferralRates   []float64 `json:"referral_rates,omitempty"`
	EcoThresholds   []float64 `json:"eco_thresholds,omitempty"`
	EcoRates        []float64 `json:"eco_rates,omitempty"`
}

const (
	DefaultStaticRate     = 0.5
	DefaultExitMultiplier = 4.0
	DefaultDirectRate     = 0.5
	DefaultAixPrice       = 1.0
	DefaultWinPrice       = 1.0
	DefaultMinSubscribe   = "100"
)

// DefaultMgmtThresholds W1→W10 小区业绩门槛（USDT）
func DefaultMgmtThresholds() []float64 {
	return []float64{5000, 20000, 50000, 200000, 500000, 1500000, 4000000, 8000000, 15000000, 30000000}
}

// DefaultMgmtRates W1→W10 管理奖比例（如 0.2 = 20%）
func DefaultMgmtRates() []float64 {
	return []float64{0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0, 1.1}
}

// NormalizeBusinessDefaults 补齐 AIX 业务参数默认值
func NormalizeBusinessDefaults(s *SystemConfigSnapshot) {
	if s == nil {
		return
	}
	if s.StaticRate <= 0 {
		s.StaticRate = DefaultStaticRate
	}
	if s.ExitMultiplier <= 0 {
		s.ExitMultiplier = DefaultExitMultiplier
	}
	if s.DirectRate <= 0 {
		s.DirectRate = DefaultDirectRate
	}
	if s.AixPriceInitial <= 0 {
		s.AixPriceInitial = DefaultAixPrice
	}
	if s.WinPrice <= 0 {
		s.WinPrice = DefaultWinPrice
	}
	if s.MinSubscribe == "" {
		s.MinSubscribe = DefaultMinSubscribe
	}
	if len(s.MgmtThresholds) != 10 {
		s.MgmtThresholds = DefaultMgmtThresholds()
	}
	if len(s.MgmtRates) != 10 {
		s.MgmtRates = DefaultMgmtRates()
	}
	// 缺省时默认计入出局
	if s.MgmtCountsTowardExitP == nil && !s.MgmtCountsTowardExit {
		// JSON 反序列化后若显式 false 会保留；首次 Normalize 时若两者皆零则置 true
		// 使用：若从未设置过业务开关，默认 true
		s.MgmtCountsTowardExit = true
	}
}
