CREATE TABLE IF NOT EXISTS lesson_assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    lesson_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    url TEXT NOT NULL,
    original_filename TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    FOREIGN KEY(lesson_id) REFERENCES lessons(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_lesson_assets_lesson ON lesson_assets(lesson_id);
