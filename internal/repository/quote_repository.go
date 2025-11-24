package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

var quoteTracer = otel.Tracer("repository.quote")

type QuoteRepository struct {
	db *database.DB
}

func NewQuoteRepository(db *database.DB) *QuoteRepository {
	return &QuoteRepository{db: db}
}

// GetQuoteForDate gets a quote for a specific date
func (r *QuoteRepository) GetQuoteForDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.MotivationalQuote, error) {
	ctx, span := quoteTracer.Start(ctx, "QuoteRepository.GetQuoteForDate")
	defer span.End()

	quote := &models.MotivationalQuote{}
	query := `
		SELECT id, user_id, quote, author, mood_type, mood_level, is_ai, created_at, date
		FROM motivational_quotes
		WHERE user_id = $1 AND date = $2
		LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, userID, date).Scan(
		&quote.ID, &quote.UserID, &quote.Quote, &quote.Author,
		&quote.MoodType, &quote.MoodLevel, &quote.IsAI, &quote.CreatedAt, &quote.Date,
	)

	if err == pgx.ErrNoRows {
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

	quote.ID = uuid.New()
	quote.CreatedAt = time.Now()
	if quote.Date.IsZero() {
		quote.Date = time.Now()
	}

	query := `
		INSERT INTO motivational_quotes (id, user_id, quote, author, mood_type, mood_level, is_ai, created_at, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.Exec(ctx, query,
		quote.ID, quote.UserID, quote.Quote, quote.Author,
		quote.MoodType, quote.MoodLevel, quote.IsAI, quote.CreatedAt, quote.Date,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

