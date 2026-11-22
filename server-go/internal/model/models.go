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
	Role         string    `db:"role" json:"role"`
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
	ID        int64      `db:"id" json:"id"`
	TitleCn   string     `db:"title_cn" json:"titleCn"`
	TitleEn   string     `db:"title_en" json:"titleEn"`
	ContentCn string     `db:"content_cn" json:"contentCn"`
	ContentEn string     `db:"content_en" json:"contentEn"`
	Tags      StringList `db:"tags" json:"tags"`
	CreatedAt time.Time  `db:"created_at" json:"-"`
	UpdatedAt time.Time  `db:"updated_at" json:"-"`
}

type Author struct {
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatarUrl"`
}

type TravelBlog struct {
	ID              int64      `db:"id" json:"id"`
	UserID          int64      `db:"user_id" json:"userId"`
	AuthorNickname  string     `db:"author_nickname" json:"-"`
	AuthorAvatarURL string     `db:"author_avatar_url" json:"-"`
	Author          *Author    `db:"-" json:"author"`
	Title           string     `db:"title" json:"title"`
	Content         string     `db:"content" json:"content"`
	Images          StringList `db:"images" json:"images"`
	Likes           int        `db:"likes" json:"likes"`
	LikedByMe       bool       `db:"liked_by_me" json:"likedByMe"`
	CommentCount    int        `db:"comment_count" json:"commentCount"`
	CreatedAt       time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updatedAt"`
}

type Comment struct {
	ID              int64     `db:"id" json:"id"`
	BlogID          int64     `db:"blog_id" json:"blogId"`
	UserID          int64     `db:"user_id" json:"userId"`
	Content         string    `db:"content" json:"content"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	AuthorNickname  string    `db:"author_nickname" json:"-"`
	AuthorAvatarURL string    `db:"author_avatar_url" json:"-"`
	Author          *Author   `db:"-" json:"author"`
}

type Notification struct {
	ID             int64     `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"-"`
	Type           string    `db:"type" json:"type"`
	IsRead         bool      `db:"is_read" json:"isRead"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	ActorID        int64     `db:"actor_id" json:"-"`
	ActorNickname  string    `db:"actor_nickname" json:"-"`
	ActorAvatarURL string    `db:"actor_avatar_url" json:"-"`
	Actor          *Author   `db:"-" json:"actor"`
	BlogID         int64     `db:"blog_id" json:"blogId,omitempty"`
	CommentID      int64     `db:"comment_id" json:"-"`
	BlogTitle      string    `db:"blog_title" json:"blogTitle,omitempty"`
	CommentContent string    `db:"comment_content" json:"commentContent,omitempty"`
}

type UserPublic struct {
	ID        int64     `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Nickname  string    `db:"nickname" json:"nickname"`
	AvatarURL string    `db:"avatar_url" json:"avatarUrl"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type ProfileStats struct {
	Posts         int `json:"posts"`
	Followers     int `json:"followers"`
	Following     int `json:"following"`
	LikesReceived int `json:"likesReceived"`
}

type UserProfile struct {
	User        UserPublic   `json:"user"`
	Stats       ProfileStats `json:"stats"`
	IsFollowing bool         `json:"isFollowing"`
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
	ID               int64     `db:"id" json:"id"`
	SessionID        int64     `db:"session_id" json:"sessionId"`
	RunID            *int64    `db:"run_id" json:"runId,omitempty"`
	Role             string    `db:"role" json:"role"`
	Content          string    `db:"content" json:"content"`
	ToolCalls        JSONText  `db:"tool_calls" json:"toolCalls,omitempty"`
	Status           string    `db:"status" json:"status"`
	Model            string    `db:"model" json:"model,omitempty"`
	PromptTokens     int       `db:"prompt_tokens" json:"promptTokens,omitempty"`
	CompletionTokens int       `db:"completion_tokens" json:"completionTokens,omitempty"`
	CreatedAt        time.Time `db:"created_at" json:"createdAt"`
}

type GuideSession struct {
	ID          int64          `db:"id" json:"id"`
	UserID      int64          `db:"user_id" json:"-"`
	Title       string         `db:"title" json:"title"`
	PlanJSON    JSONText       `db:"plan_json" json:"plan,omitempty"`
	PlanVersion int            `db:"plan_version" json:"planVersion"`
	LastRunID   *int64         `db:"last_run_id" json:"lastRunId,omitempty"`
	Messages    []GuideMessage `db:"-" json:"messages"`
	CreatedAt   time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time      `db:"updated_at" json:"-"`
}

// --- shop -----------------------------------------------------------------

type Product struct {
	ID            int64     `db:"id" json:"id"`
	NameCn        string    `db:"name_cn" json:"nameCn"`
	NameEn        string    `db:"name_en" json:"nameEn"`
	DescriptionCn string    `db:"description_cn" json:"descriptionCn"`
	DescriptionEn string    `db:"description_en" json:"descriptionEn"`
	PriceCents    int       `db:"price_cents" json:"priceCents"`
	Category      string    `db:"category" json:"category"`
	ImageURL      string    `db:"image_url" json:"imageUrl"`
	Stock         int       `db:"stock" json:"stock"`
	SoldCount     int       `db:"sold_count" json:"soldCount"`
	Rating        float64   `db:"rating" json:"rating"`
	ReviewCount   int       `db:"review_count" json:"reviewCount"`
	Status        bool      `db:"status" json:"active"`
	CreatedAt     time.Time `db:"created_at" json:"-"`
	UpdatedAt     time.Time `db:"updated_at" json:"-"`
}

type CartItem struct {
	ID         int64  `db:"id" json:"id"`
	ProductID  int64  `db:"product_id" json:"productId"`
	Quantity   int    `db:"quantity" json:"quantity"`
	NameCn     string `db:"name_cn" json:"nameCn"`
	NameEn     string `db:"name_en" json:"nameEn"`
	PriceCents int    `db:"price_cents" json:"priceCents"`
	ImageURL   string `db:"image_url" json:"imageUrl"`
	Stock      int    `db:"stock" json:"stock"`
	Category   string `db:"category" json:"category"`
}

type OrderItem struct {
	ProductID  int64  `db:"product_id" json:"productId"`
	NameCn     string `db:"name_cn" json:"nameCn"`
	NameEn     string `db:"name_en" json:"nameEn"`
	PriceCents int    `db:"price_cents" json:"priceCents"`
	Quantity   int    `db:"quantity" json:"quantity"`
	ImageURL   string `db:"image_url" json:"imageUrl"`
}

type Order struct {
	ID         int64       `db:"id" json:"id"`
	OrderNo    string      `db:"order_no" json:"orderNo"`
	UserID     int64       `db:"user_id" json:"-"`
	TotalCents int         `db:"total_cents" json:"totalCents"`
	Status     string      `db:"status" json:"status"`
	CreatedAt  time.Time   `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time   `db:"updated_at" json:"updatedAt"`
	Items      []OrderItem `db:"-" json:"items"`
}

type ProductReview struct {
	ID             int64     `db:"id" json:"id"`
	ProductID      int64     `db:"product_id" json:"productId"`
	UserID         int64     `db:"user_id" json:"userId"`
	Rating         int       `db:"rating" json:"rating"`
	Content        string    `db:"content" json:"content"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	AuthorNickname string    `db:"author_nickname" json:"-"`
	AuthorAvatarURL string   `db:"author_avatar_url" json:"-"`
	Author         *Author   `db:"-" json:"author"`
}

type Activity struct {
	ID            int64     `db:"id" json:"id"`
	TitleCn       string    `db:"title_cn" json:"titleCn"`
	TitleEn       string    `db:"title_en" json:"titleEn"`
	DescriptionCn string    `db:"description_cn" json:"descriptionCn"`
	DescriptionEn string    `db:"description_en" json:"descriptionEn"`
	Location      string    `db:"location" json:"location"`
	StartAt       time.Time `db:"start_at" json:"startAt"`
	EndAt         time.Time `db:"end_at" json:"endAt"`
	ImageURL      string    `db:"image_url" json:"imageUrl"`
	Organizer     string    `db:"organizer" json:"organizer"`
	Tag           string    `db:"tag" json:"tag"`
	Status        string    `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"-"`
	UpdatedAt     time.Time `db:"updated_at" json:"-"`
}
