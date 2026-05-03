CREATE TABLE IF NOT EXISTS course_teachers (
    course_id BIGINT NOT NULL,
    teacher_id BIGINT NOT NULL,
    created_at TEXT NOT NULL,
    PRIMARY KEY (course_id, teacher_id),
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (teacher_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_course_teachers_teacher ON course_teachers(teacher_id);

CREATE TABLE IF NOT EXISTS course_enrollments (
    course_id BIGINT NOT NULL,
    student_id BIGINT NOT NULL,
    enrolled_by BIGINT NOT NULL,
    enrolled_at TEXT NOT NULL,
    PRIMARY KEY (course_id, student_id),
    FOREIGN KEY (course_id) REFERENCES courses(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_enrollments_student ON course_enrollments(student_id);
CREATE INDEX IF NOT EXISTS idx_enrollments_course ON course_enrollments(course_id);
