package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
		model:        "", // Not used - model is hardcoded in API calls
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
			attribute.Int("conversation.length", len(conversationHistory)),
		))
	defer span.End()

	// Check if using Gemini API (by URL only - model is optional)
	isGemini := strings.Contains(c.apiURL, "generativelanguage.googleapis.com")
	
	if isGemini {
		log.Printf("🔵 Detected Gemini API (URL: %s)", c.apiURL)
		return c.chatGemini(ctx, conversationHistory, userMessage)
	}
	
	log.Printf("🔵 Using OpenAI-compatible API (URL: %s, Model: %s)", c.apiURL, c.model)

	// Build messages with system prompt (OpenAI format)
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

	// OpenAI-compatible endpoint
	endpoint := fmt.Sprintf("%s/chat/completions", c.apiURL)

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	
	// Gemini uses API key in URL, OpenAI uses Bearer token
	if c.apiURL != "https://generativelanguage.googleapis.com/v1beta" && 
	   c.apiURL != "https://generativelanguage.googleapis.com/v1" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}

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

// listGeminiModels lists available Gemini models
func (c *Client) listGeminiModels(ctx context.Context) ([]string, error) {
	apiURL := c.apiURL
	if !strings.HasSuffix(apiURL, "/v1beta") && !strings.HasSuffix(apiURL, "/v1") {
		if strings.Contains(apiURL, "generativelanguage.googleapis.com") {
			apiURL = strings.TrimSuffix(apiURL, "/") + "/v1beta"
		}
	}
	
	endpoint := fmt.Sprintf("%s/models?key=%s", apiURL, c.apiKey)
	log.Printf("🔵 Listing available Gemini models: %s", endpoint)
	
	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list models request: %w", err)
	}
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send list models request: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read list models response: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Failed to list models: %s", string(body))
		return nil, fmt.Errorf("failed to list models: status %d, %s", resp.StatusCode, string(body))
	}
	
	type ModelInfo struct {
		Name         string   `json:"name"`
		DisplayName  string   `json:"displayName"`
		SupportedMethods []string `json:"supportedGenerationMethods"`
	}
	
	type ListModelsResponse struct {
		Models []ModelInfo `json:"models"`
	}
	
	var listResp ListModelsResponse
	if err := json.Unmarshal(body, &listResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal models list: %w", err)
	}
	
	var availableModels []string
	for _, model := range listResp.Models {
		// Check if model supports generateContent
		supportsGenerateContent := false
		for _, method := range model.SupportedMethods {
			if method == "generateContent" {
				supportsGenerateContent = true
				break
			}
		}
		if supportsGenerateContent {
			// Extract model name (format: models/gemini-xxx)
			modelName := strings.TrimPrefix(model.Name, "models/")
			
			// Skip experimental models (they have stricter free tier limits)
			if strings.Contains(modelName, "-exp") || strings.Contains(modelName, "experimental") {
				log.Printf("⏭️  Skipping experimental model: %s (%s)", modelName, model.DisplayName)
				continue
			}
			
			availableModels = append(availableModels, modelName)
			log.Printf("✅ Available model: %s (%s)", modelName, model.DisplayName)
		}
	}
	
	return availableModels, nil
}

// chatGemini handles Gemini API requests (different format)
func (c *Client) chatGemini(ctx context.Context, conversationHistory []Message, userMessage string) (string, error) {
	ctx, span := llmTracer.Start(ctx, "LLM.ChatGemini")
	defer span.End()

	// Gemini uses a different request format
	type GeminiContent struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
		Role string `json:"role,omitempty"`
	}

	type GeminiRequest struct {
		Contents []GeminiContent `json:"contents"`
		SystemInstruction *struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"systemInstruction,omitempty"`
		GenerationConfig struct {
			Temperature float64 `json:"temperature,omitempty"`
			MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
		} `json:"generationConfig,omitempty"`
	}

	// Build contents array
	contents := []GeminiContent{}
	
	// Add conversation history
	// Gemini uses "user" and "model" roles (not "assistant")
	for _, msg := range conversationHistory {
		role := msg.Role
		if role == "assistant" {
			role = "model" // Gemini uses "model" instead of "assistant"
		}
		contents = append(contents, GeminiContent{
			Parts: []struct {
				Text string `json:"text"`
			}{{Text: msg.Content}},
			Role: role,
		})
	}
	
	// Add current user message
	contents = append(contents, GeminiContent{
		Parts: []struct {
			Text string `json:"text"`
		}{{Text: userMessage}},
		Role: "user",
	})

	geminiReq := GeminiRequest{
		Contents: contents,
		GenerationConfig: struct {
			Temperature float64 `json:"temperature,omitempty"`
			MaxOutputTokens int `json:"maxOutputTokens,omitempty"`
		}{
			Temperature: 0.7,
			MaxOutputTokens: 1000,
		},
	}
	
	// Add system instruction if provided (some Gemini models support it)
	if c.systemPrompt != "" {
		geminiReq.SystemInstruction = &struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			Parts: []struct {
				Text string `json:"text"`
			}{{Text: c.systemPrompt}},
		}
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to marshal Gemini request: %w", err)
	}

	// Ensure API URL ends with /v1beta (Gemini API requirement)
	apiURL := c.apiURL
	if !strings.HasSuffix(apiURL, "/v1beta") && !strings.HasSuffix(apiURL, "/v1") {
		if strings.Contains(apiURL, "generativelanguage.googleapis.com") {
			apiURL = strings.TrimSuffix(apiURL, "/") + "/v1beta"
		}
	}
	
	// Try to get available models first, then use the first one that supports generateContent
	// For free tier, prioritize: gemini-pro, gemini-1.5-flash (best for free tier)
	modelName := "gemini-pro" // Default fallback for free tier
	
	// Try to list available models
	availableModels, err := c.listGeminiModels(ctx)
	if err == nil && len(availableModels) > 0 {
		// Prefer free-tier friendly models in order of preference
		// gemini-1.5-flash is fastest and cheapest for free tier
		// gemini-pro is the original stable model
		preferredModels := []string{"gemini-1.5-flash", "gemini-pro", "gemini-1.5-pro"}
		found := false
		for _, preferred := range preferredModels {
			for _, available := range availableModels {
				if available == preferred {
					modelName = preferred
					found = true
					log.Printf("✅ Selected preferred free-tier model: %s", modelName)
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			// Use first available non-experimental model
			modelName = availableModels[0]
			log.Printf("✅ Using first available model: %s", modelName)
		}
		log.Printf("📊 Model selection: %s (from %d available models)", modelName, len(availableModels))
	} else {
		log.Printf("⚠️  Could not list models, using default: %s (error: %v)", modelName, err)
	}
	
	endpoint := fmt.Sprintf("%s/models/%s:generateContent?key=%s", apiURL, modelName, c.apiKey)
	log.Printf("🔵 Calling Gemini API: %s", endpoint)
	log.Printf("🔵 Request body: %s", string(jsonData))
	
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to create Gemini request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to send Gemini request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		return "", fmt.Errorf("failed to read Gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))
		log.Printf("❌ Gemini API error - Status: %d, Response: %s", resp.StatusCode, string(body))
		span.RecordError(fmt.Errorf("Gemini API error: %s", string(body)))
		return "", fmt.Errorf("Gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse Gemini response
	type GeminiResponse struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		log.Printf("❌ Failed to unmarshal Gemini response: %v, Body: %s", err, string(body))
		span.RecordError(err)
		return "", fmt.Errorf("failed to unmarshal Gemini response: %w", err)
	}

	if geminiResp.Error != nil {
		log.Printf("❌ Gemini API error in response: %s", geminiResp.Error.Message)
		span.RecordError(fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message))
		return "", fmt.Errorf("Gemini API error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 {
		log.Printf("❌ No candidates in Gemini response. Full response: %s", string(body))
		return "", fmt.Errorf("no candidates in Gemini response")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		log.Printf("❌ No parts in Gemini candidate. Full response: %s", string(body))
		return "", fmt.Errorf("no response content from Gemini")
	}

	response := geminiResp.Candidates[0].Content.Parts[0].Text
	span.SetAttributes(attribute.Int("response.length", len(response)))

	return response, nil
}

