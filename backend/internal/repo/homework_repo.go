package repo

import (
	"context"
	"database/sql"
	"time"
)

type HomeworkRepo struct{ db *sql.DB }

func NewHomeworkRepo(db *sql.DB) *HomeworkRepo { return &HomeworkRepo{db: db} }

func (r *HomeworkRepo) Create(ctx context.Context, h HomeworkSubmission) (HomeworkSubmission, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO homework_submissions (session_id, student_id, content, status, created_at) VALUES (?, ?, ?, ?, ?)",
		h.SessionID, h.StudentID, h.Content, h.Status, created,
	)
	if err != nil {
		return HomeworkSubmission{}, err
	}
	id, _ := res.LastInsertId()
	h.ID = id
	h.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return h, nil
}

func (r *HomeworkRepo) GetByID(ctx context.Context, id int64) (*HomeworkSubmission, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, session_id, student_id, content, status, created_at FROM homework_submissions WHERE id = ?",
		id,
	)
	var h HomeworkSubmission
	var created string
	if err := row.Scan(&h.ID, &h.SessionID, &h.StudentID, &h.Content, &h.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	h.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &h, nil
}

func (r *HomeworkRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE homework_submissions SET status = ? WHERE id = ?", status, id)
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

	q := "SELECT id, session_id, student_id, content, status, created_at FROM homework_submissions"
	args := make([]any, 0, 5)
	where := ""

	if f.SessionID > 0 {
		where += "session_id = ?"
		args = append(args, f.SessionID)
	}
	if f.StudentID > 0 {
		if where != "" {
			where += " AND "
		}
		where += "student_id = ?"
		args = append(args, f.StudentID)
	}
	if f.Status != "" {
		if where != "" {
			where += " AND "
		}
		where += "status = ?"
		args = append(args, f.Status)
	}
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY id DESC LIMIT ?"
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
		if err := rows.Scan(&h.ID, &h.SessionID, &h.StudentID, &h.Content, &h.Status, &created); err != nil {
			return nil, err
		}
		h.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, h)
	}
	return out, rows.Err()
}