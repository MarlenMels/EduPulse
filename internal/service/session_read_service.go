package service

import (
	"context"
	"fmt"

	"edupulse/internal/repo"

	"github.com/ziprecruiter/h3-go/pkg/h3"
)

type SessionReadService struct {
	sessions *repo.SessionRepo
}

func NewSessionReadService(sessions *repo.SessionRepo) *SessionReadService {
	return &SessionReadService{sessions: sessions}
}

func (s *SessionReadService) Get(ctx context.Context, id int64) (*repo.Session, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	ss, err := s.sessions.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ss == nil {
		return nil, fmt.Errorf("not found")
	}
	return ss, nil
}

func (s *SessionReadService) NearbyByH3(ctx context.Context, centerH3 string, k int, limit int) ([]repo.Session, error) {
	if centerH3 == "" {
		return nil, fmt.Errorf("h3 is required")
	}
	if k < 0 || k > 3 {
		k = 1
	}

	cell, err := h3.NewCellFromString(centerH3)
	if err != nil || !cell.Valid() {
		return nil, fmt.Errorf("invalid h3")
	}

	disk, err := cell.GridDisk(k)
	if err != nil {
		return nil, err
	}

	uniq := make(map[string]struct{}, len(disk))
	set := make([]string, 0, len(disk))
	for _, c := range disk {
		if c == 0 {
			continue
		}
		h := c.String()
		if _, ok := uniq[h]; ok {
			continue
		}
		uniq[h] = struct{}{}
		set = append(set, h)
	}

	return s.sessions.ListByH3Set(ctx, set, limit)
}