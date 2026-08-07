package pet

import "errors"

var ErrPetNotFound = errors.New("pet not found")
var ErrCooldown = errors.New("cooldown active")
