package tasks

import (
	"context"
	"database/sql"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/progression"
	"github.com/google/uuid"
)

func requirePet(ctx context.Context, db *sql.DB, userID string) error {
	var ok bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pets WHERE owner_id=$1)`, userID).Scan(&ok); err != nil {
		return err
	}
	if !ok {
		return ErrPetNotFound
	}
	return nil
}

type executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func ensureTasks(ctx context.Context, db executor, userID, date string) error {
	for _, d := range definitions() {
		if _, err := db.ExecContext(ctx, `INSERT INTO daily_tasks(id,user_id,task_date,activity_type,progress,target,xp_reward) VALUES($1,$2,$3,$4,0,$5,$6) ON CONFLICT DO NOTHING`, uuid.NewString(), userID, date, d.activity, d.target, d.xp); err != nil {
			return err
		}
	}
	return nil
}
func ensureRewards(ctx context.Context, db *sql.DB, userID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO user_rewards(id,user_id,reward_type,required_level) VALUES($1,$2,'listingPromotion',2),($3,$2,'deliveryDiscount',5) ON CONFLICT (user_id,reward_type) DO NOTHING`, uuid.NewString(), userID, uuid.NewString())
	return err
}
func addXPAndUnlock(ctx context.Context, tx *sql.Tx, userID string, xp int, now time.Time) error {
	var total, level int
	if err := tx.QueryRowContext(ctx, `SELECT total_xp,level FROM pets WHERE owner_id=$1 FOR UPDATE`, userID).Scan(&total, &level); err != nil {
		return err
	}
	newLevel, next := progression.LevelForXP(total + xp)
	if _, err := tx.ExecContext(ctx, `UPDATE pets SET total_xp=$1,level=$2,next_level_xp=$3,stage=$4,updated_at=$5 WHERE owner_id=$6`, total+xp, newLevel, next, progression.StageForLevel(newLevel), now, userID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE user_rewards SET unlocked_at=$1 WHERE user_id=$2 AND required_level <= $3 AND unlocked_at IS NULL`, now, userID, newLevel)
	return err
}
func ptr(t time.Time) *time.Time { return &t }
