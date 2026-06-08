package main

import (
	"log"

	"github.com/Bainandhika/acis/apps/backend/internal/config"
	"github.com/Bainandhika/acis/apps/backend/internal/database"
	"github.com/Bainandhika/acis/apps/backend/internal/server"
	"github.com/Bainandhika/acis/apps/backend/pkg/logger"
)

func main() {
	// 1. Initialize Logger
	logger.Init("./logs")

	// 2. Load Configuration
	cfg := config.Load()

	// 3. Initialize Database
	db, err := database.NewConnection(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close() // Pastikan koneksi DB ditutup pas app mati

	// 4. Setup and Start Server
	srv := server.NewServer(cfg, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
