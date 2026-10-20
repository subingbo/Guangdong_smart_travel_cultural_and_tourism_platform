package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type BlogRepo struct{ db *sqlx.DB }

func NewBlogRepo(db *sqlx.DB) *BlogRepo { return &BlogRepo{db: db} }

/** viewerID <= 0 means anonymous: liked_by_me is always false. */
const blogSelect = `
	SELECT b.id, b.user_id, u.nickname AS author_nickname, u.avatar_url AS author_avatar_url,
	       b.title, b.content, b.images, b.likes, b.created_at,
	       CASE WHEN ? > 0 AND EXISTS(
	           SELECT 1 FROM blog_likes bl WHERE bl.blog_id = b.id AND bl.user_id = ?
	       ) THEN TRUE ELSE FALSE END AS liked_by_me
	FROM travel_blogs b JOIN users u ON u.id = b.user_id`

func (r *BlogRepo) FindAll(ctx context.Context, viewerID int64) ([]model.TravelBlog, error) {
	var list []model.TravelBlog
	err := r.db.SelectContext(ctx, &list, blogSelect+" ORDER BY b.created_at DESC", viewerID, viewerID)
	return list, err
}

func (r *BlogRepo) FindByID(ctx context.Context, id, viewerID int64) (*model.TravelBlog, error) {
	var b model.TravelBlog
	err := r.db.GetContext(ctx, &b, blogSelect+" WHERE b.id = ?", viewerID, viewerID, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

func (r *BlogRepo) Insert(ctx context.Context, b *model.TravelBlog) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO travel_blogs (user_id, title, content, images) VALUES (?, ?, ?, ?)",
		b.UserID, b.Title, b.Content, b.Images)
	if err != nil {
		return err
	}
	b.ID, err = res.LastInsertId()
	return err
}

func (r *BlogRepo) IncrementLikes(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE travel_blogs SET likes = likes + 1 WHERE id = ?", id)
	return err
}

func (r *BlogRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM travel_blogs WHERE user_id = ?", userID)
	return n, err
}

func (r *BlogRepo) AddLike(ctx context.Context, userID, blogID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT IGNORE INTO blog_likes (user_id, blog_id) VALUES (?, ?)", userID, blogID)
	return err
}
