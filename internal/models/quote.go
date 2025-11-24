package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MotivationalQuote represents a daily motivational quote
type MotivationalQuote struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Quote     string             `bson:"quote" json:"quote"`
	Author    string             `bson:"author" json:"author"`
	MoodType  MoodType           `bson:"mood_type" json:"mood_type"`   // Quote tailored for this mood
	MoodLevel MoodLevel          `bson:"mood_level" json:"mood_level"` // Quote tailored for this mood level
	IsAI      bool               `bson:"is_ai" json:"is_ai"`           // Whether quote was AI-generated
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	Date      time.Time          `bson:"date" json:"date"` // The date this quote is for
}

// GetIDString returns the ID as a string
func (mq *MotivationalQuote) GetIDString() string {
	return mq.ID.Hex()
}

