package leaderboard

type Entry struct {
	Rank        int    `json:"rank"`
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
	Level       int    `json:"level"`
	TotalXP     int    `json:"totalXp"`
}
type Response struct {
	Entries     []*Entry `json:"entries"`
	CurrentUser *Entry   `json:"currentUser"`
}
