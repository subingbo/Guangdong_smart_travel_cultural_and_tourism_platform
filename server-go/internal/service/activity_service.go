package service

import (
	"context"

	"github.com/travelgd/server-go/internal/common"
	"github.com/travelgd/server-go/internal/model"
	"github.com/travelgd/server-go/internal/repo"
)

type ActivityPage struct {
	Items    []model.Activity `json:"items"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"pageSize"`
}

type ActivityService struct {
	activities *repo.ActivityRepo
}

func NewActivityService(activities *repo.ActivityRepo) *ActivityService {
	return &ActivityService{activities: activities}
}

func (s *ActivityService) List(ctx context.Context, status, q string, page, pageSize int) (*ActivityPage, error) {
	page, pageSize, offset := normalizePage(page, pageSize)
	items, total, err := s.activities.FindPage(ctx, status, q, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return &ActivityPage{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *ActivityService) Get(ctx context.Context, id int64) (*model.Activity, error) {
	activity, err := s.activities.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if activity == nil {
		return nil, common.NotFound("Activity not found")
	}
	return activity, nil
}
