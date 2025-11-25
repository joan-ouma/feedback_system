package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/service"
)

type FeedbackHandler struct {
	feedbackService *service.FeedbackService
	authService     *service.AuthService
	templates       *template.Template
}

type FeedbackViewData struct {
	ID               string
	TypeLabel        string
	Status           string
	Title            string
	Content          string
	CreatedAtFormatted string
}

type FeedbacksViewData struct {
	Feedbacks []FeedbackViewData
}

func NewFeedbackHandler(feedbackService *service.FeedbackService, authService *service.AuthService, templateDir string) (*FeedbackHandler, error) {
	tmpl := template.New("").Funcs(template.FuncMap{
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"le": func(a, b int) bool {
			return a <= b
		},
		"gt": func(a, b int) bool {
			return a > b
		},
		"len": func(s string) int {
			return len(s)
		},
	})
	
	// Load template files
	pattern := filepath.Join(templateDir, "*.html")
	templates, err := tmpl.ParseGlob(pattern)
	if err != nil {
		return nil, err
	}

	return &FeedbackHandler{
		feedbackService: feedbackService,
		authService:     authService,
		templates:       templates,
	}, nil
}

// SubmitFeedback handles feedback submission
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
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

	// Parse form data (HTMX sends form-encoded data)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	req := struct {
		Type    string
		Title   string
		Content string
	}{
		Type:    r.FormValue("type"),
		Title:   r.FormValue("title"),
		Content: r.FormValue("content"),
	}

	feedbackType := models.FeedbackType(req.Type)
	if feedbackType == "" {
		feedbackType = models.FeedbackTypeGeneral
	}

	feedback, err := h.feedbackService.SubmitFeedback(r.Context(), user.ID, feedbackType, req.Title, req.Content)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if HTMX request
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("HX-Trigger", "feedbackSubmitted")
		// Render success message
		if err := h.templates.ExecuteTemplate(w, "feedback_success.html", nil); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
			return
		}
		return
	}

	// Fallback to JSON for API requests
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedback)
}

// GetFeedbacks retrieves user's feedbacks
func (h *FeedbackHandler) GetFeedbacks(w http.ResponseWriter, r *http.Request) {
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

	feedbacks, err := h.feedbackService.GetUserFeedbacks(r.Context(), user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if this is an HTMX request - return HTML instead of JSON
	if r.Header.Get("HX-Request") != "" {
		w.Header().Set("Content-Type", "text/html")
		h.renderFeedbacksHTML(w, feedbacks)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedbacks)
}

// renderFeedbacksHTML renders feedbacks as HTML using templates
func (h *FeedbackHandler) renderFeedbacksHTML(w http.ResponseWriter, feedbacks []*models.Feedback) {
	viewData := FeedbacksViewData{
		Feedbacks: make([]FeedbackViewData, 0, len(feedbacks)),
	}

	for _, feedback := range feedbacks {
		typeLabel := string(feedback.Type)
		switch feedback.Type {
		case models.FeedbackTypeMentalHealth:
			typeLabel = "Mental Health"
		case models.FeedbackTypeCampus:
			typeLabel = "Campus Issue"
		case models.FeedbackTypeGeneral:
			typeLabel = "General"
		case models.FeedbackTypeOther:
			typeLabel = "Other"
		}

		viewData.Feedbacks = append(viewData.Feedbacks, FeedbackViewData{
			ID:                feedback.ID.String(),
			TypeLabel:         typeLabel,
			Status:            feedback.Status,
			Title:             feedback.Title,
			Content:           feedback.Content,
			CreatedAtFormatted: feedback.CreatedAt.Format("January 2, 2006 at 3:04 PM"),
		})
	}

	if err := h.templates.ExecuteTemplate(w, "feedback_partial.html", viewData); err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetFeedback retrieves a specific feedback
func (h *FeedbackHandler) GetFeedback(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	feedbackID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		http.Error(w, "Invalid feedback ID", http.StatusBadRequest)
		return
	}

	feedback, err := h.feedbackService.GetFeedback(r.Context(), feedbackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if feedback == nil {
		http.Error(w, "Feedback not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feedback)
}

// Dashboard renders the feedback dashboard
func (h *FeedbackHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "text/html")
	
	// Try template first, fallback to file serve
	if h.templates != nil {
		if err := h.templates.ExecuteTemplate(w, "dashboard.html", nil); err == nil {
			return
		}
	}
	
	// Fallback to file serve
	http.ServeFile(w, r, "templates/dashboard.html")
}

func (h *FeedbackHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
	router.HandleFunc("/api/feedback", h.SubmitFeedback).Methods("POST")
	router.HandleFunc("/api/feedback", h.GetFeedbacks).Methods("GET")
	router.HandleFunc("/api/feedback/{id}", h.GetFeedback).Methods("GET")
}

