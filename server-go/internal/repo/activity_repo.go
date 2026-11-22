package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type ActivityRepo struct{ db *sqlx.DB }

func NewActivityRepo(db *sqlx.DB) *ActivityRepo { return &ActivityRepo{db: db} }

func (r *ActivityRepo) FindPage(ctx context.Context, status, q string, offset, limit int) ([]model.Activity, int, error) {
	clauses := []string{"1=1"}
	args := make([]interface{}, 0, 5)
	if status = strings.TrimSpace(status); status != "" && status != "all" {
		clauses = append(clauses, "status = ?")
		args = append(args, status)
	}
	if q = strings.TrimSpace(q); q != "" {
		like := "%" + q + "%"
		clauses = append(clauses, "(title_cn LIKE ? OR title_en LIKE ? OR description_cn LIKE ? OR description_en LIKE ?)")
		args = append(args, like, like, like, like)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")
	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM activities"+where, args...); err != nil {
		return nil, 0, err
	}
	list := make([]model.Activity, 0)
	queryArgs := append(append([]interface{}{}, args...), limit, offset)
	if err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM activities"+where+" ORDER BY start_at ASC, id ASC LIMIT ? OFFSET ?", queryArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ActivityRepo) FindByID(ctx context.Context, id int64) (*model.Activity, error) {
	var a model.Activity
	err := r.db.GetContext(ctx, &a, "SELECT * FROM activities WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &a, err
}
