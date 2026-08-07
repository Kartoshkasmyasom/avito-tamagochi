package pet

import (
	"context"
	"database/sql"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/progression"
	"github.com/google/uuid"
)

type locked struct {
	ID, Name, Stage, Status, LastDecayDate    string
	Level, TotalXP, NextLevelXP, BatteryLevel int
	IsActionAvailable                         bool
	CooldownEndsAt                            *time.Time
	UpdatedAt                                 time.Time
}

func create(ctx context.Context, db *sql.DB, userID string, now time.Time) error {
	_, err := db.ExecContext(ctx, `INSERT INTO pets (id,owner_id,name,level,total_xp,next_level_xp,stage,battery_level,status,is_action_available,last_decay_date,updated_at) VALUES ($1,$2,'Кита (K1-T4)',1,0,100,'egg',100,'happy',true,$3,$4) ON CONFLICT (owner_id) DO NOTHING`, uuid.NewString(), userID, progression.MoscowDate(now), now)
	return err
}
func lock(ctx context.Context, tx *sql.Tx, userID string) (*locked, error) {
	p := &locked{}
	err := tx.QueryRowContext(ctx, `SELECT id,name,level,total_xp,next_level_xp,stage,battery_level,status,is_action_available,cooldown_ends_at,last_decay_date,updated_at FROM pets WHERE owner_id=$1 FOR UPDATE`, userID).Scan(&p.ID, &p.Name, &p.Level, &p.TotalXP, &p.NextLevelXP, &p.Stage, &p.BatteryLevel, &p.Status, &p.IsActionAvailable, &p.CooldownEndsAt, &p.LastDecayDate, &p.UpdatedAt)
	return p, err
}
func decay(p *locked, now time.Time) {
	if p.LastDecayDate < progression.MoscowDate(now) {
		p.BatteryLevel = max(0, p.BatteryLevel-20)
		p.LastDecayDate = progression.MoscowDate(now)
	}
	p.Status = progression.StatusForBattery(p.BatteryLevel)
	p.IsActionAvailable = p.CooldownEndsAt == nil || !p.CooldownEndsAt.After(now)
	p.UpdatedAt = now
}
func state(p *locked) *PetState {
	return &PetState{ID: p.ID, Name: p.Name, Level: p.Level, TotalXP: p.TotalXP, NextLevelXP: p.NextLevelXP, Stage: p.Stage, BatteryLevel: p.BatteryLevel, Status: p.Status, IsActionAvailable: p.IsActionAvailable, CooldownEndsAt: p.CooldownEndsAt, UpdatedAt: p.UpdatedAt}
}
func ensureRewards(ctx context.Context, db *sql.DB, userID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO user_rewards (id,user_id,reward_type,required_level) VALUES ($1,$2,'listingPromotion',2),($3,$2,'deliveryDiscount',5) ON CONFLICT (user_id,reward_type) DO NOTHING`, uuid.NewString(), userID, uuid.NewString())
	return err
}
func unlock(ctx context.Context, tx *sql.Tx, userID string, level int, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE user_rewards SET unlocked_at=$1 WHERE user_id=$2 AND required_level <= $3 AND unlocked_at IS NULL`, now, userID, level)
	return err
}
func ptr(t time.Time) *time.Time { return &t }
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
