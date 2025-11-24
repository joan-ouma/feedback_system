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

var moodTracer = otel.Tracer("repository.mood")

type MoodRepository struct {
	db *database.DB
}

func NewMoodRepository(db *database.DB) *MoodRepository {
	return &MoodRepository{db: db}
}

// CreateMoodEntry creates a new mood entry
func (r *MoodRepository) CreateMoodEntry(ctx context.Context, entry *models.MoodEntry) error {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.CreateMoodEntry")
	defer span.End()

	entry.ID = uuid.New()
	entry.CreatedAt = time.Now()
	if entry.Date.IsZero() {
		entry.Date = time.Now()
	}

	query := `
		INSERT INTO mood_entries (id, user_id, mood_type, mood_level, score, notes, created_at, date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (user_id, date) DO UPDATE SET
			mood_type = EXCLUDED.mood_type,
			mood_level = EXCLUDED.mood_level,
			score = EXCLUDED.score,
			notes = EXCLUDED.notes,
			created_at = EXCLUDED.created_at
	`

	_, err := r.db.Exec(ctx, query,
		entry.ID, entry.UserID, entry.MoodType, entry.MoodLevel,
		entry.Score, entry.Notes, entry.CreatedAt, entry.Date,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetMoodEntryByDate gets mood entry for a specific date
func (r *MoodRepository) GetMoodEntryByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.MoodEntry, error) {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.GetMoodEntryByDate")
	defer span.End()

	entry := &models.MoodEntry{}
	query := `
		SELECT id, user_id, mood_type, mood_level, score, notes, created_at, date
		FROM mood_entries
		WHERE user_id = $1 AND date = $2
	`

	err := r.db.QueryRow(ctx, query, userID, date).Scan(
		&entry.ID, &entry.UserID, &entry.MoodType, &entry.MoodLevel,
		&entry.Score, &entry.Notes, &entry.CreatedAt, &entry.Date,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return entry, nil
}

// GetMoodHistory gets mood entries for a user within a date range
func (r *MoodRepository) GetMoodHistory(ctx context.Context, userID uuid.UUID, days int) ([]*models.MoodEntry, error) {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.GetMoodHistory")
	defer span.End()

	startDate := time.Now().AddDate(0, 0, -days)
	query := `
		SELECT id, user_id, mood_type, mood_level, score, notes, created_at, date
		FROM mood_entries
		WHERE user_id = $1 AND date >= $2
		ORDER BY date DESC
	`

	rows, err := r.db.Query(ctx, query, userID, startDate)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var entries []*models.MoodEntry
	for rows.Next() {
		entry := &models.MoodEntry{}
		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.MoodType, &entry.MoodLevel,
			&entry.Score, &entry.Notes, &entry.CreatedAt, &entry.Date,
		)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// CreateMoodRecommendation creates a recommendation for a mood entry
func (r *MoodRepository) CreateMoodRecommendation(ctx context.Context, rec *models.MoodRecommendation) error {
	ctx, span := moodTracer.Start(ctx, "MoodRepository.CreateMoodRecommendation")
	defer span.End()

	rec.ID = uuid.New()
	rec.CreatedAt = time.Now()

	query := `
		INSERT INTO mood_recommendations (id, user_id, mood_entry_id, recommendations, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		rec.ID, rec.UserID, rec.MoodEntryID, rec.Recommendations, rec.CreatedAt,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

