package repo

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type FollowRepo struct{ db *sqlx.DB }

func NewFollowRepo(db *sqlx.DB) *FollowRepo { return &FollowRepo{db: db} }

// Follow returns true when a new row was created; following twice is a no-op.
func (r *FollowRepo) Follow(ctx context.Context, followerID, followeeID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT IGNORE INTO follows (follower_id, followee_id) VALUES (?, ?)", followerID, followeeID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *FollowRepo) Unfollow(ctx context.Context, followerID, followeeID int64) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		"DELETE FROM follows WHERE follower_id = ? AND followee_id = ?", followerID, followeeID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *FollowRepo) IsFollowing(ctx context.Context, followerID, followeeID int64) (bool, error) {
	if followerID <= 0 || followerID == followeeID {
		return false, nil
	}
	var n int
	err := r.db.GetContext(ctx, &n,
		"SELECT COUNT(*) FROM follows WHERE follower_id = ? AND followee_id = ?", followerID, followeeID)
	return n > 0, err
}

func (r *FollowRepo) CountFollowers(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM follows WHERE followee_id = ?", userID)
	return n, err
}

func (r *FollowRepo) CountFollowing(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM follows WHERE follower_id = ?", userID)
	return n, err
}

func (r *FollowRepo) ListFollowers(ctx context.Context, userID int64, offset, limit int) ([]model.UserPublic, int, error) {
	total, err := r.CountFollowers(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	list := make([]model.UserPublic, 0)
	err = r.db.SelectContext(ctx, &list, `
		SELECT u.id, u.username, u.nickname, u.avatar_url, u.created_at
		FROM follows f JOIN users u ON u.id = f.follower_id
		WHERE f.followee_id = ?
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT ? OFFSET ?`, userID, limit, offset)
	return list, total, err
}

func (r *FollowRepo) ListFollowing(ctx context.Context, userID int64, offset, limit int) ([]model.UserPublic, int, error) {
	total, err := r.CountFollowing(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	list := make([]model.UserPublic, 0)
	err = r.db.SelectContext(ctx, &list, `
		SELECT u.id, u.username, u.nickname, u.avatar_url, u.created_at
		FROM follows f JOIN users u ON u.id = f.followee_id
		WHERE f.follower_id = ?
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT ? OFFSET ?`, userID, limit, offset)
	return list, total, err
}
