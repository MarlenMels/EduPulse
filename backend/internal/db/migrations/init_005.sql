CREATE TABLE IF NOT EXISTS parent_students (
    parent_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY(parent_id, student_id),
    FOREIGN KEY(parent_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(student_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_parent_students_parent ON parent_students(parent_id);
