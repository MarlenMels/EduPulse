package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	var courseID int64
	if err := tx.QueryRowContext(ctx,
		"INSERT INTO courses (title, description, image_url, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		c.Title, c.Description, c.ImageURL, now,
	).Scan(&courseID); err != nil {
		return Course{}, nil, err
	}
	c.ID = courseID
	c.CreatedAt, _ = time.Parse(time.RFC3339, now)

	out := make([]Lesson, 0, len(lessons))
	for _, l := range lessons {
		l.CourseID = courseID
		var lid int64
		if err := tx.QueryRowContext(ctx,
			"INSERT INTO lessons (course_id, title, description, video_url, file_url, sort_order, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
			l.CourseID, l.Title, l.Description, l.VideoURL, l.FileURL, l.SortOrder, now,
		).Scan(&lid); err != nil {
			return Course{}, nil, err
		}
		l.ID = lid
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
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO lessons (course_id, title, description, video_url, file_url, sort_order, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		l.CourseID, l.Title, l.Description, l.VideoURL, l.FileURL, l.SortOrder, created,
	).Scan(&id)
	if err != nil {
		return Lesson{}, err
	}
	l.ID = id
	l.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return l, nil
}

func (r *CourseRepo) UpdateLesson(ctx context.Context, l Lesson) (Lesson, error) {
	_, err := r.db.ExecContext(ctx,
		"UPDATE lessons SET title = $1, description = $2, video_url = $3, file_url = $4, sort_order = $5 WHERE id = $6 AND course_id = $7",
		l.Title, l.Description, l.VideoURL, l.FileURL, l.SortOrder, l.ID, l.CourseID,
	)
	if err != nil {
		return Lesson{}, err
	}
	return l, nil
}

func (r *CourseRepo) List(ctx context.Context, limit int, search string) ([]Course, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	search = strings.TrimSpace(search)

	query := "SELECT id, title, description, image_url, created_at FROM courses"
	args := []any{}
	n := 0
	ph := func() string { n++; return fmt.Sprintf("$%d", n) }

	if search != "" {
		needle := "%" + strings.ToLower(search) + "%"
		query += " WHERE lower(title) LIKE " + ph() + " OR lower(description) LIKE " + ph()
		args = append(args, needle, needle)
	}
	query += " ORDER BY id DESC LIMIT " + ph()
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
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
		"SELECT id, title, description, image_url, created_at FROM courses WHERE id = $1", id,
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
		   FROM lessons WHERE course_id = $1 ORDER BY sort_order`,
		courseID,
	)
	if err != nil {
		return nil, err
	}
	out := make([]Lesson, 0, 16)
	for rows.Next() {
		var l Lesson
		var created string
		if err := rows.Scan(&l.ID, &l.CourseID, &l.Title, &l.Description, &l.VideoURL, &l.FileURL, &l.HLSUrl, &l.VideoStatus, &l.SortOrder, &created); err != nil {
			rows.Close()
			return nil, err
		}
		l.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, l)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}

	for i := range out {
		assets, err := r.AssetsByLesson(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Assets = assets
	}
	return out, nil
}

func (r *CourseRepo) CreateLessonAsset(ctx context.Context, asset LessonAsset) (LessonAsset, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := r.db.QueryRowContext(ctx,
		"INSERT INTO lesson_assets (lesson_id, type, url, original_filename, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		asset.LessonID, asset.Type, asset.URL, asset.OriginalFilename, created,
	).Scan(&id)
	if err != nil {
		return LessonAsset{}, err
	}
	asset.ID = id
	asset.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return asset, nil
}

func (r *CourseRepo) DeleteLessonAssetsByType(ctx context.Context, lessonID int64, assetType string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM lesson_assets WHERE lesson_id = $1 AND type = $2", lessonID, assetType)
	return err
}

func (r *CourseRepo) AssetsByLesson(ctx context.Context, lessonID int64) ([]LessonAsset, error) {
	rows, err := r.db.QueryContext(ctx,
		"SELECT id, lesson_id, type, url, original_filename, created_at FROM lesson_assets WHERE lesson_id = $1 ORDER BY id",
		lessonID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]LessonAsset, 0, 8)
	for rows.Next() {
		var asset LessonAsset
		var created string
		if err := rows.Scan(&asset.ID, &asset.LessonID, &asset.Type, &asset.URL, &asset.OriginalFilename, &created); err != nil {
			return nil, err
		}
		asset.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, asset)
	}
	return out, rows.Err()
}

func (r *CourseRepo) DeleteLessonAsset(ctx context.Context, lessonID, assetID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM lesson_assets WHERE id = $1 AND lesson_id = $2", assetID, lessonID)
	return err
}

func (r *CourseRepo) UpdateVideoStatus(ctx context.Context, lessonID int64, status, hlsURL string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE lessons SET hls_url = $1, video_status = $2 WHERE id = $3",
		hlsURL, status, lessonID,
	)
	return err
}

func (r *CourseRepo) ClearVideo(ctx context.Context, lessonID int64) error {
	if err := r.DeleteLessonAssetsByType(ctx, lessonID, "video"); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx,
		"UPDATE lessons SET video_url = '', hls_url = '', video_status = '' WHERE id = $1",
		lessonID,
	)
	return err
}

func (r *CourseRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM courses WHERE id = $1", id)
	return err
}

func (r *CourseRepo) DeleteLesson(ctx context.Context, courseID, lessonID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM lessons WHERE id = $1 AND course_id = $2", lessonID, courseID)
	return err
}
