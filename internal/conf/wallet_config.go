package conf

import (
	"strings"

	"github.com/shopspring/decimal"
)

// WalletConfig holds wallet and recharge settings.
type WalletConfig struct {
	DepositAddress   string   `json:"deposit_address" yaml:"deposit_address"`
	DepositAddresses []string `json:"deposit_addresses" yaml:"deposit_addresses"`
	UsdtContract     string   `json:"usdt_contract" yaml:"usdt_contract"`
	UsdtDecimals     int32    `json:"usdt_decimals" yaml:"usdt_decimals"`
	RPCURL           string   `json:"rpc_url" yaml:"rpc_url"`
	MinSubscribe     string   `json:"min_subscribe" yaml:"min_subscribe"`
	MinWithdraw      string   `json:"min_withdraw" yaml:"min_withdraw"`
	WithdrawFeeRate  float64  `json:"withdraw_fee_rate" yaml:"withdraw_fee_rate"`
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
	if w == nil {
		return ""
	}
	return w.RPCURL
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
