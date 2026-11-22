// Travel Guangdong tourist API. Gin + sqlx; listens on :8081 by default.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/travelgd/server-go/internal/agent"
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

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
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
	// The agent runs blocking LLM/AMap calls inside goroutines; keep the
	// database pool as the hard ceiling rather than letting it grow unbounded.
	db.SetMaxOpenConns(envInt("DB_MAX_OPEN_CONNS", 10))
	db.SetMaxIdleConns(envInt("DB_MAX_IDLE_CONNS", 5))
	db.SetConnMaxLifetime(time.Duration(envInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	expireHours, _ := strconv.Atoi(env("JWT_EXPIRE_HOURS", "72"))
	jwtSecret := env("JWT_SECRET", "change-me-to-a-long-random-secret-at-least-32-chars")
	if os.Getenv("GIN_MODE") == "release" &&
		(len(jwtSecret) < 32 || jwtSecret == "change-me-to-a-long-random-secret-at-least-32-chars") {
		log.Fatal("JWT_SECRET must be set to a strong, unique value in release mode")
	}
	jwtSvc := auth.NewJwtService(jwtSecret, expireHours)

	qwen := ai.NewQwenClient(
		env("DASHSCOPE_API_KEY", ""),
		env("AI_CHAT_URL", "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation/generation"),
		env("AI_MODEL", "qwen-plus"))

	visionClient := ai.NewVisionClient(
		env("DASHSCOPE_API_KEY", ""),
		env("AI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
		env("AI_VISION_MODEL", "qwen3-vl-plus"))
	wanxClient := ai.NewWanxClient(env("DASHSCOPE_API_KEY", ""))

	// Prefer AMap. Tencent stays as a fallback for environments that only have
	// the old key: AMap JS API keys cannot call the Web service API.
	var amapClient *ai.AmapClient
	var weatherClient *ai.WeatherClient
	var mapAPI ai.RouteProvider
	if amapKey := env("AMAP_WEB_KEY", ""); amapKey != "" {
		amapClient = ai.NewAmapClient(
			amapKey,
			env("AMAP_ROUTE_URL", "https://restapi.amap.com/v3/direction"),
			env("AMAP_TRANSIT_CITY", "广州"))
		weatherClient = ai.NewWeatherClient(
			amapKey,
			env("AMAP_WEATHER_URL", "https://restapi.amap.com/v3/weather/weatherInfo"))
		mapAPI = amapClient
	} else {
		mapAPI = ai.NewTencentRouteClient(
			env("TENCENT_MAP_API_KEY", ""),
			env("MAP_ROUTE_URL", "https://apis.map.qq.com/ws/direction/v1"))
	}

	// Repositories
	userRepo := repo.NewUserRepo(db)
	scenicRepo := repo.NewScenicRepo(db)
	placeRepo := repo.NewPlaceRepo(db)
	knowledgeRepo := repo.NewKnowledgeRepo(db)
	favoriteRepo := repo.NewFavoriteRepo(db)
	blogRepo := repo.NewBlogRepo(db)
	routeRepo := repo.NewRouteRepo(db)
	guideRepo := repo.NewGuideRepo(db)
	commentRepo := repo.NewCommentRepo(db)
	followRepo := repo.NewFollowRepo(db)
	notificationRepo := repo.NewNotificationRepo(db)
	productRepo := repo.NewProductRepo(db)
	cartRepo := repo.NewCartRepo(db)
	orderRepo := repo.NewOrderRepo(db)
	activityRepo := repo.NewActivityRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, jwtSvc,
		env("WX_APPID", ""), env("WX_SECRET", ""))
	notificationSvc := service.NewNotificationService(notificationRepo)
	meSvc := service.NewMeService(authSvc, userRepo, favoriteRepo, routeRepo, blogRepo, followRepo, notificationRepo)
	catalogSvc := service.NewCatalogService(scenicRepo, placeRepo)
	favoriteSvc := service.NewFavoriteService(favoriteRepo, catalogSvc)
	routeSvc := service.NewRouteService(routeRepo, qwen, mapAPI)
	blogSvc := service.NewBlogService(blogRepo, qwen, notificationSvc)
	commentSvc := service.NewCommentService(commentRepo, blogRepo, notificationSvc)
	userSvc := service.NewUserService(userRepo, blogRepo, followRepo, blogSvc, notificationSvc)
	guideSvc := service.NewGuideService(guideRepo, knowledgeRepo, qwen)
	aiSvc := service.NewAiService(qwen, visionClient, wanxClient)
	shopSvc := service.NewShopService(productRepo, cartRepo, orderRepo)
	activitySvc := service.NewActivityService(activityRepo)
	adminSvc := service.NewAdminService(productRepo, knowledgeRepo, shopSvc)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	meH := handler.NewMeHandler(meSvc)
	catalogH := handler.NewCatalogHandler(catalogSvc)
	favoriteH := handler.NewFavoriteHandler(favoriteSvc)
	routeH := handler.NewRouteHandler(routeSvc)
	blogH := handler.NewBlogHandler(blogSvc)
	commentH := handler.NewCommentHandler(commentSvc)
	userH := handler.NewUserHandler(userSvc)
	notificationH := handler.NewNotificationHandler(notificationSvc)
	uploadH := handler.NewUploadHandler(env("UPLOAD_DIR", "/opt/travel-gd/uploads"))
	guideH := handler.NewGuideHandler(guideSvc)
	aiH := handler.NewAiHandler(aiSvc)
	shopH := handler.NewShopHandler(shopSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	adminH := handler.NewAdminHandler(adminSvc)

	// --- AI agent wiring -------------------------------------------------
	// The legacy QwenClient above stays for /guide/chat and /routes/generate.
	// The agent uses the OpenAI-compatible endpoint because it needs tools,
	// streaming deltas and usage accounting.
	agentEnabled := env("AGENT_ENABLED", "true") == "true"
	agentDebug := env("AGENT_DEBUG", "false") == "true"
	var agentLockConn *sql.Conn
	if agentEnabled {
		agentLockConn = acquireAgentLock(db)
		if agentLockConn == nil {
			agentEnabled = false
			log.Printf("agent disabled: another instance holds travel_gd_agent_singleton")
		}
	}
	openaiClient := ai.NewOpenAIClient(
		env("DASHSCOPE_API_KEY", ""),
		env("AI_BASE_URL", "https://dashscope.aliyuncs.com/compatible-mode/v1"),
		env("AI_MODEL", "qwen-plus"),
		env("AI_STREAM", "true") == "true")
	if !openaiClient.Enabled() {
		agentEnabled = false
	}

	agentRepo := repo.NewAgentRepo(db)
	travelSvc := ai.NewTravelService(amapClient, envInt("AGENT_MAP_CONCURRENCY", 8))
	planSvc := service.NewAgentPlanService(agentRepo, guideRepo, placeRepo, routeRepo)
	agentTools := service.NewAgentTools(placeRepo, favoriteRepo, knowledgeRepo, routeRepo, travelSvc, planSvc, weatherClient)
	registry, registryErr := agent.NewRegistry(agentTools.All()...)
	if registryErr != nil {
		log.Fatalf("agent tools: %v", registryErr)
	}
	runner := agent.NewRunner(openaiClient, registry, agent.LoadLimits())

	maxRuns := envInt("AGENT_MAX_CONCURRENT_RUNS", 4)
	queueCapacity := envInt("AGENT_QUEUE_CAPACITY", 16)
	queueTimeout := time.Duration(envInt("AGENT_QUEUE_TIMEOUT_SECONDS", 30)) * time.Second
	var runExecutor agent.Executor
	if env("AGENT_EXECUTOR_MODE", "spawn") == "fixed" {
		runExecutor = agent.NewFixedExecutor(maxRuns, queueCapacity, queueTimeout)
	} else {
		runExecutor = agent.NewSpawnExecutor(maxRuns, queueCapacity, queueTimeout)
	}
	runSvc := service.NewAgentRunService(db, guideRepo, agentRepo, userRepo, runner, runExecutor, service.AgentRunConfig{
		Enabled:            agentEnabled,
		Model:              env("AI_MODEL", "qwen-plus"),
		MaxConcurrentRuns:  maxRuns,
		QueueCapacity:      queueCapacity,
		QueueTimeout:       queueTimeout,
		RunTimeout:         time.Duration(envInt("AGENT_RUN_TIMEOUT_SECONDS", 300)) * time.Second,
		HistoryMessages:    envInt("AGENT_HISTORY_MESSAGES", 12),
		MaxOutputBytes:     envInt("AGENT_MAX_OUTPUT_BYTES", 64*1024),
		RateLimitPerMinute: envInt("AGENT_RATE_LIMIT_PER_MINUTE", 10),
		Debug:              agentDebug,
	})
	recoverCtx, recoverCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if agentEnabled {
		if err := runSvc.RecoverInterrupted(recoverCtx); err != nil {
			log.Printf("agent recovery: %v", err)
		}
	}
	recoverCancel()
	shutdownFns = append(shutdownFns, runSvc.Shutdown)
	if agentLockConn != nil {
		shutdownFns = append(shutdownFns, func(ctx context.Context) {
			_, _ = agentLockConn.ExecContext(ctx, "SELECT RELEASE_LOCK('travel_gd_agent_singleton')")
			_ = agentLockConn.Close()
		})
	}
	agentH := handler.NewAgentHandler(runSvc, planSvc, guideSvc, agentDebug)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "database": "down"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "agent": agentEnabled, "ai": qwen.Enabled()})
	})
	// Uploaded images live under UPLOAD_DIR. Production serves /uploads via
	// nginx; this static route keeps local development self-contained.
	r.Static("/uploads", env("UPLOAD_DIR", "/opt/travel-gd/uploads"))
	api := r.Group("/api")
	{
		api.POST("/auth/register", authH.Register)
		api.POST("/auth/login", authH.Login)
		api.POST("/auth/wechat-login", authH.WechatLogin)

		api.GET("/scenics", catalogH.ListScenics)
		api.GET("/scenics/:id", catalogH.GetScenic)
		api.GET("/places", catalogH.ListPlaces)
		api.GET("/places/:id", catalogH.GetPlace)
		api.GET("/shop/products", shopH.Products)
		api.GET("/shop/products/:id", shopH.Product)
		api.GET("/shop/products/:id/reviews", shopH.Reviews)
		api.GET("/activities", activityH.List)
		api.GET("/activities/:id", activityH.Detail)

		optional := api.Group("", handler.OptionalAuth(jwtSvc))
		optional.GET("/blogs", blogH.Feed)
		optional.GET("/blogs/:id", blogH.Detail)
		optional.GET("/blogs/:id/comments", commentH.List)
		optional.GET("/users/:id", userH.Profile)
		optional.GET("/users/:id/blogs", userH.Blogs)
		optional.GET("/users/:id/followers", userH.Followers)
		optional.GET("/users/:id/following", userH.Following)

		protected := api.Group("", handler.RequireAuth(jwtSvc))
		protected.GET("/me", meH.Profile)
		protected.PUT("/me", meH.Update)
		protected.GET("/me/stats", meH.Stats)
		protected.GET("/me/blogs", blogH.MyBlogs)
		protected.GET("/favorites", favoriteH.List)
		protected.POST("/favorites", favoriteH.Add)
		protected.DELETE("/favorites/:placeId", favoriteH.Remove)
		protected.GET("/cart", shopH.Cart)
		protected.POST("/cart", shopH.AddCart)
		protected.PUT("/cart/:productId", shopH.UpdateCart)
		protected.DELETE("/cart/:productId", shopH.RemoveCart)
		protected.POST("/shop/orders", shopH.Checkout)
		protected.GET("/shop/orders", shopH.Orders)
		protected.GET("/shop/orders/:id", shopH.Order)
		protected.POST("/routes/generate", routeH.Generate)
		protected.GET("/routes", routeH.List)
		protected.GET("/routes/:id", routeH.Detail)
		protected.POST("/blogs", blogH.Publish)
		protected.PUT("/blogs/:id", blogH.Update)
		protected.DELETE("/blogs/:id", blogH.Delete)
		protected.POST("/blogs/draft", blogH.Draft)
		protected.POST("/blogs/:id/like", blogH.Like)
		protected.DELETE("/blogs/:id/like", blogH.Unlike)
		protected.POST("/blogs/:id/comments", commentH.Create)
		protected.DELETE("/comments/:id", commentH.Delete)
		protected.POST("/users/:id/follow", userH.Follow)
		protected.DELETE("/users/:id/follow", userH.Unfollow)
		protected.GET("/notifications", notificationH.List)
		protected.GET("/notifications/unread-count", notificationH.UnreadCount)
		protected.POST("/notifications/read", notificationH.MarkRead)
		protected.POST("/notifications/read-all", notificationH.MarkAllRead)
		protected.POST("/upload", uploadH.Upload)
		protected.POST("/ai/copy", aiH.Copy)
		protected.POST("/ai/poster", aiH.CreatePoster)
		protected.GET("/ai/poster/:taskId", aiH.PosterTask)
		protected.POST("/ai/menu-ocr", aiH.MenuOCR)
		protected.POST("/ai/menu-translate", aiH.MenuTranslate)
		protected.POST("/ai/dish-explain", aiH.DishExplain)
		protected.POST("/guide/chat", guideH.Chat)
		protected.GET("/guide/sessions/:id", guideH.Session)

		protected.POST("/agent/runs", agentH.CreateRun)
		protected.GET("/agent/runs/:id", agentH.GetRun)
		protected.GET("/agent/runs/:id/events", agentH.Stream)
		protected.POST("/agent/runs/:id/cancel", agentH.Cancel)
		protected.POST("/agent/runs/:id/plan/confirm", agentH.ConfirmPlan)
		protected.GET("/agent/conversations", agentH.ListConversations)
		protected.GET("/agent/conversations/:id", agentH.GetConversation)
		if agentDebug {
			protected.GET("/agent/debug/metrics", agentH.Metrics)
		}

		admin := protected.Group("/admin", handler.RequireRole(userRepo, "admin"))
		admin.GET("/knowledge", adminH.ListKnowledge)
		admin.POST("/knowledge", adminH.CreateKnowledge)
		admin.PUT("/knowledge/:id", adminH.UpdateKnowledge)
		admin.DELETE("/knowledge/:id", adminH.DeleteKnowledge)
		admin.GET("/products", adminH.ListProducts)
		admin.POST("/products", adminH.CreateProduct)
		admin.PUT("/products/:id", adminH.UpdateProduct)
		admin.DELETE("/products/:id", adminH.DeleteProduct)
		admin.GET("/orders", adminH.ListOrders)
		admin.PUT("/orders/:id/status", adminH.UpdateOrderStatus)
	}

	port := env("PORT", "8081")
	// HOST is empty by default (listen on all interfaces) so docker-compose keeps
	// working; production sets HOST=127.0.0.1 to keep the backend behind nginx.
	host := env("HOST", "")

	srv := &http.Server{
		Addr:              host + ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("travel-gd listening on %s", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server failed: %v", err)
		}
	case <-ctx.Done():
		log.Printf("shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, fn := range shutdownFns {
		fn(shutdownCtx)
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	db.Close()
}

// shutdownFns lets optional subsystems (the agent run manager) drain before
// the HTTP server closes.
var shutdownFns []func(context.Context)

// acquireAgentLock keeps the in-process agent state (SSE buffers, poster
// tasks, recovery) safe on a single instance. The lock lives on one dedicated
// connection, so it is released automatically if the process dies.
func acquireAgentLock(db *sqlx.DB) *sql.Conn {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := db.Conn(ctx)
	if err != nil {
		log.Printf("agent lock connection: %v", err)
		return nil
	}
	var got sql.NullInt64
	if err := conn.QueryRowContext(ctx,
		"SELECT GET_LOCK('travel_gd_agent_singleton', 0)").Scan(&got); err != nil ||
		!got.Valid || got.Int64 != 1 {
		_ = conn.Close()
		return nil
	}
	return conn
}
