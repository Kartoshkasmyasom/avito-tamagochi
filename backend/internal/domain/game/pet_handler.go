package game

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/pet"
	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPet(c *gin.Context) {
	value, err := h.pet.Get(c, c.GetString("userID"))
	if err != nil {
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
		if errors.Is(err, pet.ErrCooldown) {
			var cooldown *pet.CooldownError
			if errors.As(err, &cooldown) {
				c.Header("Retry-After", fmtSeconds(cooldown.RetryAfter))
			}
			response.ErrorJSON(c, http.StatusTooManyRequests, "cooldown_active", "Pet action is on cooldown")
			return
		}
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}

func fmtSeconds(seconds int) string {
	if seconds < 1 {
		seconds = 1
	}
	return strconv.Itoa(seconds)
}
