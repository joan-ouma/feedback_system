package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/models"
	"github.com/joan/feedback-sys/internal/service"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		// Convert ObjectID to hex string for template
		quizData := map[string]interface{}{
			"quiz": map[string]interface{}{
				"ID":          quiz.ID.Hex(),
				"Type":        string(quiz.Type),
				"Title":       quiz.Title,
				"Description": quiz.Description,
				"Questions":   quiz.Questions,
				"CreatedAt":   quiz.CreatedAt,
			},
		}
		if err := h.templates.ExecuteTemplate(w, "quiz.html", quizData); err != nil {
			log.Printf("❌ Template error: %v", err)
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
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

	// Support both JSON and form data
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.QuizID = r.FormValue("quiz_id")
		// Parse answers from form
		req.Answers = make(map[string]interface{})
		for key, values := range r.Form {
			if key != "quiz_id" {
				if len(values) > 0 {
					req.Answers[key] = values[0]
				}
			}
		}
	}

	quizID, err := primitive.ObjectIDFromHex(req.QuizID)
	if err != nil {
		http.Error(w, "Invalid quiz ID", http.StatusBadRequest)
		return
	}

	response, recommendation, err := h.quizService.SubmitQuiz(r.Context(), user.ID, quizID, req.Answers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Always return JSON (frontend handles rendering)
	w.Header().Set("Content-Type", "application/json")
	
	result := map[string]interface{}{
		"response": response,
	}
	
	if recommendation != nil {
		result["recommendation"] = recommendation
		log.Printf("✅ Quiz recommendation: %s", recommendation.Recommendations)
	} else {
		log.Printf("⚠️  No recommendation generated for quiz response")
	}
	
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Printf("❌ Error encoding quiz response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// QuizList renders the quiz selection page with user's quiz history
func (h *QuizHandler) QuizList(w http.ResponseWriter, r *http.Request) {
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

	// Get user's quiz history
	history, err := h.quizService.GetUserQuizHistory(r.Context(), user.ID)
	if err != nil {
		log.Printf("⚠️  Error fetching quiz history: %v", err)
		history = make(map[string]*service.QuizHistoryItem)
	}

	w.Header().Set("Content-Type", "text/html")
	if h.templates != nil {
		data := map[string]interface{}{
			"QuizHistory": history,
		}
		if err := h.templates.ExecuteTemplate(w, "quiz_list.html", data); err == nil {
			return
		}
		log.Printf("❌ Template error: %v", err)
	}
	http.ServeFile(w, r, "templates/quiz_list.html")
}

func (h *QuizHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/quizzes", h.QuizList).Methods("GET")
	router.HandleFunc("/api/quiz/{type}", h.GetQuiz).Methods("GET")
	router.HandleFunc("/api/quiz/submit", h.SubmitQuiz).Methods("POST")
}

