package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type BlogRepo struct{ db *sqlx.DB }

func NewBlogRepo(db *sqlx.DB) *BlogRepo { return &BlogRepo{db: db} }

// viewerID <= 0 means anonymous: liked_by_me is always false.
const blogColumns = `
	b.id, b.user_id, u.nickname AS author_nickname, u.avatar_url AS author_avatar_url,
	b.title, b.content, b.images, b.likes, b.created_at, b.updated_at,
	CASE WHEN ? > 0 AND EXISTS(
		SELECT 1 FROM blog_likes bl WHERE bl.blog_id = b.id AND bl.user_id = ?
	) THEN TRUE ELSE FALSE END AS liked_by_me,
	(SELECT COUNT(*) FROM comments c WHERE c.blog_id = b.id) AS comment_count`

const blogFrom = ` FROM travel_blogs b JOIN users u ON u.id = b.user_id`

func blogFilter(authorID int64, q string) (string, []interface{}) {
	clauses := []string{"1=1"}
	args := make([]interface{}, 0, 3)
	if authorID > 0 {
		clauses = append(clauses, "b.user_id = ?")
		args = append(args, authorID)
	}
	if q = strings.TrimSpace(q); q != "" {
		like := "%" + q + "%"
		clauses = append(clauses, "(b.title LIKE ? OR b.content LIKE ?)")
		args = append(args, like, like)
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func (r *BlogRepo) FindPage(ctx context.Context, viewerID, authorID int64, q string, offset, limit int) ([]model.TravelBlog, int, error) {
	total := 0
	where, filterArgs := blogFilter(authorID, q)
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM travel_blogs b"+where, filterArgs...); err != nil {
		return nil, 0, err
	}

	list := make([]model.TravelBlog, 0)
	args := make([]interface{}, 0, len(filterArgs)+4)
	args = append(args, viewerID, viewerID)
	args = append(args, filterArgs...)
	args = append(args, limit, offset)
	query := "SELECT " + blogColumns + blogFrom + where + " ORDER BY b.created_at DESC, b.id DESC LIMIT ? OFFSET ?"
	if err := r.db.SelectContext(ctx, &list, query, args...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *BlogRepo) FindByID(ctx context.Context, id, viewerID int64) (*model.TravelBlog, error) {
	var b model.TravelBlog
	err := r.db.GetContext(ctx, &b,
		"SELECT "+blogColumns+blogFrom+" WHERE b.id = ?", viewerID, viewerID, id)
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

// Update returns affected rows; 0 means not found or not the owner.
func (r *BlogRepo) Update(ctx context.Context, id, userID int64, title, content string, images model.StringList) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		"UPDATE travel_blogs SET title = ?, content = ?, images = ?, updated_at = NOW() WHERE id = ? AND user_id = ?",
		title, content, images, id, userID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Delete removes the blog and its likes; comments and notifications cascade.
func (r *BlogRepo) Delete(ctx context.Context, id, userID int64) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, "DELETE FROM travel_blogs WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return affected, err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM blog_likes WHERE blog_id = ?", id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return affected, nil
}

// AuthorID returns the owner id, or 0 when the blog does not exist.
func (r *BlogRepo) AuthorID(ctx context.Context, id int64) (int64, error) {
	var userID int64
	err := r.db.GetContext(ctx, &userID, "SELECT user_id FROM travel_blogs WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return userID, err
}

// SetLike inserts the like once and bumps the counter only when the row was
// actually created. It returns (newLike, likes).
func (r *BlogRepo) SetLike(ctx context.Context, userID, blogID int64) (bool, int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		"INSERT IGNORE INTO blog_likes (user_id, blog_id) SELECT ?, id FROM travel_blogs WHERE id = ?",
		userID, blogID)
	if err != nil {
		return false, 0, err
	}
	inserted, err := res.RowsAffected()
	if err != nil {
		return false, 0, err
	}
	if inserted > 0 {
		if _, err := tx.ExecContext(ctx, "UPDATE travel_blogs SET likes = likes + 1 WHERE id = ?", blogID); err != nil {
			return false, 0, err
		}
	}
	likes, err := likesTx(ctx, tx, blogID)
	if err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return inserted > 0, likes, nil
}

// RemoveLike deletes the like and decrements the counter only when a row was removed.
func (r *BlogRepo) RemoveLike(ctx context.Context, userID, blogID int64) (bool, int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		"DELETE FROM blog_likes WHERE user_id = ? AND blog_id = ?", userID, blogID)
	if err != nil {
		return false, 0, err
	}
	removed, err := res.RowsAffected()
	if err != nil {
		return false, 0, err
	}
	if removed > 0 {
		if _, err := tx.ExecContext(ctx,
			"UPDATE travel_blogs SET likes = GREATEST(likes - 1, 0) WHERE id = ?", blogID); err != nil {
			return false, 0, err
		}
	}
	likes, err := likesTx(ctx, tx, blogID)
	if err != nil {
		return false, 0, err
	}
	if err := tx.Commit(); err != nil {
		return false, 0, err
	}
	return removed > 0, likes, nil
}

func likesTx(ctx context.Context, tx *sqlx.Tx, blogID int64) (int, error) {
	var likes int
	err := tx.GetContext(ctx, &likes, "SELECT likes FROM travel_blogs WHERE id = ?", blogID)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return likes, err
}

func (r *BlogRepo) CountByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n, "SELECT COUNT(*) FROM travel_blogs WHERE user_id = ?", userID)
	return n, err
}

func (r *BlogRepo) SumLikesByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n,
		"SELECT COALESCE(SUM(likes), 0) FROM travel_blogs WHERE user_id = ?", userID)
	return n, err
}
