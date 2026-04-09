package service

import (
	"context"
	"errors"

	"edupulse/internal/repo"
)

type BranchService struct {
	repo     *repo.BranchRepo
	auditSvc *AuditService
}

func NewBranchService(r *repo.BranchRepo, audit *AuditService) *BranchService {
	return &BranchService{repo: r, auditSvc: audit}
}

func (s *BranchService) Create(ctx context.Context, actorID int64, name string, lat, lng float64) (repo.Branch, error) {
	if name == "" {
		return repo.Branch{}, errors.New("name is required")
	}
	b := repo.Branch{Name: name, Lat: lat, Lng: lng}
	created, err := s.repo.Create(ctx, b)
	if err != nil {
		return repo.Branch{}, err
	}
	_ = s.auditSvc.Log(ctx, actorID, "create_branch", "branch", created.ID, map[string]any{
		"name": name,
	})
	return created, nil
}
