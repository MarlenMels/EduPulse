package repo

import (
	"context"
	"database/sql"
	"time"
)

type NotificationRepo struct{ db *sql.DB }

func NewNotificationRepo(db *sql.DB) *NotificationRepo { return &NotificationRepo{db: db} }

func (r *NotificationRepo) Create(ctx context.Context, n Notification) (Notification, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO notifications (event_type, payload_json, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		n.EventType, n.PayloadJSON, n.Status, created,
	).Scan(&id)
	if err != nil {
		return Notification{}, err
	}
	n.ID = id
	n.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return n, nil
}

func (r *NotificationRepo) ListRecent(ctx context.Context, limit int) ([]Notification, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, event_type, payload_json, status, created_at FROM notifications ORDER BY id DESC LIMIT $1",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Notification, 0, 16)
	for rows.Next() {
		var n Notification
		var created string
		if err := rows.Scan(&n.ID, &n.EventType, &n.PayloadJSON, &n.Status, &created); err != nil {
			return nil, err
		}
		n.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, n)
	}
	return out, rows.Err()
}
