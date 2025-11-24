package models

import (
	"time"

	"github.com/google/uuid"
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
	ID          uuid.UUID `json:"id"`
	Type        QuizType  `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Questions   []QuizQuestion `json:"questions"`
	CreatedAt   time.Time `json:"created_at"`
}

// QuizQuestion represents a question in a quiz
type QuizQuestion struct {
	ID          uuid.UUID `json:"id"`
	QuizID      uuid.UUID `json:"quiz_id"`
	Question    string    `json:"question"`
	Options     []string  `json:"options"` // Multiple choice options
	QuestionType string   `json:"question_type"` // "multiple_choice", "scale", "text"
	Order       int       `json:"order"`
}

// QuizResponse represents a user's response to a quiz
type QuizResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	QuizID      uuid.UUID `json:"quiz_id"`
	Answers     map[string]interface{} `json:"answers"` // question_id -> answer
	Score       int       `json:"score"` // Calculated score
	Result      string    `json:"result"` // Interpretation of the score
	CreatedAt   time.Time `json:"created_at"`
}

// QuizRecommendation represents recommendations based on quiz results
type QuizRecommendation struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	QuizResponseID uuid.UUID `json:"quiz_response_id"`
	Recommendations string `json:"recommendations"` // AI-generated recommendations
	CreatedAt   time.Time `json:"created_at"`
}

