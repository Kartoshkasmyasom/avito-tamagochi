package progression

import (
	"testing"
	"time"
)

func TestLevelForXP(t *testing.T) {
	for _, tt := range []struct{ xp, level, next int }{{0, 1, 100}, {100, 2, 250}, {249, 2, 250}, {450, 4, 700}, {1350, 7, 1700}} {
		level, next := LevelForXP(tt.xp)
		if level != tt.level || next != tt.next {
			t.Errorf("LevelForXP(%d) = (%d, %d)", tt.xp, level, next)
		}
	}
}
func TestStageForLevel(t *testing.T) {
	for _, tt := range []struct {
		level int
		stage string
	}{{1, "egg"}, {2, "child"}, {4, "teen"}, {7, "adult"}} {
		if got := StageForLevel(tt.level); got != tt.stage {
			t.Errorf("StageForLevel(%d) = %q", tt.level, got)
		}
	}
}
func TestStatusForBattery(t *testing.T) {
	for _, tt := range []struct {
		battery int
		status  string
	}{{100, "happy"}, {60, "normal"}, {25, "tired"}, {0, "exhausted"}} {
		if got := StatusForBattery(tt.battery); got != tt.status {
			t.Errorf("StatusForBattery(%d) = %q", tt.battery, got)
		}
	}
}
func TestMoscowMidnight(t *testing.T) {
	now := time.Date(2026, 8, 7, 23, 30, 0, 0, time.UTC)
	if got := MoscowDate(now); got != "2026-08-08" {
		t.Fatalf("MoscowDate() = %q", got)
	}
	if got := NextMoscowMidnight(now).Format(time.RFC3339); got != "2026-08-09T00:00:00+03:00" {
		t.Fatalf("NextMoscowMidnight() = %s", got)
	}
}
