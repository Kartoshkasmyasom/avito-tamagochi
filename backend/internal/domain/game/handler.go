package game

import (
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/leaderboard"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/pet"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/rewards"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/summary"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/tasks"
	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	pet         pet.Service
	tasks       tasks.Service
	rewards     rewards.Service
	leaderboard leaderboard.Service
	summary     summary.Service
}

func NewHandler(petService pet.Service, tasksService tasks.Service, rewardsService rewards.Service, leaderboardService leaderboard.Service, summaryService summary.Service) *Handler {
	return &Handler{pet: petService, tasks: tasksService, rewards: rewardsService, leaderboard: leaderboardService, summary: summaryService}
}

func (h *Handler) GetPet(c *gin.Context) {
	value, err := h.pet.Get(c, c.GetString("userID"))
	if err != nil {
		if errors.Is(err, pet.ErrPetNotFound) {
			response.ErrorJSON(c, http.StatusNotFound, "pet_not_found", "Pet has not been created")
			return
		}
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
func (h *Handler) PetAction(c *gin.Context) {
	var req pet.PetActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	value, err := h.pet.Charge(c, c.GetString("userID"))
	if err != nil {
		if errors.Is(err, pet.ErrPetNotFound) {
			response.ErrorJSON(c, http.StatusNotFound, "pet_not_found", "Pet has not been created")
			return
		}
		if errors.Is(err, pet.ErrCooldown) {
			response.ErrorJSON(c, http.StatusTooManyRequests, "cooldown_active", "Pet action is on cooldown")
			return
		}
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
func (h *Handler) ListTasks(c *gin.Context) {
	value, err := h.tasks.List(c, c.GetString("userID"))
	if err != nil {
		if errors.Is(err, tasks.ErrPetNotFound) {
			response.ErrorJSON(c, http.StatusNotFound, "pet_not_found", "Pet has not been created")
			return
		}
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
func (h *Handler) Activity(c *gin.Context) {
	var req tasks.DemoActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.tasks.ProcessActivity(c, c.GetString("userID"), req); err != nil {
		switch {
		case errors.Is(err, tasks.ErrPetNotFound):
			response.ErrorJSON(c, http.StatusNotFound, "pet_not_found", "Pet has not been created")
		case errors.Is(err, tasks.ErrEventConflict):
			response.ErrorJSON(c, http.StatusConflict, "event_conflict", "Event ID was already used with another activity")
		default:
			response.InternalError(c)
		}
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) ListRewards(c *gin.Context) {
	value, err := h.rewards.List(c, c.GetString("userID"))
	if err != nil {
		if errors.Is(err, rewards.ErrPetNotFound) {
			response.ErrorJSON(c, http.StatusNotFound, "pet_not_found", "Pet has not been created")
			return
		}
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
func (h *Handler) ClaimReward(c *gin.Context) {
	value, err := h.rewards.Claim(c, c.GetString("userID"), c.Param("rewardId"))
	if err != nil {
		switch {
		case errors.Is(err, rewards.ErrPetNotFound):
			response.ErrorJSON(c, http.StatusNotFound, "pet_not_found", "Pet has not been created")
		case errors.Is(err, rewards.ErrRewardNotFound):
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
func (h *Handler) Leaderboard(c *gin.Context) {
	value, err := h.leaderboard.Get(c, c.GetString("userID"))
	if err != nil {
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
func (h *Handler) Summary(c *gin.Context) {
	value, err := h.summary.Get(c, c.GetString("userID"))
	if err != nil {
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
