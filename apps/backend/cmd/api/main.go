package main

import (
	"log"

	"github.com/Bainandhika/acis/apps/backend/config"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/bootstrap"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/shared/logger"
)

func main() {
	// 1. Initialize Global Logger
	logger.Init("./logs")

	// 2. Load Modular Configuration (acis-config.yaml + .env secrets)
	cfg := config.Load("acis-config.yaml")

	// 3. Initialize Database Connection Pool
	db, err := database.NewConnection(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to initialize database connection pool: %v", err)
	}
	defer db.Close()

	// 4. Instantiate & Start Modular Monolith Server
	srv := bootstrap.NewServer(cfg, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start application server: %v", err)
	}
}
