package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Bainandhika/acis/apps/backend/config"
	"github.com/Bainandhika/acis/apps/backend/infrastructure/database"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

func main() {
	// 1. Load Modular Configuration (acis-config.yaml + .env secrets)
	cfg, err := config.Load("acis-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := cfg.AdminDSN()

	// 2. Initialize Database Connection Pool
	db, err := database.NewConnection(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database connection pool: %v", err)
	}
	defer db.Close()
	log.Println(" Connected to database")

	// Create migrations tracking table
	createMigrationsTable(db.DB)

	// Resolve migrations directory from multiple candidate paths
	migrationsDir := resolveMigrationsDir()
	log.Printf("📂 Using migrations directory: %s\n", migrationsDir)

	// Run migrations
	runMigrations(db.DB, migrationsDir)
}

func resolveMigrationsDir() string {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	candidates := []string{
		"../migrations",
		"../../migrations",
		"./migrations",
		"apps/migrations",
		"/app/migrations",
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}

	return "../migrations"
}

func createMigrationsTable(db *sqlx.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}
	log.Println(" Migrations tracking table ready")
}

func runMigrations(db *sqlx.DB, migrationsDir string) {
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory %s: %v", migrationsDir, err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	applied := getAppliedMigrations(db)

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		if applied[file.Name()] {
			log.Printf("⏭️  Skipping %s (already applied)", file.Name())
			continue
		}

		log.Printf(" Applying migration: %s", file.Name())

		filePath := filepath.Join(migrationsDir, file.Name())
		contentBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file.Name(), err)
		}

		sqlContent := string(contentBytes)
		if strings.Contains(sqlContent, "-- +goose Down") {
			parts := strings.Split(sqlContent, "-- +goose Down")
			sqlContent = parts[0]
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		if _, err := tx.Exec(sqlContent); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration %s: %v", file.Name(), err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", file.Name()); err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration %s: %v", file.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit migration %s: %v", file.Name(), err)
		}

		log.Printf(" Applied %s", file.Name())
	}

	log.Println(" All migrations completed successfully")
}

func getAppliedMigrations(db *sqlx.DB) map[string]bool {
	applied := make(map[string]bool)

	var versions []string
	if err := db.Select(&versions, "SELECT version FROM schema_migrations"); err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	for _, v := range versions {
		applied[v] = true
	}

	return applied
}
