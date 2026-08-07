package leaderboard

import (
	"context"
	"database/sql"
)

type Service interface {
	Get(context.Context, string) (*Response, error)
}
type service struct{ db *sql.DB }

func NewService(db *sql.DB) Service { return &service{db: db} }
func (s *service) Get(ctx context.Context, userID string) (*Response, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,display_name,level,total_xp FROM (SELECT u.id,u.display_name,p.level,p.total_xp,ROW_NUMBER() OVER (ORDER BY p.total_xp DESC,u.id) rank FROM users u JOIN pets p ON p.owner_id=u.id) x ORDER BY rank LIMIT 10`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	r := &Response{Entries: []*Entry{}}
	rank := 1
	for rows.Next() {
		e := &Entry{Rank: rank}
		if err := rows.Scan(&e.UserID, &e.DisplayName, &e.Level, &e.TotalXP); err != nil {
			return nil, err
		}
		r.Entries = append(r.Entries, e)
		rank++
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	e := &Entry{}
	if err = s.db.QueryRowContext(ctx, `SELECT rank,id,display_name,level,total_xp FROM (SELECT u.id,u.display_name,p.level,p.total_xp,ROW_NUMBER() OVER (ORDER BY p.total_xp DESC,u.id) rank FROM users u JOIN pets p ON p.owner_id=u.id) x WHERE id=$1`, userID).Scan(&e.Rank, &e.UserID, &e.DisplayName, &e.Level, &e.TotalXP); err != nil {
		return nil, err
	}
	r.CurrentUser = e
	return r, nil
}
