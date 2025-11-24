package models

import (
	"time"

	"github.com/google/uuid"
)

// MotivationalQuote represents a daily motivational quote
type MotivationalQuote struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Quote       string    `json:"quote"`
	Author      string    `json:"author"`
	MoodType    MoodType  `json:"mood_type"` // Quote tailored for this mood
	MoodLevel   MoodLevel `json:"mood_level"` // Quote tailored for this mood level
	IsAI        bool      `json:"is_ai"` // Whether quote was AI-generated
	CreatedAt   time.Time `json:"created_at"`
	Date        time.Time `json:"date"` // The date this quote is for
}

