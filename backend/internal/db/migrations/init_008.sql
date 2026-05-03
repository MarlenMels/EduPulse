-- Beta cleanup: existing sessions and homework_submissions don't fit the new
-- course-anchored model (sessions need course_id, homework needs assignment_id).
-- We drop and recreate them. Beta data is acceptable to lose.
DROP TABLE IF EXISTS homework_submissions;
DROP TABLE IF EXISTS sessions;

CREATE TABLE sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    course_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    start_time TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE
);
CREATE INDEX idx_sessions_course ON sessions(course_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);

CREATE TABLE assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    created_by INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);
CREATE INDEX idx_assignments_session ON assignments(session_id);
CREATE INDEX idx_assignments_creator ON assignments(created_by);

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
CREATE INDEX idx_hw_assignment ON homework_submissions(assignment_id);
CREATE INDEX idx_hw_student ON homework_submissions(student_id);
