package repo

import (
	"context"
	"database/sql"
)

type AnalyticsRepo struct{ db *sql.DB }

func NewAnalyticsRepo(db *sql.DB) *AnalyticsRepo { return &AnalyticsRepo{db: db} }

func (r *AnalyticsRepo) IncrementSessionsByH3Day(ctx context.Context, h3Index, day string, delta int) error {
	if delta == 0 {
		delta = 1
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO analytics_sessions_by_h3 (h3_index, day, sessions_count)
         VALUES (?, ?, ?)
         ON CONFLICT(h3_index, day) DO UPDATE SET sessions_count = sessions_count + excluded.sessions_count`,
		h3Index, day, delta,
	)
	return err
}

func (r *AnalyticsRepo) ListSessionsByH3(ctx context.Context, h3Index, day string, limit int) ([]AnalyticsRow, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	q := "SELECT h3_index, day, sessions_count FROM analytics_sessions_by_h3"
	args := make([]any, 0, 3)
	where := ""
	if h3Index != "" {
		where += "h3_index = ?"
		args = append(args, h3Index)
	}
	if day != "" {
		if where != "" {
			where += " AND "
		}
		where += "day = ?"
		args = append(args, day)
	}
	if where != "" {
		q += " WHERE " + where
	}
	q += " ORDER BY day DESC, sessions_count DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AnalyticsRow, 0, 32)
	for rows.Next() {
		var rrow AnalyticsRow
		if err := rows.Scan(&rrow.H3Index, &rrow.Day, &rrow.SessionsCount); err != nil {
			return nil, err
		}
		out = append(out, rrow)
	}
	return out, rows.Err()
}