package service

import (
	"context"
	"fmt"

	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
)

var feedbackServiceTracer = otel.Tracer("service.feedback")

type FeedbackService struct {
	feedbackRepo *repository.FeedbackRepository
}

func NewFeedbackService(feedbackRepo *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{feedbackRepo: feedbackRepo}
}

// SubmitFeedback creates a new feedback entry
func (s *FeedbackService) SubmitFeedback(ctx context.Context, userID primitive.ObjectID, feedbackType models.FeedbackType, title, content string) (*models.Feedback, error) {
	ctx, span := feedbackServiceTracer.Start(ctx, "FeedbackService.SubmitFeedback")
	defer span.End()

	feedback := &models.Feedback{
		UserID:  userID,
		Type:    feedbackType,
		Title:   title,
		Content: content,
		Status:  "pending",
	}

	if !feedback.IsValid() {
		return nil, fmt.Errorf("invalid feedback data")
	}

	err := s.feedbackRepo.Create(ctx, feedback)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to create feedback: %w", err)
	}

	return feedback, nil
}

// GetUserFeedbacks retrieves all feedbacks for a user
func (s *FeedbackService) GetUserFeedbacks(ctx context.Context, userID primitive.ObjectID) ([]*models.Feedback, error) {
	ctx, span := feedbackServiceTracer.Start(ctx, "FeedbackService.GetUserFeedbacks")
	defer span.End()

	feedbacks, err := s.feedbackRepo.GetByUserID(ctx, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get feedbacks: %w", err)
	}

	return feedbacks, nil
}

// GetFeedback retrieves a feedback by ID
func (s *FeedbackService) GetFeedback(ctx context.Context, feedbackID primitive.ObjectID) (*models.Feedback, error) {
	ctx, span := feedbackServiceTracer.Start(ctx, "FeedbackService.GetFeedback")
	defer span.End()

	feedback, err := s.feedbackRepo.GetByID(ctx, feedbackID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	return feedback, nil
}

