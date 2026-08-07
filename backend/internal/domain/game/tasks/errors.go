package tasks

import "errors"

var (
	ErrPetNotFound   = errors.New("pet not found")
	ErrEventConflict = errors.New("event conflict")
)
