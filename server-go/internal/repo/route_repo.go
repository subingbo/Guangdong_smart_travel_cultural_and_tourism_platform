package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type RouteRepo struct{ db *sqlx.DB }

func NewRouteRepo(db *sqlx.DB) *RouteRepo { return &RouteRepo{db: db} }

func (r *RouteRepo) Insert(ctx context.Context, rt *model.Route) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO routes (user_id, route_name, duration_days, total_distance, mode, pace, interests, waypoints, polyline)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rt.UserID, rt.RouteName, rt.DurationDays, rt.TotalDistance, rt.Mode, rt.Pace,
		rt.Interests, rt.Waypoints, rt.Polyline)
	if err != nil {
		return err
	}
	rt.ID, err = res.LastInsertId()
	return err
}

func (r *RouteRepo) FindByUser(ctx context.Context, userID int64) ([]model.Route, error) {
	list := make([]model.Route, 0)
	err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM routes WHERE user_id = ? ORDER BY created_at DESC", userID)
	return list, err
}

func (r *RouteRepo) FindByID(ctx context.Context, id, userID int64) (*model.Route, error) {
	var rt model.Route
	err := r.db.GetContext(ctx, &rt,
		"SELECT * FROM routes WHERE id = ? AND user_id = ?", id, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &rt, err
}

func (r *RouteRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM routes WHERE user_id = ?", userID)
	return n, err
}
