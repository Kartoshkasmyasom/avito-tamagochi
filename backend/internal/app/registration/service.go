package registration

import (
	"context"
	"fmt"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	authservice "github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/service"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/pet"
)

type Service interface {
	Register(context.Context, auth.RegisterRequest) (*auth.User, *auth.Session, error)
}

type service struct {
	auth authservice.AuthService
	pet  pet.Service
}

func NewService(authService authservice.AuthService, petService pet.Service) Service {
	return &service{auth: authService, pet: petService}
}

func (s *service) Register(ctx context.Context, req auth.RegisterRequest) (*auth.User, *auth.Session, error) {
	user, session, err := s.auth.Register(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	if _, err := s.pet.Create(ctx, user.ID); err != nil {
		if cleanupErr := s.auth.DeleteUser(ctx, user.ID); cleanupErr != nil {
			return nil, nil, fmt.Errorf("create pet: %w; cleanup user: %v", err, cleanupErr)
		}
		return nil, nil, fmt.Errorf("create pet: %w", err)
	}
	return user, session, nil
}
