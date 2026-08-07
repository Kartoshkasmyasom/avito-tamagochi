package game

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrCooldown       = errors.New("cooldown active")
	ErrEventConflict  = errors.New("event conflict")
	ErrRewardNotFound = errors.New("reward not found")
	ErrRewardLocked   = errors.New("reward locked")
)

func ensurePet(ctx context.Context, db *sql.DB, userID string, now time.Time) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO pets (id, owner_id, name, level, total_xp, next_level_xp, stage, battery_level, status, is_action_available, last_decay_date, updated_at)
		VALUES ($1,$2,'Кита (K1-T4)',1,0,100,'egg',100,'happy',true,$3,$4)
		ON CONFLICT (owner_id) DO NOTHING
	`, uuid.NewString(), userID, MoscowDate(now), now)
	return err
}

func ensureRewards(ctx context.Context, db *sql.DB, userID string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO user_rewards (id, user_id, reward_type, required_level)
		VALUES ($1,$2,'listingPromotion',2),($3,$2,'deliveryDiscount',5)
		ON CONFLICT (user_id, reward_type) DO NOTHING
	`, uuid.NewString(), userID, uuid.NewString())
	return err
}

func addXPAndUnlock(ctx context.Context, tx *sql.Tx, userID string, xp int, now time.Time) error {
	var total, level int
	if err := tx.QueryRowContext(ctx, `SELECT total_xp,level FROM pets WHERE owner_id=$1 FOR UPDATE`, userID).Scan(&total, &level); err != nil {
		return err
	}
	newLevel, next := LevelForXP(total + xp)
	_, err := tx.ExecContext(ctx, `UPDATE pets SET total_xp=$1,level=$2,next_level_xp=$3,stage=$4,updated_at=$5 WHERE owner_id=$6`, total+xp, newLevel, next, StageForLevel(newLevel), now, userID)
	if err != nil {
		return err
	}
	return unlockRewards(ctx, tx, userID, newLevel, now)
}

func unlockRewards(ctx context.Context, tx *sql.Tx, userID string, level int, now time.Time) error {
	_, err := tx.ExecContext(ctx, `UPDATE user_rewards SET unlocked_at=$1 WHERE user_id=$2 AND required_level <= $3 AND unlocked_at IS NULL`, now, userID, level)
	return err
}

func ptrTime(t time.Time) *time.Time { return &t }

func levelXP(level int) int {
	if level <= 1 {
		return 0
	}
	return levelThresholds[level-2]
}

func taskText(activity string) (string, string) {
	switch activity {
	case ActivityListingViewed:
		return "Посмотреть объявление", "Посмотрите объявление на Авито"
	case ActivityFavoriteAdded:
		return "Добавить в избранное", "Добавьте объявление в избранное"
	default:
		return "Опубликовать объявление", "Опубликуйте объявление на Авито"
	}
}

func rewardText(rewardType string) (string, string) {
	if rewardType == "listingPromotion" {
		return "Продвижение объявления", "Бонус на продвижение объявления"
	}
	return "Авито Доставка", "Бонус на Авито Доставку"
}
