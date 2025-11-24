package models

import (
	"time"

	"github.com/google/uuid"
)

// Consultation represents a mental health consultation session with LLM
type Consultation struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Message     string    `json:"message"`
	Response    string    `json:"response"`
	CreatedAt   time.Time `json:"created_at"`
	SessionID   uuid.UUID `json:"session_id"` // Groups related consultations
}

// ConsultationSession groups related consultation messages
type ConsultationSession struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

