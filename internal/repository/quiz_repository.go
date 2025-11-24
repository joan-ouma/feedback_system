package repository

import (
	"context"
	"time"

	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel"
)

var quizTracer = otel.Tracer("repository.quiz")

type QuizRepository struct {
	db                    *database.DB
	quizzesCollection     *mongo.Collection
	questionsCollection   *mongo.Collection
	responsesCollection   *mongo.Collection
	recommendationsCollection *mongo.Collection
}

func NewQuizRepository(db *database.DB) *QuizRepository {
	return &QuizRepository{
		db:                         db,
		quizzesCollection:          db.Collection("quizzes"),
		questionsCollection:        db.Collection("quiz_questions"),
		responsesCollection:        db.Collection("quiz_responses"),
		recommendationsCollection:  db.Collection("quiz_recommendations"),
	}
}

// GetQuizByType gets a quiz by its type
func (r *QuizRepository) GetQuizByType(ctx context.Context, quizType models.QuizType) (*models.Quiz, error) {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.GetQuizByType")
	defer span.End()

	quiz := &models.Quiz{}
	err := r.quizzesCollection.FindOne(ctx, bson.M{"type": string(quizType)}).Decode(quiz)

	if err == mongo.ErrNoDocuments {
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
func (r *QuizRepository) GetQuizQuestions(ctx context.Context, quizID primitive.ObjectID) ([]models.QuizQuestion, error) {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.GetQuizQuestions")
	defer span.End()

	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := r.questionsCollection.Find(ctx, bson.M{"quiz_id": quizID}, opts)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var questions []models.QuizQuestion
	if err := cursor.All(ctx, &questions); err != nil {
		span.RecordError(err)
		return nil, err
	}

	return questions, nil
}

// CreateQuizResponse creates a quiz response
func (r *QuizRepository) CreateQuizResponse(ctx context.Context, response *models.QuizResponse) error {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.CreateQuizResponse")
	defer span.End()

	response.ID = primitive.NewObjectID()
	response.CreatedAt = time.Now()

	_, err := r.responsesCollection.InsertOne(ctx, response)
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

	rec.ID = primitive.NewObjectID()
	rec.CreatedAt = time.Now()

	_, err := r.recommendationsCollection.InsertOne(ctx, rec)
	if err != nil {
		span.RecordError(err)
		return err
	}

	return nil
}
