package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents an anonymous user in the system
// Uses cryptographic token instead of email/phone
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Token        string             `bson:"token" json:"token"`                // Unique anonymous identifier
	DisplayName  string             `bson:"display_name" json:"display_name"`  // Optional display name
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	LastActiveAt time.Time          `bson:"last_active_at" json:"last_active_at"`
}

// GetIDString returns the ID as a string
func (u *User) GetIDString() string {
	return u.ID.Hex()
}

// IsValid checks if the user token is valid
func (u *User) IsValid() bool {
	return u.Token != "" && !u.ID.IsZero()
}

