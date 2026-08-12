package conf

import (
	"strings"

	"github.com/shopspring/decimal"
)

const (
	DefaultDepositContract = "0xe11c2F7902CB03cAA38F80B27DC20702af14D5c7"
	DefaultRPCURL          = "https://rpc1.eoeo.info"
)

// WalletConfig holds wallet and recharge settings.
type WalletConfig struct {
	DepositAddress                   string   `json:"deposit_address" yaml:"deposit_address"`
	DepositAddresses                 []string `json:"deposit_addresses" yaml:"deposit_addresses"`
	DepositContract                  string   `json:"deposit_contract" yaml:"deposit_contract"`
	UsdtContract                     string   `json:"usdt_contract" yaml:"usdt_contract"`
	UsdtDecimals                     int32    `json:"usdt_decimals" yaml:"usdt_decimals"`
	RPCURL                           string   `json:"rpc_url" yaml:"rpc_url"`
	RechargeMonitorEnabled           bool     `json:"recharge_monitor_enabled" yaml:"recharge_monitor_enabled"`
	RechargeScanIntervalSeconds      int64    `json:"recharge_scan_interval_seconds" yaml:"recharge_scan_interval_seconds"`
	RechargeScanQueriesPerCycle      int32    `json:"recharge_scan_queries_per_cycle" yaml:"recharge_scan_queries_per_cycle"`
	RechargeScanQueryIntervalSeconds int64    `json:"recharge_scan_query_interval_seconds" yaml:"recharge_scan_query_interval_seconds"`
	RechargeConfirmations            uint64   `json:"recharge_confirmations" yaml:"recharge_confirmations"`
	RechargeScanStartBlock           uint64   `json:"recharge_scan_start_block" yaml:"recharge_scan_start_block"`
	RechargeScanLookbackBlocks       uint64   `json:"recharge_scan_lookback_blocks" yaml:"recharge_scan_lookback_blocks"`
	RechargeScanBatchBlocks          uint64   `json:"recharge_scan_batch_blocks" yaml:"recharge_scan_batch_blocks"`
	MinSubscribe                     string   `json:"min_subscribe" yaml:"min_subscribe"`
	MinWithdraw                      string   `json:"min_withdraw" yaml:"min_withdraw"`
	WithdrawFeeRate                  float64  `json:"withdraw_fee_rate" yaml:"withdraw_fee_rate"`
}

func (w *WalletConfig) GetDepositContract() string {
	if w == nil || strings.TrimSpace(w.DepositContract) == "" {
		return DefaultDepositContract
	}
	return strings.TrimSpace(w.DepositContract)
}

func (w *WalletConfig) GetRechargeScanIntervalSeconds() int64 {
	if w == nil || w.RechargeScanIntervalSeconds <= 0 {
		return 60
	}
	return w.RechargeScanIntervalSeconds
}

func (w *WalletConfig) GetRechargeScanQueriesPerCycle() int32 {
	if w == nil || w.RechargeScanQueriesPerCycle <= 0 {
		return 10
	}
	return w.RechargeScanQueriesPerCycle
}

func (w *WalletConfig) GetRechargeScanQueryIntervalSeconds() int64 {
	if w == nil || w.RechargeScanQueryIntervalSeconds <= 0 {
		return 5
	}
	return w.RechargeScanQueryIntervalSeconds
}

func (w *WalletConfig) GetRechargeConfirmations() uint64 {
	if w == nil || w.RechargeConfirmations == 0 {
		return 3
	}
	return w.RechargeConfirmations
}

func (w *WalletConfig) GetRechargeScanLookbackBlocks() uint64 {
	if w == nil || w.RechargeScanLookbackBlocks == 0 {
		return 5000
	}
	return w.RechargeScanLookbackBlocks
}

func (w *WalletConfig) GetRechargeScanBatchBlocks() uint64 {
	if w == nil || w.RechargeScanBatchBlocks == 0 {
		return 1000
	}
	return w.RechargeScanBatchBlocks
}

// GetDepositAddresses 返回全部收款地址（去重，顺序：deposit_addresses 再 deposit_address）。
// deposit_address 支持逗号/空白分隔多个地址。
func (w *WalletConfig) GetDepositAddresses() []string {
	if w == nil {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0, len(w.DepositAddresses)+1)
	add := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return
		}
		parts := strings.FieldsFunc(raw, func(r rune) bool {
			return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
		})
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			key := strings.ToLower(p)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, p)
		}
	}
	for _, a := range w.DepositAddresses {
		add(a)
	}
	add(w.DepositAddress)
	return out
}

func (w *WalletConfig) GetDepositAddress() string {
	addrs := w.GetDepositAddresses()
	if len(addrs) == 0 {
		return ""
	}
	return addrs[0]
}

// SetDepositAddresses 写入收款地址列表，并同步主展示地址。
func (w *WalletConfig) SetDepositAddresses(addrs []string) {
	if w == nil {
		return
	}
	w.DepositAddresses = nil
	w.DepositAddress = ""
	normalized := (&WalletConfig{DepositAddresses: addrs}).GetDepositAddresses()
	w.DepositAddresses = normalized
	if len(normalized) > 0 {
		w.DepositAddress = normalized[0]
	}
}

func (w *WalletConfig) GetUsdtContract() string {
	if w == nil {
		return ""
	}
	return w.UsdtContract
}

func (w *WalletConfig) GetUsdtDecimals() int32 {
	if w == nil || w.UsdtDecimals <= 0 {
		return 6
	}
	return w.UsdtDecimals
}

func (w *WalletConfig) GetRPCURL() string {
	if w == nil || strings.TrimSpace(w.RPCURL) == "" {
		return DefaultRPCURL
	}
	return strings.TrimSpace(w.RPCURL)
}

func (w *WalletConfig) GetMinSubscribe() string {
	if w == nil || w.MinSubscribe == "" {
		return "100"
	}
	return w.MinSubscribe
}

func (w *WalletConfig) GetMinWithdraw() string {
	if w == nil || w.MinWithdraw == "" {
		return "10"
	}
	return w.MinWithdraw
}

func (w *WalletConfig) GetWithdrawFeeRate() decimal.Decimal {
	if w == nil || w.WithdrawFeeRate <= 0 {
		return decimal.NewFromFloat(0.06)
	}
	return decimal.NewFromFloat(w.WithdrawFeeRate)
}
