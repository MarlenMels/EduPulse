package repo

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type HomeworkRepo struct{ db *sql.DB }

func NewHomeworkRepo(db *sql.DB) *HomeworkRepo { return &HomeworkRepo{db: db} }

func (r *HomeworkRepo) Create(ctx context.Context, h HomeworkSubmission) (HomeworkSubmission, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO homework_submissions (session_id, student_id, content, attachments, status, created_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id",
		h.SessionID, h.StudentID, h.Content, h.Attachments, h.Status, created,
	).Scan(&id)
	if err != nil {
		return HomeworkSubmission{}, err
	}
	h.ID = id
	h.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return h, nil
}

func (r *HomeworkRepo) GetByID(ctx context.Context, id int64) (*HomeworkSubmission, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, session_id, student_id, content, attachments, status, created_at FROM homework_submissions WHERE id = $1",
		id,
	)
	var h HomeworkSubmission
	var created string
	if err := row.Scan(&h.ID, &h.SessionID, &h.StudentID, &h.Content, &h.Attachments, &h.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	h.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &h, nil
}

func (r *HomeworkRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE homework_submissions SET status = $1 WHERE id = $2", status, id)
	return err
}

type HomeworkListFilter struct {
	SessionID int64
	StudentID int64
	Status    string
	Limit     int
}

func (r *HomeworkRepo) List(ctx context.Context, f HomeworkListFilter) ([]HomeworkSubmission, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := "SELECT id, session_id, student_id, content, attachments, status, created_at FROM homework_submissions"
	args := make([]any, 0, 5)
	where := ""
	n := 0
	ph := func() string { n++; return fmt.Sprintf("$%d", n) }

	if f.SessionID > 0 {
		where += "session_id = " + ph()
		args = append(args, f.SessionID)
	}
	if f.StudentID > 0 {
		if where != "" {
			where += " AND "
		}
		where += "student_id = " + ph()
		args = append(args, f.StudentID)
	}
	if f.Status != "" {
		if where != "" {
			where += " AND "
		}
		where += "status = " + ph()
		args = append(args, f.Status)
	}
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY id DESC LIMIT " + ph()
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]HomeworkSubmission, 0, 64)
	for rows.Next() {
		var h HomeworkSubmission
		var created string
		if err := rows.Scan(&h.ID, &h.SessionID, &h.StudentID, &h.Content, &h.Attachments, &h.Status, &created); err != nil {
			return nil, err
		}
		h.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, h)
	}
	return out, rows.Err()
}

func (r *HomeworkRepo) ListForParent(ctx context.Context, parentID int64, status string, limit int) ([]HomeworkSubmission, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	q := `
		SELECT h.id, h.session_id, h.student_id, h.content, h.attachments, h.status, h.created_at
		  FROM homework_submissions h
		  JOIN parent_students ps ON ps.student_id = h.student_id
		 WHERE ps.parent_id = $1`
	args := []any{parentID}
	n := 1
	ph := func() string { n++; return fmt.Sprintf("$%d", n) }

	if status != "" {
		q += " AND h.status = " + ph()
		args = append(args, status)
	}
	q += " ORDER BY h.id DESC LIMIT " + ph()
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]HomeworkSubmission, 0, 64)
	for rows.Next() {
		var h HomeworkSubmission
		var created string
		if err := rows.Scan(&h.ID, &h.SessionID, &h.StudentID, &h.Content, &h.Attachments, &h.Status, &created); err != nil {
			return nil, err
		}
		h.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, h)
	}
	return out, rows.Err()
}
