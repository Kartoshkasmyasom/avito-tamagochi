package pet

import "time"

type PetState struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Level             int        `json:"level"`
	TotalXP           int        `json:"totalXp"`
	NextLevelXP       int        `json:"nextLevelXp"`
	Stage             string     `json:"stage"`
	BatteryLevel      int        `json:"batteryLevel"`
	Status            string     `json:"status"`
	IsActionAvailable bool       `json:"isActionAvailable"`
	CooldownEndsAt    *time.Time `json:"cooldownEndsAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}
type PetActionRequest struct {
	ActionType string `json:"actionType" binding:"required,oneof=charge"`
}
type PetActionResult struct {
	XPAwarded    int       `json:"xpAwarded"`
	BatteryAdded int       `json:"batteryAdded"`
	LevelUp      bool      `json:"levelUp"`
	StageChanged bool      `json:"stageChanged"`
	Pet          *PetState `json:"pet"`
}
