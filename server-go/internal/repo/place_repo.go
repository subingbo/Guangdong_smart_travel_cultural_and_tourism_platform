package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type PlaceRepo struct{ db *sqlx.DB }

func NewPlaceRepo(db *sqlx.DB) *PlaceRepo { return &PlaceRepo{db: db} }

func (r *PlaceRepo) FindAll(ctx context.Context, tag string) ([]model.Place, error) {
	list := make([]model.Place, 0)
	var err error
	if tag == "" {
		err = r.db.SelectContext(ctx, &list, "SELECT * FROM places ORDER BY id ASC")
	} else {
		// JSON_CONTAINS checks one tag inside the JSON array.
		err = r.db.SelectContext(ctx, &list,
			"SELECT * FROM places WHERE JSON_CONTAINS(tags, JSON_QUOTE(?)) ORDER BY id ASC", tag)
	}
	return list, err
}

func (r *PlaceRepo) FindByID(ctx context.Context, id int64) (*model.Place, error) {
	var p model.Place
	err := r.db.GetContext(ctx, &p, "SELECT * FROM places WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

// Search backs the agent's search_places tool. Query is matched against
// names/descriptions; tag uses the JSON column; indoor is tri-state.
func (r *PlaceRepo) Search(ctx context.Context, query, tag string, indoor *bool, limit int) ([]model.Place, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	clauses := []string{"1=1"}
	args := make([]interface{}, 0, 8)
	if q := strings.TrimSpace(query); q != "" {
		pattern := "%" + q + "%"
		clauses = append(clauses, "(name_cn LIKE ? OR name_en LIKE ? OR description_cn LIKE ? OR description_en LIKE ?)")
		args = append(args, pattern, pattern, pattern, pattern)
	}
	if tag != "" {
		clauses = append(clauses, "JSON_CONTAINS(tags, JSON_QUOTE(?))")
		args = append(args, tag)
	}
	if indoor != nil {
		if *indoor {
			clauses = append(clauses, "indoor = 1")
		} else {
			clauses = append(clauses, "indoor = 0")
		}
	}
	args = append(args, limit)
	list := make([]model.Place, 0)
	sqlText := "SELECT * FROM places WHERE " + strings.Join(clauses, " AND ") + " ORDER BY id ASC LIMIT ?"
	if err := r.db.SelectContext(ctx, &list, sqlText, args...); err != nil {
		return nil, err
	}
	return list, nil
}

func (r *PlaceRepo) FindByIDs(ctx context.Context, ids []int64) ([]model.Place, error) {
	if len(ids) == 0 {
		return []model.Place{}, nil
	}
	query, args, err := sqlx.In("SELECT * FROM places WHERE id IN (?)", ids)
	if err != nil {
		return nil, err
	}
	list := make([]model.Place, 0, len(ids))
	if err := r.db.SelectContext(ctx, &list, r.db.Rebind(query), args...); err != nil {
		return nil, err
	}
	return list, nil
}
