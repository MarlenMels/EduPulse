package events

const (
	EventHomeworkSubmitted = "homework_submitted"
	EventHomeworkGraded    = "homework_graded"
)

type HomeworkSubmittedPayload struct {
	SubmissionID int64  `json:"submission_id"`
	SessionID    int64  `json:"session_id"`
	StudentID    int64  `json:"student_id"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

type HomeworkGradedPayload struct {
	SubmissionID int64  `json:"submission_id"`
	SessionID    int64  `json:"session_id"`
	StudentID    int64  `json:"student_id"`
	OldStatus    string `json:"old_status"`
	NewStatus    string `json:"new_status"`
	GradedAt     string `json:"graded_at"`
}