package repo

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/travelgd/server-go/internal/model"
)

var (
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrProductUnavailable = errors.New("product unavailable")
)

type ProductRepo struct{ db *sqlx.DB }

func NewProductRepo(db *sqlx.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) FindPage(ctx context.Context, category, q string, offset, limit int) ([]model.Product, int, error) {
	clauses := []string{"status = 1"}
	args := make([]interface{}, 0, 5)
	if category = strings.TrimSpace(category); category != "" && category != "all" {
		clauses = append(clauses, "category = ?")
		args = append(args, category)
	}
	if q = strings.TrimSpace(q); q != "" {
		like := "%" + q + "%"
		clauses = append(clauses, "(name_cn LIKE ? OR name_en LIKE ? OR description_cn LIKE ? OR description_en LIKE ?)")
		args = append(args, like, like, like, like)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")

	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM products"+where, args...); err != nil {
		return nil, 0, err
	}
	list := make([]model.Product, 0)
	queryArgs := append(append([]interface{}{}, args...), limit, offset)
	if err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM products"+where+" ORDER BY id ASC LIMIT ? OFFSET ?", queryArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ProductRepo) FindByID(ctx context.Context, id int64) (*model.Product, error) {
	var p model.Product
	err := r.db.GetContext(ctx, &p, "SELECT * FROM products WHERE id = ?", id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

type CartRepo struct{ db *sqlx.DB }

func NewCartRepo(db *sqlx.DB) *CartRepo { return &CartRepo{db: db} }

const cartSelect = `
	SELECT c.id, c.product_id, c.quantity,
	       p.name_cn, p.name_en, p.price_cents, p.image_url, p.stock, p.category
	FROM cart_items c JOIN products p ON p.id = c.product_id
	WHERE c.user_id = ? AND p.status = 1
	ORDER BY c.id ASC`

func (r *CartRepo) List(ctx context.Context, userID int64) ([]model.CartItem, error) {
	list := make([]model.CartItem, 0)
	err := r.db.SelectContext(ctx, &list, cartSelect, userID)
	return list, err
}

func (r *CartRepo) Add(ctx context.Context, userID, productID int64, quantity int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cart_items (user_id, product_id, quantity) VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE quantity = quantity + VALUES(quantity), updated_at = NOW()`,
		userID, productID, quantity)
	return err
}

func (r *CartRepo) UpdateQuantity(ctx context.Context, userID, productID int64, quantity int) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		"UPDATE cart_items SET quantity = ?, updated_at = NOW() WHERE user_id = ? AND product_id = ?",
		quantity, userID, productID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *CartRepo) Remove(ctx context.Context, userID, productID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM cart_items WHERE user_id = ? AND product_id = ?", userID, productID)
	return err
}

type OrderRepo struct{ db *sqlx.DB }

func NewOrderRepo(db *sqlx.DB) *OrderRepo { return &OrderRepo{db: db} }

const orderColumns = `id, order_no, user_id, total_cents, status, created_at, updated_at`

func (r *OrderRepo) FindByIdempotency(ctx context.Context, userID int64, key string) (*model.Order, error) {
	if key == "" {
		return nil, nil
	}
	var o model.Order
	err := r.db.GetContext(ctx, &o,
		"SELECT "+orderColumns+" FROM orders WHERE user_id = ? AND idempotency_key = ?", userID, key)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) FindByID(ctx context.Context, userID, orderID int64) (*model.Order, error) {
	var o model.Order
	err := r.db.GetContext(ctx, &o,
		"SELECT "+orderColumns+" FROM orders WHERE id = ? AND user_id = ?", orderID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.loadItems(ctx, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepo) FindPage(ctx context.Context, userID int64, offset, limit int) ([]model.Order, int, error) {
	total := 0
	if err := r.db.GetContext(ctx, &total,
		"SELECT COUNT(*) FROM orders WHERE user_id = ?", userID); err != nil {
		return nil, 0, err
	}
	list := make([]model.Order, 0)
	if err := r.db.SelectContext(ctx, &list,
		"SELECT "+orderColumns+" FROM orders WHERE user_id = ? ORDER BY id DESC LIMIT ? OFFSET ?",
		userID, limit, offset); err != nil {
		return nil, 0, err
	}
	for i := range list {
		if err := r.loadItems(ctx, &list[i]); err != nil {
			return nil, 0, err
		}
	}
	return list, total, nil
}

func (r *OrderRepo) loadItems(ctx context.Context, o *model.Order) error {
	items := make([]model.OrderItem, 0)
	if err := r.db.SelectContext(ctx, &items,
		"SELECT product_id, name_cn, name_en, price_cents, quantity, image_url FROM order_items WHERE order_id = ? ORDER BY id ASC",
		o.ID); err != nil {
		return err
	}
	o.Items = items
	return nil
}

// Create decrements stock and writes the order in one transaction. If the
// caller supplies an idempotency key and the order already exists, it returns
// the previous order instead of charging/creating twice.
func (r *OrderRepo) Create(ctx context.Context, userID int64, idemKey string, items []model.OrderItem, totalCents int, orderNo string) (*model.Order, error) {
	if idemKey != "" {
		var existing model.Order
		err := r.db.GetContext(ctx, &existing,
			"SELECT "+orderColumns+" FROM orders WHERE user_id = ? AND idempotency_key = ?", userID, idemKey)
		if err == nil {
			if err := r.loadItems(ctx, &existing); err != nil {
				return nil, err
			}
			return &existing, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, item := range items {
		res, err := tx.ExecContext(ctx,
			"UPDATE products SET stock = stock - ?, sold_count = sold_count + ?, updated_at = NOW() WHERE id = ? AND status = 1 AND stock >= ?",
			item.Quantity, item.Quantity, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return nil, ErrInsufficientStock
		}
	}

	res, err := tx.ExecContext(ctx,
		"INSERT INTO orders (order_no, user_id, total_cents, status, idempotency_key) VALUES (?, ?, ?, 'pending', ?)",
		orderNo, userID, totalCents, nullableKey(idemKey))
	if err != nil {
		// A concurrent retry may have committed the same idempotency key while
		// this transaction was running. Roll back the stock change and return
		// the winner's order.
		var mysqlErr *mysql.MySQLError
		if idemKey != "" && errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			_ = tx.Rollback()
			return r.FindByIdempotency(ctx, userID, idemKey)
		}
		return nil, err
	}
	orderID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO order_items (order_id, product_id, name_cn, name_en, price_cents, quantity, image_url) VALUES (?, ?, ?, ?, ?, ?, ?)",
			orderID, item.ProductID, item.NameCn, item.NameEn, item.PriceCents, item.Quantity, item.ImageURL); err != nil {
			return nil, err
		}
		if _, err := tx.ExecContext(ctx,
			"DELETE FROM cart_items WHERE user_id = ? AND product_id = ?", userID, item.ProductID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindByID(ctx, userID, orderID)
}

func nullableKey(key string) interface{} {
	if key == "" {
		return nil
	}
	return key
}

// AdminList returns orders across all users (used by the admin console).
func (r *OrderRepo) AdminList(ctx context.Context, status string, offset, limit int) ([]model.Order, int, error) {
	where := ""
	args := make([]interface{}, 0, 3)
	if status = strings.TrimSpace(status); status != "" && status != "all" {
		where = " WHERE status = ?"
		args = append(args, status)
	}
	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM orders"+where, args...); err != nil {
		return nil, 0, err
	}
	list := make([]model.Order, 0)
	queryArgs := append(append([]interface{}{}, args...), limit, offset)
	if err := r.db.SelectContext(ctx, &list,
		"SELECT "+orderColumns+" FROM orders"+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...); err != nil {
		return nil, 0, err
	}
	for i := range list {
		if err := r.loadItems(ctx, &list[i]); err != nil {
			return nil, 0, err
		}
	}
	return list, total, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, orderID int64, status string) (bool, error) {
	res, err := r.db.ExecContext(ctx,
		"UPDATE orders SET status = ?, updated_at = NOW() WHERE id = ?", status, orderID)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// --- admin ---------------------------------------------------------------

func (r *ProductRepo) AdminFindPage(ctx context.Context, q, category string, offset, limit int) ([]model.Product, int, error) {
	clauses := []string{"1=1"}
	args := make([]interface{}, 0, 6)
	if q = strings.TrimSpace(q); q != "" {
		like := "%" + q + "%"
		clauses = append(clauses, "(name_cn LIKE ? OR name_en LIKE ?)")
		args = append(args, like, like)
	}
	if category = strings.TrimSpace(category); category != "" && category != "all" {
		clauses = append(clauses, "category = ?")
		args = append(args, category)
	}
	where := " WHERE " + strings.Join(clauses, " AND ")
	total := 0
	if err := r.db.GetContext(ctx, &total, "SELECT COUNT(*) FROM products"+where, args...); err != nil {
		return nil, 0, err
	}
	list := make([]model.Product, 0)
	queryArgs := append(append([]interface{}{}, args...), limit, offset)
	if err := r.db.SelectContext(ctx, &list,
		"SELECT * FROM products"+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...); err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *ProductRepo) Insert(ctx context.Context, p *model.Product) error {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO products (name_cn, name_en, description_cn, description_en, price_cents, category, image_url, stock, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.NameCn, p.NameEn, p.DescriptionCn, p.DescriptionEn, p.PriceCents, p.Category, p.ImageURL, p.Stock, p.Status)
	if err != nil {
		return err
	}
	p.ID, err = res.LastInsertId()
	return err
}

func (r *ProductRepo) Update(ctx context.Context, p *model.Product) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE products SET name_cn = ?, name_en = ?, description_cn = ?, description_en = ?,
		 price_cents = ?, category = ?, image_url = ?, stock = ?, status = ?, updated_at = NOW()
		 WHERE id = ?`,
		p.NameCn, p.NameEn, p.DescriptionCn, p.DescriptionEn, p.PriceCents, p.Category, p.ImageURL, p.Stock, p.Status, p.ID)
	return err
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) (bool, error) {
	res, err := r.db.ExecContext(ctx, "UPDATE products SET status = 0, updated_at = NOW() WHERE id = ?", id)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

func (r *ProductRepo) FindReviews(ctx context.Context, productID int64, limit int) ([]model.ProductReview, error) {
	if limit <= 0 || limit > 20 {
		limit = 6
	}
	list := make([]model.ProductReview, 0)
	err := r.db.SelectContext(ctx, &list, `
		SELECT r.id, r.product_id, r.user_id, r.rating, r.content, r.created_at,
		       u.nickname AS author_nickname, u.avatar_url AS author_avatar_url
		FROM product_reviews r JOIN users u ON u.id = r.user_id
		WHERE r.product_id = ?
		ORDER BY r.id DESC
		LIMIT ?`, productID, limit)
	for i := range list {
		list[i].Author = &model.Author{
			ID:        list[i].UserID,
			Nickname:  list[i].AuthorNickname,
			AvatarURL: list[i].AuthorAvatarURL,
		}
	}
	return list, err
}
