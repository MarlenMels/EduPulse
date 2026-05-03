package service

import (
	"context"
	"errors"
	"strings"

	"edupulse/internal/auth"
	"edupulse/internal/repo"
)

type AssignmentService struct {
	assignments *repo.AssignmentRepo
	sessions    *repo.SessionRepo
	teachers    *repo.CourseTeachersRepo
	enrollments *repo.EnrollmentsRepo
	audit       *AuditService
}

func NewAssignmentService(
	a *repo.AssignmentRepo,
	sessions *repo.SessionRepo,
	teachers *repo.CourseTeachersRepo,
	enrollments *repo.EnrollmentsRepo,
	audit *AuditService,
) *AssignmentService {
	return &AssignmentService{
		assignments: a,
		sessions:    sessions,
		teachers:    teachers,
		enrollments: enrollments,
		audit:       audit,
	}
}

type CreateAssignmentInput struct {
	SessionID   int64
	Title       string
	Description string
	ActorRole   string
}

func (s *AssignmentService) Create(ctx context.Context, actorID int64, in CreateAssignmentInput) (repo.Assignment, error) {
	if in.SessionID <= 0 {
		return repo.Assignment{}, errors.New("session_id is required")
	}
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return repo.Assignment{}, errors.New("title is required")
	}
	if len(in.Title) > 200 {
		return repo.Assignment{}, errors.New("title must be at most 200 characters")
	}
	in.Description = strings.TrimSpace(in.Description)
	if len(in.Description) > 5000 {
		return repo.Assignment{}, errors.New("description must be at most 5000 characters")
	}

	sess, err := s.sessions.GetByID(ctx, in.SessionID)
	if err != nil {
		return repo.Assignment{}, err
	}
	if sess == nil {
		return repo.Assignment{}, errors.New("session not found")
	}

	// A teacher can only attach assignments to sessions in courses they teach.
	if in.ActorRole == auth.RoleTeacher {
		isTeacher, err := s.teachers.IsTeacher(ctx, sess.CourseID, actorID)
		if err != nil {
			return repo.Assignment{}, err
		}
		if !isTeacher {
			return repo.Assignment{}, errors.New("you are not a teacher of this course")
		}
	}

	a, err := s.assignments.Create(ctx, repo.Assignment{
		SessionID:   in.SessionID,
		CreatedBy:   actorID,
		Title:       in.Title,
		Description: in.Description,
	})
	if err != nil {
		return repo.Assignment{}, err
	}

	_ = s.audit.Log(ctx, actorID, "create_assignment", "assignment", a.ID, map[string]any{
		"session_id": a.SessionID,
		"course_id":  sess.CourseID,
	})

	return a, nil
}

// ListForActor returns assignments the actor can see based on their role.
// - admin/manager: all
// - teacher: assignments they created OR in courses they teach
// - student: assignments in courses they are enrolled in
func (s *AssignmentService) ListForActor(ctx context.Context, actorID int64, role string, limit int) ([]repo.AssignmentRow, error) {
	switch role {
	case auth.RoleAdmin, auth.RoleManager:
		return s.assignments.List(ctx, repo.AssignmentFilter{Limit: limit})
	case auth.RoleTeacher:
		return s.assignments.List(ctx, repo.AssignmentFilter{TeacherID: actorID, Limit: limit})
	case auth.RoleStudent:
		ids, err := s.enrollments.CourseIDsByStudent(ctx, actorID)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			return []repo.AssignmentRow{}, nil
		}
		return s.assignments.List(ctx, repo.AssignmentFilter{CourseIDs: ids, Limit: limit})
	default:
		return []repo.AssignmentRow{}, nil
	}
}

func (s *AssignmentService) GetByID(ctx context.Context, id int64) (*repo.Assignment, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.assignments.GetByID(ctx, id)
}

func (s *AssignmentService) Delete(ctx context.Context, actorID int64, role string, id int64) error {
	a, err := s.assignments.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if a == nil {
		return errors.New("assignment not found")
	}
	if role == auth.RoleTeacher && a.CreatedBy != actorID {
		return errors.New("you can only delete assignments you created")
	}
	if err := s.assignments.Delete(ctx, id); err != nil {
		return err
	}
	_ = s.audit.Log(ctx, actorID, "delete_assignment", "assignment", id, map[string]any{})
	return nil
}
