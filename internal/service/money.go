package service

import "github.com/shopspring/decimal"

// normalizeExitMultiplier returns a positive exit multiplier (default 1).
func normalizeExitMultiplier(mul decimal.Decimal) decimal.Decimal {
	if !mul.IsPositive() {
		return decimal.NewFromInt(1)
	}
	return mul
}

// displayMoney returns amount × exitMultiplier (same display formula as admin money column).
func displayMoney(amount, exitMultiplier decimal.Decimal) string {
	return amount.Mul(normalizeExitMultiplier(exitMultiplier)).String()
}

func parseExitMultiplier(raw string) decimal.Decimal {
	mul, err := decimal.NewFromString(raw)
	if err != nil {
		return decimal.NewFromInt(1)
	}
	return normalizeExitMultiplier(mul)
}
