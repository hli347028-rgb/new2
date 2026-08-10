package release

import (
	"math"
	"testing"

	"github.com/shopspring/decimal"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.001
}

func TestNaturalUpWithoutWithdraw(t *testing.T) {
	index, up := 0, true
	want := []float64{0.6, 0.65, 0.7, 0.75, 0.8, 0.85}
	for i, w := range want {
		got := SettlementRate(index, false)
		if !almostEqual(got, w) {
			t.Fatalf("day %d: got %.2f want %.2f", i+1, got, w)
		}
		index, up = NextRateState(index, up, false)
	}
	if index != 6 || !up {
		t.Fatalf("after day6 advance index=%d up=%v want 6 true", index, up)
	}
}

func TestFullNaturalCycle33Days(t *testing.T) {
	index, up := 0, true
	// 上升 17 天：0.6 → 1.4
	for day := 1; day <= 17; day++ {
		got := SettlementRate(index, false)
		want := RateFromIndex(day - 1)
		if !almostEqual(got, want) {
			t.Fatalf("up day %d: got %.2f want %.2f", day, got, want)
		}
		index, up = NextRateState(index, up, false)
	}
	if !almostEqual(RateFromIndex(MaxIndex), MaxRate) {
		t.Fatalf("max rate want 1.4")
	}
	// 第 17 天结算后应进入回落：index=15, goingUp=false
	if index != MaxIndex-1 || up {
		t.Fatalf("after peak: index=%d up=%v want 15 false", index, up)
	}
	// 下降 16 天：1.35 → 0.6（day18 index15 … day33 index0）
	for day := 18; day <= 33; day++ {
		got := SettlementRate(index, false)
		want := RateFromIndex(33 - day)
		if !almostEqual(got, want) {
			t.Fatalf("down day %d: got %.2f want %.2f (index=%d)", day, got, want, index)
		}
		index, up = NextRateState(index, up, false)
	}
	// 回落后从 0.6 再上升
	if index != 1 || !up {
		t.Fatalf("after cycle: index=%d up=%v want 1 true", index, up)
	}
}

func TestWithdrawDayUsesMaxRate(t *testing.T) {
	index, up := 5, true // 当前约 0.85%
	got := SettlementRate(index, true)
	if !almostEqual(got, MaxRate) {
		t.Fatalf("withdraw day got %.2f want 1.4", got)
	}
	index, up = NextRateState(index, up, true)
	if index != MaxIndex-1 || up {
		t.Fatalf("after withdraw: index=%d up=%v want 15 false", index, up)
	}
	// 次日开始回落
	got = SettlementRate(index, false)
	if !almostEqual(got, 1.35) {
		t.Fatalf("next day got %.2f want 1.35", got)
	}
}

func TestDailyAmount_ReleaseCoefTimesPrincipal(t *testing.T) {
	// 每日静态释放 = 释放系数 × 认购金额；0.6% → 系数 0.006
	principal := decimal.NewFromInt(1000)
	got := DailyAmount(principal, 0.6)
	want := decimal.NewFromInt(6) // 1000 * 0.006
	if !got.Equal(want) {
		t.Fatalf("0.6%% of 1000: got %s want %s", got.String(), want.String())
	}
	got = DailyAmount(principal, 1.4)
	want = decimal.RequireFromString("14") // 1000 * 0.014
	if !got.Equal(want) {
		t.Fatalf("1.4%% of 1000: got %s want %s", got.String(), want.String())
	}
	got = DailyAmount(decimal.NewFromInt(100), 0.6)
	want = decimal.RequireFromString("0.6")
	if !got.Equal(want) {
		t.Fatalf("0.6%% of 100: got %s want %s", got.String(), want.String())
	}
}

func TestPeakThenDown(t *testing.T) {
	index, up := MaxIndex, true
	got := SettlementRate(index, false)
	if !almostEqual(got, 1.4) {
		t.Fatalf("peak got %.2f", got)
	}
	index, up = NextRateState(index, up, false)
	if !almostEqual(RateFromIndex(index), 1.35) {
		t.Fatalf("after peak got %.2f want 1.35", RateFromIndex(index))
	}
	if up {
		t.Fatal("should go down after peak")
	}
}
