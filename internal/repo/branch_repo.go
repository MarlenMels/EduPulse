package repo

import (
	"context"
	"database/sql"
	"time"
)

type BranchRepo struct{ db *sql.DB }

func NewBranchRepo(db *sql.DB) *BranchRepo { return &BranchRepo{db: db} }

func (r *BranchRepo) Create(ctx context.Context, b Branch) (Branch, error) {
	created := time.Now().UTC().Format(time.RFC3339)
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO branches (name, lat, lng, created_at) VALUES (?, ?, ?, ?)",
		b.Name, b.Lat, b.Lng, created,
	)
	if err != nil {
		return Branch{}, err
	}
	id, _ := res.LastInsertId()
	b.ID = id
	b.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return b, nil
}

func (r *BranchRepo) GetByID(ctx context.Context, id int64) (*Branch, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, name, lat, lng, created_at FROM branches WHERE id = ?", id)
	var b Branch
	var created string
	if err := row.Scan(&b.ID, &b.Name, &b.Lat, &b.Lng, &created); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	b.CreatedAt, _ = time.Parse(time.RFC3339, created)
	return &b, nil
}

func (r *BranchRepo) List(ctx context.Context, q string, limit int) ([]Branch, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	sqlText := "SELECT id, name, lat, lng, created_at FROM branches"
	args := make([]any, 0, 2)

	if q != "" {
		sqlText += " WHERE name LIKE ?"
		args = append(args, "%"+q+"%")
	}
	sqlText += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Branch, 0, 32)
	for rows.Next() {
		var b Branch
		var created string
		if err := rows.Scan(&b.ID, &b.Name, &b.Lat, &b.Lng, &created); err != nil {
			return nil, err
		}
		b.CreatedAt, _ = time.Parse(time.RFC3339, created)
		out = append(out, b)
	}
	return out, rows.Err()
}
