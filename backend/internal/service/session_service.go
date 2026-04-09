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
	BranchID  int64
	TeacherID int64
	Title     string
	StartTime time.Time
	Lat       float64
	Lng       float64
	ActorRole string
}

func (s *SessionService) Create(ctx context.Context, actorID int64, in CreateSessionInput) (repo.Session, error) {
	if in.BranchID <= 0 {
		return repo.Session{}, errors.New("branch_id is required")
	}
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
		BranchID:  in.BranchID,
		TeacherID: teacherID,
		Title:     in.Title,
		StartTime: in.StartTime.UTC(),
		Lat:       in.Lat,
		Lng:       in.Lng,
	}
	created, err := s.repo.Create(ctx, sess)
	if err != nil {
		return repo.Session{}, err
	}

	_ = s.auditSvc.Log(ctx, actorID, "create_session", "session", created.ID, map[string]any{
		"branch_id":  created.BranchID,
		"teacher_id": created.TeacherID,
		"start_time": created.StartTime.UTC().Format(time.RFC3339),
	})

	return created, nil
}

func (s *SessionService) List(ctx context.Context, limit int) ([]repo.Session, error) {
	return s.repo.List(ctx, limit)
}
