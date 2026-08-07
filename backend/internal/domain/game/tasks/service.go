package tasks

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/progression"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Service interface {
	List(context.Context, string) (*TasksResponse, error)
	ProcessActivity(context.Context, string, DemoActivityRequest) error
}
type service struct {
	db  *sql.DB
	now func() time.Time
}

func NewService(db *sql.DB) Service { return &service{db: db, now: time.Now} }
func (s *service) List(ctx context.Context, userID string) (*TasksResponse, error) {
	now := s.now()
	if err := requirePet(ctx, s.db, userID); err != nil {
		return nil, err
	}
	date := progression.MoscowDate(now)
	if err := ensureTasks(ctx, s.db, userID, date); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,activity_type,progress,target,xp_reward,completed_at FROM daily_tasks WHERE user_id=$1 AND task_date=$2 ORDER BY activity_type`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	r := &TasksResponse{Items: []*Task{}, ResetsAt: progression.NextMoscowMidnight(now)}
	for rows.Next() {
		t := &Task{}
		if err := rows.Scan(&t.ID, &t.ActivityType, &t.Progress, &t.Target, &t.XPReward, &t.CompletedAt); err != nil {
			return nil, err
		}
		t.Title, t.Description = taskText(t.ActivityType)
		t.Status = "active"
		if t.CompletedAt != nil {
			t.Status = "completed"
		}
		r.Items = append(r.Items, t)
	}
	return r, rows.Err()
}
func (s *service) ProcessActivity(ctx context.Context, userID string, req DemoActivityRequest) error {
	now := s.now()
	if err := requirePet(ctx, s.db, userID); err != nil {
		return err
	}
	if err := ensureRewards(ctx, s.db, userID); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT activity_type FROM activity_events WHERE id=$1 AND user_id=$2`, req.EventID, userID).Scan(&existing)
	if err == nil {
		if existing != req.ActivityType {
			return ErrEventConflict
		}
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO activity_events(id,user_id,activity_type,occurred_at) VALUES($1,$2,$3,$4)`, req.EventID, userID, req.ActivityType, now); err != nil {
		if pge, ok := err.(*pq.Error); ok && pge.Code == "23505" {
			return nil
		}
		return err
	}
	date := progression.MoscowDate(now)
	var id string
	var progress, target, xp int
	var completed *time.Time
	err = tx.QueryRowContext(ctx, `SELECT id,progress,target,xp_reward,completed_at FROM daily_tasks WHERE user_id=$1 AND task_date=$2 AND activity_type=$3 FOR UPDATE`, userID, date, req.ActivityType).Scan(&id, &progress, &target, &xp, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		if err = ensureTasks(ctx, tx, userID, date); err != nil {
			return err
		}
		err = tx.QueryRowContext(ctx, `SELECT id,progress,target,xp_reward,completed_at FROM daily_tasks WHERE user_id=$1 AND task_date=$2 AND activity_type=$3 FOR UPDATE`, userID, date, req.ActivityType).Scan(&id, &progress, &target, &xp, &completed)
	}
	if err != nil {
		return err
	}
	if completed == nil {
		progress++
		if progress >= target {
			completed = ptr(now)
			if _, err = tx.ExecContext(ctx, `INSERT INTO xp_events(id,user_id,source,xp_awarded,occurred_at) VALUES($1,$2,'task',$3,$4)`, uuid.NewString(), userID, xp, now); err != nil {
				return err
			}
			if err = addXPAndUnlock(ctx, tx, userID, xp, now); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE daily_tasks SET progress=$1,completed_at=$2 WHERE id=$3`, progress, completed, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
