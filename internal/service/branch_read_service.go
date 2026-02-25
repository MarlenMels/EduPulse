package service

import (
	"context"
	"fmt"

	"edupulse/internal/repo"
)

type BranchReadService struct {
	branches *repo.BranchRepo
}

func NewBranchReadService(branches *repo.BranchRepo) *BranchReadService {
	return &BranchReadService{branches: branches}
}

func (s *BranchReadService) Get(ctx context.Context, id int64) (*repo.Branch, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	b, err := s.branches.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b == nil {
		return nil, fmt.Errorf("not found")
	}
	return b, nil
}

func (s *BranchReadService) List(ctx context.Context, h3Index, q string, limit int) ([]repo.Branch, error) {
	return s.branches.List(ctx, h3Index, q, limit)
}