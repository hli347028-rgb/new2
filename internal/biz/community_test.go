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
