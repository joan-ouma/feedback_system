package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuizType represents the type of quiz
type QuizType string

const (
	QuizTypeMoodAssessment QuizType = "mood_assessment"
	QuizTypeStressLevel    QuizType = "stress_level"
	QuizTypeAnxietyCheck   QuizType = "anxiety_check"
	QuizTypeWellness       QuizType = "wellness"
)

// Quiz represents a quiz/test
type Quiz struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type        QuizType           `bson:"type" json:"type"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Questions   []QuizQuestion     `bson:"questions" json:"questions"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// GetIDString returns the ID as a string
func (q *Quiz) GetIDString() string {
	return q.ID.Hex()
}

// QuizQuestion represents a question in a quiz
type QuizQuestion struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	QuizID       primitive.ObjectID `bson:"quiz_id" json:"quiz_id"`
	Question     string             `bson:"question" json:"question"`
	Options      []string           `bson:"options" json:"options"` // Multiple choice options
	QuestionType string             `bson:"question_type" json:"question_type"` // "multiple_choice", "scale", "text"
	Order        int                `bson:"order" json:"order"`
}

// QuizResponse represents a user's response to a quiz
type QuizResponse struct {
	ID          primitive.ObjectID      `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID      `bson:"user_id" json:"user_id"`
	QuizID      primitive.ObjectID      `bson:"quiz_id" json:"quiz_id"`
	Answers     map[string]interface{}  `bson:"answers" json:"answers"` // question_id -> answer
	Score       int                     `bson:"score" json:"score"` // Calculated score
	Result      string                  `bson:"result" json:"result"` // Interpretation of the score
	CreatedAt   time.Time               `bson:"created_at" json:"created_at"`
}

// GetIDString returns the ID as a string
func (qr *QuizResponse) GetIDString() string {
	return qr.ID.Hex()
}

// QuizRecommendation represents recommendations based on quiz results
type QuizRecommendation struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id"`
	QuizResponseID primitive.ObjectID `bson:"quiz_response_id" json:"quiz_response_id"`
	Recommendations string           `bson:"recommendations" json:"recommendations"` // AI-generated recommendations
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
}

// GetIDString returns the ID as a string
func (qr *QuizRecommendation) GetIDString() string {
	return qr.ID.Hex()
}

