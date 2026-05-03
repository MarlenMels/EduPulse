-- Drop and recreate homework_submissions table with correct structure
DROP TABLE IF EXISTS homework_submissions CASCADE;

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
