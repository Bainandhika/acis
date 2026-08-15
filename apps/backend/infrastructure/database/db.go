package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// AppDB wraps sqlx.DB to add custom logging with trace ID
type AppDB struct {
	*sqlx.DB
}

func NewConnection(dsn string) (*AppDB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool initialized successfully")

	return &AppDB{DB: db}, nil
}

func NewAppDB(db *sqlx.DB) *AppDB {
	return &AppDB{DB: db}
}

func logQuery(ctx context.Context, query string, args ...interface{}) {
	var traceID string
	if tid, ok := ctx.Value("X-Transaction-ID").(string); ok {
		traceID = tid
	} else {
		traceID = "system"
	}

	start := time.Now()

	defer func() {
		duration := time.Since(start).Milliseconds()
		slog.Info("DB Query Executed",
			slog.String("trace_id", traceID),
			slog.String("query", query),
			slog.Any("args", args),
			slog.Int64("duration_ms", duration),
		)
	}()
}

func (db *AppDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logQuery(ctx, query, args...)
	return db.DB.SelectContext(ctx, dest, query, args...)
}

func (db *AppDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logQuery(ctx, query, args...)
	return db.DB.GetContext(ctx, dest, query, args...)
}

func (db *AppDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	logQuery(ctx, query, args...)
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *AppDB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	logQuery(ctx, query, args...)
	return db.DB.QueryxContext(ctx, query, args...)
}

func (db *AppDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	logQuery(ctx, query, args...)
	return db.DB.QueryRowContext(ctx, query, args...)
}
