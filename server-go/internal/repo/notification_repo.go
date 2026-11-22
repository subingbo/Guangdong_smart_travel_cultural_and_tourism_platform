package repo

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type NotificationRepo struct{ db *sqlx.DB }

func NewNotificationRepo(db *sqlx.DB) *NotificationRepo { return &NotificationRepo{db: db} }

func (r *NotificationRepo) Insert(ctx context.Context, n *model.Notification) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO notifications (user_id, actor_id, type, blog_id, comment_id) VALUES (?, ?, ?, ?, ?)",
		n.UserID, n.ActorID, n.Type, nullableID(n.BlogID), nullableID(n.CommentID))
	if err != nil {
		return err
	}
	n.ID, err = res.LastInsertId()
	return err
}

func nullableID(id int64) interface{} {
	if id <= 0 {
		return nil
	}
	return id
}

const notificationSelect = `
	SELECT n.id, n.type, n.is_read, n.created_at, n.actor_id,
	       u.nickname AS actor_nickname, u.avatar_url AS actor_avatar_url,
	       COALESCE(n.blog_id, 0) AS blog_id,
	       COALESCE(b.title, '') AS blog_title,
	       COALESCE(c.content, '') AS comment_content
	FROM notifications n
	JOIN users u ON u.id = n.actor_id
	LEFT JOIN travel_blogs b ON b.id = n.blog_id
	LEFT JOIN comments c ON c.id = n.comment_id`

func (r *NotificationRepo) List(ctx context.Context, userID int64, unreadOnly bool, offset, limit int) ([]model.Notification, int, error) {
	where := " WHERE n.user_id = ?"
	args := []interface{}{userID}
	if unreadOnly {
		where += " AND n.is_read = 0"
	}
	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM notifications n"+where, args...); err != nil {
		return nil, 0, err
	}
	list := make([]model.Notification, 0)
	args = append(args, limit, offset)
	err := r.db.SelectContext(ctx, &list,
		notificationSelect+where+" ORDER BY n.created_at DESC, n.id DESC LIMIT ? OFFSET ?", args...)
	return list, total, err
}

func (r *NotificationRepo) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.GetContext(ctx, &n,
		"SELECT COUNT(*) FROM notifications WHERE user_id = ? AND is_read = 0", userID)
	return n, err
}

func (r *NotificationRepo) MarkRead(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query, args, err := sqlx.In(
		"UPDATE notifications SET is_read = 1 WHERE user_id = ? AND id IN (?)", userID, ids)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, r.db.Rebind(query), args...)
	return err
}

func (r *NotificationRepo) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE notifications SET is_read = 1 WHERE user_id = ? AND is_read = 0", userID)
	return err
}
