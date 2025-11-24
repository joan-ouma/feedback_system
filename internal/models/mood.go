package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	MoodType    MoodType           `bson:"mood_type" json:"mood_type"`
	MoodLevel   MoodLevel          `bson:"mood_level" json:"mood_level"`
	Score       int                `bson:"score" json:"score"` // 1-10 scale
	Notes       string             `bson:"notes" json:"notes"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	Date        time.Time          `bson:"date" json:"date"` // The date this mood entry is for
}

// GetIDString returns the ID as a string
func (m *MoodEntry) GetIDString() string {
	return m.ID.Hex()
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
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	MoodEntryID    primitive.ObjectID `bson:"mood_entry_id" json:"mood_entry_id"`
	Recommendations string            `bson:"recommendations" json:"recommendations"` // AI-generated recommendations
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// GetIDString returns the ID as a string
func (mr *MoodRecommendation) GetIDString() string {
	return mr.ID.Hex()
}

