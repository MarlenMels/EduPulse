package service

import (
	"context"
	"errors"

	"edupulse/internal/repo"
)

type BranchService struct {
	repo     *repo.BranchRepo
	auditSvc *AuditService
	h3Res    int
}

func NewBranchService(r *repo.BranchRepo, audit *AuditService) *BranchService {
	return &BranchService{repo: r, auditSvc: audit, h3Res: 9}
}

func (s *BranchService) Create(ctx context.Context, actorID int64, name string, lat, lng float64) (repo.Branch, error) {
	if name == "" {
		return repo.Branch{}, errors.New("name is required")
	}
	h3idx, err := H3FromLatLng(lat, lng, s.h3Res)
	if err != nil {
		return repo.Branch{}, err
	}
	b := repo.Branch{Name: name, Lat: lat, Lng: lng, H3Index: h3idx}
	created, err := s.repo.Create(ctx, b)
	if err != nil {
		return repo.Branch{}, err
	}
	_ = s.auditSvc.Log(ctx, actorID, "create_branch", "branch", created.ID, map[string]any{
		"name": name,
		"h3":   h3idx,
	})
	return created, nil
}