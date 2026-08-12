package biz

import (
	"github.com/shopspring/decimal"
)

// ManagementRewardSplit is retained for callers that aggregate differential
// management rewards. Peer is always zero because flat-level rewards were
// removed.
type ManagementRewardSplit struct {
	Base decimal.Decimal
	Peer decimal.Decimal
}

// ManagementBaseLine records one daily-static source contributing a
// differential management reward to a claimant.
type ManagementBaseLine struct {
	SourceID int64
	Amount   decimal.Decimal
	Static   decimal.Decimal
	GapRate  decimal.Decimal
}

type ManagementOrderLine struct {
	UserID int64
	Rate   decimal.Decimal
	Amount decimal.Decimal
}

// ListManagementOrderLines calculates one-time differential entitlements for
// a downline subscription. Equal/lower uplines receive zero; only a new
// highest rate receives the positive difference from the highest lower rate.
func ListManagementOrderLines(sourceUserID int64, principal decimal.Decimal, parent map[int64]int64, levels map[int64]int) []ManagementOrderLine {
	if !principal.IsPositive() {
		return nil
	}
	currentID, ok := parent[sourceUserID]
	if !ok {
		return nil
	}
	highestLowerRate := decimal.Zero
	seen := map[int64]bool{sourceUserID: true}
	lines := make([]ManagementOrderLine, 0)
	for currentID > 0 && !seen[currentID] {
		seen[currentID] = true
		rate := mgmtRateForRank(levels[currentID])
		gap := rate.Sub(highestLowerRate)
		if gap.IsPositive() {
			amount := principal.Mul(gap).Round(8)
			if amount.IsPositive() {
				lines = append(lines, ManagementOrderLine{UserID: currentID, Rate: gap, Amount: amount})
			}
		}
		if rate.GreaterThan(highestLowerRate) {
			highestLowerRate = rate
		}
		nextID, exists := parent[currentID]
		if !exists {
			break
		}
		currentID = nextID
	}
	return lines
}

func CalcManagementRelease(principalCapacity, alreadyReleased, pending decimal.Decimal) decimal.Decimal {
	capacity := principalCapacity.Sub(alreadyReleased)
	if !capacity.IsPositive() || !pending.IsPositive() {
		return decimal.Zero
	}
	if pending.LessThan(capacity) {
		return pending
	}
	return capacity
}
func mgmtRateForRank(rank int) decimal.Decimal {
	if rank <= 0 {
		return decimal.Zero
	}
	return decimal.NewFromFloat(MgmtRateForLevel(int32(rank)))
}

// hasMgmtBreak reports whether a source is blocked by an intermediate user
// whose level is equal to or higher than the claimant's level.
func hasMgmtBreak(sourceID, claimantID int64, claimantLevel int, parent map[int64]int64, levels map[int64]int) bool {
	if claimantLevel <= 0 {
		return false
	}
	current, ok := parent[sourceID]
	if !ok {
		return false
	}
	seen := map[int64]bool{sourceID: true}
	for current != claimantID {
		if seen[current] {
			return true
		}
		seen[current] = true
		if levels[current] >= claimantLevel {
			return true
		}
		next, exists := parent[current]
		if !exists {
			return false
		}
		current = next
	}
	return false
}

func collectMgmtSubtreeIDs(root int64, children map[int64][]int64) []int64 {
	out := []int64{root}
	stack := append([]int64(nil), children[root]...)
	seen := map[int64]bool{root: true}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
		stack = append(stack, children[n]...)
	}
	return out
}

func maxMgmtRateBetween(sourceID, claimantID int64, parent map[int64]int64, levels map[int64]int) decimal.Decimal {
	current, ok := parent[sourceID]
	if !ok {
		return decimal.Zero
	}
	bestLevel := 0
	seen := map[int64]bool{sourceID: true}
	for current != claimantID {
		if seen[current] {
			break
		}
		seen[current] = true
		if levels[current] > bestLevel {
			bestLevel = levels[current]
		}
		next, exists := parent[current]
		if !exists {
			break
		}
		current = next
	}
	return mgmtRateForRank(bestLevel)
}

// ListManagementBaseLines applies the new project's tree-based differential
// rule using this project's W1-W10 levels and rates:
//   - a direct downline's static reward uses the claimant's full rate;
//   - deeper static rewards use claimant rate minus the highest intermediate rate;
//   - an intermediate equal/higher level blocks the rewards below it.
func ListManagementBaseLines(
	claimantID int64,
	children map[int64][]int64,
	parent map[int64]int64,
	levels map[int64]int,
	todayStatic map[int64]decimal.Decimal,
) []ManagementBaseLine {
	claimantLevel := levels[claimantID]
	claimantRate := mgmtRateForRank(claimantLevel)
	if !claimantRate.IsPositive() {
		return nil
	}

	var lines []ManagementBaseLine
	for _, directID := range children[claimantID] {
		for _, sourceID := range collectMgmtSubtreeIDs(directID, children) {
			direct := sourceID == directID
			if !direct && hasMgmtBreak(sourceID, claimantID, claimantLevel, parent, levels) {
				continue
			}
			staticAmount := todayStatic[sourceID]
			if !staticAmount.IsPositive() {
				continue
			}
			gapRate := claimantRate
			if !direct {
				gapRate = claimantRate.Sub(maxMgmtRateBetween(sourceID, claimantID, parent, levels))
			}
			if !gapRate.IsPositive() {
				continue
			}
			amount := staticAmount.Mul(gapRate).Round(8)
			if amount.IsPositive() {
				lines = append(lines, ManagementBaseLine{
					SourceID: sourceID,
					Amount:   amount,
					Static:   staticAmount,
					GapRate:  gapRate,
				})
			}
		}
	}
	return lines
}

// ComputeAllManagementRewards computes differential rewards only.
func ComputeAllManagementRewards(
	userIDs []int64,
	children map[int64][]int64,
	parent map[int64]int64,
	levels map[int64]int,
	todayStatic map[int64]decimal.Decimal,
) map[int64]ManagementRewardSplit {
	out := make(map[int64]ManagementRewardSplit, len(userIDs))
	for _, userID := range userIDs {
		for _, line := range ListManagementBaseLines(userID, children, parent, levels, todayStatic) {
			split := out[userID]
			split.Base = split.Base.Add(line.Amount)
			out[userID] = split
		}
	}
	for userID, split := range out {
		split.Base = split.Base.Round(8)
		split.Peer = split.Peer.Round(8)
		out[userID] = split
	}
	return out
}

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
