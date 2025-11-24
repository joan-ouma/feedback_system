package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("database")

// DB wraps pgxpool.Pool with tracing capabilities
type DB struct {
	*pgxpool.Pool
}

// NewDB creates a new database connection pool
func NewDB(ctx context.Context, databaseURL string) (*DB, error) {
	ctx, span := tracer.Start(ctx, "database.NewDB")
	defer span.End()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	span.SetAttributes(attribute.String("db.system", "postgresql"))
	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// QueryContext wraps pgxpool query with tracing
func (db *DB) QueryContext(ctx context.Context, queryName, query string, args ...interface{}) (interface{}, error) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("database.%s", queryName),
		trace.WithAttributes(
			attribute.String("db.statement", query),
		))
	defer span.End()

	// This is a placeholder - actual implementation would use pgx methods
	return nil, nil
}

