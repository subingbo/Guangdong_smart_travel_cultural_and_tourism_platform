package repo

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type CommentRepo struct{ db *sqlx.DB }

func NewCommentRepo(db *sqlx.DB) *CommentRepo { return &CommentRepo{db: db} }

const commentSelect = `
	SELECT c.id, c.blog_id, c.user_id, c.content, c.created_at,
	       u.nickname AS author_nickname, u.avatar_url AS author_avatar_url
	FROM comments c JOIN users u ON u.id = c.user_id`

func (r *CommentRepo) ListByBlog(ctx context.Context, blogID int64, offset, limit int) ([]model.Comment, int, error) {
	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM comments WHERE blog_id = ?", blogID); err != nil {
		return nil, 0, err
	}
	list := make([]model.Comment, 0)
	err := r.db.SelectContext(ctx, &list,
		commentSelect+" WHERE c.blog_id = ? ORDER BY c.created_at ASC, c.id ASC LIMIT ? OFFSET ?",
		blogID, limit, offset)
	return list, total, err
}

func (r *CommentRepo) FindByID(ctx context.Context, id int64) (*model.Comment, error) {
	var c model.Comment
	err := r.db.GetContext(ctx, &c, commentSelect+" WHERE c.id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *CommentRepo) Insert(ctx context.Context, c *model.Comment) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO comments (blog_id, user_id, content) VALUES (?, ?, ?)",
		c.BlogID, c.UserID, c.Content)
	if err != nil {
		return err
	}
	c.ID, err = res.LastInsertId()
	return err
}

func (r *CommentRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM comments WHERE id = ?", id)
	return err
}

func (r *CommentRepo) CountByBlog(ctx context.Context, blogID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM comments WHERE blog_id = ?", blogID)
	return n, err
}
