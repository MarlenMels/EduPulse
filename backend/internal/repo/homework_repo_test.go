package repo

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestListForParentReturnsOnlyLinkedStudentHomework(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	schema := `
		CREATE TABLE homework_submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			content TEXT NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE TABLE parent_students (
			parent_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			created_at TEXT NOT NULL,
			PRIMARY KEY(parent_id, student_id)
		);
		INSERT INTO parent_students (parent_id, student_id, created_at) VALUES (10, 20, '2026-05-02T00:00:00Z');
		INSERT INTO homework_submissions (session_id, student_id, content, status, created_at)
		VALUES (1, 20, 'linked', 'submitted', '2026-05-02T00:00:00Z'),
		       (1, 30, 'unlinked', 'submitted', '2026-05-02T00:00:00Z');
	`
	if _, err := db.Exec(schema); err != nil {
		t.Fatal(err)
	}

	items, err := NewHomeworkRepo(db).ListForParent(context.Background(), 10, "", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one linked homework item, got %d", len(items))
	}
	if items[0].StudentID != 20 || items[0].Content != "linked" {
		t.Fatalf("unexpected homework item: %#v", items[0])
	}
}
