package game

import "time"

const (
	ActivityListingViewed    = "listingViewed"
	ActivityFavoriteAdded    = "favoriteAdded"
	ActivityListingPublished = "listingPublished"
)

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

type DemoActivityRequest struct {
	EventID      string `json:"eventId" binding:"required,uuid"`
	ActivityType string `json:"activityType" binding:"required,oneof=listingViewed favoriteAdded listingPublished"`
}

type Task struct {
	ID           string     `json:"id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	ActivityType string     `json:"activityType"`
	Progress     int        `json:"progress"`
	Target       int        `json:"target"`
	XPReward     int        `json:"xpReward"`
	Status       string     `json:"status"`
	CompletedAt  *time.Time `json:"completedAt"`
}

type TasksResponse struct {
	Items    []*Task   `json:"items"`
	ResetsAt time.Time `json:"resetsAt"`
}

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

type LeaderboardEntry struct {
	Rank        int    `json:"rank"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Level       int    `json:"level"`
	TotalXP     int    `json:"totalXp"`
}

type Leaderboard struct {
	Entries     []*LeaderboardEntry `json:"entries"`
	CurrentUser *LeaderboardEntry   `json:"currentUser"`
}

type DailySummary struct {
	Date                 string   `json:"date"`
	XPEarned             int      `json:"xpEarned"`
	ChargesPerformed     int      `json:"chargesPerformed"`
	CompletedTaskTitles  []string `json:"completedTaskTitles"`
	UnlockedRewardTitles []string `json:"unlockedRewardTitles"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
