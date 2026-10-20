// Travel Guangdong tourist API - Go implementation (Gin + sqlx).
// Implements the same REST contract as server-java; run on port 8081 so the
// two backends can run side by side against one MySQL database.
package main

import (
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/travelgd/server-go/internal/ai"
	"github.com/travelgd/server-go/internal/auth"
	"github.com/travelgd/server-go/internal/handler"
	"github.com/travelgd/server-go/internal/repo"
	"github.com/travelgd/server-go/internal/service"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	dbHost := env("DB_HOST", "localhost")
	dbPort := env("DB_PORT", "3306")
	dbUser := env("DB_USER", "travelgd")
	dbPass := env("DB_PASSWORD", "travelgd")
	dbName := env("DB_NAME", "travel_gd")
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName +
		"?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"

	db := sqlx.MustConnect("mysql", dsn)

	expireHours, _ := strconv.Atoi(env("JWT_EXPIRE_HOURS", "72"))
	jwtSvc := auth.NewJwtService(
		env("JWT_SECRET", "change-me-to-a-long-random-secret-at-least-32-chars"), expireHours)

	qwen := ai.NewQwenClient(
		env("DASHSCOPE_API_KEY", ""),
		env("AI_CHAT_URL", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"))
	mapAPI := ai.NewTencentRouteClient(env("TENCENT_MAP_API_KEY", ""),
		env("MAP_ROUTE_URL", "https://apis.map.qq.com/ws/direction/v1"))

	// Repositories (= MyBatis mappers)
	userRepo := repo.NewUserRepo(db)
	scenicRepo := repo.NewScenicRepo(db)
	placeRepo := repo.NewPlaceRepo(db)
	knowledgeRepo := repo.NewKnowledgeRepo(db)
	favoriteRepo := repo.NewFavoriteRepo(db)
	blogRepo := repo.NewBlogRepo(db)
	routeRepo := repo.NewRouteRepo(db)
	guideRepo := repo.NewGuideRepo(db)

	// Services (same layering as the Java project)
	authSvc := service.NewAuthService(userRepo, jwtSvc)
	meSvc := service.NewMeService(authSvc, userRepo, favoriteRepo, routeRepo, blogRepo)
	catalogSvc := service.NewCatalogService(scenicRepo, placeRepo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo, catalogSvc)
	routeSvc := service.NewRouteService(routeRepo, qwen, mapAPI)
	blogSvc := service.NewBlogService(blogRepo, qwen)
	guideSvc := service.NewGuideService(guideRepo, knowledgeRepo, qwen)

	// Handlers (= controllers)
	authH := handler.NewAuthHandler(authSvc)
	meH := handler.NewMeHandler(meSvc)
	catalogH := handler.NewCatalogHandler(catalogSvc)
	favoriteH := handler.NewFavoriteHandler(favoriteSvc)
	routeH := handler.NewRouteHandler(routeSvc)
	blogH := handler.NewBlogHandler(blogSvc)
	guideH := handler.NewGuideHandler(guideSvc)

	r := gin.Default()
	api := r.Group("/api")
	{
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)

		api.GET("/scenics", catalogH.ListScenics)
		api.GET("/scenics/:id", catalogH.GetScenic)
		api.GET("/places", catalogH.ListPlaces)
		api.GET("/places/:id", catalogH.GetPlace)

		optional := api.Group("", handler.OptionalAuth(jwtSvc))
		optional.GET("/blogs", blogH.Feed)
		optional.GET("/blogs/:id", blogH.Detail)
		api.POST("/blogs/:id/like", blogH.Like)

		protected := api.Group("", handler.RequireAuth(jwtSvc))
		protected.GET("/me", meH.Profile)
		protected.PUT("/me", meH.Update)
		protected.GET("/me/stats", meH.Stats)
		protected.GET("/favorites", favoriteH.List)
		protected.POST("/favorites", favoriteH.Add)
		protected.DELETE("/favorites/:placeId", favoriteH.Remove)
		protected.POST("/routes/generate", routeH.Generate)
		protected.GET("/routes", routeH.List)
		protected.GET("/routes/:id", routeH.Detail)
		protected.POST("/blogs", blogH.Publish)
		protected.POST("/blogs/draft", blogH.Draft)
		protected.POST("/guide/chat", guideH.Chat)
		protected.GET("/guide/sessions/:id", guideH.Session)
	}

	port := env("PORT", "8081")
	_ = r.Run(":" + port)
}
