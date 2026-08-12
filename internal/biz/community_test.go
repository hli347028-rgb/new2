package biz

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestCalcAreaPerformance(t *testing.T) {
	tests := []struct {
		name     string
		branches []decimal.Decimal
		large    string
		small    string
		team     string
	}{
		{name: "no referrals", large: "0", small: "0", team: "0"},
		{name: "one branch", branches: []decimal.Decimal{decimal.NewFromInt(2500)}, large: "2500", small: "0", team: "2500"},
		{name: "multiple branches", branches: []decimal.Decimal{
			decimal.NewFromInt(2500), decimal.NewFromInt(1000), decimal.NewFromInt(500),
		}, large: "2500", small: "1500", team: "4000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			large, small, team := CalcAreaPerformance(tt.branches)
			if large.String() != tt.large || small.String() != tt.small || team.String() != tt.team {
				t.Fatalf("got large=%s small=%s team=%s", large, small, team)
			}
		})
	}
}

func TestIsLinealRelation(t *testing.T) {
	// 1 -> 2 -> 4, 1 -> 3 -> 5, and 6 is unrelated.
	parents := map[int64]int64{2: 1, 3: 1, 4: 2, 5: 3}
	tests := []struct {
		name string
		a, b int64
		want bool
	}{
		{name: "direct downline", a: 1, b: 2, want: true},
		{name: "direct upline", a: 2, b: 1, want: true},
		{name: "deep lineal", a: 1, b: 4, want: true},
		{name: "siblings", a: 2, b: 3, want: false},
		{name: "different branches", a: 4, b: 5, want: false},
		{name: "unrelated", a: 4, b: 6, want: false},
		{name: "self", a: 2, b: 2, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLinealRelation(tt.a, tt.b, parents); got != tt.want {
				t.Fatalf("IsLinealRelation(%d, %d)=%v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestManagementBaseDirectGetsFullRate(t *testing.T) {
	children := map[int64][]int64{1: {2}}
	parent := map[int64]int64{2: 1}
	levels := map[int64]int{1: 3, 2: 0}
	static := map[int64]decimal.Decimal{2: decimal.NewFromInt(100)}

	lines := ListManagementBaseLines(1, children, parent, levels, static)
	if len(lines) != 1 || lines[0].Amount.StringFixed(2) != "40.00" {
		t.Fatalf("got %+v, want W3 full rate reward 40.00", lines)
	}
}

func TestManagementBaseUsesGradeGap(t *testing.T) {
	// W5 -> W3 -> source. W3 receives 40%; W5 receives 60%-40%=20%.
	children := map[int64][]int64{1: {2}, 2: {3}}
	parent := map[int64]int64{2: 1, 3: 2}
	levels := map[int64]int{1: 5, 2: 3, 3: 0}
	static := map[int64]decimal.Decimal{3: decimal.NewFromInt(100)}

	lower := ListManagementBaseLines(2, children, parent, levels, static)
	upper := ListManagementBaseLines(1, children, parent, levels, static)
	if len(lower) != 1 || lower[0].Amount.StringFixed(2) != "40.00" {
		t.Fatalf("lower got %+v, want 40.00", lower)
	}
	if len(upper) != 1 || upper[0].Amount.StringFixed(2) != "20.00" {
		t.Fatalf("upper got %+v, want 20.00", upper)
	}
}

func TestManagementBaseBlockedByEqualOrHigherIntermediate(t *testing.T) {
	children := map[int64][]int64{1: {2}, 2: {3}}
	parent := map[int64]int64{2: 1, 3: 2}
	levels := map[int64]int{1: 3, 2: 5, 3: 0}
	static := map[int64]decimal.Decimal{3: decimal.NewFromInt(100)}

	if lines := ListManagementBaseLines(1, children, parent, levels, static); len(lines) != 0 {
		t.Fatalf("upper reward should be blocked, got %+v", lines)
	}
}

func TestManagementOrderRewardUsesLevelDifferenceAndNoPeerReward(t *testing.T) {
	// source -> W3 -> W3 -> W5: 100 order gives W3 40, equal W3 zero,
	// and W5 receives only the 20% difference.
	parent := map[int64]int64{4: 3, 3: 2, 2: 1}
	levels := map[int64]int{1: 5, 2: 3, 3: 3, 4: 0}
	lines := ListManagementOrderLines(4, decimal.NewFromInt(100), parent, levels)
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2: %+v", len(lines), lines)
	}
	if lines[0].UserID != 3 || lines[0].Rate.StringFixed(2) != "0.40" || lines[0].Amount.StringFixed(2) != "40.00" {
		t.Fatalf("nearest W3 line=%+v", lines[0])
	}
	if lines[1].UserID != 1 || lines[1].Rate.StringFixed(2) != "0.20" || lines[1].Amount.StringFixed(2) != "20.00" {
		t.Fatalf("upper W5 line=%+v", lines[1])
	}
}

func TestManagementRewardPendingContinuesAfterNewSubscription(t *testing.T) {
	pending := decimal.NewFromInt(300)
	first := CalcManagementRelease(decimal.NewFromInt(100), decimal.Zero, pending)
	if first.StringFixed(2) != "100.00" {
		t.Fatalf("first release=%s, want 100", first)
	}
	second := CalcManagementRelease(decimal.NewFromInt(250), first, pending.Sub(first))
	if second.StringFixed(2) != "150.00" {
		t.Fatalf("second release=%s, want 150", second)
	}
	remaining := pending.Sub(first).Sub(second)
	if remaining.StringFixed(2) != "50.00" {
		t.Fatalf("remaining=%s, want 50", remaining)
	}
}
