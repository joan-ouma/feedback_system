package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Consultation represents a mental health consultation session with LLM
type Consultation struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Message     string             `bson:"message" json:"message"`
	Response    string             `bson:"response" json:"response"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	SessionID   primitive.ObjectID `bson:"session_id" json:"session_id"` // Groups related consultations
}

// ConsultationSession groups related consultation messages
type ConsultationSession struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// GetIDString returns the ID as a string
func (c *Consultation) GetIDString() string {
	return c.ID.Hex()
}

// GetIDString returns the ID as a string
func (cs *ConsultationSession) GetIDString() string {
	return cs.ID.Hex()
}

