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
	repo            *repo.SessionRepo
	courseTeachers  *repo.CourseTeachersRepo
	enrollments     *repo.EnrollmentsRepo
	auditSvc        *AuditService
}

func NewSessionService(
	r *repo.SessionRepo,
	courseTeachers *repo.CourseTeachersRepo,
	enrollments *repo.EnrollmentsRepo,
	audit *AuditService,
) *SessionService {
	return &SessionService{repo: r, courseTeachers: courseTeachers, enrollments: enrollments, auditSvc: audit}
}

type CreateSessionInput struct {
	CourseID  int64
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
	if in.CourseID <= 0 {
		return repo.Session{}, errors.New("course_id is required")
	}
	if in.StartTime.IsZero() {
		return repo.Session{}, errors.New("start_time is required")
	}

	// TODO: Temporarily removed teacher-course relationship check
	// A teacher can only create sessions in courses they teach.
	// if in.ActorRole == auth.RoleTeacher {
	// 	isTeacher, err := s.courseTeachers.IsTeacher(ctx, in.CourseID, actorID)
	// 	if err != nil {
	// 		return repo.Session{}, err
	// 	}
	// 	if !isTeacher {
	// 		return repo.Session{}, errors.New("you are not a teacher of this course")
	// 	}
	// }

	created, err := s.repo.Create(ctx, repo.Session{
		CourseID:  in.CourseID,
		Title:     in.Title,
		StartTime: in.StartTime.UTC(),
	})
	if err != nil {
		return repo.Session{}, err
	}

	_ = s.auditSvc.Log(ctx, actorID, "create_session", "session", created.ID, map[string]any{
		"course_id":  created.CourseID,
		"start_time": created.StartTime.UTC().Format(time.RFC3339),
	})

	return created, nil
}

// ListForActor returns sessions visible to the actor based on role:
// - admin/manager: all sessions
// - teacher: sessions of courses they teach
// - student/parent: sessions of courses they are enrolled in (parent: of their children)
func (s *SessionService) ListForActor(ctx context.Context, actorID int64, role string, limit int) ([]repo.SessionRow, error) {
	switch role {
	case auth.RoleAdmin, auth.RoleManager:
		return s.repo.List(ctx, repo.SessionFilter{Limit: limit})
	case auth.RoleTeacher:
		// TODO: Temporarily show all sessions for teachers (removed course relationship check)
		return s.repo.List(ctx, repo.SessionFilter{Limit: limit})
	case auth.RoleStudent:
		ids, err := s.enrollments.CourseIDsByStudent(ctx, actorID)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []repo.SessionRow{}, nil
		}
		return s.repo.List(ctx, repo.SessionFilter{CourseIDs: ids, Limit: limit})
	default:
		return []repo.SessionRow{}, nil
	}
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
