package summary

type DailySummary struct {
	Date                 string   `json:"date"`
	XPEarned             int      `json:"xpEarned"`
	ChargesPerformed     int      `json:"chargesPerformed"`
	CompletedTaskTitles  []string `json:"completedTaskTitles"`
	UnlockedRewardTitles []string `json:"unlockedRewardTitles"`
}
