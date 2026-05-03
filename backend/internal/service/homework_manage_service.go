package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"edupulse/internal/auth"
	"edupulse/internal/events"
	"edupulse/internal/repo"
)

type HomeworkManageService struct {
	hw          *repo.HomeworkRepo
	assignments *repo.AssignmentRepo
	sessions    *repo.SessionRepo
	teachers    *repo.CourseTeachersRepo
	audit       *AuditService
	bus         *events.Bus
}

func NewHomeworkManageService(
	hw *repo.HomeworkRepo,
	assignments *repo.AssignmentRepo,
	sessions *repo.SessionRepo,
	teachers *repo.CourseTeachersRepo,
	audit *AuditService,
	bus *events.Bus,
) *HomeworkManageService {
	return &HomeworkManageService{
		hw:          hw,
		assignments: assignments,
		sessions:    sessions,
		teachers:    teachers,
		audit:       audit,
		bus:         bus,
	}
}

func (s *HomeworkManageService) Submissions(ctx context.Context, assignmentID int64) ([]repo.SubmissionRow, error) {
	return s.assignments.Submissions(ctx, assignmentID)
}

// ListAll returns all homework submissions (admin/manager only)
func (s *HomeworkManageService) ListAll(ctx context.Context, limit int) ([]repo.HomeworkSubmission, error) {
	return s.hw.List(ctx, repo.HomeworkFilter{Limit: limit})
}

// ListByTeacher returns homework submissions for teacher's courses
// TODO: Implement proper filtering by teacher's courses
func (s *HomeworkManageService) ListByTeacher(ctx context.Context, teacherID int64, limit int) ([]repo.HomeworkSubmission, error) {
	// For now, return all submissions (temporarily simplified)
	return s.hw.List(ctx, repo.HomeworkFilter{Limit: limit})
}

// UpdateStatus changes a submission's status. Teachers can only grade submissions
// for assignments belonging to courses they teach (or assignments they created).
// Admin/manager can grade anything.
func (s *HomeworkManageService) UpdateStatus(
	ctx context.Context,
	actorID int64,
	actorRole string,
	submissionID int64,
	newStatus string,
) (*repo.HomeworkSubmission, error) {
	if submissionID <= 0 {
		return nil, errors.New("invalid id")
	}
	if newStatus != "accepted" && newStatus != "rejected" && newStatus != "needs_fix" {
		return nil, errors.New("status must be one of: accepted, rejected, needs_fix")
	}

	cur, err := s.hw.GetByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, errors.New("submission not found")
	}

	if actorRole == auth.RoleTeacher {
		ok, err := s.canTeacherGrade(ctx, actorID, cur.AssignmentID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("you can only grade homework in courses you teach")
		}
	}

	old := cur.Status
	if err := s.hw.UpdateStatus(ctx, submissionID, newStatus); err != nil {
		return nil, err
	}
	updated, _ := s.hw.GetByID(ctx, submissionID)

	// Look up the session for audit/event payload.
	a, _ := s.assignments.GetByID(ctx, cur.AssignmentID)
	var sessionID int64
	if a != nil {
		sessionID = a.SessionID
	}

	_ = s.audit.Log(ctx, actorID, "grade_homework", "homework_submission", submissionID, map[string]any{
		"assignment_id": cur.AssignmentID,
		"student_id":    cur.StudentID,
		"old_status":    old,
		"new_status":    newStatus,
	})

	if s.bus != nil {
		s.bus.Publish(ctx, events.Event{
			Type: events.EventHomeworkGraded,
			Payload: events.HomeworkGradedPayload{
				SubmissionID: submissionID,
				SessionID:    sessionID,
				StudentID:    cur.StudentID,
				OldStatus:    old,
				NewStatus:    newStatus,
				GradedAt:     time.Now().UTC().Format(time.RFC3339),
			},
			CreatedAt: time.Now().UTC(),
		})
	}

	return updated, nil
}

func (s *HomeworkManageService) canTeacherGrade(ctx context.Context, teacherID, assignmentID int64) (bool, error) {
	a, err := s.assignments.GetByID(ctx, assignmentID)
	if err != nil {
		return false, err
	}
	if a == nil {
		return false, fmt.Errorf("assignment %d not found", assignmentID)
	}
	if a.CreatedBy == teacherID {
		return true, nil
	}
	sess, err := s.sessions.GetByID(ctx, a.SessionID)
	if err != nil {
		return false, err
	}
	if sess == nil {
		return false, nil
	}
	return s.teachers.IsTeacher(ctx, sess.CourseID, teacherID)
}
