package service

import (
	"context"
	"errors"
	"strings"
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
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return repo.Session{}, errors.New("title is required")
	}
	if len(in.Title) > 120 {
		return repo.Session{}, errors.New("title must be 120 characters or less")
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

func (s *SessionService) Delete(ctx context.Context, actorID, sessionID int64) error {
	if sessionID <= 0 {
		return errors.New("invalid session id")
	}
	if err := s.repo.Delete(ctx, sessionID); err != nil {
		return err
	}
	_ = s.auditSvc.Log(ctx, actorID, "delete_session", "session", sessionID, map[string]any{})
	return nil
}
