package models

import (
	"time"

	"github.com/google/uuid"
)

// MoodLevel represents the intensity of a mood
type MoodLevel string

const (
	MoodLevelVeryLow    MoodLevel = "very_low"    // 1-2
	MoodLevelLow        MoodLevel = "low"          // 3-4
	MoodLevelModerate   MoodLevel = "moderate"     // 5-6
	MoodLevelGood       MoodLevel = "good"         // 7-8
	MoodLevelExcellent  MoodLevel = "excellent"    // 9-10
)

// MoodType represents the type of mood
type MoodType string

const (
	MoodTypeHappy      MoodType = "happy"
	MoodTypeSad        MoodType = "sad"
	MoodTypeAnxious    MoodType = "anxious"
	MoodTypeStressed    MoodType = "stressed"
	MoodTypeCalm        MoodType = "calm"
	MoodTypeEnergetic   MoodType = "energetic"
	MoodTypeTired       MoodType = "tired"
	MoodTypeFrustrated  MoodType = "frustrated"
	MoodTypeNeutral     MoodType = "neutral"
)

// MoodEntry represents a daily mood tracking entry
type MoodEntry struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	MoodType    MoodType  `json:"mood_type"`
	MoodLevel   MoodLevel `json:"mood_level"`
	Score       int       `json:"score"` // 1-10 scale
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
	Date        time.Time `json:"date"` // The date this mood entry is for
}

// GetMoodLevelFromScore converts a score (1-10) to a MoodLevel
func GetMoodLevelFromScore(score int) MoodLevel {
	if score <= 2 {
		return MoodLevelVeryLow
	} else if score <= 4 {
		return MoodLevelLow
	} else if score <= 6 {
		return MoodLevelModerate
	} else if score <= 8 {
		return MoodLevelGood
	}
	return MoodLevelExcellent
}

// MoodRecommendation represents AI-generated recommendations based on mood
type MoodRecommendation struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	MoodEntryID uuid.UUID `json:"mood_entry_id"`
	Recommendations string `json:"recommendations"` // AI-generated recommendations
	CreatedAt   time.Time `json:"created_at"`
}

