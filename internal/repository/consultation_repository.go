// +build ignore

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

var consultationTracer = otel.Tracer("repository.consultation")

type ConsultationRepository struct {
	db *database.DB
}

func NewConsultationRepository(db *database.DB) *ConsultationRepository {
	return &ConsultationRepository{db: db}
}

// CreateSession creates a new consultation session
func (r *ConsultationRepository) CreateSession(ctx context.Context, userID uuid.UUID) (*models.ConsultationSession, error) {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.CreateSession")
	defer span.End()

	session := &models.ConsultationSession{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	query := `
		INSERT INTO consultation_sessions (id, user_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, created_at, updated_at
	`

	err := r.db.QueryRow(ctx, query,
		session.ID, session.UserID, session.CreatedAt, session.UpdatedAt,
	).Scan(
		&session.ID, &session.UserID, &session.CreatedAt, &session.UpdatedAt,
	)

	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return session, nil
}

// GetOrCreateSession gets an existing session or creates a new one
func (r *ConsultationRepository) GetOrCreateSession(ctx context.Context, userID uuid.UUID, sessionID *uuid.UUID) (*models.ConsultationSession, error) {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.GetOrCreateSession")
	defer span.End()

	if sessionID != nil {
		session := &models.ConsultationSession{}
		query := `
			SELECT id, user_id, created_at, updated_at
			FROM consultation_sessions
			WHERE id = $1 AND user_id = $2
		`

		err := r.db.QueryRow(ctx, query, *sessionID, userID).Scan(
			&session.ID, &session.UserID, &session.CreatedAt, &session.UpdatedAt,
		)

		if err == nil {
			return session, nil
		}
		if err != pgx.ErrNoRows {
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

	span.SetAttributes(attribute.String("consultation.session_id", consultation.SessionID.String()))

	consultation.ID = uuid.New()
	consultation.CreatedAt = time.Now()

	query := `
		INSERT INTO consultations (id, session_id, user_id, message, response, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		consultation.ID, consultation.SessionID, consultation.UserID,
		consultation.Message, consultation.Response, consultation.CreatedAt,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	// Update session updated_at
	updateQuery := `
		UPDATE consultation_sessions
		SET updated_at = $1
		WHERE id = $2
	`
	_, err = r.db.Exec(ctx, updateQuery, time.Now(), consultation.SessionID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// GetBySessionID retrieves all consultations for a session
func (r *ConsultationRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) ([]*models.Consultation, error) {
	ctx, span := consultationTracer.Start(ctx, "ConsultationRepository.GetBySessionID")
	defer span.End()

	query := `
		SELECT id, session_id, user_id, message, response, created_at
		FROM consultations
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var consultations []*models.Consultation
	for rows.Next() {
		consultation := &models.Consultation{}
		err := rows.Scan(
			&consultation.ID, &consultation.SessionID, &consultation.UserID,
			&consultation.Message, &consultation.Response, &consultation.CreatedAt,
		)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}
		consultations = append(consultations, consultation)
	}

	return consultations, rows.Err()
}

