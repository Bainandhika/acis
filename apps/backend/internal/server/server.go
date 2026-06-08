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

	"github.com/Bainandhika/acis/apps/backend/internal/config"
	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/handler"
	"github.com/Bainandhika/acis/apps/backend/internal/middleware"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/Bainandhika/acis/apps/backend/internal/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg    *config.Config
	db     *database.AppDB
	router *gin.Engine
}

// NewServer melakukan Dependency Injection dan setup routing.
func NewServer(cfg *config.Config, db *database.AppDB) *Server {
	r := gin.Default()

	// Setup CORS
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "X-Transaction-ID"},
		AllowCredentials: true,
	}
	r.Use(cors.New(corsConfig))
	r.Use(middleware.TraceID())

	s := &Server{
		cfg:    cfg,
		db:     db,
		router: r,
	}

	// Setup routes
	s.setupRoutes()
	return s
}

// setupRoutes mendaftarkan semua endpoint dan melakukan DI untuk Handler.
func (s *Server) setupRoutes() {
	// --- Dependency Injection ---
	userRepo := repository.NewUserRepository(s.db)
	walletRepo := repository.NewWalletRepository()
	proposalRepo := repository.NewProposalRepository()
	authRepo := repository.NewAuthRepository(s.db)

	authService := service.NewAuthService(authRepo, userRepo, s.cfg.JWTSecret)
	walletService := service.NewWalletService(walletRepo, s.db)
	proposalService := service.NewProposalService(proposalRepo, walletRepo, s.db)

	authHandler := handler.NewAuthHandler(authService)
	walletHandler := handler.NewWalletHandler(walletService)
	proposalHandler := handler.NewProposalHandler(proposalService)

	// --- Health Check ---
	s.router.GET("/health", func(c *gin.Context) {
		if err := s.db.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection failed",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "ACIS API is running smoothly",
		})
	})

	// --- PUBLIC ROUTES ---
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.POST("/auth/request-otp", authHandler.RequestOTP)
		v1.POST("/auth/verify-otp", authHandler.VerifyOTP)
	}

	// --- PROTECTED ROUTES ---
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(s.cfg.JWTSecret))
	{
		protected.POST("/wallets", walletHandler.CreateWallet)
		protected.GET("/wallets", walletHandler.GetWallets)
		protected.POST("/proposals", proposalHandler.CreateProposal)
		protected.POST("/:id/reject", proposalHandler.RejectProposal)
	}
}

// Start menjalankan server HTTP dengan graceful shutdown.
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.cfg.Port),
		Handler: s.router,
	}

	// Goroutine buat jalanin server
	go func() {
		log.Printf("Server starting on port %s\n", s.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	// Graceful Shutdown: Nunggu sinyal interrupt (Ctrl+C) atau terminate dari OS
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	// Kasih timeout 5 detik buat nge-drain existing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Forced to Shutdown:", err)
	}

	log.Println("Server exiting")
	return nil
}
