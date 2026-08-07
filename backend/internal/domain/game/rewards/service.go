package rewards

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/progression"
)

type Service interface {
	List(context.Context, string) (*RewardsResponse, error)
	Claim(context.Context, string, string) (*Reward, error)
}
type service struct {
	db  *sql.DB
	now func() time.Time
}

func NewService(db *sql.DB) Service { return &service{db: db, now: time.Now} }
func (s *service) List(ctx context.Context, userID string) (*RewardsResponse, error) {
	if err := requirePet(ctx, s.db, userID); err != nil {
		return nil, err
	}
	if err := ensure(ctx, s.db, userID); err != nil {
		return nil, err
	}
	var xp int
	if err := s.db.QueryRowContext(ctx, `SELECT total_xp FROM pets WHERE owner_id=$1`, userID).Scan(&xp); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id,reward_type,required_level,claimed_at FROM user_rewards WHERE user_id=$1 ORDER BY required_level`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	r := &RewardsResponse{Items: []*Reward{}}
	for rows.Next() {
		x := &Reward{}
		var required int
		if err := rows.Scan(&x.ID, &x.Type, &required, &x.ClaimedAt); err != nil {
			return nil, err
		}
		x.RequiredLevel = required
		x.Title, x.Description = rewardText(x.Type)
		x.CurrentXP = xp
		x.RequiredXP = levelXP(required)
		if x.ClaimedAt != nil {
			x.Status = "claimed"
		} else if levelOnly(xp) >= required {
			x.Status = "available"
		} else {
			x.Status = "locked"
		}
		r.Items = append(r.Items, x)
	}
	return r, rows.Err()
}
func (s *service) Claim(ctx context.Context, userID, rewardID string) (*Reward, error) {
	now := s.now()
	if err := requirePet(ctx, s.db, userID); err != nil {
		return nil, err
	}
	if err := ensure(ctx, s.db, userID); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var typ string
	var required int
	var claimed *time.Time
	err = tx.QueryRowContext(ctx, `SELECT reward_type,required_level,claimed_at FROM user_rewards WHERE id=$1 AND user_id=$2 FOR UPDATE`, rewardID, userID).Scan(&typ, &required, &claimed)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRewardNotFound
	}
	if err != nil {
		return nil, err
	}
	var level, xp int
	if err = tx.QueryRowContext(ctx, `SELECT level,total_xp FROM pets WHERE owner_id=$1`, userID).Scan(&level, &xp); err != nil {
		return nil, err
	}
	if claimed == nil && level < required {
		return nil, ErrRewardLocked
	}
	if claimed == nil {
		claimed = &now
		if _, err = tx.ExecContext(ctx, `UPDATE user_rewards SET claimed_at=$1 WHERE id=$2`, claimed, rewardID); err != nil {
			return nil, err
		}
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	r := &Reward{ID: rewardID, Type: typ, RequiredLevel: required, CurrentXP: xp, RequiredXP: levelXP(required), ClaimedAt: claimed, Status: "claimed"}
	r.Title, r.Description = rewardText(typ)
	return r, nil
}
func levelOnly(xp int) int { level, _ := progression.LevelForXP(xp); return level }
func levelXP(level int) int {
	return progression.XPForLevel(level)
}
