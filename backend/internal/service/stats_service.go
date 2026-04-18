package service

import (
	"context"

	"edupulse/internal/repo"
)

type StatsService struct {
	users *repo.UserRepo
}

func NewStatsService(users *repo.UserRepo) *StatsService {
	return &StatsService{users: users}
}

type StatsResult struct {
	Roles      []repo.RoleCount `json:"roles"`
	TotalUsers int              `json:"total_users"`
	TotalOnline int             `json:"total_online"`
}

func (s *StatsService) Get(ctx context.Context) (StatsResult, error) {
	roles, err := s.users.Stats(ctx)
	if err != nil {
		return StatsResult{}, err
	}

	var total, online int
	for _, rc := range roles {
		total += rc.Total
		online += rc.Online
	}

	return StatsResult{
		Roles:       roles,
		TotalUsers:  total,
		TotalOnline: online,
	}, nil
}
