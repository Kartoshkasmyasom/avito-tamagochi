package middleware

import (
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/repository"
	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

const sessionCookieName = "session_id"

func RequireSession(repo repository.AuthRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie(sessionCookieName)
		if err != nil || sessionID == "" {
			unauthorized(c)
			return
		}
		session, err := repo.FindSession(c.Request.Context(), sessionID)
		if err != nil {
			if errors.Is(err, repository.ErrSessionNotFound) {
				unauthorized(c)
				return
			}
			response.ErrorJSON(c, http.StatusInternalServerError, "internal_error", "Internal server error")
			c.Abort()
			return
		}
		c.Set("userID", session.UserID)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	response.ErrorJSON(c, http.StatusUnauthorized, "unauthorized", "Authentication is required")
	c.Abort()
}
