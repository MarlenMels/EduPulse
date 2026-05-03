package repo

import (
	"context"
	"database/sql"
	"time"
)

type HomeworkRepo struct{ db *sql.DB }

func NewHomeworkRepo(db *sql.DB) *HomeworkRepo { return &HomeworkRepo{db: db} }

func (r *HomeworkRepo) Create(ctx context.Context, h HomeworkSubmission) (HomeworkSubmission, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO homework_submissions (assignment_id, student_id, content, attachments, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		h.AssignmentID, h.StudentID, h.Content, h.Attachments, h.Status, created,
	).Scan(&id)
	if err != nil {
		return HomeworkSubmission{}, err
	}
	h.ID = id
	h.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return h, nil
}

func (r *HomeworkRepo) GetByID(ctx context.Context, id int64) (*HomeworkSubmission, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, assignment_id, student_id, content, attachments, status, created_at FROM homework_submissions WHERE id = $1",
		id,
	)
	var h HomeworkSubmission
	var created string
	if err := row.Scan(&h.ID, &h.AssignmentID, &h.StudentID, &h.Content, &h.Attachments, &h.Status, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	h.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &h, nil
}

func (r *HomeworkRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE homework_submissions SET status = $1 WHERE id = $2", status, id)
	return err
}

// HasSubmission returns true if the student already submitted for the assignment.
func (r *HomeworkRepo) HasSubmission(ctx context.Context, assignmentID, studentID int64) (bool, error) {
	var n int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM homework_submissions WHERE assignment_id = $1 AND student_id = $2",
		assignmentID, studentID,
	).Scan(&n)
	return n > 0, err
}

// MineByStudent returns submissions a student made, joined with assignment info.
type MineRow struct {
	HomeworkSubmission
	AssignmentTitle string `json:"assignment_title"`
	SessionID       int64  `json:"session_id"`
	SessionTitle    string `json:"session_title"`
	CourseID        int64  `json:"course_id"`
	CourseTitle     string `json:"course_title"`
}

func (r *HomeworkRepo) MineByStudent(ctx context.Context, studentID int64, limit int) ([]MineRow, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT h.id, h.assignment_id, h.student_id, h.content, h.attachments, h.status, h.created_at,
		       a.title, s.id, s.title, c.id, c.title
		  FROM homework_submissions h
		  JOIN assignments a ON a.id = h.assignment_id
		  JOIN sessions    s ON s.id = a.session_id
		  JOIN courses     c ON c.id = s.course_id
		 WHERE h.student_id = $1
		 ORDER BY h.id DESC
		 LIMIT $2
	`, studentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]MineRow, 0, 16)
	for rows.Next() {
		var m MineRow
		var created string
		if err := rows.Scan(
			&m.ID, &m.AssignmentID, &m.StudentID, &m.Content, &m.Attachments, &m.Status, &created,
			&m.AssignmentTitle, &m.SessionID, &m.SessionTitle, &m.CourseID, &m.CourseTitle,
		); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, m)
	}
	return out, rows.Err()
}
