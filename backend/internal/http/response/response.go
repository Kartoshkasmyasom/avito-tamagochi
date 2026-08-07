package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, status int, body any) {
	c.JSON(status, body)
}

func ErrorJSON(c *gin.Context, status int, code, message string) {
	JSON(c, status, Error{Code: code, Message: message})
}

func InternalError(c *gin.Context) {
	ErrorJSON(c, http.StatusInternalServerError, "internal_error", "Internal server error")
}
