package service

import "github.com/travelgd/server-go/internal/model"

const (
	defaultPageSize = 10
	maxPageSize     = 50
)

// normalizePage clamps the 1-based page and page size used by list endpoints.
func normalizePage(page, pageSize int) (int, int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > maxPageSize {
		pageSize = defaultPageSize
	}
	return page, pageSize, (page - 1) * pageSize
}

type BlogPage struct {
	Items    []model.TravelBlog `json:"items"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

type CommentPage struct {
	Items    []model.Comment `json:"items"`
	Total    int             `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

type UserPage struct {
	Items    []model.UserPublic `json:"items"`
	Total    int                `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"pageSize"`
}

type ConversationPage struct {
	Items    []model.GuideSession `json:"items"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

type NotificationPage struct {
	Items    []model.Notification `json:"items"`
	Total    int                  `json:"total"`
	Unread   int                  `json:"unread"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}
