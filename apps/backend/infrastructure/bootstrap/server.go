package bootstrap

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bainandhika/acis/apps/backend/config"
	"github.com/Bainandhika/acis/apps/backend/domain/authentication"
	"github.com/Bainandhika/acis/apps/backend/domain/family"
	"github.com/Bainandhika/acis/apps/backend/domain/telegram"
	"github.com/Bainandhika/acis/apps/backend/domain/transaction"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/middleware"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/notification"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/worker"
	"github.com/Bainandhika/acis/apps/backend/shared/cache"
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

	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Transaction-ID"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))
	r.Use(middleware.TraceID())

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
	// Outbox Repository & Worker Pool Setup
	outboxRepo := notification.NewOutboxRepository(s.db)
	s.workerPool = worker.NewWorkerPool(outboxRepo, 3, 100)
	s.workerPool.Start(context.Background())

	s.poller = worker.NewOutboxPoller(outboxRepo, s.workerPool, 2*time.Second, 20)
	s.poller.Start(context.Background())

	// Telegram Client Setup
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgClient := telegram.NewClient(tgToken)

	// Register Outbox Notification Handlers
	s.workerPool.RegisterHandler("email_otp", func(ctx context.Context, job notification.NotificationJob) error {
		log.Printf("📧 Outbox worker sending Email OTP to %s: %s\n", job.Recipient, string(job.Payload))
		return nil
	})
	s.workerPool.RegisterHandler("proposal_approved", func(ctx context.Context, job notification.NotificationJob) error {
		log.Printf("🔔 Outbox worker sending Proposal Approved notification to %s: %s\n", job.Recipient, string(job.Payload))
		return nil
	})
	s.workerPool.RegisterHandler("telegram_alert", func(ctx context.Context, job notification.NotificationJob) error {
		log.Printf("⚠️ Outbox worker sending Telegram Alert to %s: %s\n", job.Recipient, string(job.Payload))
		return nil
	})

	// Domain Modules Dependency Injection
	authRepo := authentication.NewRepository(s.db)
	authSvc := authentication.NewService(authRepo, outboxRepo, s.otpCache, s.db, s.cfg.JWT.Secret)
	authHandler := authentication.NewAuthHandler(authSvc)

	familyRepo := family.NewRepository(s.db)
	familySvc := family.NewService(familyRepo, s.db)
	familyHandler := family.NewHandler(familySvc)

	txRepo := transaction.NewRepository(s.db)
	txSvc := transaction.NewService(txRepo, outboxRepo, s.db)
	txHandler := transaction.NewHandler(txSvc)

	tgTxAdapter := NewTelegramTxAdapter(txSvc)
	tgFamAdapter := NewTelegramFamilyAdapter(familySvc)

	botService := telegram.NewBotService(tgTxAdapter, tgFamAdapter, tgClient)
	webhookHandler := telegram.NewWebhookHandler(botService, tgClient)

	s.reminderWorker = telegram.NewLowBalanceWorker(tgFamAdapter, outboxRepo, s.db, 1*time.Hour)
	s.reminderWorker.Start(context.Background())

	// Public Routes
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "ACIS Modular Monolith API is running"})
		})
		v1.POST("/auth/request-otp", authHandler.RequestOTP)
		v1.POST("/auth/verify-otp", authHandler.VerifyOTP)

		v1.POST("/telegram/webhook", webhookHandler.HandleWebhook)
	}

	// Protected Routes
	protected := v1.Group("")
	jwtSecret := s.cfg.JWT.Secret
	if jwtSecret == "" {
		jwtSecret = s.cfg.JWT.Secret
	}
	protected.Use(middleware.AuthMiddleware(jwtSecret))
	{
		protected.POST("/wallets", middleware.RequireRole("admin"), familyHandler.CreateWallet)
		protected.GET("/wallets", familyHandler.GetWallets)

		protected.POST("/proposals", txHandler.CreateProposal)
		protected.POST("/proposals/:id/approve", middleware.RequireRole("admin"), txHandler.ApproveProposal)
		protected.POST("/proposals/:id/reject", middleware.RequireRole("admin"), txHandler.RejectProposal)

		protected.POST("/transactions", middleware.RequireRole("admin"), txHandler.CreateTransaction)
		protected.GET("/transactions", txHandler.GetTransactions)

		protected.POST("/families", familyHandler.CreateFamily)
		protected.POST("/families/join", familyHandler.JoinFamily)
		protected.GET("/families/me", familyHandler.GetMyFamily)
	}
}

func (s *Server) Start() error {
	port := s.cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: s.router,
	}

	go func() {
		log.Printf("Server starting on port %s\n", port)
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
