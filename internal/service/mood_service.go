// +build ignore

package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/joan/feedback-sys/internal/llm"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/repository"
	"go.opentelemetry.io/otel"
)

var moodServiceTracer = otel.Tracer("service.mood")

type MoodService struct {
	moodRepo    *repository.MoodRepository
	quoteRepo   *repository.QuoteRepository
	llmClient   *llm.Client
}

func NewMoodService(moodRepo *repository.MoodRepository, quoteRepo *repository.QuoteRepository, llmClient *llm.Client) *MoodService {
	return &MoodService{
		moodRepo:  moodRepo,
		quoteRepo: quoteRepo,
		llmClient: llmClient,
	}
}

// RecordMood records a mood entry and generates recommendations
func (s *MoodService) RecordMood(ctx context.Context, userID uuid.UUID, moodType models.MoodType, score int, notes string) (*models.MoodEntry, *models.MoodRecommendation, error) {
	ctx, span := moodServiceTracer.Start(ctx, "MoodService.RecordMood")
	defer span.End()

	moodLevel := models.GetMoodLevelFromScore(score)
	entry := &models.MoodEntry{
		UserID:    userID,
		MoodType:  moodType,
		MoodLevel: moodLevel,
		Score:     score,
		Notes:     notes,
		Date:      time.Now(),
	}

	if err := s.moodRepo.CreateMoodEntry(ctx, entry); err != nil {
		span.RecordError(err)
		return nil, nil, fmt.Errorf("failed to create mood entry: %w", err)
	}

	// Generate recommendations using LLM
	var recommendation *models.MoodRecommendation
	if s.llmClient != nil {
		recText, err := s.generateRecommendations(ctx, entry)
		if err == nil {
			recommendation = &models.MoodRecommendation{
				UserID:        userID,
				MoodEntryID:   entry.ID,
				Recommendations: recText,
			}
			if err := s.moodRepo.CreateMoodRecommendation(ctx, recommendation); err != nil {
				// Log error but don't fail
				span.RecordError(err)
			}
		}
	}

	// Generate daily quote based on mood
	go func() {
		_ = s.generateDailyQuote(context.Background(), userID, moodType, moodLevel)
	}()

	return entry, recommendation, nil
}

// generateRecommendations generates personalized recommendations using LLM
func (s *MoodService) generateRecommendations(ctx context.Context, entry *models.MoodEntry) (string, error) {
	if s.llmClient == nil {
		return "", fmt.Errorf("LLM client not configured")
	}

	prompt := fmt.Sprintf(`Based on the following mood entry, provide 3-5 personalized, actionable recommendations to help improve their mental well-being:

Mood Type: %s
Mood Level: %s (Score: %d/10)
Notes: %s

Provide practical, empathetic recommendations that are:
- Specific and actionable
- Appropriate for their current mood level
- Focused on self-care and mental health
- Encouraging but realistic

Format as a numbered list.`, entry.MoodType, entry.MoodLevel, entry.Score, entry.Notes)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	response, err := s.llmClient.Chat(ctx, messages, prompt)
	if err != nil {
		return "", err
	}

	return response, nil
}

// generateDailyQuote generates a motivational quote based on mood
func (s *MoodService) generateDailyQuote(ctx context.Context, userID uuid.UUID, moodType models.MoodType, moodLevel models.MoodLevel) error {
	// Check if quote already exists for today
	existing, _ := s.quoteRepo.GetQuoteForDate(ctx, userID, time.Now())
	if existing != nil {
		return nil // Quote already exists
	}

	var quoteText, author string
	var err error

	if s.llmClient != nil {
		// Generate AI quote
		prompt := fmt.Sprintf(`Generate a short, uplifting motivational quote (1-2 sentences) appropriate for someone feeling %s with a mood level of %s. 
The quote should be:
- Inspiring and positive
- Relevant to their emotional state
- Encouraging but not dismissive
- Suitable for a student

Respond with only the quote text, no additional explanation.`, moodType, moodLevel)

		messages := []llm.Message{
			{Role: "user", Content: prompt},
		}

		quoteText, err = s.llmClient.Chat(ctx, messages, prompt)
		if err == nil {
			author = "AI Generated"
		}
	}

	// Fallback to curated quotes if LLM fails
	if quoteText == "" {
		quoteText, author = s.getCuratedQuote(moodType, moodLevel)
	}

	quote := &models.MotivationalQuote{
		UserID:    userID,
		Quote:     quoteText,
		Author:    author,
		MoodType:  moodType,
		MoodLevel: moodLevel,
		IsAI:      s.llmClient != nil && err == nil,
		Date:      time.Now(),
	}

	return s.quoteRepo.CreateQuote(ctx, quote)
}

// getCuratedQuote returns a curated quote based on mood
func (s *MoodService) getCuratedQuote(moodType models.MoodType, moodLevel models.MoodLevel) (string, string) {
	quotes := map[string][]struct {
		quote  string
		author string
	}{
		"very_low": {
			{"It's okay to not be okay. Every storm runs out of rain.", "Maya Angelou"},
			{"You are stronger than you think. This feeling will pass.", "Unknown"},
			{"Even the darkest night will end and the sun will rise.", "Victor Hugo"},
		},
		"low": {
			{"Progress, not perfection. Small steps forward are still steps forward.", "Unknown"},
			{"You don't have to be great to start, but you have to start to be great.", "Zig Ziglar"},
			{"The only way out is through. Keep going.", "Robert Frost"},
		},
		"moderate": {
			{"Today is a new day. Make it count.", "Unknown"},
			{"You are capable of amazing things.", "Unknown"},
			{"Small progress is still progress.", "Unknown"},
		},
		"good": {
			{"Keep up the great work! You're doing amazing.", "Unknown"},
			{"Your positive energy is contagious. Keep shining!", "Unknown"},
		},
		"excellent": {
			{"You're unstoppable! Keep that energy going!", "Unknown"},
			{"Your positivity is inspiring. Keep spreading joy!", "Unknown"},
		},
	}

	levelQuotes, exists := quotes[string(moodLevel)]
	if !exists || len(levelQuotes) == 0 {
		levelQuotes = quotes["moderate"]
	}

	// Simple selection based on day
	selected := levelQuotes[time.Now().Day()%len(levelQuotes)]
	return selected.quote, selected.author
}

// GetMoodHistory gets mood history for a user
func (s *MoodService) GetMoodHistory(ctx context.Context, userID uuid.UUID, days int) ([]*models.MoodEntry, error) {
	ctx, span := moodServiceTracer.Start(ctx, "MoodService.GetMoodHistory")
	defer span.End()

	return s.moodRepo.GetMoodHistory(ctx, userID, days)
}

// GetTodayMood gets today's mood entry
func (s *MoodService) GetTodayMood(ctx context.Context, userID uuid.UUID) (*models.MoodEntry, error) {
	ctx, span := moodServiceTracer.Start(ctx, "MoodService.GetTodayMood")
	defer span.End()

	return s.moodRepo.GetMoodEntryByDate(ctx, userID, time.Now())
}

// GetDailyQuote gets today's motivational quote
func (s *MoodService) GetDailyQuote(ctx context.Context, userID uuid.UUID) (*models.MotivationalQuote, error) {
	ctx, span := moodServiceTracer.Start(ctx, "MoodService.GetDailyQuote")
	defer span.End()

	return s.quoteRepo.GetQuoteForDate(ctx, userID, time.Now())
}

