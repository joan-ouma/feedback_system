package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("database")

// DB wraps mongo.Client with tracing capabilities
type DB struct {
	*mongo.Client
	Database *mongo.Database
}

// NewDB creates a new MongoDB connection
func NewDB(ctx context.Context, mongoURI string) (*DB, error) {
	ctx, span := tracer.Start(ctx, "database.NewDB")
	defer span.End()

	clientOptions := options.Client().ApplyURI(mongoURI)
	clientOptions.SetConnectTimeout(10 * time.Second)
	clientOptions.SetServerSelectionTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Extract database name from URI or use default
	dbName := "feedback_sys"
	if dbNameFromURI := extractDatabaseName(mongoURI); dbNameFromURI != "" {
		dbName = dbNameFromURI
	}

	database := client.Database(dbName)

	span.SetAttributes(attribute.String("db.system", "mongodb"))
	span.SetAttributes(attribute.String("db.name", dbName))

	return &DB{
		Client:   client,
		Database: database,
	}, nil
}

// extractDatabaseName extracts database name from MongoDB URI
func extractDatabaseName(uri string) string {
	// Simple extraction - MongoDB URI format: mongodb://host:port/dbname
	// For production URIs, this might be more complex
	// This is a simple implementation
	return ""
}

// Close closes the MongoDB connection
func (db *DB) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db.Client.Disconnect(ctx)
}

// Collection returns a MongoDB collection
func (db *DB) Collection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}

