package game

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type TasksService interface {
	List(context.Context, string) (*TasksResponse, error)
	ProcessActivity(context.Context, string, DemoActivityRequest) error
}

type tasksService struct {
	db  *sql.DB
	now func() time.Time
}

func NewTasksService(db *sql.DB) TasksService { return &tasksService{db: db, now: time.Now} }

func (s *tasksService) List(ctx context.Context, userID string) (*TasksResponse, error) {
	now := s.now()
	if err := ensurePet(ctx, s.db, userID, now); err != nil {
		return nil, err
	}
	date := MoscowDate(now)
	if err := ensureTasks(ctx, s.db, userID, date); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,activity_type,progress,target,xp_reward,completed_at FROM daily_tasks WHERE user_id=$1 AND task_date=$2 ORDER BY activity_type`, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	resp := &TasksResponse{Items: []*Task{}, ResetsAt: NextMoscowMidnight(now)}
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
		resp.Items = append(resp.Items, t)
	}
	return resp, rows.Err()
}

func (s *tasksService) ProcessActivity(ctx context.Context, userID string, req DemoActivityRequest) error {
	now := s.now()
	if err := ensurePet(ctx, s.db, userID, now); err != nil {
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
	date := MoscowDate(now)
	var taskID string
	var progress, target, xp int
	var completedAt *time.Time
	err = tx.QueryRowContext(ctx, `SELECT id,progress,target,xp_reward,completed_at FROM daily_tasks WHERE user_id=$1 AND task_date=$2 AND activity_type=$3 FOR UPDATE`, userID, date, req.ActivityType).Scan(&taskID, &progress, &target, &xp, &completedAt)
	if errors.Is(err, sql.ErrNoRows) {
		if err = ensureTasksTx(ctx, tx, userID, date); err != nil {
			return err
		}
		err = tx.QueryRowContext(ctx, `SELECT id,progress,target,xp_reward,completed_at FROM daily_tasks WHERE user_id=$1 AND task_date=$2 AND activity_type=$3 FOR UPDATE`, userID, date, req.ActivityType).Scan(&taskID, &progress, &target, &xp, &completedAt)
	}
	if err != nil {
		return err
	}
	if completedAt == nil {
		progress++
		if progress >= target {
			completedAt = ptrTime(now)
			if _, err = tx.ExecContext(ctx, `INSERT INTO xp_events(id,user_id,source,xp_awarded,occurred_at) VALUES($1,$2,'task',$3,$4)`, uuid.NewString(), userID, xp, now); err != nil {
				return err
			}
			if err = addXPAndUnlock(ctx, tx, userID, xp, now); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE daily_tasks SET progress=$1,completed_at=$2 WHERE id=$3`, progress, completedAt, taskID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func ensureTasks(ctx context.Context, db *sql.DB, userID, date string) error {
	for _, d := range taskDefinitions() {
		if _, err := db.ExecContext(ctx, `INSERT INTO daily_tasks(id,user_id,task_date,activity_type,progress,target,xp_reward) VALUES($1,$2,$3,$4,0,$5,$6) ON CONFLICT DO NOTHING`, uuid.NewString(), userID, date, d.activity, d.target, d.xp); err != nil {
			return err
		}
	}
	return nil
}
func ensureTasksTx(ctx context.Context, tx *sql.Tx, userID, date string) error {
	for _, d := range taskDefinitions() {
		if _, err := tx.ExecContext(ctx, `INSERT INTO daily_tasks(id,user_id,task_date,activity_type,progress,target,xp_reward) VALUES($1,$2,$3,$4,0,$5,$6) ON CONFLICT DO NOTHING`, uuid.NewString(), userID, date, d.activity, d.target, d.xp); err != nil {
			return err
		}
	}
	return nil
}
func taskDefinitions() []struct {
	activity   string
	target, xp int
} {
	return []struct {
		activity   string
		target, xp int
	}{{ActivityListingViewed, 1, 40}, {ActivityFavoriteAdded, 1, 70}, {ActivityListingPublished, 1, 100}}
}
