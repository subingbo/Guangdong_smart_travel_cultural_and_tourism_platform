package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type PlaceRepo struct{ db *sqlx.DB }

func NewPlaceRepo(db *sqlx.DB) *PlaceRepo { return &PlaceRepo{db: db} }

func (r *PlaceRepo) FindAll(ctx context.Context, tag string) ([]model.Place, error) {
	var list []model.Place
	var err error
	if tag == "" {
		err = r.db.SelectContext(ctx, &list, "SELECT * FROM places ORDER BY id ASC")
	} else {
		// JSON_CONTAINS mirrors the MyBatis XML tag filter.
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
