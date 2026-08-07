package game

import (
	"testing"
	"time"
)

func TestLevelForXP(t *testing.T) {
	tests := []struct {
		xp        int
		level     int
		nextLevel int
	}{
		{xp: 0, level: 1, nextLevel: 100},
		{xp: 100, level: 2, nextLevel: 250},
		{xp: 249, level: 2, nextLevel: 250},
		{xp: 450, level: 4, nextLevel: 700},
		{xp: 1350, level: 7, nextLevel: 1700},
	}
	for _, tt := range tests {
		level, next := LevelForXP(tt.xp)
		if level != tt.level || next != tt.nextLevel {
			t.Errorf("LevelForXP(%d) = (%d, %d), want (%d, %d)", tt.xp, level, next, tt.level, tt.nextLevel)
		}
	}
}

func TestStageForLevel(t *testing.T) {
	for _, tt := range []struct {
		level int
		stage string
	}{
		{1, "egg"}, {2, "child"}, {4, "teen"}, {7, "adult"},
	} {
		if got := StageForLevel(tt.level); got != tt.stage {
			t.Errorf("StageForLevel(%d) = %q, want %q", tt.level, got, tt.stage)
		}
	}
}

func TestStatusForBattery(t *testing.T) {
	for _, tt := range []struct {
		battery int
		status  string
	}{
		{100, "happy"}, {60, "normal"}, {25, "tired"}, {0, "exhausted"},
	} {
		if got := StatusForBattery(tt.battery); got != tt.status {
			t.Errorf("StatusForBattery(%d) = %q, want %q", tt.battery, got, tt.status)
		}
	}
}

func TestMoscowMidnight(t *testing.T) {
	now := time.Date(2026, 8, 7, 23, 30, 0, 0, time.UTC)
	if got := MoscowDate(now); got != "2026-08-08" {
		t.Fatalf("MoscowDate() = %q, want 2026-08-08", got)
	}
	if got := NextMoscowMidnight(now); got.Format(time.RFC3339) != "2026-08-09T00:00:00+03:00" {
		t.Fatalf("NextMoscowMidnight() = %s", got.Format(time.RFC3339))
	}
}
