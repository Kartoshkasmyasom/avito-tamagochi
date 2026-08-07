package game

import (
	"context"
	"database/sql"
	"time"
)

type SummaryService interface {
	Get(context.Context, string) (*DailySummary, error)
}
type summaryService struct {
	db  *sql.DB
	now func() time.Time
}

func NewSummaryService(db *sql.DB) SummaryService { return &summaryService{db: db, now: time.Now} }
func (s *summaryService) Get(ctx context.Context, userID string) (*DailySummary, error) {
	now := s.now()
	start := MoscowMidnight(now).UTC()
	result := &DailySummary{Date: MoscowDate(now), CompletedTaskTitles: []string{}, UnlockedRewardTitles: []string{}}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(xp_awarded),0),COUNT(*) FILTER (WHERE source='charge') FROM xp_events WHERE user_id=$1 AND occurred_at >= $2 AND occurred_at <= $3`, userID, start, now).Scan(&result.XPEarned, &result.ChargesPerformed); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT activity_type FROM daily_tasks WHERE user_id=$1 AND task_date=$2 AND completed_at IS NOT NULL ORDER BY activity_type`, userID, MoscowDate(now))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var activity string
		if err := rows.Scan(&activity); err != nil {
			rows.Close()
			return nil, err
		}
		title, _ := taskText(activity)
		result.CompletedTaskTitles = append(result.CompletedTaskTitles, title)
	}
	rows.Close()
	rows, err = s.db.QueryContext(ctx, `SELECT reward_type FROM user_rewards WHERE user_id=$1 AND unlocked_at >= $2 AND unlocked_at <= $3 ORDER BY required_level`, userID, start, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var typ string
		if err := rows.Scan(&typ); err != nil {
			return nil, err
		}
		title, _ := rewardText(typ)
		result.UnlockedRewardTitles = append(result.UnlockedRewardTitles, title)
	}
	return result, rows.Err()
}
