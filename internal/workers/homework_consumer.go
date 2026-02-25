package workers

import (
	"context"
	"encoding/json"

	"edupulse/internal/events"
	"edupulse/internal/service"
)

type HomeworkConsumer struct {
	notif *service.NotificationService
	audit *service.AuditService
}

func NewHomeworkConsumer(notif *service.NotificationService, audit *service.AuditService) *HomeworkConsumer {
	return &HomeworkConsumer{notif: notif, audit: audit}
}

func (c *HomeworkConsumer) Handle(ctx context.Context, e events.Event) error {
	b, err := json.Marshal(e.Payload)
	if err != nil {
		b = []byte(`{"marshal_error":true}`)
	}

	_, _ = c.notif.Create(ctx, e.Type, string(b))

	// actor_user_id = 0 => system
	_ = c.audit.Log(ctx, 0, "event_consumed", "event", 0, map[string]any{
		"type":    e.Type,
		"payload": json.RawMessage(b),
	})
	return nil
}