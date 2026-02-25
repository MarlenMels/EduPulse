package auth

import (
	"context"
	"errors"
	"time"

	"edupulse/internal/repo"
)

type Service struct {
	users  *repo.UserRepo
	secret string
}

func NewService(users *repo.UserRepo, secret string) *Service {
	return &Service{users: users, secret: secret}
}

type LoginResult struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func (s *Service) Login(ctx context.Context, email, password string) (LoginResult, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, err
	}
	if u == nil || !CheckPassword(u.PasswordHash, password) {
		return LoginResult{}, errors.New("invalid credentials")
	}
	tok, err := NewToken(u.ID, u.Role, s.secret, 24*time.Hour)
	if err != nil {
		return LoginResult{}, err
	}
	return LoginResult{Token: tok, Role: u.Role}, nil
}