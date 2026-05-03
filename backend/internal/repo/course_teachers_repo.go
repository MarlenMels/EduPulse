package repo

import (
	"context"
	"database/sql"
	"time"
)

type CourseTeachersRepo struct{ db *sql.DB }

func NewCourseTeachersRepo(db *sql.DB) *CourseTeachersRepo { return &CourseTeachersRepo{db: db} }

func (r *CourseTeachersRepo) Add(ctx context.Context, courseID, teacherID int64) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO course_teachers (course_id, teacher_id, created_at)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (course_id, teacher_id) DO NOTHING`,
		courseID, teacherID, now,
	)
	return err
}

func (r *CourseTeachersRepo) Remove(ctx context.Context, courseID, teacherID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM course_teachers WHERE course_id = $1 AND teacher_id = $2",
		courseID, teacherID,
	)
	return err
}

func (r *CourseTeachersRepo) IsTeacher(ctx context.Context, courseID, teacherID int64) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM course_teachers WHERE course_id = $1 AND teacher_id = $2",
		courseID, teacherID,
	).Scan(&n)
	return n > 0, err
}

// TeachersByCourse returns the user records of teachers assigned to a course.
func (r *CourseTeachersRepo) TeachersByCourse(ctx context.Context, courseID int64) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT u.id, u.email, u.role, u.created_at
		   FROM users u
		   JOIN course_teachers ct ON ct.teacher_id = u.id
		  WHERE ct.course_id = $1
		  ORDER BY u.email`,
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]User, 0, 4)
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

// CourseIDsByTeacher returns the IDs of courses a teacher is assigned to.
func (r *CourseTeachersRepo) CourseIDsByTeacher(ctx context.Context, teacherID int64) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT course_id FROM course_teachers WHERE teacher_id = $1",
		teacherID,
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
