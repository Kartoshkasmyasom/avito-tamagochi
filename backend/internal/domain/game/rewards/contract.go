package rewards

import "time"

type Reward struct {
	ID            string     `json:"id"`
	Type          string     `json:"type"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	RequiredLevel int        `json:"requiredLevel"`
	CurrentXP     int        `json:"currentXp"`
	RequiredXP    int        `json:"requiredXp"`
	Status        string     `json:"status"`
	ClaimedAt     *time.Time `json:"claimedAt"`
}
type RewardsResponse struct {
	Items []*Reward `json:"items"`
}
