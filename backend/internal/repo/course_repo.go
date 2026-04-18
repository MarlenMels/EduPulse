package repo

import (
	"context"
	"database/sql"
	"time"
)

type CourseRepo struct{ db *sql.DB }

func NewCourseRepo(db *sql.DB) *CourseRepo { return &CourseRepo{db: db} }

func (r *CourseRepo) CreateWithLessons(ctx context.Context, c Course, lessons []Lesson) (Course, []Lesson, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Course{}, nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)

	res, err := tx.ExecContext(ctx,
		"INSERT INTO courses (title, description, image_url, created_at) VALUES (?, ?, ?, ?)",
		c.Title, c.Description, c.ImageURL, now,
	)
	if err != nil {
		return Course{}, nil, err
	}
	courseID, _ := res.LastInsertId()
	c.ID = courseID
	c.CreatedAt, _ = time.Parse(time.RFC3339, now)

	out := make([]Lesson, 0, len(lessons))
	for _, l := range lessons {
		l.CourseID = courseID
		lRes, err := tx.ExecContext(ctx,
			"INSERT INTO lessons (course_id, title, description, video_url, file_url, sort_order, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			l.CourseID, l.Title, l.Description, l.VideoURL, l.FileURL, l.SortOrder, now,
		)
		if err != nil {
			return Course{}, nil, err
		}
		l.ID, _ = lRes.LastInsertId()
		l.CreatedAt, _ = time.Parse(time.RFC3339, now)
		out = append(out, l)
	}

	if err := tx.Commit(); err != nil {
		return Course{}, nil, err
	}
	return c, out, nil
}

func (r *CourseRepo) CreateLesson(ctx context.Context, l Lesson) (Lesson, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO lessons (course_id, title, description, video_url, file_url, sort_order, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		l.CourseID, l.Title, l.Description, l.VideoURL, l.FileURL, l.SortOrder, created,
	)
	if err != nil {
		return Lesson{}, err
	}
	l.ID, _ = res.LastInsertId()
	l.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return l, nil
}

func (r *CourseRepo) UpdateLesson(ctx context.Context, l Lesson) (Lesson, error) {
	_, err := r.db.ExecContext(ctx,
		"UPDATE lessons SET title = ?, description = ?, video_url = ?, file_url = ?, sort_order = ? WHERE id = ? AND course_id = ?",
		l.Title, l.Description, l.VideoURL, l.FileURL, l.SortOrder, l.ID, l.CourseID,
	)
	if err != nil {
		return Lesson{}, err
	}
	return l, nil
}

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
		`SELECT id, course_id, title, description, video_url, file_url, hls_url, video_status, sort_order, created_at
		   FROM lessons WHERE course_id = ? ORDER BY sort_order`,
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
		if err := rows.Scan(&l.ID, &l.CourseID, &l.Title, &l.Description, &l.VideoURL, &l.FileURL, &l.HLSUrl, &l.VideoStatus, &l.SortOrder, &created); err != nil {
			return nil, err
		}
		l.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *CourseRepo) UpdateVideoStatus(ctx context.Context, lessonID int64, status, hlsURL string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE lessons SET hls_url = ?, video_status = ? WHERE id = ?",
		hlsURL, status, lessonID,
	)
	return err
}
