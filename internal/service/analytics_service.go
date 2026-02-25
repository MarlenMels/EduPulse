package service

import (
	"context"

	"edupulse/internal/repo"
)

type AnalyticsService struct{ repo *repo.AnalyticsRepo }

func NewAnalyticsService(r *repo.AnalyticsRepo) *AnalyticsService { return &AnalyticsService{repo: r} }

func (s *AnalyticsService) ListSessionsByH3(ctx context.Context, h3Index, day string, limit int) ([]repo.AnalyticsRow, error) {
	return s.repo.ListSessionsByH3(ctx, h3Index, day, limit)
}