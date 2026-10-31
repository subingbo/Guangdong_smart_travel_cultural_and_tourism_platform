package repo

import (
	"context"

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
