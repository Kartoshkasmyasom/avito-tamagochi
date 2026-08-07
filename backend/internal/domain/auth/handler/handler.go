package handler

import (
	"context"
	"errors"
	"net/http"
	"os"
	"regexp"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/service"
	"github.com/accelolabs/avito-tamagochi/backend/internal/http/response"
	"github.com/gin-gonic/gin"
)

const (
	sessionCookieName = "session_id"
	sessionMaxAge     = 604800 // 7 days in seconds
)

var displayNamePattern = regexp.MustCompile(`^[A-Za-z_-]+$`)

type AuthHandler struct {
	service      service.AuthService
	registration RegistrationService
}

type RegistrationService interface {
	Register(context.Context, auth.RegisterRequest) (*auth.User, *auth.Session, error)
}

func NewAuthHandler(service service.AuthService, registration RegistrationService) *AuthHandler {
	return &AuthHandler{service: service, registration: registration}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if !displayNamePattern.MatchString(req.DisplayName) {
		response.ErrorJSON(c, http.StatusBadRequest, "validation_error", "displayName must contain only latin letters, underscores, or hyphens")
		return
	}

	user, session, err := h.registration.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.ErrorJSON(c, http.StatusConflict, "email_already_exists", "Email is already registered")
			return
		}
		response.ErrorJSON(c, http.StatusInternalServerError, "internal_error", "Failed to register user")
		return
	}

	h.setSessionCookie(c, session)
	c.JSON(http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorJSON(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	user, session, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.ErrorJSON(c, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
			return
		}
		response.ErrorJSON(c, http.StatusInternalServerError, "internal_error", "Failed to login")
		return
	}

	h.setSessionCookie(c, session)
	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie(sessionCookieName)
	if err == nil { // Only attempt to delete from DB if cookie exists
		if err := h.service.Logout(c.Request.Context(), sessionID); err != nil {
			// Log the error but don't fail the request, as the goal is to ensure the client is logged out.
			// In a real app, you might want to log this for monitoring.
			_ = c.Error(err) // Use c.Error to log the error with Gin's logger
		}
	}

	h.clearSessionCookie(c)
	c.Status(http.StatusNoContent) // 204 No Content as per spec
}

// --- Helper Functions ---

func (h *AuthHandler) setSessionCookie(c *gin.Context, session *auth.Session) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	// MaxAge is calculated from now until expiration, ensuring it's always positive.
	// If session.ExpiresAt is in the past, MaxAge will be 0 or negative, effectively expiring the cookie.
	c.SetCookie(sessionCookieName, session.ID, sessionMaxAge, "/", "", isSecure, true)
}

func (h *AuthHandler) clearSessionCookie(c *gin.Context) {
	isSecure := os.Getenv("GIN_MODE") == "release"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookieName, "", -1, "/", "", isSecure, true) // MaxAge -1 clears the cookie
}
