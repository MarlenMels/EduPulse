-- Drop existing tables to avoid conflicts
DROP TABLE IF EXISTS homework_submissions CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;
DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Create tables with correct structure
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    last_seen_at TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);

CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    teacher_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    start_time TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(teacher_id) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);

CREATE TABLE homework_submissions (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY(student_id) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX idx_hw_session ON homework_submissions(session_id);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT NOT NULL DEFAULT 0,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id BIGINT NOT NULL,
    meta_json TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX idx_audit_created_at ON audit_logs(created_at);

CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload_json TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX idx_notifications_status ON notifications(status);
