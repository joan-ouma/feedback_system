package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/service"
)

type QuizHandler struct {
	quizService *service.QuizService
	authService *service.AuthService
	templates   *template.Template
}

func NewQuizHandler(quizService *service.QuizService, authService *service.AuthService, templates *template.Template) *QuizHandler {
	return &QuizHandler{
		quizService: quizService,
		authService: authService,
		templates:   templates,
	}
}

// GetQuiz gets a quiz by type
func (h *QuizHandler) GetQuiz(w http.ResponseWriter, r *http.Request) {
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
	quizType := models.QuizType(vars["type"])

	quiz, err := h.quizService.GetQuiz(r.Context(), quizType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if quiz == nil {
		http.Error(w, "Quiz not found", http.StatusNotFound)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"quiz": quiz,
		}
		if err := h.templates.ExecuteTemplate(w, "quiz.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quiz)
}

// SubmitQuiz handles quiz submission
func (h *QuizHandler) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
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
		QuizID  string                 `json:"quiz_id"`
		Answers map[string]interface{} `json:"answers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	quizID, err := uuid.Parse(req.QuizID)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	response, recommendation, err := h.quizService.SubmitQuiz(r.Context(), user.ID, quizID, req.Answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html")
		data := map[string]interface{}{
			"response":      response,
			"recommendation": recommendation,
		}
		if err := h.templates.ExecuteTemplate(w, "quiz_results.html", data); err != nil {
			http.Error(w, "Template error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"response":      response,
		"recommendation": recommendation,
	})
}

// QuizList renders the quiz selection page
func (h *QuizHandler) QuizList(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "templates/quiz_list.html")
}

func (h *QuizHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/quizzes", h.QuizList).Methods("GET")
	router.HandleFunc("/api/quiz/{type}", h.GetQuiz).Methods("GET")
	router.HandleFunc("/api/quiz/submit", h.SubmitQuiz).Methods("POST")
}

