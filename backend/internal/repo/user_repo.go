package repo

import (
	"context"
	"database/sql"
	"time"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role, created_at FROM users WHERE email = $1", email)
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
	row := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role, created_at FROM users WHERE id = $1", id)
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

func (r *UserRepo) FirstByRole(ctx context.Context, role string) (*User, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, email, password_hash, role, created_at FROM users WHERE role = $1 ORDER BY id ASC LIMIT 1", role)
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
	_, _ = r.db.ExecContext(ctx, "UPDATE users SET last_seen_at = $1 WHERE id = $2", now, id)
}

func (r *UserRepo) Stats(ctx context.Context) ([]RoleCount, error) {
	threshold := time.Now().UTC().Add(-5 * time.Minute).Format(time.RFC3339)
	rows, err := r.db.QueryContext(ctx, `
		SELECT role,
		       COUNT(*) AS total,
		       SUM(CASE WHEN last_seen_at > $1 THEN 1 ELSE 0 END) AS online
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
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO users (email, password_hash, role, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		email, passwordHash, role, now,
	).Scan(&id)
	if err != nil {
		return User{}, err
	}
	t, _ := time.Parse(time.RFC3339, now)
	return User{ID: id, Email: email, Role: role, CreatedAt: t}, nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET password_hash = $1 WHERE id = $2", passwordHash, id)
	return err
}

func (r *UserRepo) List(ctx context.Context, limit int) ([]User, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, email, role, created_at FROM users ORDER BY id DESC LIMIT $1",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]User, 0, 32)
	for rows.Next() {
		var u User
		var created string
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &created); err != nil {
			return nil, err
		}
		u.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, u)
	}
	return out, rows.Err()
}

func (r *UserRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
