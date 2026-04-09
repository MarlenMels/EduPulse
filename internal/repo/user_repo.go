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
