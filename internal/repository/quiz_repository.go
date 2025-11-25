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

// GetUserQuizResponses gets all quiz responses for a user, grouped by quiz type
func (r *QuizRepository) GetUserQuizResponses(ctx context.Context, userID primitive.ObjectID) (map[string]*models.QuizResponse, error) {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.GetUserQuizResponses")
	defer span.End()

	// Get all responses for this user, sorted by created_at descending
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.responsesCollection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var responses []models.QuizResponse
	if err := cursor.All(ctx, &responses); err != nil {
		span.RecordError(err)
		return nil, err
	}

	// Get quiz types for each response
	quizMap := make(map[primitive.ObjectID]models.QuizType)
	for _, resp := range responses {
		if _, exists := quizMap[resp.QuizID]; !exists {
			quiz := &models.Quiz{}
			if err := r.quizzesCollection.FindOne(ctx, bson.M{"_id": resp.QuizID}).Decode(quiz); err == nil {
				quizMap[resp.QuizID] = quiz.Type
			}
		}
	}

	// Group by quiz type, keeping only the most recent response for each type
	result := make(map[string]*models.QuizResponse)
	for i := range responses {
		quizType := quizMap[responses[i].QuizID]
		typeStr := string(quizType)
		// Only keep the first (most recent) response for each quiz type
		if _, exists := result[typeStr]; !exists {
			result[typeStr] = &responses[i]
		}
	}

	return result, nil
}

// GetQuizRecommendation gets the recommendation for a quiz response
func (r *QuizRepository) GetQuizRecommendation(ctx context.Context, responseID primitive.ObjectID) (*models.QuizRecommendation, error) {
	ctx, span := quizTracer.Start(ctx, "QuizRepository.GetQuizRecommendation")
	defer span.End()

	rec := &models.QuizRecommendation{}
	err := r.recommendationsCollection.FindOne(ctx, bson.M{"quiz_response_id": responseID}).Decode(rec)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	return rec, nil
}
