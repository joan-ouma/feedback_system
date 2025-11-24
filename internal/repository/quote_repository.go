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
)

var quoteTracer = otel.Tracer("repository.quote")

type QuoteRepository struct {
	db              *database.DB
	quotesCollection *mongo.Collection
}

func NewQuoteRepository(db *database.DB) *QuoteRepository {
	return &QuoteRepository{
		db:               db,
		quotesCollection: db.Collection("motivational_quotes"),
	}
}

// GetQuoteForDate gets a quote for a specific date
func (r *QuoteRepository) GetQuoteForDate(ctx context.Context, userID primitive.ObjectID, date time.Time) (*models.MotivationalQuote, error) {
	ctx, span := quoteTracer.Start(ctx, "QuoteRepository.GetQuoteForDate")
	defer span.End()

	quote := &models.MotivationalQuote{}
	err := r.quotesCollection.FindOne(ctx, bson.M{
		"user_id": userID,
		"date":    date,
	}).Decode(quote)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return quote, nil
}

// CreateQuote creates a new motivational quote
func (r *QuoteRepository) CreateQuote(ctx context.Context, quote *models.MotivationalQuote) error {
	ctx, span := quoteTracer.Start(ctx, "QuoteRepository.CreateQuote")
	defer span.End()

	quote.ID = primitive.NewObjectID()
	quote.CreatedAt = time.Now()
	if quote.Date.IsZero() {
		quote.Date = time.Now()
	}

	// Use upsert to avoid duplicates
	filter := bson.M{
		"user_id": quote.UserID,
		"date":    quote.Date,
	}
	update := bson.M{
		"$set": bson.M{
			"quote":      quote.Quote,
			"author":     quote.Author,
			"mood_type":  quote.MoodType,
			"mood_level": quote.MoodLevel,
			"is_ai":      quote.IsAI,
			"created_at": quote.CreatedAt,
		},
		"$setOnInsert": bson.M{
			"_id": quote.ID,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := r.quotesCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
