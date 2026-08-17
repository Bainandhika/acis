package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bainandhika/acis/apps/backend/config"
	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/domain/bot"
	"github.com/Bainandhika/acis/apps/backend/domain/family"
	"github.com/Bainandhika/acis/apps/backend/domain/transaction"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/middleware"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/telegramclient"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/worker"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type roleFinderAdapter struct {
	repo family.FamilyRepository
}

func NewRoleFinderAdapter(repo family.FamilyRepository) authentication.RoleFinder {
	return &roleFinderAdapter{repo: repo}
}

func (a *roleFinderAdapter) FindRoleByUserID(ctx context.Context, userID string) (string, error) {
	member, err := a.repo.FindMemberByUserID(ctx, userID)
	if err != nil {
		return "", err
	}
	if member == nil {
		return "", nil
	}
	return member.Role, nil
}

type Server struct {
	cfg         *config.Config
	db          *database.AppDB
	redisClient *redis.Client
	router      *gin.Engine
	workerPool  *worker.WorkerPool
	poller      *worker.OutboxPoller
	otpCache    *cache.OTPCache
	limiter     *cache.TokenBucketLimiter
	authLimiter *cache.TokenBucketLimiter
}

func NewServer(cfg *config.Config, db *database.AppDB) *Server {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())

	// 1. Security Headers
	r.Use(middleware.SecurityHeadersMiddleware())

	// 2. Dynamic CORS Configuration
	allowedOrigins := cfg.CORS.AllowedOrigins
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:5173"}
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Transaction-ID", "X-Bot-Secret"},
		ExposeHeaders:    []string{"Content-Length", "X-Transaction-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsConfig))

	// 3. Request Tracing & Payload Limiters
	r.Use(middleware.TraceID())
	r.Use(middleware.RequestSizeLimiter(1 << 20)) // 1 MB payload limit
	r.Use(middleware.TimeoutMiddleware(15 * time.Second))

	// 4. Rate Limiters (General API: 30 rps/burst 50; Auth: strict 5 req/min)
	generalLimiter := cache.NewTokenBucketLimiter(30, 50, 5*time.Minute)
	r.Use(middleware.NativeRateLimitMiddleware(generalLimiter))

	authLimiter := cache.NewTokenBucketLimiter(1, 5, 5*time.Minute)

	// 5. Initialize Redis Client
	var redisClient *redis.Client
	if cfg.Redis.URL != "" {
		opt, err := redis.ParseURL(cfg.Redis.URL)
		if err == nil {
			redisClient = redis.NewClient(opt)
		}
	}
	if redisClient == nil {
		redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
		if cfg.Redis.Host == "" || cfg.Redis.Port == "" {
			redisAddr = "localhost:6379"
		}
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisAddr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
	}

	otpCache := cache.NewOTPCache(redisClient, cfg.OTP.EncryptionKey)
	tokenStore := cache.NewRefreshTokenStore(redisClient)

	s := &Server{
		cfg:         cfg,
		db:          db,
		redisClient: redisClient,
		router:      r,
		otpCache:    otpCache,
		limiter:     generalLimiter,
		authLimiter: authLimiter,
	}

	s.setupRoutes(tokenStore)
	return s
}

func (s *Server) setupRoutes(tokenStore *cache.RefreshTokenStore) {
	// Outbox Repository & Worker Pool Setup
	outboxRepo := notification.NewOutboxRepository(s.db, s.redisClient)
	s.workerPool = worker.NewWorkerPool(outboxRepo, 3, 100)
	s.workerPool.Start(context.Background())

	s.poller = worker.NewOutboxPoller(outboxRepo, s.workerPool, 1*time.Minute, 20, s.redisClient)
	s.poller.Start(context.Background())

	tgClient := telegramclient.NewClient(s.cfg.Telegram.BotToken)

	// Register Outbox Notification Handlers
	s.workerPool.RegisterHandler("proposal_approved", func(ctx context.Context, job notification.NotificationJob) error {
		log.Printf("🔔 Outbox worker sending Proposal Approved notification to %s\n", job.Recipient)
		return nil
	})
	s.workerPool.RegisterHandler("telegram_alert", func(ctx context.Context, job notification.NotificationJob) error {
		var payload map[string]interface{}
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}
		var chatID int64
		if _, err := fmt.Sscanf(job.Recipient, "%d", &chatID); err == nil && chatID != 0 {
			msg := fmt.Sprintf("⚠️ *Peringatan Saldo Dompet Rendah*\n\nDompet: *%v*\nSaldo: Rp %.2f\nLimit Min: Rp %.2f",
				payload["wallet_name"], payload["current_balance"], payload["minimum_limit"])
			return tgClient.SendMessage(ctx, chatID, msg)
		}
		return nil
	})

	// Domain Modules Dependency Injection
	authRepo := authentication.NewRepository(s.db)
	familyRepo := family.NewRepository(s.db)

	roleFinder := NewRoleFinderAdapter(familyRepo)
	otpTTL, err := time.ParseDuration(s.cfg.OTP.TTL)
	if err != nil {
		otpTTL = 5 * time.Minute
	}
	authSvc := authentication.NewService(
		authRepo,
		roleFinder,
		outboxRepo,
		s.otpCache,
		tokenStore,
		tgClient,
		s.db,
		s.cfg.JWT.Secret,
		s.cfg.Telegram.BotUsername,
		otpTTL,
	)
	isProduction := s.cfg.Server.Mode == "release"
	authHandler := authentication.NewAuthHandler(authSvc, isProduction)

	familySvc := family.NewService(familyRepo, s.db)
	familyHandler := family.NewHandler(familySvc)

	txRepo := transaction.NewRepository(s.db)
	txSvc := transaction.NewService(txRepo, outboxRepo, s.db)
	txHandler := transaction.NewHandler(txSvc)

	botHandler := bot.NewBotHandler(familySvc, txSvc)

	// Public Routes
	v1 := s.router.Group("/api/v1")
	{
		// Enhanced Health Check with DB & Redis Ping
		v1.GET("/health", func(c *gin.Context) {
			dbStatus := "healthy"
			ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
			defer cancel()

			startDB := time.Now()
			if err := s.db.PingContext(ctx); err != nil {
				dbStatus = fmt.Sprintf("unhealthy: %v", err)
			}
			dbLatencyMs := time.Since(startDB).Milliseconds()

			redisStatus := "healthy"
			startRedis := time.Now()
			if err := s.redisClient.Ping(ctx).Err(); err != nil {
				redisStatus = fmt.Sprintf("unhealthy: %v", err)
			}
			redisLatencyMs := time.Since(startRedis).Milliseconds()

			statusCode := http.StatusOK
			if dbStatus != "healthy" {
				statusCode = http.StatusServiceUnavailable
			}

			c.JSON(statusCode, gin.H{
				"status":      "ok",
				"version":     "1.4.0",
				"environment": s.cfg.Server.Mode,
				"database": gin.H{
					"status":     dbStatus,
					"latency_ms": dbLatencyMs,
				},
				"redis": gin.H{
					"status":     redisStatus,
					"latency_ms": redisLatencyMs,
				},
			})
		})

		// Auth group with strict rate limiting
		authGroup := v1.Group("/authentication")
		authGroup.Use(middleware.AuthRateLimitMiddleware(s.authLimiter))
		{
			authGroup.POST("/request-otp", authHandler.RequestOTP)
			authGroup.POST("/verify-otp", authHandler.VerifyOTP)
			authGroup.POST("/refresh", authHandler.RefreshToken)
			authGroup.POST("/logout", authHandler.Logout)
		}

		// Bot Internal API routes
		botAPI := v1.Group("/bot")
		botAPI.Use(middleware.BotSecretMiddleware(s.cfg.Bot.Secret))
		{
			botAPI.POST("/link", botHandler.Link)
			botAPI.GET("/family", botHandler.GetFamily)
			botAPI.GET("/balance", botHandler.Balance)
			botAPI.POST("/transaction", botHandler.RecordTransaction)
		}
	}

	// Protected Routes — Family Setup (no family context needed)
	familySetup := v1.Group("")
	familySetup.Use(middleware.AuthMiddleware(s.cfg.JWT.Secret))
	{
		familySetup.POST("/family", familyHandler.CreateFamily)
		familySetup.POST("/family/join", familyHandler.JoinFamily)
		familySetup.GET("/family/me", familyHandler.GetMyFamily)
	}

	// Protected Routes — Requires family membership (family_id injected into context)
	familyProtected := v1.Group("")
	familyProtected.Use(middleware.AuthMiddleware(s.cfg.JWT.Secret))
	familyProtected.Use(middleware.FamilyContextMiddleware(s.db))
	{
		familyProtected.PATCH("/family", middleware.RequireRole("admin"), familyHandler.UpdateFamily)
		familyProtected.PATCH("/family/settings", middleware.RequireRole("admin"), familyHandler.UpdateSettings)
		familyProtected.POST("/family/telegram/disconnect", middleware.RequireRole("admin"), familyHandler.DisconnectTelegram)

		familyProtected.POST("/family/wallets", middleware.RequireRole("admin"), familyHandler.CreateWallet)
		familyProtected.PATCH("/family/wallets/:id", middleware.RequireRole("admin"), familyHandler.UpdateWallet)
		familyProtected.DELETE("/family/wallets/:id", middleware.RequireRole("admin"), familyHandler.DeleteWallet)
		familyProtected.GET("/family/wallets", familyHandler.GetWallets)
		familyProtected.DELETE("/family/members/:id", middleware.RequireRole("admin"), familyHandler.RemoveMember)

		familyProtected.POST("/transaction", middleware.RequireRole("admin"), txHandler.CreateTransaction)
		familyProtected.PATCH("/transaction/:id", middleware.RequireRole("admin"), txHandler.UpdateTransaction)
		familyProtected.DELETE("/transaction/:id", middleware.RequireRole("admin"), txHandler.DeleteTransaction)
		familyProtected.GET("/transaction", txHandler.GetTransactions)
		familyProtected.POST("/transaction/proposals", txHandler.CreateProposal)
		familyProtected.GET("/transaction/proposals", txHandler.GetProposals)
		familyProtected.POST("/transaction/proposals/:id/approve", middleware.RequireRole("admin"), txHandler.ApproveProposal)
		familyProtected.POST("/transaction/proposals/:id/reject", middleware.RequireRole("admin"), txHandler.RejectProposal)
	}
}

func (s *Server) Start() error {
	port := s.cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s (mode: %s)\n", port, s.cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.poller != nil {
		s.poller.Stop()
	}
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
	if s.limiter != nil {
		s.limiter.Close()
	}
	if s.authLimiter != nil {
		s.authLimiter.Close()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Forced to Shutdown:", err)
	}

	log.Println("Server exiting")
	return nil
}
