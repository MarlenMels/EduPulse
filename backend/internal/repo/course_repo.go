package repo

import (
	"context"
	"database/sql"
	"time"
)

type CourseRepo struct{ db *sql.DB }

func NewCourseRepo(db *sql.DB) *CourseRepo { return &CourseRepo{db: db} }

func (r *CourseRepo) List(ctx context.Context, limit int) ([]Course, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, title, description, image_url, created_at FROM courses ORDER BY id DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Course, 0, 16)
	for rows.Next() {
		var c Course
		var created string
		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.ImageURL, &created); err != nil {
			return nil, err
		}
		c.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CourseRepo) GetByID(ctx context.Context, id int64) (*Course, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, title, description, image_url, created_at FROM courses WHERE id = ?", id,
	)
	var c Course
	var created string
	if err := row.Scan(&c.ID, &c.Title, &c.Description, &c.ImageURL, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &c, nil
}

func (r *CourseRepo) LessonsByCourse(ctx context.Context, courseID int64) ([]Lesson, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, course_id, title, video_url, sort_order, created_at FROM lessons WHERE course_id = ? ORDER BY sort_order",
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Lesson, 0, 16)
	for rows.Next() {
		var l Lesson
		var created string
		if err := rows.Scan(&l.ID, &l.CourseID, &l.Title, &l.VideoURL, &l.SortOrder, &created); err != nil {
			return nil, err
		}
		l.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *CourseRepo) ReviewsByCourse(ctx context.Context, courseID int64) ([]Review, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, course_id, user_id, rating, comment, created_at FROM reviews WHERE course_id = ? ORDER BY id DESC",
		courseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Review, 0, 16)
	for rows.Next() {
		var rv Review
		var created string
		if err := rows.Scan(&rv.ID, &rv.CourseID, &rv.UserID, &rv.Rating, &rv.Comment, &created); err != nil {
			return nil, err
		}
		rv.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, rv)
	}
	return out, rows.Err()
}
