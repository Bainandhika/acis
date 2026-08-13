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
	"github.com/redis/go-redis/v9"
)

type Server struct {
	cfg            *config.Config
	db             *database.AppDB
	redisClient    *redis.Client
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

	redisAddr := fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port)
	if cfg.Redis.Host == "" || cfg.Redis.Port == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	otpCache := cache.NewOTPCache(redisClient, cfg.OTP.EncryptionKey)

	s := &Server{
		cfg:         cfg,
		db:          db,
		redisClient: redisClient,
		router:      r,
		otpCache:    otpCache,
		limiter:     limiter,
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

	// Telegram & Resend Client Setup
	tgToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	tgClient := telegram.NewClient(tgToken)
	resendApiKey := os.Getenv("RESEND_API_KEY")
	resendSender := notification.NewResendSender(resendApiKey)

	// Register Outbox Notification Handlers
	s.workerPool.RegisterHandler("email_otp", func(ctx context.Context, job notification.NotificationJob) error {
		var payload map[string]string
		if err := json.Unmarshal(job.Payload, &payload); err != nil {
			return err
		}
		log.Printf("📧 Outbox worker sending Email OTP to %s\n", job.Recipient)
		return resendSender.SendOTP(ctx, job.Recipient, payload["code"])
	})
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
		log.Printf("⚠️ Outbox worker processed Telegram Alert for recipient: %s\n", job.Recipient)
		return nil
	})

	// Domain Modules Dependency Injection
	authRepo := authentication.NewRepository(s.db)
	familyRepo := family.NewRepository(s.db)

	// RoleFinder adapter: bridges authentication.RoleFinder interface to family repo
	roleFinder := NewRoleFinderAdapter(familyRepo)
	otpTTL, err := time.ParseDuration(s.cfg.OTP.TTL)
	if err != nil {
		otpTTL = 5 * time.Minute
	}
	authSvc := authentication.NewService(authRepo, roleFinder, outboxRepo, s.otpCache, s.db, s.cfg.JWT.Secret, otpTTL)
	isProduction := s.cfg.Server.Mode == "release"
	authHandler := authentication.NewAuthHandler(authSvc, isProduction)


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

		v1.POST("/authentication/request-otp", authHandler.RequestOTP)
		v1.POST("/authentication/verify-otp", authHandler.VerifyOTP)

		v1.POST("/telegram/webhook", webhookHandler.HandleWebhook)
	}

	// Protected Routes — Family Setup (no family context needed)
	familySetup := v1.Group("")
	familySetup.Use(middleware.AuthMiddleware(s.cfg.JWT.Secret))
	{
		familySetup.POST("/authentication/logout", authHandler.Logout)
		familySetup.POST("/family", familyHandler.CreateFamily)
		familySetup.POST("/family/join", familyHandler.JoinFamily)
		familySetup.GET("/family/me", familyHandler.GetMyFamily)
	}

	// Protected Routes — Requires family membership (family_id injected into context)
	familyProtected := v1.Group("")
	familyProtected.Use(middleware.AuthMiddleware(s.cfg.JWT.Secret))
	familyProtected.Use(middleware.FamilyContextMiddleware(s.db))
	{
		familyProtected.PATCH("/family/settings", middleware.RequireRole("admin"), familyHandler.UpdateSettings)
		familyProtected.POST("/family/telegram/disconnect", middleware.RequireRole("admin"), familyHandler.DisconnectTelegram)

		familyProtected.POST("/family/wallets", middleware.RequireRole("admin"), familyHandler.CreateWallet)
		familyProtected.GET("/family/wallets", familyHandler.GetWallets)

		familyProtected.POST("/transaction", middleware.RequireRole("admin"), txHandler.CreateTransaction)
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
	if s.limiter != nil {
		s.limiter.Close()
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
