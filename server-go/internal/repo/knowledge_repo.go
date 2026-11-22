package repo

import (
	"context"
	"database/sql"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

type KnowledgeRepo struct{ db *sqlx.DB }

func NewKnowledgeRepo(db *sqlx.DB) *KnowledgeRepo { return &KnowledgeRepo{db: db} }

func (r *KnowledgeRepo) SearchByKeyword(ctx context.Context, keyword string, limit int) ([]model.Knowledge, error) {
	list := make([]model.Knowledge, 0)
	pattern := "%" + keyword + "%"
	err := r.db.SelectContext(ctx, &list,
		`SELECT * FROM knowledge
		 WHERE title_cn LIKE ? OR title_en LIKE ? OR content_cn LIKE ? OR content_en LIKE ? OR tags LIKE ?
		 LIMIT ?`,
		pattern, pattern, pattern, pattern, pattern, limit)
	return list, err
}

func (r *KnowledgeRepo) FindPage(ctx context.Context, q string, offset, limit int) ([]model.Knowledge, int, error) {
	where := ""
	args := make([]interface{}, 0, 6)
	if q = strings.TrimSpace(q); q != "" {
		pattern := "%" + q + "%"
		where = " WHERE title_cn LIKE ? OR title_en LIKE ? OR content_cn LIKE ? OR content_en LIKE ?"
		args = append(args, pattern, pattern, pattern, pattern)
	}
	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM knowledge"+where, args...); err != nil {
		return nil, 0, err
	}
	list := make([]model.Knowledge, 0)
	queryArgs := append(append([]interface{}{}, args...), limit, offset)
	if err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM knowledge"+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *KnowledgeRepo) FindByID(ctx context.Context, id int64) (*model.Knowledge, error) {
	var k model.Knowledge
	err := r.db.GetContext(ctx, &k, "SELECT * FROM knowledge WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &k, err
}

func (r *KnowledgeRepo) Insert(ctx context.Context, k *model.Knowledge) error {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO knowledge (title_cn, title_en, content_cn, content_en, tags) VALUES (?, ?, ?, ?, ?)",
		k.TitleCn, k.TitleEn, k.ContentCn, k.ContentEn, k.Tags)
	if err != nil {
		return err
	}
	k.ID, err = res.LastInsertId()
	return err
}

func (r *KnowledgeRepo) Update(ctx context.Context, k *model.Knowledge) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE knowledge SET title_cn = ?, title_en = ?, content_cn = ?, content_en = ?, tags = ?, updated_at = NOW() WHERE id = ?",
		k.TitleCn, k.TitleEn, k.ContentCn, k.ContentEn, k.Tags, k.ID)
	return err
}

func (r *KnowledgeRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, "DELETE FROM knowledge WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}
