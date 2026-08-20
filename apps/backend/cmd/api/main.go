package main

import (
	"log"

	"github.com/Bainandhika/acis/apps/backend/config"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/bootstrap"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/Bainandhika/acis/apps/backend/shared/logger"
)

func main() {
	// 1. Load Modular Configuration (acis-config.yaml + .env secrets)
	cfg, err := config.Load("acis-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Global Logger with YAML log level & monthly rotation
	logger.Init(cfg.Log.Dir, cfg.Log.Level)
	defer logger.Close()

	// 3. Initialize Dual Database Connection Pools (userDB for RLS + adminDB for internal/workers)
	db, err := database.NewDualPool(cfg.AppDSN(), cfg.AdminDSN())
	if err != nil {
		log.Fatalf("Failed to initialize dual database connection pools: %v", err)
	}
	defer db.Close()

	// 4. Instantiate & Start Modular Monolith Server
	srv := bootstrap.NewServer(cfg, db)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start application server: %v", err)
	}
}
