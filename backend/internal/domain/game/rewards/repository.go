package rewards

import (
	"context"
	"database/sql"

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
func ensure(ctx context.Context, db *sql.DB, userID string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO user_rewards(id,user_id,reward_type,required_level) VALUES($1,$2,'listingPromotion',2),($3,$2,'deliveryDiscount',5) ON CONFLICT (user_id,reward_type) DO NOTHING`, uuid.NewString(), userID, uuid.NewString())
	return err
}
