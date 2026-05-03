package repo

import (
	"context"
	"database/sql"
	"time"
)

type EnrollmentsRepo struct{ db *sql.DB }

func NewEnrollmentsRepo(db *sql.DB) *EnrollmentsRepo { return &EnrollmentsRepo{db: db} }

func (r *EnrollmentsRepo) Enroll(ctx context.Context, courseID, studentID, enrolledBy int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO course_enrollments (course_id, student_id, enrolled_by, enrolled_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (course_id, student_id) DO NOTHING`,
		courseID, studentID, enrolledBy, now,
	)
	return err
}

func (r *EnrollmentsRepo) Unenroll(ctx context.Context, courseID, studentID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM course_enrollments WHERE course_id = $1 AND student_id = $2",
		courseID, studentID,
	)
	return err
}

func (r *EnrollmentsRepo) IsEnrolled(ctx context.Context, courseID, studentID int64) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM course_enrollments WHERE course_id = $1 AND student_id = $2",
		courseID, studentID,
	).Scan(&n)
	return n > 0, err
}

// StudentsByCourse returns the user records of students enrolled in a course.
func (r *EnrollmentsRepo) StudentsByCourse(ctx context.Context, courseID int64) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.role, u.created_at
		   FROM users u
		   JOIN course_enrollments e ON e.student_id = u.id
		  WHERE e.course_id = $1
		  ORDER BY u.email`,
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]User, 0, 16)
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

// CourseIDsByStudent returns the IDs of courses a student is enrolled in.
func (r *EnrollmentsRepo) CourseIDsByStudent(ctx context.Context, studentID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT course_id FROM course_enrollments WHERE student_id = $1",
		studentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]int64, 0, 4)
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// StudentsByTeacher returns distinct students enrolled in any of the teacher's courses.
func (r *EnrollmentsRepo) StudentsByTeacher(ctx context.Context, teacherID int64) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT u.id, u.email, u.role, u.created_at
		   FROM users u
		   JOIN course_enrollments e ON e.student_id = u.id
		   JOIN course_teachers ct  ON ct.course_id = e.course_id
		  WHERE ct.teacher_id = $1
		  ORDER BY u.email`,
		teacherID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]User, 0, 16)
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
