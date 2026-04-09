package db

import (
	"database/sql"
	"time"

	"edupulse/internal/auth"
)

type seedUser struct {
	Email    string
	Password string
	Role     string
}

var seedUsers = []seedUser{
	{Email: "admin@edupulse.local", Password: "adminpass", Role: auth.RoleAdmin},
	{Email: "manager@edupulse.local", Password: "managerpass", Role: auth.RoleManager},
	{Email: "teacher@edupulse.local", Password: "teacherpass", Role: auth.RoleTeacher},
	{Email: "student@edupulse.local", Password: "studentpass", Role: auth.RoleStudent},
	{Email: "parent@edupulse.local", Password: "parentpass", Role: auth.RoleParent},
}

func Seed(db *sql.DB) error {
	ctx, cancel := contextWithTimeout(5 * time.Second)
	defer cancel()

	var cnt int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(1) FROM users").Scan(&cnt); err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}

	tx, err := db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	for _, u := range seedUsers {
		hash, err := auth.HashPassword(u.Password)
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx,
			"INSERT INTO users (email, password_hash, role, created_at) VALUES (?, ?, ?, ?)",
			u.Email, hash, u.Role, now,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}