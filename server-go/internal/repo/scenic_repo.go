package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type ScenicRepo struct{ db *sqlx.DB }

func NewScenicRepo(db *sqlx.DB) *ScenicRepo { return &ScenicRepo{db: db} }

func (r *ScenicRepo) FindAll(ctx context.Context) ([]model.ScenicImage, error) {
	list := make([]model.ScenicImage, 0)
	err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM scenic_images ORDER BY sort_order ASC, id ASC")
	return list, err
}

func (r *ScenicRepo) FindByID(ctx context.Context, id int64) (*model.ScenicImage, error) {
	var s model.ScenicImage
	err := r.db.GetContext(ctx, &s, "SELECT * FROM scenic_images WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}
