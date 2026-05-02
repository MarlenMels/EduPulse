package service

import (
	"context"
	"fmt"

	"edupulse/internal/auth"
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
	u.PasswordHash = ""
	return u, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, currentPassword, newPassword string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("user not found")
	}
	if !auth.CheckPassword(u.PasswordHash, currentPassword) {
		return fmt.Errorf("current password is incorrect")
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.users.UpdatePassword(ctx, userID, hash)
}
