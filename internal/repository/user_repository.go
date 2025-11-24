package repository

import (
	"context"
	"time"

	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var userTracer = otel.Tracer("repository.user")

type UserRepository struct {
	db         *database.DB
	collection *mongo.Collection
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{
		db:         db,
		collection: db.Collection("users"),
	}
}

// Create creates a new anonymous user with a cryptographic token
func (r *UserRepository) Create(ctx context.Context, token, displayName string) (*models.User, error) {
	ctx, span := userTracer.Start(ctx, "UserRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.display_name", displayName),
	)

	now := time.Now()
	user := &models.User{
		ID:           primitive.NewObjectID(),
		Token:        token,
		DisplayName:  displayName,
		CreatedAt:    now,
		LastActiveAt: now,
	}

	_, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return user, nil
}

// GetByToken retrieves a user by their anonymous token
func (r *UserRepository) GetByToken(ctx context.Context, token string) (*models.User, error) {
	ctx, span := userTracer.Start(ctx, "UserRepository.GetByToken")
	defer span.End()

	span.SetAttributes(attribute.String("user.token", token))

	user := &models.User{}
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(user)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return user, nil
}

// UpdateLastActive updates the user's last active timestamp
func (r *UserRepository) UpdateLastActive(ctx context.Context, userID primitive.ObjectID) error {
	ctx, span := userTracer.Start(ctx, "UserRepository.UpdateLastActive")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.Hex()))

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"last_active_at": time.Now()}},
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

