-- See SQLite init_009 for rationale.
DROP TABLE IF EXISTS homework_submissions;
CREATE TABLE homework_submissions (
    id BIGSERIAL PRIMARY KEY,
    assignment_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    attachments TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_hw_assignment ON homework_submissions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_hw_student ON homework_submissions(student_id);
