package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ResponseData represents structured data for the professional response template
type ResponseData struct {
	EmpathyTitle       string
	EmpathyMessage     string
	ShowCrisisResources bool
	Tips               []Tip
}

// Tip represents an actionable tip in the response
type Tip struct {
	Title       string
	Description string
	Icon        string
	IconColor   string
	ColorClass  string
}

type ConsultationHandler struct {
	consultationService *service.ConsultationService
	authService         *service.AuthService
	templateDir         string
}

func NewConsultationHandler(consultationService *service.ConsultationService, authService *service.AuthService, templateDir string) *ConsultationHandler {
	return &ConsultationHandler{
		consultationService: consultationService,
		authService:         authService,
		templateDir:         templateDir,
	}
}

// Chat renders the consultation chat page
func (h *ConsultationHandler) Chat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	templatePath := filepath.Join(h.templateDir, "consultation.html")
	http.ServeFile(w, r, templatePath)
}

// SendMessage handles sending a message to the LLM
func (h *ConsultationHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetTokenFromContext(r.Context())
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.Authenticate(r.Context(), token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		SessionID string `json:"session_id" form:"session_id"`
		Message   string `json:"message" form:"message"`
	}

	// Support both JSON and form-encoded data with improved parsing
	contentType := r.Header.Get("Content-Type")
	log.Printf("🔵 Request Content-Type: %s, Method: %s", contentType, r.Method)
	
	// Check Content-Type to decide parsing method
	if strings.Contains(contentType, "application/json") {
		// Parse as JSON
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields() // Reject unknown fields for better security
		if err := decoder.Decode(&req); err != nil {
			log.Printf("❌ JSON decode error: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Invalid JSON request: " + err.Error(),
			})
			return
		}
		log.Printf("✅ Parsed as JSON: session_id=%s, message length=%d", req.SessionID, len(req.Message))
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") || 
		strings.Contains(contentType, "multipart/form-data") {
		// Parse as form data
		if err := r.ParseForm(); err != nil {
			log.Printf("❌ Form parse error: %v, Content-Type: %s", err, contentType)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Invalid form data: " + err.Error(),
			})
			return
		}
		req.SessionID = strings.TrimSpace(r.FormValue("session_id"))
		req.Message = strings.TrimSpace(r.FormValue("message"))
		log.Printf("✅ Parsed as form data: session_id=%s, message length=%d", req.SessionID, len(req.Message))
	} else {
		// Try to parse as form data as fallback
		if err := r.ParseForm(); err == nil {
			req.SessionID = strings.TrimSpace(r.FormValue("session_id"))
			req.Message = strings.TrimSpace(r.FormValue("message"))
			log.Printf("✅ Parsed as form data (fallback): session_id=%s, message length=%d", req.SessionID, len(req.Message))
		} else {
			log.Printf("❌ Unknown Content-Type and form parse failed: %s", contentType)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Unsupported Content-Type. Please use application/json or application/x-www-form-urlencoded",
			})
			return
		}
	}

	// Validate and sanitize message
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		log.Printf("❌ Empty message in request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Message cannot be empty",
		})
		return
	}

	// Validate message length (prevent abuse)
	const maxMessageLength = 5000
	if len(req.Message) > maxMessageLength {
		log.Printf("❌ Message too long: %d characters", len(req.Message))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Message is too long. Maximum length is %d characters", maxMessageLength),
		})
		return
	}

	// Validate and parse session ID
	var sessionIDPtr *primitive.ObjectID
	if req.SessionID != "" {
		// Sanitize session ID
		req.SessionID = strings.TrimSpace(req.SessionID)
		
		// Validate ObjectID format (24 hex characters)
		if len(req.SessionID) != 24 {
			log.Printf("❌ Invalid session ID length: %d (expected 24)", len(req.SessionID))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Invalid session ID format",
			})
			return
		}
		
		sessionID, err := primitive.ObjectIDFromHex(req.SessionID)
		if err != nil {
			log.Printf("❌ Invalid session ID: %s, error: %v", req.SessionID, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "Invalid session ID: " + err.Error(),
			})
			return
		}
		sessionIDPtr = &sessionID
	}
	session, err := h.consultationService.GetOrCreateSession(r.Context(), user.ID, sessionIDPtr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send message with improved error handling
	consultation, err := h.consultationService.SendMessage(r.Context(), user.ID, session.ID, req.Message)
	if err != nil {
		log.Printf("❌ Consultation error: %v", err)
		
		// Categorize errors for better user experience
		errorMsg := err.Error()
		statusCode := http.StatusInternalServerError
		userFriendlyMsg := "I'm sorry, I'm having trouble responding right now. Please try again in a moment."
		
		// Check for specific error types
		if strings.Contains(errorMsg, "LLM API key is not configured") || 
		   strings.Contains(errorMsg, "API key") ||
		   strings.Contains(errorMsg, "not configured") {
			statusCode = http.StatusServiceUnavailable
			userFriendlyMsg = "I'm sorry, but the consultation service is currently unavailable. Please contact your campus mental health services directly for support."
		} else if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "context deadline exceeded") {
			statusCode = http.StatusGatewayTimeout
			userFriendlyMsg = "The request took too long to process. Please try again with a shorter message."
		} else if strings.Contains(errorMsg, "rate limit") || strings.Contains(errorMsg, "quota") {
			statusCode = http.StatusTooManyRequests
			userFriendlyMsg = "Too many requests. Please wait a moment before trying again."
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		
		response := map[string]interface{}{
			"error": errorMsg,
			"consultation": map[string]interface{}{
				"response": userFriendlyMsg,
			},
			"session_id": session.ID.Hex(),
		}
		
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("❌ Failed to encode error response: %v", err)
		}
		return
	}

	// Validate consultation response before sending
	if consultation == nil {
		log.Printf("❌ Consultation is nil after successful service call")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Failed to generate response",
			"consultation": map[string]interface{}{
				"response": "I'm sorry, I'm having trouble responding right now. Please try again in a moment.",
			},
			"session_id": session.ID.Hex(),
		})
		return
	}

	// Ensure response is not empty
	if consultation.Response == "" {
		log.Printf("❌ Empty consultation response")
		consultation.Response = "I'm sorry, I didn't receive a proper response. Could you please rephrase your question?"
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"consultation": consultation,
		"session_id":   session.ID.Hex(),
	}); err != nil {
		log.Printf("❌ Failed to encode success response: %v", err)
	}
}

// GetHistory retrieves consultation history for a session
func (h *ConsultationHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetTokenFromContext(r.Context())
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := h.authService.Authenticate(r.Context(), token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	sessionID, err := primitive.ObjectIDFromHex(vars["session_id"])
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	history, err := h.consultationService.GetSessionHistory(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// StartSession starts a new consultation session
func (h *ConsultationHandler) StartSession(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetTokenFromContext(r.Context())
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.authService.Authenticate(r.Context(), token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	session, err := h.consultationService.StartSession(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// ShowResponse renders the professional response template
func (h *ConsultationHandler) ShowResponse(w http.ResponseWriter, r *http.Request) {
	token := middleware.GetTokenFromContext(r.Context())
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	_, err := h.authService.Authenticate(r.Context(), token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Load the response template
	tmpl, err := template.ParseFiles(filepath.Join(h.templateDir, "response.html"))
	if err != nil {
		log.Printf("❌ Failed to load response template: %v", err)
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}

	// Example data - in production, this would come from LLM analysis or user input
	data := ResponseData{
		EmpathyTitle: "I'm truly sorry you're feeling this way",
		EmpathyMessage: "That sounds incredibly difficult, and it takes courage to even talk about it. You're not alone, and there are people who want to help you through this.",
		ShowCrisisResources: true,
		Tips: []Tip{
			{
				Title:       "Break things down",
				Description: "When overwhelmed, split tasks into tiny, manageable steps. Focus on just one thing at a time.",
				Icon:        "fas fa-puzzle-piece",
				IconColor:   "text-green-500",
				ColorClass:  "border-green-400",
			},
			{
				Title:       "Do something you enjoy",
				Description: "Listen to a favorite song or watch a funny video, even for just a few minutes. Small moments of joy matter.",
				Icon:        "fas fa-music",
				IconColor:   "text-purple-500",
				ColorClass:  "border-purple-400",
			},
			{
				Title:       "Connect with others",
				Description: "Reach out to a friend, family member, or support group. You don't have to go through this alone.",
				Icon:        "fas fa-users",
				IconColor:   "text-blue-500",
				ColorClass:  "border-blue-400",
			},
			{
				Title:       "Practice self-care",
				Description: "Take a warm bath, go for a walk, or do some deep breathing. Your well-being matters.",
				Icon:        "fas fa-spa",
				IconColor:   "text-pink-500",
				ColorClass:  "border-pink-400",
			},
		},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("❌ Failed to execute response template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

// GenerateResponseData creates ResponseData from LLM response and user message
// This can be enhanced with LLM analysis to detect crisis situations and generate appropriate tips
func GenerateResponseData(llmResponse string, userMessage string) ResponseData {
	// Detect if crisis resources should be shown based on keywords
	showCrisis := strings.Contains(strings.ToLower(userMessage), "suicide") ||
		strings.Contains(strings.ToLower(userMessage), "hurt myself") ||
		strings.Contains(strings.ToLower(userMessage), "end it all") ||
		strings.Contains(strings.ToLower(userMessage), "not worth living")

	// Default empathy message
	empathyTitle := "I hear you, and I'm here to help"
	empathyMessage := "Thank you for sharing what you're going through. It takes strength to reach out, and I want you to know that support is available."

	// Default tips - can be customized based on LLM response analysis
	tips := []Tip{
		{
			Title:       "Break things down",
			Description: "When overwhelmed, split tasks into tiny, manageable steps. Focus on just one thing at a time.",
			Icon:        "fas fa-puzzle-piece",
			IconColor:   "text-green-500",
			ColorClass:  "border-green-400",
		},
		{
			Title:       "Practice mindfulness",
			Description: "Take a few deep breaths. Try the 4-7-8 technique: inhale for 4, hold for 7, exhale for 8.",
			Icon:        "fas fa-leaf",
			IconColor:   "text-teal-500",
			ColorClass:  "border-teal-400",
		},
		{
			Title:       "Connect with others",
			Description: "Reach out to a friend, family member, or support group. You don't have to go through this alone.",
			Icon:        "fas fa-users",
			IconColor:   "text-blue-500",
			ColorClass:  "border-blue-400",
		},
		{
			Title:       "Seek professional help",
			Description: "Consider speaking with a counselor or therapist. Professional support can make a significant difference.",
			Icon:        "fas fa-user-md",
			IconColor:   "text-indigo-500",
			ColorClass:  "border-indigo-400",
		},
	}

	return ResponseData{
		EmpathyTitle:       empathyTitle,
		EmpathyMessage:     empathyMessage,
		ShowCrisisResources: showCrisis,
		Tips:               tips,
	}
}

func (h *ConsultationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/consultation", h.Chat).Methods("GET")
	router.HandleFunc("/consultation/response", h.ShowResponse).Methods("GET")
	router.HandleFunc("/api/consultation/session", h.StartSession).Methods("POST")
	router.HandleFunc("/api/consultation/message", h.SendMessage).Methods("POST")
	router.HandleFunc("/api/consultation/session/{session_id}/history", h.GetHistory).Methods("GET")
}

