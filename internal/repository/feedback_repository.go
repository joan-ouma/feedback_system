package repository

import (
	"context"
	"time"

	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var feedbackTracer = otel.Tracer("repository.feedback")

type FeedbackRepository struct {
	db         *database.DB
	collection *mongo.Collection
}

func NewFeedbackRepository(db *database.DB) *FeedbackRepository {
	return &FeedbackRepository{
		db:         db,
		collection: db.Collection("feedbacks"),
	}
}

// Create creates a new feedback entry
func (r *FeedbackRepository) Create(ctx context.Context, feedback *models.Feedback) error {
	ctx, span := feedbackTracer.Start(ctx, "FeedbackRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("feedback.type", string(feedback.Type)),
		attribute.String("feedback.status", feedback.Status),
	)

	feedback.ID = primitive.NewObjectID()
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, feedback)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetByUserID retrieves all feedbacks for a user
func (r *FeedbackRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Feedback, error) {
	ctx, span := feedbackTracer.Start(ctx, "FeedbackRepository.GetByUserID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.Hex()))

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var feedbacks []*models.Feedback
	if err := cursor.All(ctx, &feedbacks); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return feedbacks, nil
}

// GetByID retrieves a feedback by ID
func (r *FeedbackRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Feedback, error) {
	ctx, span := feedbackTracer.Start(ctx, "FeedbackRepository.GetByID")
	defer span.End()

	feedback := &models.Feedback{}
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(feedback)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return feedback, nil
}
