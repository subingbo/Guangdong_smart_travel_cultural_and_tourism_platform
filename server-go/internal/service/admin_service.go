package service

import (
	"context"
	"strings"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type AdminService struct {
	products  *repo.ProductRepo
	knowledge *repo.KnowledgeRepo
	shop      *ShopService
}

func NewAdminService(products *repo.ProductRepo, knowledge *repo.KnowledgeRepo, shop *ShopService) *AdminService {
	return &AdminService{products: products, knowledge: knowledge, shop: shop}
}

type ProductInput struct {
	NameCn        string `json:"nameCn"`
	NameEn        string `json:"nameEn"`
	DescriptionCn string `json:"descriptionCn"`
	DescriptionEn string `json:"descriptionEn"`
	PriceCents    int    `json:"priceCents"`
	Category      string `json:"category"`
	ImageURL      string `json:"imageUrl"`
	Stock         int    `json:"stock"`
	Active        bool   `json:"active"`
}

func (s *AdminService) ListProducts(ctx context.Context, q, category string, page, pageSize int) (*ProductPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.products.AdminFindPage(ctx, q, category, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &ProductPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *AdminService) CreateProduct(ctx context.Context, in ProductInput) (*model.Product, error) {
	p, err := productFromInput(in)
	if err != nil {
		return nil, err
	}
	if err := s.products.Insert(ctx, p); err != nil {
		return nil, err
	}
	return s.products.FindByID(ctx, p.ID)
}

func (s *AdminService) UpdateProduct(ctx context.Context, id int64, in ProductInput) (*model.Product, error) {
	existing, err := s.products.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, common.NotFound("Product not found")
	}
	p, err := productFromInput(in)
	if err != nil {
		return nil, err
	}
	p.ID = id
	if err := s.products.Update(ctx, p); err != nil {
		return nil, err
	}
	return s.products.FindByID(ctx, id)
}

func (s *AdminService) DeleteProduct(ctx context.Context, id int64) error {
	ok, err := s.products.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return common.NotFound("Product not found")
	}
	return nil
}

func productFromInput(in ProductInput) (*model.Product, error) {
	in.NameCn = strings.TrimSpace(in.NameCn)
	in.NameEn = strings.TrimSpace(in.NameEn)
	in.Category = strings.TrimSpace(in.Category)
	if in.NameCn == "" || in.NameEn == "" {
		return nil, common.BadRequest("nameCn and nameEn are required")
	}
	if len([]rune(in.NameCn)) > 120 || len([]rune(in.NameEn)) > 160 {
		return nil, common.BadRequest("product name is too long")
	}
	if in.PriceCents < 0 {
		return nil, common.BadRequest("priceCents must be >= 0")
	}
	if in.Stock < 0 {
		return nil, common.BadRequest("stock must be >= 0")
	}
	switch in.Category {
	case "ticket", "snack", "souvenir", "experience", "coupon":
	default:
		return nil, common.BadRequest("invalid category")
	}
	return &model.Product{
		NameCn:        in.NameCn,
		NameEn:        in.NameEn,
		DescriptionCn: strings.TrimSpace(in.DescriptionCn),
		DescriptionEn: strings.TrimSpace(in.DescriptionEn),
		PriceCents:    in.PriceCents,
		Category:      in.Category,
		ImageURL:      strings.TrimSpace(in.ImageURL),
		Stock:         in.Stock,
		Status:        in.Active,
	}, nil
}

type KnowledgeInput struct {
	TitleCn   string   `json:"titleCn"`
	TitleEn   string   `json:"titleEn"`
	ContentCn string   `json:"contentCn"`
	ContentEn string   `json:"contentEn"`
	Tags      []string `json:"tags"`
}

func (s *AdminService) ListKnowledge(ctx context.Context, q string, page, pageSize int) ([]model.Knowledge, int, int, int, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.knowledge.FindPage(ctx, q, offset, pageSize)
	return items, total, page, pageSize, err
}

func (s *AdminService) CreateKnowledge(ctx context.Context, in KnowledgeInput) (*model.Knowledge, error) {
	k, err := knowledgeFromInput(in)
	if err != nil {
		return nil, err
	}
	if err := s.knowledge.Insert(ctx, k); err != nil {
		return nil, err
	}
	return s.knowledge.FindByID(ctx, k.ID)
}

func (s *AdminService) UpdateKnowledge(ctx context.Context, id int64, in KnowledgeInput) (*model.Knowledge, error) {
	existing, err := s.knowledge.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, common.NotFound("Knowledge entry not found")
	}
	k, err := knowledgeFromInput(in)
	if err != nil {
		return nil, err
	}
	k.ID = id
	if err := s.knowledge.Update(ctx, k); err != nil {
		return nil, err
	}
	return s.knowledge.FindByID(ctx, id)
}

func (s *AdminService) DeleteKnowledge(ctx context.Context, id int64) error {
	ok, err := s.knowledge.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !ok {
		return common.NotFound("Knowledge entry not found")
	}
	return nil
}

func knowledgeFromInput(in KnowledgeInput) (*model.Knowledge, error) {
	in.TitleCn = strings.TrimSpace(in.TitleCn)
	in.TitleEn = strings.TrimSpace(in.TitleEn)
	if in.TitleCn == "" || in.TitleEn == "" {
		return nil, common.BadRequest("titleCn and titleEn are required")
	}
	if len([]rune(in.TitleCn)) > 100 || len([]rune(in.TitleEn)) > 100 {
		return nil, common.BadRequest("title is too long")
	}
	tags := make([]string, 0, len(in.Tags))
	for _, tag := range in.Tags {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	return &model.Knowledge{
		TitleCn:   in.TitleCn,
		TitleEn:   in.TitleEn,
		ContentCn: strings.TrimSpace(in.ContentCn),
		ContentEn: strings.TrimSpace(in.ContentEn),
		Tags:      tags,
	}, nil
}

func (s *AdminService) ListOrders(ctx context.Context, status string, page, pageSize int) (*OrderPage, error) {
	return s.shop.AdminListOrders(ctx, status, page, pageSize)
}

func (s *AdminService) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	return s.shop.AdminUpdateOrderStatus(ctx, id, status)
}
