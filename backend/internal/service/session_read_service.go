package service

import (
	"context"
	"fmt"

	"edupulse/internal/repo"
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
