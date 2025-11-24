package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/joan/feedback-sys/internal/llm"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
)

var quizServiceTracer = otel.Tracer("service.quiz")

type QuizService struct {
	quizRepo  *repository.QuizRepository
	llmClient *llm.Client
}

func NewQuizService(quizRepo *repository.QuizRepository, llmClient *llm.Client) *QuizService {
	return &QuizService{
		quizRepo:  quizRepo,
		llmClient: llmClient,
	}
}

// GetQuiz gets a quiz by type
func (s *QuizService) GetQuiz(ctx context.Context, quizType models.QuizType) (*models.Quiz, error) {
	ctx, span := quizServiceTracer.Start(ctx, "QuizService.GetQuiz")
	defer span.End()

	return s.quizRepo.GetQuizByType(ctx, quizType)
}

// SubmitQuiz processes quiz answers and generates recommendations
func (s *QuizService) SubmitQuiz(ctx context.Context, userID primitive.ObjectID, quizID primitive.ObjectID, answers map[string]interface{}) (*models.QuizResponse, *models.QuizRecommendation, error) {
	ctx, span := quizServiceTracer.Start(ctx, "QuizService.SubmitQuiz")
	defer span.End()

	// Calculate score based on answers
	score, result := s.calculateScore(answers)

	response := &models.QuizResponse{
		UserID:  userID,
		QuizID:  quizID,
		Answers: answers,
		Score:   score,
		Result:  result,
	}

	if err := s.quizRepo.CreateQuizResponse(ctx, response); err != nil {
		span.RecordError(err)
		return nil, nil, fmt.Errorf("failed to create quiz response: %w", err)
	}

	// Generate recommendations using LLM
	var recommendation *models.QuizRecommendation
	if s.llmClient != nil {
		recText, err := s.generateQuizRecommendations(ctx, response, answers)
		if err == nil {
			recommendation = &models.QuizRecommendation{
				UserID:        userID,
				QuizResponseID: response.ID,
				Recommendations: recText,
			}
			if err := s.quizRepo.CreateQuizRecommendation(ctx, recommendation); err != nil {
				span.RecordError(err)
			}
		}
	}

	return response, recommendation, nil
}

// calculateScore calculates a score from quiz answers
func (s *QuizService) calculateScore(answers map[string]interface{}) (int, string) {
	totalScore := 0
	count := 0

	for _, answer := range answers {
		if str, ok := answer.(string); ok {
			// Try to extract number from answer
			if score := s.extractScoreFromAnswer(str); score > 0 {
				totalScore += score
				count++
			}
		} else if num, ok := answer.(float64); ok {
			totalScore += int(num)
			count++
		} else if num, ok := answer.(int); ok {
			totalScore += num
			count++
		}
	}

	if count == 0 {
		return 0, "Unable to calculate score"
	}

	avgScore := totalScore / count

	var result string
	if avgScore <= 2 {
		result = "Very Low - Consider seeking professional support"
	} else if avgScore <= 4 {
		result = "Low - Focus on self-care and support"
	} else if avgScore <= 6 {
		result = "Moderate - Some areas need attention"
	} else if avgScore <= 8 {
		result = "Good - You're doing well"
	} else {
		result = "Excellent - Keep up the great work!"
	}

	return avgScore, result
}

// extractScoreFromAnswer extracts a numeric score from answer text
func (s *QuizService) extractScoreFromAnswer(answer string) int {
	answer = strings.ToLower(answer)
	
	// Check for explicit numbers
	if strings.Contains(answer, "1") || strings.Contains(answer, "very low") || strings.Contains(answer, "very poor") {
		return 1
	}
	if strings.Contains(answer, "2") || strings.Contains(answer, "low") || strings.Contains(answer, "poor") {
		return 2
	}
	if strings.Contains(answer, "3") || strings.Contains(answer, "moderate") || strings.Contains(answer, "fair") {
		return 3
	}
	if strings.Contains(answer, "4") || strings.Contains(answer, "good") {
		return 4
	}
	if strings.Contains(answer, "5") || strings.Contains(answer, "excellent") || strings.Contains(answer, "very high") {
		return 5
	}

	return 0
}

// generateQuizRecommendations generates personalized recommendations based on quiz results
func (s *QuizService) generateQuizRecommendations(ctx context.Context, response *models.QuizResponse, answers map[string]interface{}) (string, error) {
	if s.llmClient == nil {
		return "", fmt.Errorf("LLM client not configured")
	}

	answersStr := ""
	for qID, ans := range answers {
		answersStr += fmt.Sprintf("Q%s: %v\n", qID, ans)
	}

	prompt := fmt.Sprintf(`Based on this quiz assessment, provide personalized recommendations:

Score: %d/10
Result: %s
Answers:
%s

Provide 3-5 actionable recommendations that are:
- Specific and practical
- Appropriate for the score level
- Focused on mental health and wellness
- Encouraging and supportive

Format as a numbered list.`, response.Score, response.Result, answersStr)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	recText, err := s.llmClient.Chat(ctx, messages, prompt)
	if err != nil {
		return "", err
	}

	return recText, nil
}

