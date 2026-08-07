package game

import (
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/tasks"
	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTasks(c *gin.Context) {
	value, err := h.tasks.List(c, c.GetString("userID"))
	if err != nil {
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
		case errors.Is(err, tasks.ErrEventConflict):
			response.ErrorJSON(c, http.StatusConflict, "event_conflict", "Event ID was already used with another activity")
		default:
			response.InternalError(c)
		}
		return
	}
	c.Status(http.StatusNoContent)
}
