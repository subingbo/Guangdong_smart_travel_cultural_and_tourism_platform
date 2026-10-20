package repo

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type FavoriteRepo struct{ db *sqlx.DB }

func NewFavoriteRepo(db *sqlx.DB) *FavoriteRepo { return &FavoriteRepo{db: db} }

func (r *FavoriteRepo) FindPlacesByUser(ctx context.Context, userID int64) ([]model.Place, error) {
	var list []model.Place
	err := r.db.SelectContext(ctx, &list,
		`SELECT p.* FROM favorites f JOIN places p ON p.id = f.place_id
		 WHERE f.user_id = ? ORDER BY f.created_at DESC`, userID)
	return list, err
}

func (r *FavoriteRepo) Insert(ctx context.Context, userID, placeID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO favorites (user_id, place_id) VALUES (?, ?)", userID, placeID)
	return err
}

func (r *FavoriteRepo) Delete(ctx context.Context, userID, placeID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM favorites WHERE user_id = ? AND place_id = ?", userID, placeID)
	return err
}

func (r *FavoriteRepo) Exists(ctx context.Context, userID, placeID int64) (bool, error) {
	var n int
	err := r.db.GetContext(ctx, &n,
		"SELECT COUNT(*) FROM favorites WHERE user_id = ? AND place_id = ?", userID, placeID)
	return n > 0, err
}

func (r *FavoriteRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM favorites WHERE user_id = ?", userID)
	return n, err
}
