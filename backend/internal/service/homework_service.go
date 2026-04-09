package service

import (
	"context"
	"errors"
	"time"

	"edupulse/internal/events"
	"edupulse/internal/repo"
)

type HomeworkService struct {
	repo     *repo.HomeworkRepo
	auditSvc *AuditService
	bus      *events.Bus
}

func NewHomeworkService(r *repo.HomeworkRepo, audit *AuditService, bus *events.Bus) *HomeworkService {
	return &HomeworkService{repo: r, auditSvc: audit, bus: bus}
}

type SubmitHomeworkInput struct {
	SessionID int64
	Content   string
}

func (s *HomeworkService) Submit(ctx context.Context, actorStudentID int64, in SubmitHomeworkInput) (repo.HomeworkSubmission, error) {
	if in.SessionID <= 0 {
		return repo.HomeworkSubmission{}, errors.New("session_id is required")
	}
	if in.Content == "" {
		return repo.HomeworkSubmission{}, errors.New("content is required")
	}

	sub, err := s.repo.Create(ctx, repo.HomeworkSubmission{
		SessionID: in.SessionID,
		StudentID: actorStudentID,
		Content:   in.Content,
		Status:    "submitted",
	})
	if err != nil {
		return repo.HomeworkSubmission{}, err
	}

	_ = s.auditSvc.Log(ctx, actorStudentID, "submit_homework", "homework_submission", sub.ID, map[string]any{
		"session_id": sub.SessionID,
		"status":     sub.Status,
		"at":         time.Now().UTC().Format(time.RFC3339),
	})

	payload := events.HomeworkSubmittedPayload{
		SubmissionID: sub.ID,
		SessionID:    sub.SessionID,
		StudentID:    sub.StudentID,
		Status:       sub.Status,
		CreatedAt:    sub.CreatedAt.UTC().Format(time.RFC3339),
	}
	s.bus.Publish(ctx, events.Event{
		Type:      events.EventHomeworkSubmitted,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	})

	return sub, nil
}