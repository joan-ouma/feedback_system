package models

import (
	"time"

	"github.com/google/uuid"
)

// FeedbackType represents the type of feedback
type FeedbackType string

const (
	FeedbackTypeGeneral    FeedbackType = "general"
	FeedbackTypeMentalHealth FeedbackType = "mental_health"
	FeedbackTypeCampus     FeedbackType = "campus"
	FeedbackTypeOther      FeedbackType = "other"
)

// Feedback represents an anonymous feedback submission
type Feedback struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"user_id"`
	Type        FeedbackType `json:"type"`
	Title       string      `json:"title"`
	Content     string      `json:"content"`
	Status      string      `json:"status"` // pending, reviewed, resolved
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// IsValid checks if the feedback is valid
func (f *Feedback) IsValid() bool {
	return f.Title != "" && f.Content != "" && f.UserID != uuid.Nil
}

