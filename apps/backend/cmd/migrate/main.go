package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Config struct {
	DatabaseURL string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPassword  string
	DBName      string
	DBSSLMode   string
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		_ = godotenv.Load("../.env")
		_ = godotenv.Load("../../.env")
	}

	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBPort:      getEnv("DB_PORT", "5432"),
		DBUser:      getEnv("DB_USER", "acis_user"),
		DBPassword:  getEnv("DB_PASSWORD", "acis_secret_password"),
		DBName:      getEnv("DB_NAME", "acis_db"),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
	}

	var dsn string
	if config.DatabaseURL != "" {
		dsn = config.DatabaseURL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, config.DBSSLMode,
		)
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println(" Connected to database")

	// Create migrations tracking table
	createMigrationsTable(db)

	// Resolve migrations directory from multiple candidate paths
	migrationsDir := resolveMigrationsDir()
	log.Printf("📂 Using migrations directory: %s\n", migrationsDir)

	// Run migrations
	runMigrations(db, migrationsDir)
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
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", file.Name(), err)
		}

		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
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

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}