package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents an anonymous user in the system
// Uses cryptographic token instead of email/phone
type User struct {
	ID           uuid.UUID `json:"id"`
	Token        string    `json:"token"`        // Unique anonymous identifier
	DisplayName  string    `json:"display_name"` // Optional display name
	CreatedAt    time.Time `json:"created_at"`
	LastActiveAt time.Time `json:"last_active_at"`
}

// IsValid checks if the user token is valid
func (u *User) IsValid() bool {
	return u.Token != "" && u.ID != uuid.Nil
}

