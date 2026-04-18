ALTER TABLE lessons ADD COLUMN hls_url TEXT NOT NULL DEFAULT '';
ALTER TABLE lessons ADD COLUMN video_status TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS video_uploads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lesson_id INTEGER NOT NULL,
    original_filename TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    hls_path TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    error_message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    finished_at TEXT NOT NULL DEFAULT '',
    FOREIGN KEY(lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_video_uploads_lesson ON video_uploads(lesson_id);
