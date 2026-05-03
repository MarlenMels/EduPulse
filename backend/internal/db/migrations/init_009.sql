-- Defensive re-recreate: prod Postgres reported homework_submissions still
-- carrying the legacy session_id column even after init_008 ran. Drop and
-- recreate it again in a dedicated migration so the new statement-per-Exec
-- migrator definitely applies it.
DROP TABLE IF EXISTS homework_submissions;
CREATE TABLE homework_submissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    attachments TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (assignment_id) REFERENCES assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX IF NOT EXISTS idx_hw_assignment ON homework_submissions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_hw_student ON homework_submissions(student_id);
