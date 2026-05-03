package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type SessionRepo struct{ db *sql.DB }

func NewSessionRepo(db *sql.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s Session) (Session, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO sessions (course_id, title, start_time, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		s.CourseID, s.Title, s.StartTime.UTC().Format(time.RFC3339), created,
	).Scan(&id)
	if err != nil {
		return Session{}, err
	}
	s.ID = id
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return s, nil
}

// SessionRow extends Session with denormalized course title for list views.
type SessionRow struct {
	Session
	CourseTitle string `json:"course_title"`
}

type SessionFilter struct {
	// CourseIDs limits to sessions whose course is in this set. Empty means no filter.
	CourseIDs []int64
	Limit     int
}

func (r *SessionRepo) List(ctx context.Context, f SessionFilter) ([]SessionRow, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	conds := []string{}
	args := []any{}
	n := 0
	ph := func() string { n++; return fmt.Sprintf("$%d", n) }

	if len(f.CourseIDs) > 0 {
		placeholders := make([]string, len(f.CourseIDs))
		for i, id := range f.CourseIDs {
			placeholders[i] = ph()
			args = append(args, id)
		}
		conds = append(conds, "s.course_id IN ("+strings.Join(placeholders, ",")+")")
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	args = append(args, limit)
	limitPH := ph()

	rows, err := r.db.QueryContext(ctx, `
		SELECT s.id, s.course_id, s.title, s.start_time, s.created_at, c.title
		  FROM sessions s
		  JOIN courses c ON c.id = s.course_id
		  `+where+`
		 ORDER BY s.start_time DESC
		 LIMIT `+limitPH,
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]SessionRow, 0, 16)
	for rows.Next() {
		var s SessionRow
		var start, created string
		if err := rows.Scan(&s.ID, &s.CourseID, &s.Title, &start, &created, &s.CourseTitle); err != nil {
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
		"SELECT id, course_id, title, start_time, created_at FROM sessions WHERE id = $1",
		id,
	)
	var s Session
	var start, created string
	if err := row.Scan(&s.ID, &s.CourseID, &s.Title, &start, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.StartTime, _ = time.Parse(time.RFC3339, start)
	s.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &s, nil
}

func (r *SessionRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM sessions WHERE id = $1", id)
	return err
}
