package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type AssignmentRepo struct{ db *sql.DB }

func NewAssignmentRepo(db *sql.DB) *AssignmentRepo { return &AssignmentRepo{db: db} }

func (r *AssignmentRepo) Create(ctx context.Context, a Assignment) (Assignment, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO assignments (session_id, created_by, title, description, created_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		a.SessionID, a.CreatedBy, a.Title, a.Description, now,
	).Scan(&id)
	if err != nil {
		return Assignment{}, err
	}
	a.ID = id
	a.CreatedAt, _ = time.Parse(time.RFC3339, now)
	return a, nil
}

func (r *AssignmentRepo) GetByID(ctx context.Context, id int64) (*Assignment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, session_id, created_by, title, description, created_at
		   FROM assignments WHERE id = $1`,
		id,
	)
	var a Assignment
	var created string
	if err := row.Scan(&a.ID, &a.SessionID, &a.CreatedBy, &a.Title, &a.Description, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	a.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &a, nil
}

func (r *AssignmentRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM assignments WHERE id = $1", id)
	return err
}

// AssignmentRow extends an assignment with denormalized session/course/teacher fields
// so list endpoints can render them without follow-up joins.
type AssignmentRow struct {
	Assignment
	SessionTitle    string `json:"session_title"`
	SessionStart    string `json:"session_start_time"`
	CourseID        int64  `json:"course_id"`
	CourseTitle     string `json:"course_title"`
	CreatorEmail    string `json:"creator_email"`
	SubmissionCount int    `json:"submission_count"`
}

type AssignmentFilter struct {
	// SessionID limits to one session.
	SessionID int64
	// TeacherID limits to assignments either created by the teacher OR
	// belonging to a course they teach.
	TeacherID int64
	// CourseIDs limits to assignments whose session belongs to one of these courses.
	// Used for student visibility.
	CourseIDs []int64
	Limit     int
}

func (r *AssignmentRepo) List(ctx context.Context, f AssignmentFilter) ([]AssignmentRow, error) {
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 100
	}

	conds := []string{}
	args := []any{}
	n := 0
	ph := func() string { n++; return fmt.Sprintf("$%d", n) }

	if f.SessionID > 0 {
		conds = append(conds, "a.session_id = "+ph())
		args = append(args, f.SessionID)
	}
	if f.TeacherID > 0 {
		// Either the teacher created the assignment, or they teach the course
		// the session belongs to.
		conds = append(conds,
			"(a.created_by = "+ph()+" OR EXISTS (SELECT 1 FROM course_teachers ct WHERE ct.course_id = s.course_id AND ct.teacher_id = "+ph()+"))",
		)
		args = append(args, f.TeacherID, f.TeacherID)
	}
	if len(f.CourseIDs) > 0 {
		placeholders := make([]string, len(f.CourseIDs))
		for i, id := range f.CourseIDs {
			placeholders[i] = ph()
			args = append(args, id)
		}
		conds = append(conds, "s.course_id IN ("+strings.Join(placeholders, ",")+")")
	}

	where := ""
	if len(conds) > 0 {
		where = "WHERE " + strings.Join(conds, " AND ")
	}

	args = append(args, limit)
	limitPH := ph()

	q := `
		SELECT a.id, a.session_id, a.created_by, a.title, a.description, a.created_at,
		       s.title, s.start_time, s.course_id, c.title, u.email,
		       (SELECT COUNT(1) FROM homework_submissions hs WHERE hs.assignment_id = a.id) AS submission_count
		  FROM assignments a
		  JOIN sessions s ON s.id = a.session_id
		  JOIN courses  c ON c.id = s.course_id
		  JOIN users    u ON u.id = a.created_by
		` + where + `
		ORDER BY a.id DESC
		LIMIT ` + limitPH

	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]AssignmentRow, 0, 32)
	for rows.Next() {
		var a AssignmentRow
		var created string
		if err := rows.Scan(
			&a.ID, &a.SessionID, &a.CreatedBy, &a.Title, &a.Description, &created,
			&a.SessionTitle, &a.SessionStart, &a.CourseID, &a.CourseTitle, &a.CreatorEmail,
			&a.SubmissionCount,
		); err != nil {
			return nil, err
		}
		a.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, a)
	}
	return out, rows.Err()
}

// SubmissionRow extends a homework submission with denormalized student email
// for list views.
type SubmissionRow struct {
	HomeworkSubmission
	StudentEmail string `json:"student_email"`
}

func (r *AssignmentRepo) Submissions(ctx context.Context, assignmentID int64) ([]SubmissionRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT h.id, h.assignment_id, h.student_id, h.content, h.attachments, h.status, h.created_at, u.email
		   FROM homework_submissions h
		   JOIN users u ON u.id = h.student_id
		  WHERE h.assignment_id = $1
		  ORDER BY h.id DESC`,
		assignmentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SubmissionRow, 0, 16)
	for rows.Next() {
		var s SubmissionRow
		var created string
		if err := rows.Scan(&s.ID, &s.AssignmentID, &s.StudentID, &s.Content, &s.Attachments, &s.Status, &created, &s.StudentEmail); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, s)
	}
	return out, rows.Err()
}
