package model

import (
	"time"
)

type User struct {
	ID           int64     `db:"id" json:"id"`
	Username     string    `db:"username" json:"username"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Nickname     string    `db:"nickname" json:"nickname"`
	AvatarURL    string    `db:"avatar_url" json:"avatarUrl"`
	Language     string    `db:"language" json:"language"`
	CreatedAt    time.Time `db:"created_at" json:"-"`
	UpdatedAt    time.Time `db:"updated_at" json:"-"`
}

type ScenicImage struct {
	ID            int64     `db:"id" json:"id"`
	ImageURL      string    `db:"image_url" json:"imageUrl"`
	TitleCn       string    `db:"title_cn" json:"titleCn"`
	TitleEn       string    `db:"title_en" json:"titleEn"`
	Location      string    `db:"location" json:"location"`
	Tag           string    `db:"tag" json:"tag"`
	DescriptionCn string    `db:"description_cn" json:"descriptionCn"`
	DescriptionEn string    `db:"description_en" json:"descriptionEn"`
	SortOrder     int       `db:"sort_order" json:"-"`
	CreatedAt     time.Time `db:"created_at" json:"-"`
	UpdatedAt     time.Time `db:"updated_at" json:"-"`
}

type Place struct {
	ID            int64      `db:"id" json:"id"`
	NameCn        string     `db:"name_cn" json:"nameCn"`
	NameEn        string     `db:"name_en" json:"nameEn"`
	Lat           float64    `db:"lat" json:"lat"`
	Lng           float64    `db:"lng" json:"lng"`
	Tags          StringList `db:"tags" json:"tags"`
	Indoor        bool       `db:"indoor" json:"indoor"`
	Duration      int        `db:"duration" json:"duration"`
	DescriptionCn string     `db:"description_cn" json:"descriptionCn"`
	DescriptionEn string     `db:"description_en" json:"descriptionEn"`
	CreatedAt     time.Time  `db:"created_at" json:"-"`
	UpdatedAt     time.Time  `db:"updated_at" json:"-"`
}

type Knowledge struct {
	ID        int64      `db:"id" json:"-"`
	TitleCn   string     `db:"title_cn" json:"-"`
	TitleEn   string     `db:"title_en" json:"-"`
	ContentCn string     `db:"content_cn" json:"-"`
	ContentEn string     `db:"content_en" json:"-"`
	Tags      StringList `db:"tags" json:"-"`
	CreatedAt time.Time  `db:"created_at" json:"-"`
	UpdatedAt time.Time  `db:"updated_at" json:"-"`
}

type TravelBlog struct {
	ID              int64      `db:"id" json:"id"`
	UserID          int64      `db:"user_id" json:"userId"`
	AuthorNickname  string     `db:"author_nickname" json:"-"`
	AuthorAvatarURL string     `db:"author_avatar_url" json:"-"`
	Title           string     `db:"title" json:"title"`
	Content         string     `db:"content" json:"content"`
	Images          StringList `db:"images" json:"images"`
	Likes           int        `db:"likes" json:"likes"`
	LikedByMe       bool       `db:"liked_by_me" json:"likedByMe"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
}

type Waypoint struct {
	Name        string  `json:"name"`
	NameEn      string  `json:"nameEn"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	StayTime    string  `json:"stayTime"`
	Description string  `json:"description"`
}

type Coordinate struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Route struct {
	ID            int64          `db:"id" json:"id"`
	UserID        int64          `db:"user_id" json:"-"`
	RouteName     string         `db:"route_name" json:"routeName"`
	DurationDays  int            `db:"duration_days" json:"durationDays"`
	TotalDistance string         `db:"total_distance" json:"totalDistance"`
	Mode          string         `db:"mode" json:"mode"`
	Pace          string         `db:"pace" json:"pace"`
	Interests     StringList     `db:"interests" json:"interests"`
	Waypoints     WaypointList   `db:"waypoints" json:"waypoints"`
	Polyline      CoordinateList `db:"polyline" json:"polyline"`
	CreatedAt     time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updatedAt"`
}

type GuideMessage struct {
	ID        int64     `db:"id" json:"id"`
	SessionID int64     `db:"session_id" json:"sessionId"`
	Role      string    `db:"role" json:"role"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type GuideSession struct {
	ID        int64          `db:"id" json:"id"`
	UserID    int64          `db:"user_id" json:"-"`
	Title     string         `db:"title" json:"title"`
	Messages  []GuideMessage `db:"-" json:"messages"`
	CreatedAt time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time      `db:"updated_at" json:"-"`
}
