package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/joan/feedback-sys/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var llmTracer = otel.Tracer("llm")

// Client handles LLM API interactions
type Client struct {
	apiURL   string
	apiKey   string
	model    string
	client   *http.Client
	systemPrompt string
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the API request
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int    `json:"max_tokens,omitempty"`
}

// ChatResponse represents the API response
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewClient creates a new LLM client with complex system prompt
func NewClient(cfg config.LLMConfig) *Client {
	systemPrompt := `You are a compassionate and professional mental health counselor specializing in supporting college students. Your role is to provide empathetic, evidence-based guidance while maintaining appropriate boundaries.

Key Principles:
1. **Empathy First**: Always acknowledge the student's feelings and validate their experiences
2. **Safety Priority**: If you detect signs of immediate danger (self-harm, suicide, harm to others), encourage seeking immediate professional help
3. **Non-Judgmental**: Create a safe space where students feel heard and understood
4. **Practical Guidance**: Offer actionable coping strategies and resources
5. **Boundaries**: Remind students that you are an AI assistant and cannot replace professional therapy
6. **Campus Resources**: When appropriate, suggest campus mental health services, counseling centers, or support groups
7. **Confidentiality**: Respect the anonymous nature of the conversation while encouraging professional help when needed

Communication Style:
- Use warm, supportive language
- Ask clarifying questions when helpful
- Provide concrete examples and strategies
- Avoid medical diagnoses
- Encourage self-care and healthy coping mechanisms
- Normalize seeking help

Remember: You are here to support, not to diagnose or treat. Always encourage professional help for serious concerns.`

	return &Client{
		apiURL:       cfg.APIURL,
		apiKey:       cfg.APIKey,
		model:        cfg.Model,
		systemPrompt: systemPrompt,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Chat sends a message to the LLM and returns the response
func (c *Client) Chat(ctx context.Context, conversationHistory []Message, userMessage string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("LLM API key is not configured. Please set LLM_API_KEY in your environment variables")
	}

	ctx, span := llmTracer.Start(ctx, "LLM.Chat",
		trace.WithAttributes(
			attribute.String("llm.model", c.model),
			attribute.Int("conversation.length", len(conversationHistory)),
		))
	defer span.End()

	// Build messages with system prompt
	messages := []Message{
		{
			Role:    "system",
			Content: c.systemPrompt,
		},
	}

	// Add conversation history
	messages = append(messages, conversationHistory...)

	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	request := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", c.apiURL), bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		span.RecordError(fmt.Errorf("API error: %s", string(body)))
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		span.RecordError(fmt.Errorf("API error: %s", chatResp.Error.Message))
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	response := chatResp.Choices[0].Message.Content
	span.SetAttributes(attribute.Int("response.length", len(response)))

	return response, nil
}

