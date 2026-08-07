package game

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type RewardsService interface {
	List(context.Context, string) (*RewardsResponse, error)
	Claim(context.Context, string, string) (*Reward, error)
}
type rewardsService struct {
	db  *sql.DB
	now func() time.Time
}

func NewRewardsService(db *sql.DB) RewardsService { return &rewardsService{db: db, now: time.Now} }
func (s *rewardsService) List(ctx context.Context, userID string) (*RewardsResponse, error) {
	now := s.now()
	if err := ensurePet(ctx, s.db, userID, now); err != nil {
		return nil, err
	}
	if err := ensureRewards(ctx, s.db, userID); err != nil {
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
	result := &RewardsResponse{Items: []*Reward{}}
	for rows.Next() {
		r := &Reward{}
		var required int
		if err := rows.Scan(&r.ID, &r.Type, &required, &r.ClaimedAt); err != nil {
			return nil, err
		}
		r.RequiredLevel = required
		r.Title, r.Description = rewardText(r.Type)
		r.CurrentXP = xp
		r.RequiredXP = levelXP(required)
		if r.ClaimedAt != nil {
			r.Status = "claimed"
		} else if levelForXPOnly(xp) >= required {
			r.Status = "available"
		} else {
			r.Status = "locked"
		}
		result.Items = append(result.Items, r)
	}
	return result, rows.Err()
}
func (s *rewardsService) Claim(ctx context.Context, userID, rewardID string) (*Reward, error) {
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
	var typ string
	var required int
	var claimed *time.Time
	if err = tx.QueryRowContext(ctx, `SELECT reward_type,required_level,claimed_at FROM user_rewards WHERE id=$1 AND user_id=$2 FOR UPDATE`, rewardID, userID).Scan(&typ, &required, &claimed); errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRewardNotFound
	} else if err != nil {
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
		claimed = ptrTime(now)
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
func levelForXPOnly(xp int) int { level, _ := LevelForXP(xp); return level }
