package service

import (
	"context"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type CatalogService struct {
	scenics *repo.ScenicRepo
	places  *repo.PlaceRepo
}

func NewCatalogService(scenics *repo.ScenicRepo, places *repo.PlaceRepo) *CatalogService {
	return &CatalogService{scenics: scenics, places: places}
}

func (s *CatalogService) ListScenics(ctx context.Context) ([]model.ScenicImage, error) {
	return s.scenics.FindAll(ctx)
}

func (s *CatalogService) GetScenic(ctx context.Context, id int64) (*model.ScenicImage, error) {
	scenic, err := s.scenics.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if scenic == nil {
		return nil, common.NotFound("Scenic image not found")
	}
	return scenic, nil
}

func (s *CatalogService) ListPlaces(ctx context.Context, tag string) ([]model.Place, error) {
	return s.places.FindAll(ctx, tag)
}

func (s *CatalogService) GetPlace(ctx context.Context, id int64) (*model.Place, error) {
	place, err := s.places.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if place == nil {
		return nil, common.NotFound("Place not found")
	}
	return place, nil
}

type FavoriteService struct {
	favorites *repo.FavoriteRepo
	catalog   *CatalogService
}

func NewFavoriteService(favorites *repo.FavoriteRepo, catalog *CatalogService) *FavoriteService {
	return &FavoriteService{favorites: favorites, catalog: catalog}
}

func (s *FavoriteService) List(ctx context.Context, userID int64) ([]model.Place, error) {
	return s.favorites.FindPlacesByUser(ctx, userID)
}

func (s *FavoriteService) Add(ctx context.Context, userID, placeID int64) error {
	if _, err := s.catalog.GetPlace(ctx, placeID); err != nil {
		return err
	}
	exists, err := s.favorites.Exists(ctx, userID, placeID)
	if err != nil {
		return err
	}
	if exists {
		return common.Conflict("Place already favorited")
	}
	return s.favorites.Insert(ctx, userID, placeID)
}

func (s *FavoriteService) Remove(ctx context.Context, userID, placeID int64) error {
	return s.favorites.Delete(ctx, userID, placeID)
}
