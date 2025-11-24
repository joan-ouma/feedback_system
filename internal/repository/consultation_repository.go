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

var consultationTracer = otel.Tracer("repository.consultation")

type ConsultationRepository struct {
	db                *database.DB
	sessionsCollection *mongo.Collection
	consultationsCollection *mongo.Collection
}

func NewConsultationRepository(db *database.DB) *ConsultationRepository {
	return &ConsultationRepository{
		db:                      db,
		sessionsCollection:      db.Collection("consultation_sessions"),
		consultationsCollection: db.Collection("consultations"),
	}
}

// CreateSession creates a new consultation session
func (r *ConsultationRepository) CreateSession(ctx context.Context, userID primitive.ObjectID) (*models.ConsultationSession, error) {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.CreateSession")
	defer span.End()

	now := time.Now()
	session := &models.ConsultationSession{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := r.sessionsCollection.InsertOne(ctx, session)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return session, nil
}

// GetOrCreateSession gets an existing session or creates a new one
func (r *ConsultationRepository) GetOrCreateSession(ctx context.Context, userID primitive.ObjectID, sessionID *primitive.ObjectID) (*models.ConsultationSession, error) {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.GetOrCreateSession")
	defer span.End()

	if sessionID != nil {
		session := &models.ConsultationSession{}
		err := r.sessionsCollection.FindOne(ctx, bson.M{
			"_id":     *sessionID,
			"user_id": userID,
		}).Decode(session)

		if err == nil {
			return session, nil
		}
		if err != mongo.ErrNoDocuments {
			span.RecordError(err)
			return nil, err
		}
	}

	return r.CreateSession(ctx, userID)
}

// Create creates a new consultation message
func (r *ConsultationRepository) Create(ctx context.Context, consultation *models.Consultation) error {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.Create")
	defer span.End()

	span.SetAttributes(attribute.String("consultation.session_id", consultation.SessionID.Hex()))

	consultation.ID = primitive.NewObjectID()
	consultation.CreatedAt = time.Now()

	_, err := r.consultationsCollection.InsertOne(ctx, consultation)
	if err != nil {
		span.RecordError(err)
		return err
	}

	// Update session updated_at
	_, err = r.sessionsCollection.UpdateOne(
		ctx,
		bson.M{"_id": consultation.SessionID},
		bson.M{"$set": bson.M{"updated_at": time.Now()}},
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetBySessionID retrieves all consultations for a session
func (r *ConsultationRepository) GetBySessionID(ctx context.Context, sessionID primitive.ObjectID) ([]*models.Consultation, error) {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.GetBySessionID")
	defer span.End()

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	cursor, err := r.consultationsCollection.Find(ctx, bson.M{"session_id": sessionID}, opts)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var consultations []*models.Consultation
	if err := cursor.All(ctx, &consultations); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return consultations, nil
}
