package tasks

import "time"

const (
	ActivityListingViewed    = "listingViewed"
	ActivityFavoriteAdded    = "favoriteAdded"
	ActivityListingPublished = "listingPublished"
)

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
