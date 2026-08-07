package game

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	pet         PetService
	tasks       TasksService
	rewards     RewardsService
	leaderboard LeaderboardService
	summary     SummaryService
}

func NewHandler(pet PetService, tasks TasksService, rewards RewardsService, leaderboard LeaderboardService, summary SummaryService) *Handler {
	return &Handler{pet: pet, tasks: tasks, rewards: rewards, leaderboard: leaderboard, summary: summary}
}

func (h *Handler) GetPet(c *gin.Context) {
	v, err := h.pet.Get(c, c.GetString("userID"))
	if err != nil {
		writeInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *Handler) PetAction(c *gin.Context) {
	var req PetActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	v, err := h.pet.Charge(c, c.GetString("userID"))
	if err != nil {
		if errors.Is(err, ErrCooldown) {
			writeError(c, http.StatusTooManyRequests, "cooldown_active", "Pet action is on cooldown")
			return
		}
		writeInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *Handler) ListTasks(c *gin.Context) {
	v, err := h.tasks.List(c, c.GetString("userID"))
	if err != nil {
		writeInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *Handler) Activity(c *gin.Context) {
	var req DemoActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.tasks.ProcessActivity(c, c.GetString("userID"), req); err != nil {
		if errors.Is(err, ErrEventConflict) {
			writeError(c, http.StatusConflict, "event_conflict", "Event ID was already used with another activity")
			return
		}
		writeInternal(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *Handler) ListRewards(c *gin.Context) {
	v, err := h.rewards.List(c, c.GetString("userID"))
	if err != nil {
		writeInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *Handler) ClaimReward(c *gin.Context) {
	v, err := h.rewards.Claim(c, c.GetString("userID"), c.Param("rewardId"))
	if err != nil {
		switch {
		case errors.Is(err, ErrRewardNotFound):
			writeError(c, http.StatusNotFound, "not_found", "Reward not found")
		case errors.Is(err, ErrRewardLocked):
			writeError(c, http.StatusConflict, "reward_locked", "Reward is locked")
		default:
			writeInternal(c, err)
		}
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *Handler) Leaderboard(c *gin.Context) {
	v, err := h.leaderboard.Get(c, c.GetString("userID"))
	if err != nil {
		writeInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
func (h *Handler) Summary(c *gin.Context) {
	v, err := h.summary.Get(c, c.GetString("userID"))
	if err != nil {
		writeInternal(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func writeInternal(c *gin.Context, err error) {
	writeError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
}
func writeError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Code: code, Message: message})
}
