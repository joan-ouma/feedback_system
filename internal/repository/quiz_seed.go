package repository

import (
	"context"

	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SeedQuizzes seeds default quizzes into MongoDB
func SeedQuizzes(ctx context.Context, db *database.DB) error {
	quizzesCollection := db.Collection("quizzes")
	questionsCollection := db.Collection("quiz_questions")

	// Check if quizzes already exist
	count, _ := quizzesCollection.CountDocuments(ctx, bson.M{})
	if count > 0 {
		return nil // Already seeded
	}

	quizzes := []struct {
		quiz      models.Quiz
		questions []models.QuizQuestion
	}{
		{
			quiz: models.Quiz{
				ID:          primitive.NewObjectID(),
				Type:        models.QuizTypeMoodAssessment,
				Title:       "Daily Mood Assessment",
				Description: "Comprehensive assessment to understand your current mood and emotional state",
			},
			questions: []models.QuizQuestion{
				{
					ID:           primitive.NewObjectID(),
					Question:     "How would you rate your overall mood today?",
					QuestionType: "scale",
					Order:        1,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How energetic do you feel?",
					QuestionType: "scale",
					Order:        2,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How well did you sleep last night?",
					QuestionType: "scale",
					Order:        3,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How would you rate your ability to concentrate today?",
					QuestionType: "scale",
					Order:        4,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How connected do you feel to others today?",
					QuestionType: "scale",
					Order:        5,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How would you describe your stress level?",
					QuestionType: "scale",
					Order:        6,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How optimistic do you feel about today?",
					QuestionType: "scale",
					Order:        7,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How would you rate your overall emotional well-being?",
					QuestionType: "scale",
					Order:        8,
				},
			},
		},
		{
			quiz: models.Quiz{
				ID:          primitive.NewObjectID(),
				Type:        models.QuizTypeStressLevel,
				Title:       "Stress Level Check",
				Description: "Evaluate your current stress levels and identify potential stressors",
			},
			questions: []models.QuizQuestion{
				{
					ID:           primitive.NewObjectID(),
					Question:     "How stressed do you feel right now?",
					QuestionType: "scale",
					Order:        1,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How well are you sleeping?",
					QuestionType: "scale",
					Order:        2,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How would you rate your ability to concentrate?",
					QuestionType: "scale",
					Order:        3,
				},
			},
		},
		{
			quiz: models.Quiz{
				ID:          primitive.NewObjectID(),
				Type:        models.QuizTypeAnxietyCheck,
				Title:       "Anxiety Check",
				Description: "Assess your anxiety levels and get personalized recommendations",
			},
			questions: []models.QuizQuestion{
				{
					ID:           primitive.NewObjectID(),
					Question:     "How anxious do you feel?",
					QuestionType: "scale",
					Order:        1,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "Are you experiencing physical symptoms of anxiety?",
					QuestionType: "multiple_choice",
					Options:      []string{"Not at all", "Mildly", "Moderately", "Severely"},
					Order:        2,
				},
			},
		},
		{
			quiz: models.Quiz{
				ID:          primitive.NewObjectID(),
				Type:        models.QuizTypeWellness,
				Title:       "Wellness Check",
				Description: "Comprehensive wellness assessment covering multiple aspects",
			},
			questions: []models.QuizQuestion{
				{
					ID:           primitive.NewObjectID(),
					Question:     "How would you rate your overall wellness?",
					QuestionType: "scale",
					Order:        1,
				},
				{
					ID:           primitive.NewObjectID(),
					Question:     "How satisfied are you with your current lifestyle?",
					QuestionType: "scale",
					Order:        2,
				},
			},
		},
	}

	// Insert quizzes and questions
	for _, qz := range quizzes {
		// Insert quiz
		if _, err := quizzesCollection.InsertOne(ctx, qz.quiz); err != nil {
			if mongo.IsDuplicateKeyError(err) {
				continue // Already exists
			}
			return err
		}

		// Insert questions
		for _, q := range qz.questions {
			q.QuizID = qz.quiz.ID
			if _, err := questionsCollection.InsertOne(ctx, q); err != nil {
				if mongo.IsDuplicateKeyError(err) {
					continue
				}
				return err
			}
		}
	}

	return nil
}

