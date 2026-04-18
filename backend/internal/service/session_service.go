package service

import (
	"context"
	"errors"
	"time"

	"edupulse/internal/auth"
	"edupulse/internal/repo"
)

type SessionService struct {
	repo     *repo.SessionRepo
	auditSvc *AuditService
}

func NewSessionService(r *repo.SessionRepo, audit *AuditService) *SessionService {
	return &SessionService{repo: r, auditSvc: audit}
}

type CreateSessionInput struct {
	TeacherID int64
	Title     string
	StartTime time.Time
	ActorRole string
}

func (s *SessionService) Create(ctx context.Context, actorID int64, in CreateSessionInput) (repo.Session, error) {
	if in.Title == "" {
		return repo.Session{}, errors.New("title is required")
	}
	if in.StartTime.IsZero() {
		return repo.Session{}, errors.New("start_time is required")
	}

	teacherID := in.TeacherID
	if in.ActorRole == auth.RoleTeacher {
		teacherID = actorID
	}
	if teacherID <= 0 {
		teacherID = actorID
	}

	sess := repo.Session{
		TeacherID: teacherID,
		Title:     in.Title,
		StartTime: in.StartTime.UTC(),
	}
	created, err := s.repo.Create(ctx, sess)
	if err != nil {
		return repo.Session{}, err
	}

	_ = s.auditSvc.Log(ctx, actorID, "create_session", "session", created.ID, map[string]any{
		"teacher_id": created.TeacherID,
		"start_time": created.StartTime.UTC().Format(time.RFC3339),
	})

	return created, nil
}

func (s *SessionService) List(ctx context.Context, limit int) ([]repo.Session, error) {
	return s.repo.List(ctx, limit)
}
