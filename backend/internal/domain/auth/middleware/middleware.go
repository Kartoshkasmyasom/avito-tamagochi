package middleware

import (
	"errors"
	"net/http"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/repository"
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, auth.ErrorResponse{Code: "internal_error", Message: "Internal server error"})
			return
		}
		c.Set("userID", session.UserID)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, auth.ErrorResponse{Code: "unauthorized", Message: "Authentication is required"})
}
