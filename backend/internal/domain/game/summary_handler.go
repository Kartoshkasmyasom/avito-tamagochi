package game

import (
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Summary(c *gin.Context) {
	value, err := h.summary.Get(c, c.GetString("userID"))
	if err != nil {
		response.InternalError(c)
		return
	}
	c.JSON(http.StatusOK, value)
}
