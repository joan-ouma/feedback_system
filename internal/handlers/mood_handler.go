// +build ignore

package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/service"
)

type MoodHandler struct {
	moodService *service.MoodService
	authService *service.AuthService
	templates   *template.Template
}

func NewMoodHandler(moodService *service.MoodService, authService *service.AuthService, templates *template.Template) *MoodHandler {
	return &MoodHandler{
		moodService: moodService,
		authService: authService,
		templates:   templates,
	}
}

// RecordMood handles mood entry submission
func (h *MoodHandler) RecordMood(w http.ResponseWriter, r *http.Request) {
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

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	score, _ := strconv.Atoi(r.FormValue("score"))
	moodType := models.MoodType(r.FormValue("mood_type"))
	notes := r.FormValue("notes")

	if score < 1 || score > 10 {
		http.Error(w, "Score must be between 1 and 10", http.StatusBadRequest)
		return
	}

	entry, recommendation, err := h.moodService.RecordMood(r.Context(), user.ID, moodType, score, notes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"entry":         entry,
			"recommendation": recommendation,
		}
		if err := h.templates.ExecuteTemplate(w, "mood_success.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"entry":         entry,
		"recommendation": recommendation,
	})
}

// GetMoodHistory gets mood history
func (h *MoodHandler) GetMoodHistory(w http.ResponseWriter, r *http.Request) {
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

	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsedDays, err := strconv.Atoi(d); err == nil {
			days = parsedDays
		}
	}

	history, err := h.moodService.GetMoodHistory(r.Context(), user.ID, days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"history": history,
		}
		if err := h.templates.ExecuteTemplate(w, "mood_history.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// GetDailyQuote gets today's motivational quote
func (h *MoodHandler) GetDailyQuote(w http.ResponseWriter, r *http.Request) {
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

	quote, err := h.moodService.GetDailyQuote(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"quote": quote,
		}
		if err := h.templates.ExecuteTemplate(w, "daily_quote.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quote)
}

// MoodDashboard renders the mood tracking dashboard
func (h *MoodHandler) MoodDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "mood_dashboard.html", nil); err == nil {
			return
		}
	}
	http.ServeFile(w, r, "templates/mood_dashboard.html")
}

func (h *MoodHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/mood", h.MoodDashboard).Methods("GET")
	router.HandleFunc("/api/mood", h.RecordMood).Methods("POST")
	router.HandleFunc("/api/mood/history", h.GetMoodHistory).Methods("GET")
	router.HandleFunc("/api/mood/quote", h.GetDailyQuote).Methods("GET")
}

