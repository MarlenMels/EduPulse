package repo

import (
	"context"
	"database/sql"
	"time"
)

type SessionRepo struct{ db *sql.DB }

func NewSessionRepo(db *sql.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s Session) (Session, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO sessions (teacher_id, title, start_time, created_at) VALUES (?, ?, ?, ?)",
		s.TeacherID, s.Title, s.StartTime.UTC().Format(time.RFC3339), created,
	)
	if err != nil {
		return Session{}, err
	}
	id, _ := res.LastInsertId()
	s.ID = id
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return s, nil
}

func (r *SessionRepo) List(ctx context.Context, limit int) ([]Session, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, teacher_id, title, start_time, created_at FROM sessions ORDER BY start_time DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Session, 0, 16)
	for rows.Next() {
		var s Session
		var start, created string
		if err := rows.Scan(&s.ID, &s.TeacherID, &s.Title, &start, &created); err != nil {
			return nil, err
		}
		s.StartTime, _ = time.Parse(time.RFC3339, start)
		s.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SessionRepo) GetByID(ctx context.Context, id int64) (*Session, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, teacher_id, title, start_time, created_at FROM sessions WHERE id = ?",
		id,
	)
	var s Session
	var start, created string
	if err := row.Scan(&s.ID, &s.TeacherID, &s.Title, &start, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.StartTime, _ = time.Parse(time.RFC3339, start)
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &s, nil
}
