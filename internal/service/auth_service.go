package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/repository"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var authTracer = otel.Tracer("service.auth")

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

// GenerateAnonymousToken generates a cryptographically secure anonymous token
// This is a clever way to sign up without email/phone - users get a unique token
// that they can save to their browser/local storage to maintain their identity
func (s *AuthService) GenerateAnonymousToken() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(bytes)
	return token, nil
}

// SignUp creates a new anonymous user with a generated token
func (s *AuthService) SignUp(ctx context.Context, displayName string) (*models.User, string, error) {
	ctx, span := authTracer.Start(ctx, "AuthService.SignUp")
	defer span.End()

	// Generate a unique token
	token, err := s.GenerateAnonymousToken()
	if err != nil {
		span.RecordError(err)
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	span.SetAttributes(attribute.String("user.display_name", displayName))

	// Create user with the token
	user, err := s.userRepo.Create(ctx, token, displayName)
	if err != nil {
		span.RecordError(err)
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	return user, token, nil
}

// Authenticate retrieves a user by their token
func (s *AuthService) Authenticate(ctx context.Context, token string) (*models.User, error) {
	ctx, span := authTracer.Start(ctx, "AuthService.Authenticate")
	defer span.End()

	user, err := s.userRepo.GetByToken(ctx, token)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("invalid token")
	}

	// Update last active timestamp
	go func() {
		_ = s.userRepo.UpdateLastActive(context.Background(), user.ID)
	}()

	return user, nil
}

// GetUser retrieves a user by ID
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	ctx, span := authTracer.Start(ctx, "AuthService.GetUser")
	defer span.End()

	// This would need a GetByID method in repository, but for now we'll use token
	// In a real scenario, you'd add GetByID to the repository
	return nil, fmt.Errorf("not implemented")
}

