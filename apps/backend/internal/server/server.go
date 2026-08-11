package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/auth"
	"github.com/Bainandhika/acis/apps/backend/internal/config"
	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/domain"
	"github.com/Bainandhika/acis/apps/backend/internal/family"
	"github.com/Bainandhika/acis/apps/backend/internal/middleware"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/Bainandhika/acis/apps/backend/internal/shared/cache"
	"github.com/Bainandhika/acis/apps/backend/internal/telegram"
	"github.com/Bainandhika/acis/apps/backend/internal/transaction"
	"github.com/Bainandhika/acis/apps/backend/internal/worker"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg            *config.Config
	db             *database.AppDB
	router         *gin.Engine
	workerPool     *worker.WorkerPool
	poller         *worker.OutboxPoller
	reminderWorker *telegram.LowBalanceWorker
	otpCache       *cache.OTPCache
	limiter        *cache.TokenBucketLimiter
}

func NewServer(cfg *config.Config, db *database.AppDB) *Server {
	r := gin.Default()

	// CORS Setup
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Transaction-ID"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))
	r.Use(middleware.TraceID())

	// Native In-Memory Middlewares
	r.Use(middleware.RequestSizeLimiter(1 << 20))
	r.Use(middleware.TimeoutMiddleware(10 * time.Second))

	limiter := cache.NewTokenBucketLimiter(2, 5, 5*time.Minute)
	r.Use(middleware.NativeRateLimitMiddleware(limiter))

	otpCache := cache.NewOTPCache(1 * time.Minute)

	s := &Server{
		cfg:      cfg,
		db:       db,
		router:   r,
		otpCache: otpCache,
		limiter:  limiter,
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// 1. Outbox Repository & Worker Pool Setup
	outboxRepo := repository.NewOutboxRepository(s.db)
	s.workerPool = worker.NewWorkerPool(outboxRepo, 3, 100)
	s.workerPool.Start(context.Background())

	s.poller = worker.NewOutboxPoller(outboxRepo, s.workerPool, 2*time.Second, 20)
	s.poller.Start(context.Background())

	// Telegram Client Setup
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgClient := telegram.NewClient(tgToken)

	// Register Outbox Notification Handlers
	s.workerPool.RegisterHandler("email_otp", func(ctx context.Context, job domain.NotificationJob) error {
		log.Printf("📧 Outbox worker sending Email OTP to %s: %s\n", job.Recipient, string(job.Payload))
		return nil
	})
	s.workerPool.RegisterHandler("proposal_approved", func(ctx context.Context, job domain.NotificationJob) error {
		log.Printf("🔔 Outbox worker sending Proposal Approved notification to %s: %s\n", job.Recipient, string(job.Payload))
		return nil
	})
	s.workerPool.RegisterHandler("telegram_alert", func(ctx context.Context, job domain.NotificationJob) error {
		log.Printf("⚠️ Outbox worker sending Telegram Alert to %s: %s\n", job.Recipient, string(job.Payload))
		return nil
	})

	// 2. Domain Modules Dependency Injection
	authRepo := auth.NewRepository(s.db)
	authSvc := auth.NewService(authRepo, outboxRepo, s.otpCache, s.db, s.cfg.JWTSecret)
	authHandler := auth.NewHandler(authSvc)

	familyRepo := family.NewRepository(s.db)
	familySvc := family.NewService(familyRepo, s.db)
	familyHandler := family.NewHandler(familySvc)

	txRepo := transaction.NewRepository(s.db)
	txSvc := transaction.NewService(txRepo, outboxRepo, s.db)
	txHandler := transaction.NewHandler(txSvc)

	// Telegram Module Dependency Injection
	botService := telegram.NewBotService(txSvc, familySvc, tgClient)
	webhookHandler := telegram.NewWebhookHandler(botService, tgClient)

	// Start Low Balance Cron Worker
	s.reminderWorker = telegram.NewLowBalanceWorker(familySvc, outboxRepo, s.db, 1*time.Hour)
	s.reminderWorker.Start(context.Background())

	// 3. Public Routes
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "ACIS Modular Monolith API is running"})
		})
		v1.POST("/auth/request-otp", authHandler.RequestOTP)
		v1.POST("/auth/verify-otp", authHandler.VerifyOTP)

		// Telegram Webhook Channel Endpoint
		v1.POST("/telegram/webhook", webhookHandler.HandleWebhook)
	}

	// 4. Protected Routes
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
	{
		// Wallets (Family module)
		protected.POST("/wallets", middleware.RequireRole("admin"), familyHandler.CreateWallet)
		protected.GET("/wallets", familyHandler.GetWallets)

		// Proposals & Transactions (Transaction module)
		protected.POST("/proposals", txHandler.CreateProposal)
		protected.POST("/proposals/:id/approve", middleware.RequireRole("admin"), txHandler.ApproveProposal)
		protected.POST("/proposals/:id/reject", middleware.RequireRole("admin"), txHandler.RejectProposal)

		protected.POST("/transactions", middleware.RequireRole("admin"), txHandler.CreateTransaction)
		protected.GET("/transactions", txHandler.GetTransactions)

		// Family routes (Family module)
		protected.POST("/families", familyHandler.CreateFamily)
		protected.POST("/families/join", familyHandler.JoinFamily)
		protected.GET("/families/me", familyHandler.GetMyFamily)
	}
}

func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.cfg.Port),
		Handler: s.router,
	}

	go func() {
		log.Printf("Server starting on port %s\n", s.cfg.Port)
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

	if s.reminderWorker != nil {
		s.reminderWorker.Stop()
	}
	if s.poller != nil {
		s.poller.Stop()
	}
	if s.workerPool != nil {
		s.workerPool.Stop()
	}
	if s.otpCache != nil {
		s.otpCache.Close()
	}
	if s.limiter != nil {
		s.limiter.Close()
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Forced to Shutdown:", err)
	}

	log.Println("Server exiting")
	return nil
}
