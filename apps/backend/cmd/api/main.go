package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/handler"
	"github.com/Bainandhika/acis/apps/backend/internal/middleware"
	"github.com/Bainandhika/acis/apps/backend/internal/repository"
	"github.com/Bainandhika/acis/apps/backend/internal/service"
	"github.com/Bainandhika/acis/apps/backend/pkg/logger"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string
}

func main() {
	logger.Init("./logs")

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	config := Config{
		Port:       getEnv("PORT", "8080"),
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "acis_user"),
		DBPassword: getEnv("DB_PASSWORD", "acis_secret_password"),
		DBName:     getEnv("DB_NAME", "acis_db"),
		JWTSecret:  getEnv("JWT_SECRET", "acis_jwt_secret"),
	}

	rawDB := initDB(config)
	db := database.NewAppDB(rawDB)
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	walletRepo := repository.NewWalletRepository(db)
	authRepo := repository.NewAuthRepository(db)

	authService := service.NewAuthService(authRepo, userRepo, config.JWTSecret)
	walletService := service.NewWalletService(walletRepo)

	authHandler := handler.NewAuthHandler(authService)
	walletHandler := handler.NewWalletHandler(walletService)

	r := gin.Default()

	corsConfig := cors.Config{
        AllowOrigins:     []string{"http://localhost:5173"}, // Spesifik ke frontend Vue
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Transaction-ID"},
        AllowCredentials: true,
    }
    r.Use(cors.New(corsConfig))
	r.Use(middleware.TraceID())

	r.GET("/health", func(c *gin.Context) {
		if err := db.Ping(); err != nil {
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
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.POST("/auth/request-otp", authHandler.RequestOTP)
		v1.POST("/auth/verify-otp", authHandler.VerifyOTP)
	}

	// --- PROTECTED ROUTES (Require Auth) ---
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(config.JWTSecret))
	{
		protected.POST("/wallets", walletHandler.CreateWallet)
		protected.GET("/wallets", walletHandler.GetWallets)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.Port),
		Handler: r,
	}

	// Graceful Shutdown
	go func() {
		log.Printf(" Server starting on port %s\n", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Forced to Shutdown:", err)
	}
	log.Println("Server exiting")
}

func initDB(config Config) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Database connected and pool initialized")
	return db
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
