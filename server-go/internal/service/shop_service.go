package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type ProductPage struct {
	Items    []model.Product `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type CartView struct {
	Items      []model.CartItem `json:"items"`
	TotalCents int              `json:"totalCents"`
	Count      int              `json:"count"`
}

type OrderPage struct {
	Items    []model.Order `json:"items"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

type ShopService struct {
	products *repo.ProductRepo
	cart     *repo.CartRepo
	orders   *repo.OrderRepo
}

func NewShopService(products *repo.ProductRepo, cart *repo.CartRepo, orders *repo.OrderRepo) *ShopService {
	return &ShopService{products: products, cart: cart, orders: orders}
}

func (s *ShopService) ListProducts(ctx context.Context, category, q string, page, pageSize int) (*ProductPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.products.FindPage(ctx, category, q, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &ProductPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ShopService) GetProduct(ctx context.Context, id int64) (*model.Product, error) {
	p, err := s.products.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil || !p.Status {
		return nil, common.NotFound("Product not found")
	}
	return p, nil
}

func (s *ShopService) GetCart(ctx context.Context, userID int64) (*CartView, error) {
	items, err := s.cart.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return cartView(items), nil
}

func (s *ShopService) AddToCart(ctx context.Context, userID, productID int64, quantity int) (*CartView, error) {
	if quantity < 1 || quantity > 99 {
		return nil, common.BadRequest("quantity must be between 1 and 99")
	}
	p, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil || !p.Status {
		return nil, common.NotFound("Product not found")
	}
	if p.Stock <= 0 {
		return nil, common.Conflict("Product is out of stock")
	}
	items, err := s.cart.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	current := 0
	for _, item := range items {
		if item.ProductID == productID {
			current = item.Quantity
			break
		}
	}
	if current+quantity > p.Stock {
		return nil, common.Conflict(fmt.Sprintf("Only %d left in stock", p.Stock))
	}
	if err := s.cart.Add(ctx, userID, productID, quantity); err != nil {
		return nil, err
	}
	return s.GetCart(ctx, userID)
}

func (s *ShopService) UpdateCart(ctx context.Context, userID, productID int64, quantity int) (*CartView, error) {
	if quantity <= 0 {
		if err := s.cart.Remove(ctx, userID, productID); err != nil {
			return nil, err
		}
		return s.GetCart(ctx, userID)
	}
	if quantity > 99 {
		return nil, common.BadRequest("quantity must be at most 99")
	}
	p, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil || !p.Status {
		return nil, common.NotFound("Product not found")
	}
	if quantity > p.Stock {
		return nil, common.Conflict(fmt.Sprintf("Only %d left in stock", p.Stock))
	}
	ok, err := s.cart.UpdateQuantity(ctx, userID, productID, quantity)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, common.NotFound("Cart item not found")
	}
	return s.GetCart(ctx, userID)
}

func (s *ShopService) RemoveFromCart(ctx context.Context, userID, productID int64) (*CartView, error) {
	if err := s.cart.Remove(ctx, userID, productID); err != nil {
		return nil, err
	}
	return s.GetCart(ctx, userID)
}

// Checkout turns the current cart into a pending order, decrementing stock in
// the same transaction. Payment is intentionally out of scope for the demo.
func (s *ShopService) Checkout(ctx context.Context, userID int64, idempotencyKey string) (*model.Order, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if len(idempotencyKey) > 64 {
		idempotencyKey = idempotencyKey[:64]
	}
	// A retried request must return the original order even though the cart
	// was already cleared by the first successful checkout.
	if existing, err := s.orders.FindByIdempotency(ctx, userID, idempotencyKey); err != nil {
		return nil, err
	} else if existing != nil {
		return existing, nil
	}
	items, err := s.cart.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, common.BadRequest("Cart is empty")
	}
	orderItems := make([]model.OrderItem, 0, len(items))
	total := 0
	for _, item := range items {
		if item.Quantity > item.Stock {
			return nil, common.Conflict(fmt.Sprintf("Only %d left for %s", item.Stock, item.NameEn))
		}
		orderItems = append(orderItems, model.OrderItem{
			ProductID:  item.ProductID,
			NameCn:     item.NameCn,
			NameEn:     item.NameEn,
			PriceCents: item.PriceCents,
			Quantity:   item.Quantity,
			ImageURL:   item.ImageURL,
		})
		total += item.PriceCents * item.Quantity
	}
	orderNo := fmt.Sprintf("TG%s%04d", time.Now().Format("20060102150405"), time.Now().Nanosecond()/1e6)
	order, err := s.orders.Create(ctx, userID, idempotencyKey, orderItems, total, orderNo)
	if err != nil {
		if errors.Is(err, repo.ErrInsufficientStock) {
			return nil, common.Conflict("One or more items just sold out; please refresh your cart")
		}
		return nil, err
	}
	return order, nil
}

func (s *ShopService) ListOrders(ctx context.Context, userID int64, page, pageSize int) (*OrderPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.orders.FindPage(ctx, userID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &OrderPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ShopService) GetOrder(ctx context.Context, userID, orderID int64) (*model.Order, error) {
	order, err := s.orders.FindByID(ctx, userID, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, common.NotFound("Order not found")
	}
	return order, nil
}

func cartView(items []model.CartItem) *CartView {
	total := 0
	count := 0
	for _, item := range items {
		total += item.PriceCents * item.Quantity
		count += item.Quantity
	}
	return &CartView{Items: items, TotalCents: total, Count: count}
}

// --- admin ---------------------------------------------------------------

func (s *ShopService) AdminListOrders(ctx context.Context, status string, page, pageSize int) (*OrderPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.orders.AdminList(ctx, status, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &OrderPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ShopService) AdminUpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	switch status {
	case "pending", "paid", "completed", "canceled":
	default:
		return common.BadRequest("invalid order status")
	}
	ok, err := s.orders.UpdateStatus(ctx, orderID, status)
	if err != nil {
		return err
	}
	if !ok {
		return common.NotFound("Order not found")
	}
	return nil
}

func (s *ShopService) Reviews(ctx context.Context, productID int64) ([]model.ProductReview, error) {
	if _, err := s.GetProduct(ctx, productID); err != nil {
		return nil, err
	}
	return s.products.FindReviews(ctx, productID, 8)
}
