package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var feedbackTracer = otel.Tracer("repository.feedback")

type FeedbackRepository struct {
	db *database.DB
}

func NewFeedbackRepository(db *database.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create creates a new feedback entry
func (r *FeedbackRepository) Create(ctx context.Context, feedback *models.Feedback) error {
	ctx, span := feedbackTracer.Start(ctx, "FeedbackRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("feedback.type", string(feedback.Type)),
		attribute.String("feedback.status", feedback.Status),
	)

	feedback.ID = uuid.New()
	feedback.CreatedAt = time.Now()
	feedback.UpdatedAt = time.Now()

	query := `
		INSERT INTO feedbacks (id, user_id, type, title, content, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		feedback.ID, feedback.UserID, feedback.Type, feedback.Title,
		feedback.Content, feedback.Status, feedback.CreatedAt, feedback.UpdatedAt,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetByUserID retrieves all feedbacks for a user
func (r *FeedbackRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*models.Feedback, error) {
	ctx, span := feedbackTracer.Start(ctx, "FeedbackRepository.GetByUserID")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	query := `
		SELECT id, user_id, type, title, content, status, created_at, updated_at
		FROM feedbacks
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var feedbacks []*models.Feedback
	for rows.Next() {
		feedback := &models.Feedback{}
		err := rows.Scan(
			&feedback.ID, &feedback.UserID, &feedback.Type, &feedback.Title,
			&feedback.Content, &feedback.Status, &feedback.CreatedAt, &feedback.UpdatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, rows.Err()
}

// GetByID retrieves a feedback by ID
func (r *FeedbackRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Feedback, error) {
	ctx, span := feedbackTracer.Start(ctx, "FeedbackRepository.GetByID")
	defer span.End()

	feedback := &models.Feedback{}
	query := `
		SELECT id, user_id, type, title, content, status, created_at, updated_at
		FROM feedbacks
		WHERE id = $1
	`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&feedback.ID, &feedback.UserID, &feedback.Type, &feedback.Title,
		&feedback.Content, &feedback.Status, &feedback.CreatedAt, &feedback.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return feedback, nil
}

