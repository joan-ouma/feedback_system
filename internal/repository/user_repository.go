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

var userTracer = otel.Tracer("repository.user")

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new anonymous user with a cryptographic token
func (r *UserRepository) Create(ctx context.Context, token, displayName string) (*models.User, error) {
	ctx, span := userTracer.Start(ctx, "UserRepository.Create")
	defer span.End()

	span.SetAttributes(
		attribute.String("user.display_name", displayName),
	)

	user := &models.User{
		ID:          uuid.New(),
		Token:       token,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
		LastActiveAt: time.Now(),
	}

	query := `
		INSERT INTO users (id, token, display_name, created_at, last_active_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, token, display_name, created_at, last_active_at
	`

	err := r.db.QueryRow(ctx, query,
		user.ID, user.Token, user.DisplayName, user.CreatedAt, user.LastActiveAt,
	).Scan(
		&user.ID, &user.Token, &user.DisplayName, &user.CreatedAt, &user.LastActiveAt,
	)

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
	query := `
		SELECT id, token, display_name, created_at, last_active_at
		FROM users
		WHERE token = $1
	`

	err := r.db.QueryRow(ctx, query, token).Scan(
		&user.ID, &user.Token, &user.DisplayName, &user.CreatedAt, &user.LastActiveAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return user, nil
}

// UpdateLastActive updates the user's last active timestamp
func (r *UserRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	ctx, span := userTracer.Start(ctx, "UserRepository.UpdateLastActive")
	defer span.End()

	span.SetAttributes(attribute.String("user.id", userID.String()))

	query := `
		UPDATE users
		SET last_active_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, time.Now(), userID)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

