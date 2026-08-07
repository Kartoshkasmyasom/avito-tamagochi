package game

import (
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/rewards"
	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListRewards(c *gin.Context) {
	value, err := h.rewards.List(c, c.GetString("userID"))
	if err != nil {
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}

func (h *Handler) ClaimReward(c *gin.Context) {
	value, err := h.rewards.Claim(c, c.GetString("userID"), c.Param("rewardId"))
	if err != nil {
		switch {
		case errors.Is(err, rewards.ErrRewardNotFound), errors.Is(err, rewards.ErrPetNotFound):
			response.ErrorJSON(c, http.StatusNotFound, "not_found", "Reward not found")
		case errors.Is(err, rewards.ErrRewardLocked):
			response.ErrorJSON(c, http.StatusConflict, "reward_locked", "Reward is locked")
		default:
			response.InternalError(c)
		}
		return
	}
	c.JSON(http.StatusOK, value)
}
