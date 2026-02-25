package service

import (
	"context"
	"errors"
	"time"

	"edupulse/internal/auth"
	"edupulse/internal/repo"
)

type SessionService struct {
	repo         *repo.SessionRepo
	analytics    *repo.AnalyticsRepo
	auditSvc     *AuditService
	h3Resolution int
}

func NewSessionService(r *repo.SessionRepo, analytics *repo.AnalyticsRepo, audit *AuditService, h3Resolution int) *SessionService {
	if h3Resolution <= 0 {
		h3Resolution = 9
	}
	return &SessionService{
		repo:         r,
		analytics:    analytics,
		auditSvc:     audit,
		h3Resolution: h3Resolution,
	}
}

type CreateSessionInput struct {
	BranchID  int64
	TeacherID int64 // optional for manager; required for manager if not 0?
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
		// allow manager to set teacher_id explicitly; default to manager as pseudo-teacher for demo
		teacherID = actorID
	}

	h3idx, err := H3FromLatLng(in.Lat, in.Lng, s.h3Resolution)
	if err != nil {
		return repo.Session{}, err
	}

	sess := repo.Session{
		BranchID:  in.BranchID,
		TeacherID: teacherID,
		Title:     in.Title,
		StartTime: in.StartTime.UTC(),
		Lat:       in.Lat,
		Lng:       in.Lng,
		H3Index:   h3idx,
	}
	created, err := s.repo.Create(ctx, sess)
	if err != nil {
		return repo.Session{}, err
	}

	day := created.StartTime.UTC().Format("2006-01-02")
	_ = s.analytics.IncrementSessionsByH3Day(ctx, created.H3Index, day, 1)

	_ = s.auditSvc.Log(ctx, actorID, "create_session", "session", created.ID, map[string]any{
		"branch_id":  created.BranchID,
		"teacher_id": created.TeacherID,
		"h3":         created.H3Index,
		"start_time": created.StartTime.UTC().Format(time.RFC3339),
	})

	return created, nil
}

func (s *SessionService) List(ctx context.Context, h3Index string, limit int) ([]repo.Session, error) {
	return s.repo.List(ctx, h3Index, limit)
}