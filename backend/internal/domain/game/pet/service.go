package pet

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/progression"
	"github.com/google/uuid"
)

type Service interface {
	Create(context.Context, string) (*PetState, error)
	Get(context.Context, string) (*PetState, error)
	Charge(context.Context, string) (*PetActionResult, error)
}
type service struct {
	db  *sql.DB
	now func() time.Time
}

func NewService(db *sql.DB) Service { return &service{db: db, now: time.Now} }
func (s *service) Create(ctx context.Context, userID string) (*PetState, error) {
	if err := create(ctx, s.db, userID, s.now()); err != nil {
		return nil, err
	}
	return s.Get(ctx, userID)
}
func (s *service) Get(ctx context.Context, userID string) (*PetState, error) {
	now := s.now()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	p, err := lock(ctx, tx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPetNotFound
	}
	if err != nil {
		return nil, err
	}
	decay(p, now)
	if _, err = tx.ExecContext(ctx, `UPDATE pets SET battery_level=$1,status=$2,last_decay_date=$3,updated_at=$4 WHERE id=$5`, p.BatteryLevel, p.Status, p.LastDecayDate, p.UpdatedAt, p.ID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return state(p), nil
}
func (s *service) Charge(ctx context.Context, userID string) (*PetActionResult, error) {
	now := s.now()
	if err := ensureRewards(ctx, s.db, userID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	p, err := lock(ctx, tx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPetNotFound
	}
	if err != nil {
		return nil, err
	}
	decay(p, now)
	if p.CooldownEndsAt != nil && p.CooldownEndsAt.After(now) {
		return nil, NewCooldownError(now, *p.CooldownEndsAt)
	}
	oldLevel, oldStage := p.Level, p.Stage
	added := min(25, 100-p.BatteryLevel)
	p.BatteryLevel += added
	p.TotalXP += 15
	p.Level, p.NextLevelXP = progression.LevelForXP(p.TotalXP)
	p.Stage, p.Status = progression.StageForLevel(p.Level), progression.StatusForBattery(p.BatteryLevel)
	p.CooldownEndsAt, p.IsActionAvailable, p.UpdatedAt = ptr(now.Add(time.Hour)), false, now
	if _, err = tx.ExecContext(ctx, `UPDATE pets SET level=$1,total_xp=$2,next_level_xp=$3,stage=$4,battery_level=$5,status=$6,is_action_available=$7,cooldown_ends_at=$8,last_decay_date=$9,updated_at=$10 WHERE id=$11`, p.Level, p.TotalXP, p.NextLevelXP, p.Stage, p.BatteryLevel, p.Status, p.IsActionAvailable, p.CooldownEndsAt, p.LastDecayDate, p.UpdatedAt, p.ID); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO xp_events(id,user_id,source,xp_awarded,occurred_at) VALUES($1,$2,'charge',15,$3)`, uuid.NewString(), userID, now); err != nil {
		return nil, err
	}
	if err = unlock(ctx, tx, userID, p.Level, now); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &PetActionResult{XPAwarded: 15, BatteryAdded: added, LevelUp: p.Level > oldLevel, StageChanged: p.Stage != oldStage, Pet: state(p)}, nil
}
