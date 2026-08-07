package summary

import (
	"context"
	"database/sql"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/progression"
)

type Service interface {
	Get(context.Context, string) (*DailySummary, error)
}
type service struct {
	db  *sql.DB
	now func() time.Time
}

func NewService(db *sql.DB) Service { return &service{db: db, now: time.Now} }
func (s *service) Get(ctx context.Context, userID string) (*DailySummary, error) {
	now := s.now()
	start := progression.MoscowMidnight(now).UTC()
	r := &DailySummary{Date: progression.MoscowDate(now), CompletedTaskTitles: []string{}, UnlockedRewardTitles: []string{}}
	if err := s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(xp_awarded),0),COUNT(*) FILTER (WHERE source='charge') FROM xp_events WHERE user_id=$1 AND occurred_at >= $2 AND occurred_at <= $3`, userID, start, now).Scan(&r.XPEarned, &r.ChargesPerformed); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT activity_type FROM daily_tasks WHERE user_id=$1 AND task_date=$2 AND completed_at IS NOT NULL ORDER BY activity_type`, userID, progression.MoscowDate(now))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			rows.Close()
			return nil, err
		}
		r.CompletedTaskTitles = append(r.CompletedTaskTitles, taskText(a))
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
		r.UnlockedRewardTitles = append(r.UnlockedRewardTitles, rewardText(typ))
	}
	return r, rows.Err()
}
func taskText(a string) string {
	switch a {
	case "listingViewed":
		return "Посмотреть объявление"
	case "favoriteAdded":
		return "Добавить в избранное"
	default:
		return "Опубликовать объявление"
	}
}
func rewardText(t string) string {
	if t == "listingPromotion" {
		return "Продвижение объявления"
	}
	return "Авито Доставка"
}
