package service

import (
	"context"
	"fmt"

	"edupulse/internal/repo"
)

type UserService struct {
	users *repo.UserRepo
}

func NewUserService(users *repo.UserRepo) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Me(ctx context.Context, userID int64) (*repo.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user id")
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}
	u.PasswordHash = "" // на всякий, хотя json:"-" и так скрывает
	return u, nil
}