package service

import (
	"context"

	"edupulse/internal/repo"
)

type NotificationService struct{ repo *repo.NotificationRepo }

func NewNotificationService(r *repo.NotificationRepo) *NotificationService { return &NotificationService{repo: r} }

func (s *NotificationService) Create(ctx context.Context, eventType string, payloadJSON string) (repo.Notification, error) {
	return s.repo.Create(ctx, repo.Notification{
		EventType:   eventType,
		PayloadJSON: payloadJSON,
		Status:      "pending",
	})
}

func (s *NotificationService) ListRecent(ctx context.Context, limit int) ([]repo.Notification, error) {
	return s.repo.ListRecent(ctx, limit)
}