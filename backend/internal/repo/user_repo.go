package repo

import (
	"context"
	"database/sql"
	"time"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role, created_at FROM users WHERE email = ?", email)
	var u User
	var created string
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, created)
	u.CreatedAt = t
	return &u, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role, created_at FROM users WHERE id = ?", id)
	var u User
	var created string
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	t, _ := time.Parse(time.RFC3339, created)
	u.CreatedAt = t
	return &u, nil
}

type RoleCount struct {
	Role   string `json:"role"`
	Total  int    `json:"total"`
	Online int    `json:"online"`
}

func (r *UserRepo) UpdateLastSeen(ctx context.Context, id int64) {
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = r.db.ExecContext(ctx, "UPDATE users SET last_seen_at = ? WHERE id = ?", now, id)
}

func (r *UserRepo) Stats(ctx context.Context) ([]RoleCount, error) {
	threshold := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx, `
		SELECT role,
		       COUNT(*) AS total,
		       SUM(CASE WHEN last_seen_at > ? THEN 1 ELSE 0 END) AS online
		FROM users
		GROUP BY role
		ORDER BY role
	`, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []RoleCount
	for rows.Next() {
		var rc RoleCount
		if err := rows.Scan(&rc.Role, &rc.Total, &rc.Online); err != nil {
			return nil, err
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}

func (r *UserRepo) Create(ctx context.Context, email, passwordHash, role string) (User, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
		email, passwordHash, role, now,
	)
	if err != nil {
		return User{}, err
	}
	id, _ := res.LastInsertId()
	t, _ := time.Parse(time.RFC3339, now)
	return User{ID: id, Email: email, Role: role, CreatedAt: t}, nil
}
