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
	db.SetConnMaxIdleTime(2 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Database connection pool initialized successfully")

	return &AppDB{DB: db}, nil
}

func NewAppDB(db *sqlx.DB) *AppDB {
	return &AppDB{DB: db}
}

func logExecution(ctx context.Context, query string, duration time.Duration, err error) {
	var traceID string
	if tid, ok := ctx.Value("X-Transaction-ID").(string); ok {
		traceID = tid
	} else {
		traceID = "system"
	}

	durationMs := duration.Milliseconds()

	if err != nil {
		slog.Error("DB Query Failed",
			slog.String("trace_id", traceID),
			slog.String("query", query),
			slog.Int64("duration_ms", durationMs),
			slog.Any("error", err),
		)
	} else {
		slog.Info("DB Query Executed",
			slog.String("trace_id", traceID),
			slog.String("query", query),
			slog.Int64("duration_ms", durationMs),
		)
	}
}

func (db *AppDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := db.DB.SelectContext(ctx, dest, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return err
}

func (db *AppDB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	start := time.Now()
	err := db.DB.GetContext(ctx, dest, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return err
}

func (db *AppDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := db.DB.ExecContext(ctx, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return res, err
}

func (db *AppDB) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	start := time.Now()
	rows, err := db.DB.QueryxContext(ctx, query, args...)
	logExecution(ctx, query, time.Since(start), err)
	return rows, err
}

func (db *AppDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := db.DB.QueryRowContext(ctx, query, args...)
	logExecution(ctx, query, time.Since(start), nil)
	return row
}
