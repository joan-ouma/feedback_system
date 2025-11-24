package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/joan/feedback-sys/internal/llm"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/repository"
	"go.opentelemetry.io/otel"
)

var consultationServiceTracer = otel.Tracer("service.consultation")

type ConsultationService struct {
	consultationRepo *repository.ConsultationRepository
	llmClient        *llm.Client
}

func NewConsultationService(consultationRepo *repository.ConsultationRepository, llmClient *llm.Client) *ConsultationService {
	return &ConsultationService{
		consultationRepo: consultationRepo,
		llmClient:        llmClient,
	}
}

// StartSession creates a new consultation session
func (s *ConsultationService) StartSession(ctx context.Context, userID uuid.UUID) (*models.ConsultationSession, error) {
	ctx, span := consultationServiceTracer.Start(ctx, "ConsultationService.StartSession")
	defer span.End()

	session, err := s.consultationRepo.CreateSession(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetOrCreateSession gets an existing session or creates a new one
func (s *ConsultationService) GetOrCreateSession(ctx context.Context, userID uuid.UUID, sessionID *uuid.UUID) (*models.ConsultationSession, error) {
	ctx, span := consultationServiceTracer.Start(ctx, "ConsultationService.GetOrCreateSession")
	defer span.End()

	session, err := s.consultationRepo.GetOrCreateSession(ctx, userID, sessionID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get or create session: %w", err)
	}

	return session, nil
}

// SendMessage sends a message to the LLM and saves the consultation
func (s *ConsultationService) SendMessage(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID, message string) (*models.Consultation, error) {
	ctx, span := consultationServiceTracer.Start(ctx, "ConsultationService.SendMessage")
	defer span.End()

	// Get conversation history
	history, err := s.consultationRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}

	// Convert to LLM message format
	conversationHistory := make([]llm.Message, 0, len(history)*2)
	for _, consultation := range history {
		conversationHistory = append(conversationHistory,
			llm.Message{Role: "user", Content: consultation.Message},
			llm.Message{Role: "assistant", Content: consultation.Response},
		)
	}

	// Get LLM response
	response, err := s.llmClient.Chat(ctx, conversationHistory, message)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	// Save consultation
	consultation := &models.Consultation{
		SessionID: sessionID,
		UserID:    userID,
		Message:   message,
		Response:  response,
	}

	err = s.consultationRepo.Create(ctx, consultation)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to save consultation: %w", err)
	}

	return consultation, nil
}

// GetSessionHistory retrieves all messages in a session
func (s *ConsultationService) GetSessionHistory(ctx context.Context, sessionID uuid.UUID) ([]*models.Consultation, error) {
	ctx, span := consultationServiceTracer.Start(ctx, "ConsultationService.GetSessionHistory")
	defer span.End()

	consultations, err := s.consultationRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get session history: %w", err)
	}

	return consultations, nil
}

