package service

import (
	"context"
	"fmt"
	"time"

	"edupulse/internal/events"
	"edupulse/internal/repo"
)

type HomeworkManageService struct {
	hw    *repo.HomeworkRepo
	audit *AuditService
	bus   *events.Bus
}

func NewHomeworkManageService(hw *repo.HomeworkRepo, audit *AuditService, bus *events.Bus) *HomeworkManageService {
	return &HomeworkManageService{hw: hw, audit: audit, bus: bus}
}

func (s *HomeworkManageService) List(ctx context.Context, f repo.HomeworkListFilter) ([]repo.HomeworkSubmission, error) {
	return s.hw.List(ctx, f)
}

func (s *HomeworkManageService) ListMine(ctx context.Context, studentID int64, status string, limit int) ([]repo.HomeworkSubmission, error) {
	if studentID <= 0 {
		return nil, fmt.Errorf("invalid student id")
	}
	return s.hw.List(ctx, repo.HomeworkListFilter{
		StudentID: studentID,
		Status:    status,
		Limit:     limit,
	})
}

func (s *HomeworkManageService) UpdateStatus(ctx context.Context, actorTeacherID int64, submissionID int64, newStatus string) (*repo.HomeworkSubmission, error) {
	if submissionID <= 0 {
		return nil, fmt.Errorf("invalid id")
	}
	if newStatus != "accepted" && newStatus != "rejected" && newStatus != "needs_fix" {
		return nil, fmt.Errorf("status must be one of: accepted, rejected, needs_fix")
	}

	cur, err := s.hw.GetByID(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, fmt.Errorf("not found")
	}

	old := cur.Status
	if err := s.hw.UpdateStatus(ctx, submissionID, newStatus); err != nil {
		return nil, err
	}

	updated, _ := s.hw.GetByID(ctx, submissionID)

	// audit: grade_homework
	_ = s.audit.Log(ctx, actorTeacherID, "grade_homework", "homework_submission", submissionID, map[string]any{
		"session_id":  cur.SessionID,
		"student_id":  cur.StudentID,
		"old_status":  old,
		"new_status":  newStatus,
	})

	// event: homework_graded
	s.bus.Publish(ctx, events.Event{
		Type: events.EventHomeworkGraded,
		Payload: events.HomeworkGradedPayload{
			SubmissionID: submissionID,
			SessionID:    cur.SessionID,
			StudentID:    cur.StudentID,
			OldStatus:    old,
			NewStatus:    newStatus,
			GradedAt:     time.Now().UTC().Format(time.RFC3339),
		},
		CreatedAt: time.Now().UTC(),
	})

	return updated, nil
}