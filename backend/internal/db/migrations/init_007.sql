CREATE TABLE IF NOT EXISTS course_teachers (
    course_id INTEGER NOT NULL,
    teacher_id INTEGER NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (course_id, teacher_id),
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_course_teachers_teacher ON course_teachers(teacher_id);

CREATE TABLE IF NOT EXISTS course_enrollments (
    course_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    enrolled_by INTEGER NOT NULL,
    enrolled_at TEXT NOT NULL,
    PRIMARY KEY (course_id, student_id),
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON course_enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course ON course_enrollments(course_id);
