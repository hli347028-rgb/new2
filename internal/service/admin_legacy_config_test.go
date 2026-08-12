package service

import (
	"testing"

	"backend/internal/conf"
)

func TestLegacyMgmtThresholdConfigValue(t *testing.T) {
	snapshot := &conf.SystemConfigSnapshot{
		MgmtThresholds: conf.DefaultMgmtThresholds(),
		MgmtRates:      conf.DefaultMgmtRates(),
	}

	if got := legacyConfigValue(snapshot, nil, 21); got != "5000" {
		t.Fatalf("W1 threshold value=%s, want 5000", got)
	}
	if got := legacyConfigValue(snapshot, nil, 30); got != "30000000" {
		t.Fatalf("W10 threshold value=%s, want 30000000", got)
	}
}

func TestApplyLegacyMgmtThresholdConfig(t *testing.T) {
	snapshot := &conf.SystemConfigSnapshot{
		MgmtThresholds: conf.DefaultMgmtThresholds(),
		MgmtRates:      conf.DefaultMgmtRates(),
	}

	if err := applyLegacyConfigUpdate(snapshot, nil, 21, "6000"); err != nil {
		t.Fatalf("update W1 threshold: %v", err)
	}
	if snapshot.MgmtThresholds[0] != 6000 {
		t.Fatalf("W1 threshold=%v, want 6000", snapshot.MgmtThresholds[0])
	}
}

func TestApplyLegacyMgmtThresholdRequiresAscendingValues(t *testing.T) {
	tests := []struct {
		name  string
		id    int
		value string
	}{
		{name: "not positive", id: 21, value: "0"},
		{name: "not numeric", id: 21, value: "abc"},
		{name: "not above previous", id: 22, value: "5000"},
		{name: "not below next", id: 22, value: "50000"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := &conf.SystemConfigSnapshot{
				MgmtThresholds: conf.DefaultMgmtThresholds(),
				MgmtRates:      conf.DefaultMgmtRates(),
			}
			if err := applyLegacyConfigUpdate(snapshot, nil, test.id, test.value); err == nil {
				t.Fatalf("expected validation error for id=%d value=%s", test.id, test.value)
			}
		})
	}
}
