package game

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PetService interface {
	Get(context.Context, string) (*PetState, error)
	Charge(context.Context, string) (*PetActionResult, error)
}

type petService struct {
	db  *sql.DB
	now func() time.Time
}

func NewPetService(db *sql.DB) PetService { return &petService{db: db, now: time.Now} }

func (s *petService) Get(ctx context.Context, userID string) (*PetState, error) {
	now := s.now()
	if err := ensurePet(ctx, s.db, userID, now); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	pet, err := lockPet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	applyDailyDecay(pet, now)
	if _, err = tx.ExecContext(ctx, `UPDATE pets SET battery_level=$1,status=$2,last_decay_date=$3,updated_at=$4 WHERE id=$5`, pet.BatteryLevel, pet.Status, pet.LastDecayDate, pet.UpdatedAt, pet.ID); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return petState(pet), nil
}

func (s *petService) Charge(ctx context.Context, userID string) (*PetActionResult, error) {
	now := s.now()
	if err := ensurePet(ctx, s.db, userID, now); err != nil {
		return nil, err
	}
	if err := ensureRewards(ctx, s.db, userID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	pet, err := lockPet(ctx, tx, userID)
	if err != nil {
		return nil, err
	}
	applyDailyDecay(pet, now)
	if pet.CooldownEndsAt != nil && pet.CooldownEndsAt.After(now) {
		return nil, ErrCooldown
	}
	oldLevel, oldStage := pet.Level, pet.Stage
	batteryAdded := min(25, 100-pet.BatteryLevel)
	pet.BatteryLevel += batteryAdded
	pet.TotalXP += 15
	pet.Level, pet.NextLevelXP = LevelForXP(pet.TotalXP)
	pet.Stage, pet.Status = StageForLevel(pet.Level), StatusForBattery(pet.BatteryLevel)
	pet.CooldownEndsAt, pet.IsActionAvailable, pet.UpdatedAt = ptrTime(now.Add(time.Hour)), false, now
	if _, err = tx.ExecContext(ctx, `UPDATE pets SET level=$1,total_xp=$2,next_level_xp=$3,stage=$4,battery_level=$5,status=$6,is_action_available=$7,cooldown_ends_at=$8,last_decay_date=$9,updated_at=$10 WHERE id=$11`, pet.Level, pet.TotalXP, pet.NextLevelXP, pet.Stage, pet.BatteryLevel, pet.Status, pet.IsActionAvailable, pet.CooldownEndsAt, pet.LastDecayDate, pet.UpdatedAt, pet.ID); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO xp_events(id,user_id,source,xp_awarded,occurred_at) VALUES($1,$2,'charge',15,$3)`, uuid.NewString(), userID, now); err != nil {
		return nil, err
	}
	if err = unlockRewards(ctx, tx, userID, pet.Level, now); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return &PetActionResult{XPAwarded: 15, BatteryAdded: batteryAdded, LevelUp: pet.Level > oldLevel, StageChanged: pet.Stage != oldStage, Pet: petState(pet)}, nil
}

type lockedPet struct {
	ID, Name, Stage, Status, LastDecayDate    string
	Level, TotalXP, NextLevelXP, BatteryLevel int
	IsActionAvailable                         bool
	CooldownEndsAt                            *time.Time
	UpdatedAt                                 time.Time
}

func lockPet(ctx context.Context, tx *sql.Tx, userID string) (*lockedPet, error) {
	p := &lockedPet{}
	err := tx.QueryRowContext(ctx, `SELECT id,name,level,total_xp,next_level_xp,stage,battery_level,status,is_action_available,cooldown_ends_at,last_decay_date,updated_at FROM pets WHERE owner_id=$1 FOR UPDATE`, userID).Scan(&p.ID, &p.Name, &p.Level, &p.TotalXP, &p.NextLevelXP, &p.Stage, &p.BatteryLevel, &p.Status, &p.IsActionAvailable, &p.CooldownEndsAt, &p.LastDecayDate, &p.UpdatedAt)
	return p, err
}

func applyDailyDecay(p *lockedPet, now time.Time) {
	if p.LastDecayDate < MoscowDate(now) {
		p.BatteryLevel = max(0, p.BatteryLevel-20)
		p.LastDecayDate = MoscowDate(now)
	}
	p.Status = StatusForBattery(p.BatteryLevel)
	p.IsActionAvailable = p.CooldownEndsAt == nil || !p.CooldownEndsAt.After(now)
	p.UpdatedAt = now
}
func petState(p *lockedPet) *PetState {
	return &PetState{ID: p.ID, Name: p.Name, Level: p.Level, TotalXP: p.TotalXP, NextLevelXP: p.NextLevelXP, Stage: p.Stage, BatteryLevel: p.BatteryLevel, Status: p.Status, IsActionAvailable: p.IsActionAvailable, CooldownEndsAt: p.CooldownEndsAt, UpdatedAt: p.UpdatedAt}
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
