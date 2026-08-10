package release

import "github.com/shopspring/decimal"

var (
	Step     = 0.05
	MinRate  = 0.6
	MaxRate  = 1.4
	MinIndex = 0
	MaxIndex = 16 // 0.6% + 16*0.05 = 1.4%
	// 自然周期：上升 (MaxIndex+1) 天，下降 MaxIndex 天
)

// Configure 热更新静态释放区间（单位：%）。step 仍固定 0.05。
func Configure(minRate, maxRate float64) {
	if minRate <= 0 || maxRate <= 0 || maxRate < minRate {
		return
	}
	MinRate = minRate
	MaxRate = maxRate
	steps := int((maxRate - minRate) / Step)
	if steps < 0 {
		steps = 0
	}
	MaxIndex = steps
	if MaxIndex < 1 {
		MaxIndex = 1
	}
}

// RateFromIndex 根据比例档位计算日释放比例（单位：%），index 0=MinRate。
func RateFromIndex(index int) float64 {
	if index < MinIndex {
		index = MinIndex
	}
	if index > MaxIndex {
		index = MaxIndex
	}
	return MinRate + float64(index)*Step
}

// SettlementRate 计算当日释放比例（单位：%）。
// withdrawReset=true：提取当天固定按最高 MaxRate 释放。
func SettlementRate(index int, withdrawReset bool) float64 {
	if withdrawReset {
		return MaxRate
	}
	return RateFromIndex(index)
}

// Coefficient 释放系数 = 日释放比例(%) / 100。
func Coefficient(ratePercent float64) decimal.Decimal {
	if ratePercent <= 0 {
		return decimal.Zero
	}
	return decimal.NewFromFloat(ratePercent).Div(decimal.NewFromInt(100))
}

// DailyAmount 每日静态释放金额 = 释放系数 × 认购金额。
func DailyAmount(principal decimal.Decimal, ratePercent float64) decimal.Decimal {
	if principal.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}
	return principal.Mul(Coefficient(ratePercent)).Round(8)
}

// NextRateState 结算后更新档位与方向。
func NextRateState(index int, goingUp, withdrawReset bool) (int, bool) {
	if withdrawReset {
		if MaxIndex <= 0 {
			return 0, true
		}
		return MaxIndex - 1, false
	}
	if goingUp {
		if index >= MaxIndex {
			if MaxIndex <= 0 {
				return 0, true
			}
			return MaxIndex - 1, false
		}
		return index + 1, true
	}
	if index <= MinIndex {
		return 1, true
	}
	return index - 1, false
}

// LegacyCycleDayToIndex 将旧版 cycle_day(1~33) 迁移为档位
func LegacyCycleDayToIndex(cycleDay int) (int, bool) {
	if cycleDay <= 0 {
		return 0, true
	}
	day := ((cycleDay - 1) % 33) + 1
	if day <= 17 {
		return day - 1, true
	}
	return 16 - (day - 17), false
}
