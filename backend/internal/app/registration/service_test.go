package registration

import (
	"context"
	"errors"
	"testing"

	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth"
	authservice "github.com/accelolabs/avito-tamagochi/backend/internal/domain/auth/service"
	"github.com/accelolabs/avito-tamagochi/backend/internal/domain/game/pet"
)

func TestServiceRegisterCreatesPetOutsideAuth(t *testing.T) {
	user := &auth.User{ID: "user-id"}
	session := &auth.Session{ID: "session-id"}
	authStub := &authServiceStub{user: user, session: session}
	petStub := &petServiceStub{}

	service := NewService(authStub, petStub)
	gotUser, gotSession, err := service.Register(context.Background(), auth.RegisterRequest{})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if gotUser != user || gotSession != session {
		t.Fatal("registration result was changed")
	}
	if !petStub.created {
		t.Fatal("pet was not created")
	}
	if authStub.deletedUserID != "" {
		t.Fatal("user cleanup was called on successful registration")
	}
}

func TestServiceRegisterCleansUpUserWhenPetCreationFails(t *testing.T) {
	authStub := &authServiceStub{user: &auth.User{ID: "user-id"}, session: &auth.Session{ID: "session-id"}}
	petStub := &petServiceStub{createErr: errors.New("pet storage unavailable")}

	service := NewService(authStub, petStub)
	if _, _, err := service.Register(context.Background(), auth.RegisterRequest{}); err == nil {
		t.Fatal("expected pet creation error")
	}
	if authStub.deletedUserID != "user-id" {
		t.Fatalf("deleted user = %q, want user-id", authStub.deletedUserID)
	}
}

type authServiceStub struct {
	user          *auth.User
	session       *auth.Session
	deletedUserID string
}

func (s *authServiceStub) Register(context.Context, auth.RegisterRequest) (*auth.User, *auth.Session, error) {
	return s.user, s.session, nil
}

func (s *authServiceStub) Login(context.Context, auth.LoginRequest) (*auth.User, *auth.Session, error) {
	return nil, nil, nil
}

func (s *authServiceStub) Logout(context.Context, string) error { return nil }

func (s *authServiceStub) DeleteUser(_ context.Context, userID string) error {
	s.deletedUserID = userID
	return nil
}

type petServiceStub struct {
	created   bool
	createErr error
}

func (s *petServiceStub) Create(context.Context, string) (*pet.PetState, error) {
	s.created = true
	return nil, s.createErr
}

func (s *petServiceStub) Get(context.Context, string) (*pet.PetState, error) {
	return nil, nil
}

func (s *petServiceStub) Charge(context.Context, string) (*pet.PetActionResult, error) {
	return nil, nil
}

var _ authservice.AuthService = (*authServiceStub)(nil)
var _ pet.Service = (*petServiceStub)(nil)
