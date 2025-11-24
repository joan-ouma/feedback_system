package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Type        FeedbackType       `bson:"type" json:"type"`
	Title       string             `bson:"title" json:"title"`
	Content     string             `bson:"content" json:"content"`
	Status      string             `bson:"status" json:"status"` // pending, reviewed, resolved
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// IsValid checks if the feedback is valid
func (f *Feedback) IsValid() bool {
	return f.Title != "" && f.Content != "" && !f.UserID.IsZero()
}

// GetIDString returns the ID as a string
func (f *Feedback) GetIDString() string {
	return f.ID.Hex()
}

