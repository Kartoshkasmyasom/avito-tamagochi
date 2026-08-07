package pet

import (
	"errors"
	"time"
)

var ErrPetNotFound = errors.New("pet not found")
var ErrCooldown = errors.New("cooldown active")

type CooldownError struct {
	RetryAfter int
}

func (e *CooldownError) Error() string { return ErrCooldown.Error() }
func (e *CooldownError) Unwrap() error { return ErrCooldown }

func NewCooldownError(now, until time.Time) error {
	seconds := int((until.Sub(now) + time.Second - 1) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return &CooldownError{RetryAfter: seconds}
}
