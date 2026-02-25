package repo

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Branch struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	H3Index   string    `json:"h3_index"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	BranchID  int64     `json:"branch_id"`
	TeacherID int64     `json:"teacher_id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	H3Index   string    `json:"h3_index"`
	CreatedAt time.Time `json:"created_at"`
}

type HomeworkSubmission struct {
	ID        int64     `json:"id"`
	SessionID int64     `json:"session_id"`
	StudentID int64     `json:"student_id"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type AuditLog struct {
	ID          int64     `json:"id"`
	ActorUserID int64     `json:"actor_user_id"`
	Action      string    `json:"action"`
	EntityType  string    `json:"entity_type"`
	EntityID    int64     `json:"entity_id"`
	MetaJSON    string    `json:"meta_json"`
	CreatedAt   time.Time `json:"created_at"`
}

type Notification struct {
	ID          int64     `json:"id"`
	EventType   string    `json:"event_type"`
	PayloadJSON string    `json:"payload_json"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type AnalyticsRow struct {
	H3Index       string `json:"h3_index"`
	Day           string `json:"day"`
	SessionsCount int    `json:"sessions_count"`
}