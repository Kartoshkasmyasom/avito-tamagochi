package game

import (
	"context"
	"database/sql"
)

type LeaderboardService interface {
	Get(context.Context, string) (*Leaderboard, error)
}
type leaderboardService struct{ db *sql.DB }

func NewLeaderboardService(db *sql.DB) LeaderboardService { return &leaderboardService{db: db} }

func (s *leaderboardService) Get(ctx context.Context, userID string) (*Leaderboard, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,display_name,level,total_xp FROM (SELECT u.id,u.display_name,p.level,p.total_xp,ROW_NUMBER() OVER (ORDER BY p.total_xp DESC,u.id) rank FROM users u JOIN pets p ON p.owner_id=u.id) x ORDER BY rank LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := &Leaderboard{Entries: []*LeaderboardEntry{}}
	rank := 1
	for rows.Next() {
		e := &LeaderboardEntry{Rank: rank}
		if err := rows.Scan(&e.UserID, &e.DisplayName, &e.Level, &e.TotalXP); err != nil {
			return nil, err
		}
		result.Entries = append(result.Entries, e)
		rank++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	e := &LeaderboardEntry{}
	if err = s.db.QueryRowContext(ctx, `SELECT rank,id,display_name,level,total_xp FROM (SELECT u.id,u.display_name,p.level,p.total_xp,ROW_NUMBER() OVER (ORDER BY p.total_xp DESC,u.id) rank FROM users u JOIN pets p ON p.owner_id=u.id) x WHERE id=$1`, userID).Scan(&e.Rank, &e.UserID, &e.DisplayName, &e.Level, &e.TotalXP); err != nil {
		return nil, err
	}
	result.CurrentUser = e
	return result, nil
}
