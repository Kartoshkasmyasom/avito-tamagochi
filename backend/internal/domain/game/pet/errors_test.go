package pet

import (
	"errors"
	"testing"
	"time"
)

func TestNewCooldownErrorRoundsUpRetryAfter(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	err := NewCooldownError(now, now.Add(1500*time.Millisecond))

	var cooldown *CooldownError
	if !errors.As(err, &cooldown) {
		t.Fatal("expected CooldownError")
	}
	if cooldown.RetryAfter != 2 {
		t.Fatalf("RetryAfter = %d, want 2", cooldown.RetryAfter)
	}
	if !errors.Is(err, ErrCooldown) {
		t.Fatal("cooldown error must match ErrCooldown")
	}
}
