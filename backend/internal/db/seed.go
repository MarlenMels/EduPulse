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
			"INSERT INTO users (email, password_hash, role, created_at) VALUES ($1, $2, $3, $4)",
			u.Email, hash, u.Role, now,
		)
		if err != nil {
			return err
		}
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO parent_students (parent_id, student_id, created_at)
		SELECT p.id, s.id, $1
		  FROM users p, users s
		 WHERE p.email = $2 AND s.email = $3
		ON CONFLICT (parent_id, student_id) DO NOTHING
	`, now, "parent@edupulse.local", "student@edupulse.local")
	if err != nil {
		return err
	}

	return tx.Commit()
}
