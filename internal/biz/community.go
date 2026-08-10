package biz

import (
	"github.com/shopspring/decimal"
)

// CalcSubtreeStake 计算以 root 为根的子树业绩（含本人）
func CalcSubtreeStake(root int64, stake map[int64]decimal.Decimal, children map[int64][]int64, memo map[int64]decimal.Decimal) decimal.Decimal {
	if v, ok := memo[root]; ok {
		return v
	}
	total := stake[root]
	for _, c := range children[root] {
		total = total.Add(CalcSubtreeStake(c, stake, children, memo))
	}
	memo[root] = total
	return total
}

// CalcBranchStakes 各直推线业绩
func CalcBranchStakes(root int64, stake map[int64]decimal.Decimal, children map[int64][]int64, memo map[int64]decimal.Decimal) []decimal.Decimal {
	dirs := children[root]
	out := make([]decimal.Decimal, 0, len(dirs))
	for _, c := range dirs {
		out = append(out, CalcSubtreeStake(c, stake, children, memo))
	}
	return out
}

// SumBranchStakes 团队总业绩
func SumBranchStakes(branches []decimal.Decimal) decimal.Decimal {
	total := decimal.Zero
	for _, b := range branches {
		total = total.Add(b)
	}
	return total
}

// CalcCommunityStake 小区业绩 = 总 - 最大区
func CalcCommunityStake(branches []decimal.Decimal) decimal.Decimal {
	if len(branches) == 0 {
		return decimal.Zero
	}
	total := SumBranchStakes(branches)
	max := decimal.Zero
	for _, b := range branches {
		if b.GreaterThan(max) {
			max = b
		}
	}
	return total.Sub(max)
}

// CalcAreaPerformance returns the largest direct branch, all remaining
// branches, and the sum of every direct branch.
func CalcAreaPerformance(branches []decimal.Decimal) (large, small, team decimal.Decimal) {
	team = SumBranchStakes(branches)
	for _, branch := range branches {
		if branch.GreaterThan(large) {
			large = branch
		}
	}
	small = team.Sub(large)
	return large, small, team
}

// IsLinealRelation reports whether one user is an ancestor of the other.
// Siblings and users in different invitation branches are not lineal.
func IsLinealRelation(a, b int64, parentByUser map[int64]int64) bool {
	if a <= 0 || b <= 0 || a == b {
		return false
	}
	isAncestor := func(ancestor, node int64) bool {
		seen := map[int64]bool{node: true}
		current := node
		for {
			parent, ok := parentByUser[current]
			if !ok || seen[parent] {
				return false
			}
			if parent == ancestor {
				return true
			}
			seen[parent] = true
			current = parent
		}
	}
	return isAncestor(a, b) || isAncestor(b, a)
}

// Legacy eco helpers (no-op / defaults for admin config compatibility)
func ApplyEcoLevels(thresholds, rates []float64) {}

func CurrentEcoThresholdsRates() (thresholds, rates []float64) {
	return append([]float64(nil), MgmtThresholds...), append([]float64(nil), MgmtRates...)
}

func GetLevelByStake(stake decimal.Decimal) (string, decimal.Decimal, bool) {
	lv := MgmtLevelByPerf(stake)
	if lv <= 0 {
		return "", decimal.Zero, false
	}
	return "W" + itoa(int(lv)), decimal.NewFromFloat(MgmtRateForLevel(lv)), true
}

func NormalizeEcoLevel(level string) string { return level }
