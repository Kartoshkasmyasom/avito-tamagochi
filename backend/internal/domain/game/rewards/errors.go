package rewards

import "errors"

var (
	ErrPetNotFound    = errors.New("pet not found")
	ErrRewardNotFound = errors.New("reward not found")
	ErrRewardLocked   = errors.New("reward locked")
)
