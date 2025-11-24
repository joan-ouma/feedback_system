// +build ignore

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/service"
)

type ConsultationHandler struct {
	consultationService *service.ConsultationService
	authService         *service.AuthService
}

func NewConsultationHandler(consultationService *service.ConsultationService, authService *service.AuthService) *ConsultationHandler {
	return &ConsultationHandler{
		consultationService: consultationService,
		authService:         authService,
	}
}

// Chat renders the consultation chat page
func (h *ConsultationHandler) Chat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	http.ServeFile(w, r, "templates/consultation.html")
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
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var sessionID uuid.UUID
	if req.SessionID != "" {
		sessionID, err = uuid.Parse(req.SessionID)
		if err != nil {
			http.Error(w, "Invalid session ID", http.StatusBadRequest)
			return
		}
	}

	// Get or create session
	session, err := h.consultationService.GetOrCreateSession(r.Context(), user.ID, &sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send message
	consultation, err := h.consultationService.SendMessage(r.Context(), user.ID, session.ID, req.Message)
	if err != nil {
		// Check if it's an API key error
		if err.Error() == "LLM API key is not configured. Please set LLM_API_KEY in your environment variables" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": "LLM service is not configured. Please contact the administrator.",
				"consultation": map[string]interface{}{
					"response": "I'm sorry, but the consultation service is currently unavailable. Please contact your campus mental health services directly for support.",
				},
				"session_id": session.ID.String(),
			})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"consultation": consultation,
		"session_id":   session.ID.String(),
	})
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
	sessionID, err := uuid.Parse(vars["session_id"])
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

func (h *ConsultationHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/consultation", h.Chat).Methods("GET")
	router.HandleFunc("/api/consultation/session", h.StartSession).Methods("POST")
	router.HandleFunc("/api/consultation/message", h.SendMessage).Methods("POST")
	router.HandleFunc("/api/consultation/session/{session_id}/history", h.GetHistory).Methods("GET")
}

