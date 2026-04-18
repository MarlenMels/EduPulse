package repo

import (
	"context"
	"database/sql"
	"time"
)

type VideoRepo struct{ db *sql.DB }

func NewVideoRepo(db *sql.DB) *VideoRepo { return &VideoRepo{db: db} }

func (r *VideoRepo) Create(ctx context.Context, u VideoUpload) (VideoUpload, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO video_uploads
			(lesson_id, original_filename, stored_path, hls_path, status, error_message, created_at, finished_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		u.LessonID, u.OriginalFilename, u.StoredPath, u.HLSPath, u.Status, u.ErrorMessage, now, "",
	)
	if err != nil {
		return VideoUpload{}, err
	}
	u.ID, _ = res.LastInsertId()
	u.CreatedAt, _ = time.Parse(time.RFC3339, now)
	return u, nil
}

func (r *VideoRepo) UpdateStatus(ctx context.Context, id int64, status, hlsPath, errMsg string) error {
	finished := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`UPDATE video_uploads
		   SET status = ?, hls_path = ?, error_message = ?, finished_at = ?
		 WHERE id = ?`,
		status, hlsPath, errMsg, finished, id,
	)
	return err
}

func (r *VideoRepo) GetByLesson(ctx context.Context, lessonID int64) (*VideoUpload, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, lesson_id, original_filename, stored_path, hls_path, status, error_message, created_at, finished_at
		   FROM video_uploads
		  WHERE lesson_id = ?
		  ORDER BY id DESC
		  LIMIT 1`, lessonID,
	)
	var u VideoUpload
	var created, finished string
	if err := row.Scan(&u.ID, &u.LessonID, &u.OriginalFilename, &u.StoredPath, &u.HLSPath, &u.Status, &u.ErrorMessage, &created, &finished); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	u.CreatedAt, _ = time.Parse(time.RFC3339, created)
	if finished != "" {
		u.FinishedAt, _ = time.Parse(time.RFC3339, finished)
	}
	return &u, nil
}
