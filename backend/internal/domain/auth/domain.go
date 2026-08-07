package auth

import "time"

// --- REQUEST/RESPONSE DTOs ---

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email,max=254"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	DisplayName string `json:"displayName" binding:"required,min=2,max=32"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=254"`
	Password string `json:"password" binding:"required,min=8,max=128"`
}

// --- DOMAIN MODELS ---

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Session struct {
	ID        string
	UserID    string
	ExpiresAt time.Time
	CreatedAt time.Time
}
