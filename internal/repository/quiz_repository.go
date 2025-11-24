package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/otel"
)

var quizTracer = otel.Tracer("repository.quiz")

type QuizRepository struct {
	db *database.DB
}

func NewQuizRepository(db *database.DB) *QuizRepository {
	return &QuizRepository{db: db}
}

// GetQuizByType gets a quiz by its type
func (r *QuizRepository) GetQuizByType(ctx context.Context, quizType models.QuizType) (*models.Quiz, error) {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.GetQuizByType")
	defer span.End()

	quiz := &models.Quiz{}
	query := `
		SELECT id, type, title, description, created_at
		FROM quizzes
		WHERE type = $1
		LIMIT 1
	`

	err := r.db.QueryRow(ctx, query, string(quizType)).Scan(
		&quiz.ID, &quiz.Type, &quiz.Title, &quiz.Description, &quiz.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Load questions
	questions, err := r.GetQuizQuestions(ctx, quiz.ID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	quiz.Questions = questions

	return quiz, nil
}

// GetQuizQuestions gets all questions for a quiz
func (r *QuizRepository) GetQuizQuestions(ctx context.Context, quizID uuid.UUID) ([]models.QuizQuestion, error) {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.GetQuizQuestions")
	defer span.End()

	query := `
		SELECT id, quiz_id, question, options, question_type, "order"
		FROM quiz_questions
		WHERE quiz_id = $1
		ORDER BY "order" ASC
	`

	rows, err := r.db.Query(ctx, query, quizID)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer rows.Close()

	var questions []models.QuizQuestion
	for rows.Next() {
		q := models.QuizQuestion{}
		var optionsJSON []byte
		err := rows.Scan(
			&q.ID, &q.QuizID, &q.Question, &optionsJSON, &q.QuestionType, &q.Order,
		)
		if err != nil {
			span.RecordError(err)
			return nil, err
		}

		// Parse JSON options
		if len(optionsJSON) > 0 {
			if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
				span.RecordError(err)
				return nil, err
			}
		}

		questions = append(questions, q)
	}

	return questions, rows.Err()
}

// CreateQuizResponse creates a quiz response
func (r *QuizRepository) CreateQuizResponse(ctx context.Context, response *models.QuizResponse) error {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.CreateQuizResponse")
	defer span.End()

	response.ID = uuid.New()
	response.CreatedAt = time.Now()

	answersJSON, err := json.Marshal(response.Answers)
	if err != nil {
		span.RecordError(err)
		return err
	}

	query := `
		INSERT INTO quiz_responses (id, user_id, quiz_id, answers, score, result, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.Exec(ctx, query,
		response.ID, response.UserID, response.QuizID, answersJSON,
		response.Score, response.Result, response.CreatedAt,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

// CreateQuizRecommendation creates a recommendation for a quiz response
func (r *QuizRepository) CreateQuizRecommendation(ctx context.Context, rec *models.QuizRecommendation) error {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.CreateQuizRecommendation")
	defer span.End()

	rec.ID = uuid.New()
	rec.CreatedAt = time.Now()

	query := `
		INSERT INTO quiz_recommendations (id, user_id, quiz_response_id, recommendations, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(ctx, query,
		rec.ID, rec.UserID, rec.QuizResponseID, rec.Recommendations, rec.CreatedAt,
	)

	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}

