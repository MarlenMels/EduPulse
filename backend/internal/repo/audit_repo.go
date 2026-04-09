package repo

import (
	"context"
	"database/sql"
	"time"
)

type AuditRepo struct{ db *sql.DB }

func NewAuditRepo(db *sql.DB) *AuditRepo { return &AuditRepo{db: db} }

func (r *AuditRepo) Create(ctx context.Context, a AuditLog) error {
	created := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, meta_json, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		a.ActorUserID, a.Action, a.EntityType, a.EntityID, a.MetaJSON, created,
	)
	return err
}

func (r *AuditRepo) ListRecent(ctx context.Context, limit int) ([]AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, actor_user_id, action, entity_type, entity_id, meta_json, created_at FROM audit_logs ORDER BY id DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AuditLog, 0, 16)
	for rows.Next() {
		var a AuditLog
		var created string
		if err := rows.Scan(&a.ID, &a.ActorUserID, &a.Action, &a.EntityType, &a.EntityID, &a.MetaJSON, &created); err != nil {
			return nil, err
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, a)
	}
	return out, rows.Err()
}