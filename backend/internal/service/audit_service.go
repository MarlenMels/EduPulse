package service

import (
	"context"
	"encoding/json"

	"edupulse/internal/repo"
)

type AuditService struct{ repo *repo.AuditRepo }

func NewAuditService(r *repo.AuditRepo) *AuditService { return &AuditService{repo: r} }

func (s *AuditService) Log(ctx context.Context, actorID int64, action, entityType string, entityID int64, meta any) error {
	b, err := json.Marshal(meta)
	if err != nil {
		b = []byte(`{"marshal_error":true}`)
	}
	return s.repo.Create(ctx, repo.AuditLog{
		ActorUserID: actorID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		MetaJSON:    string(b),
	})
}

func (s *AuditService) ListRecent(ctx context.Context, limit int) ([]repo.AuditLog, error) {
	return s.repo.ListRecent(ctx, limit)
}