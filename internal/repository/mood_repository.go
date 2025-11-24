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

var moodTracer = otel.Tracer("repository.mood")

type MoodRepository struct {
	db                    *database.DB
	entriesCollection     *mongo.Collection
	recommendationsCollection *mongo.Collection
}

func NewMoodRepository(db *database.DB) *MoodRepository {
	return &MoodRepository{
		db:                         db,
		entriesCollection:          db.Collection("mood_entries"),
		recommendationsCollection:  db.Collection("mood_recommendations"),
	}
}

// CreateMoodEntry creates a new mood entry
func (r *MoodRepository) CreateMoodEntry(ctx context.Context, entry *models.MoodEntry) error {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.CreateMoodEntry")
	defer span.End()

	entry.ID = primitive.NewObjectID()
	entry.CreatedAt = time.Now()
	if entry.Date.IsZero() {
		entry.Date = time.Now()
	}

	// Use upsert to handle duplicate entries for same user/date
	filter := bson.M{
		"user_id": entry.UserID,
		"date":    entry.Date,
	}
	update := bson.M{
		"$set": bson.M{
			"mood_type":  entry.MoodType,
			"mood_level": entry.MoodLevel,
			"score":      entry.Score,
			"notes":      entry.Notes,
			"created_at": entry.CreatedAt,
		},
		"$setOnInsert": bson.M{
			"_id": entry.ID,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := r.entriesCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetMoodEntryByDate gets mood entry for a specific date
func (r *MoodRepository) GetMoodEntryByDate(ctx context.Context, userID primitive.ObjectID, date time.Time) (*models.MoodEntry, error) {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.GetMoodEntryByDate")
	defer span.End()

	entry := &models.MoodEntry{}
	err := r.entriesCollection.FindOne(ctx, bson.M{
		"user_id": userID,
		"date":    date,
	}).Decode(entry)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return entry, nil
}

// GetMoodHistory gets mood entries for a user within a date range
func (r *MoodRepository) GetMoodHistory(ctx context.Context, userID primitive.ObjectID, days int) ([]*models.MoodEntry, error) {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.GetMoodHistory")
	defer span.End()

	startDate := time.Now().AddDate(0, 0, -days)
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: -1}})
	cursor, err := r.entriesCollection.Find(ctx, bson.M{
		"user_id": userID,
		"date":    bson.M{"$gte": startDate},
	}, opts)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*models.MoodEntry
	if err := cursor.All(ctx, &entries); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return entries, nil
}

// CreateMoodRecommendation creates a recommendation for a mood entry
func (r *MoodRepository) CreateMoodRecommendation(ctx context.Context, rec *models.MoodRecommendation) error {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.CreateMoodRecommendation")
	defer span.End()

	rec.ID = primitive.NewObjectID()
	rec.CreatedAt = time.Now()

	_, err := r.recommendationsCollection.InsertOne(ctx, rec)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
