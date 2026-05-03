package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"edupulse/internal/events"
	"edupulse/internal/repo"
)

type HomeworkService struct {
	repo        *repo.HomeworkRepo
	assignments *repo.AssignmentRepo
	sessions    *repo.SessionRepo
	enrollments *repo.EnrollmentsRepo
	auditSvc    *AuditService
	bus         *events.Bus
}

func NewHomeworkService(
	r *repo.HomeworkRepo,
	assignments *repo.AssignmentRepo,
	sessions *repo.SessionRepo,
	enrollments *repo.EnrollmentsRepo,
	audit *AuditService,
	bus *events.Bus,
) *HomeworkService {
	return &HomeworkService{
		repo:        r,
		assignments: assignments,
		sessions:    sessions,
		enrollments: enrollments,
		auditSvc:    audit,
		bus:         bus,
	}
}

type SubmitHomeworkInput struct {
	AssignmentID int64
	Content      string
	Attachments  string
}

func (s *HomeworkService) Submit(ctx context.Context, studentID int64, in SubmitHomeworkInput) (repo.HomeworkSubmission, error) {
	if in.AssignmentID <= 0 {
		return repo.HomeworkSubmission{}, errors.New("assignment_id is required")
	}
	in.Content = strings.TrimSpace(in.Content)
	if in.Content == "" {
		return repo.HomeworkSubmission{}, errors.New("content is required")
	}
	if len(in.Content) > 5000 {
		return repo.HomeworkSubmission{}, errors.New("content must be at most 5000 characters")
	}

	a, err := s.assignments.GetByID(ctx, in.AssignmentID)
	if err != nil {
		return repo.HomeworkSubmission{}, err
	}
	if a == nil {
		return repo.HomeworkSubmission{}, errors.New("assignment not found")
	}
	sess, err := s.sessions.GetByID(ctx, a.SessionID)
	if err != nil {
		return repo.HomeworkSubmission{}, err
	}
	if sess == nil {
		return repo.HomeworkSubmission{}, errors.New("session not found")
	}

	// The student must be enrolled in the course this assignment belongs to.
	enrolled, err := s.enrollments.IsEnrolled(ctx, sess.CourseID, studentID)
	if err != nil {
		return repo.HomeworkSubmission{}, err
	}
	if !enrolled {
		return repo.HomeworkSubmission{}, errors.New("you are not enrolled in this course")
	}

	// Disallow duplicates so each student has exactly one submission per assignment.
	already, err := s.repo.HasSubmission(ctx, in.AssignmentID, studentID)
	if err != nil {
		return repo.HomeworkSubmission{}, err
	}
	if already {
		return repo.HomeworkSubmission{}, errors.New("you already submitted this assignment")
	}

	sub, err := s.repo.Create(ctx, repo.HomeworkSubmission{
		AssignmentID: in.AssignmentID,
		StudentID:    studentID,
		Content:      in.Content,
		Attachments:  in.Attachments,
		Status:       "submitted",
	})
	if err != nil {
		return repo.HomeworkSubmission{}, err
	}

	_ = s.auditSvc.Log(ctx, studentID, "submit_homework", "homework_submission", sub.ID, map[string]any{
		"assignment_id": sub.AssignmentID,
		"status":        sub.Status,
		"at":            time.Now().UTC().Format(time.RFC3339),
	})

	if s.bus != nil {
		payload := events.HomeworkSubmittedPayload{
			SubmissionID: sub.ID,
			SessionID:    sess.ID,
			StudentID:    sub.StudentID,
			Status:       sub.Status,
			CreatedAt:    sub.CreatedAt.UTC().Format(time.RFC3339),
		}
		s.bus.Publish(ctx, events.Event{
			Type:      events.EventHomeworkSubmitted,
			Payload:   payload,
			CreatedAt: time.Now().UTC(),
		})
	}

	return sub, nil
}

func (s *HomeworkService) Mine(ctx context.Context, studentID int64, limit int) ([]repo.MineRow, error) {
	return s.repo.MineByStudent(ctx, studentID, limit)
}
