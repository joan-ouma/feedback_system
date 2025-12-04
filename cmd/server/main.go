package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"html/template"

	"github.com/gorilla/mux"
	"github.com/joan/feedback-sys/internal/config"
	"github.com/joan/feedback-sys/internal/database"
	"github.com/joan/feedback-sys/internal/handlers"
	"github.com/joan/feedback-sys/internal/llm"
	"github.com/joan/feedback-sys/internal/middleware"
	"github.com/joan/feedback-sys/internal/repository"
	"github.com/joan/feedback-sys/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config:", err)
	}

	ctx := context.Background()
	db, err := database.NewDB(ctx, cfg.Database.URI)
	if err != nil {
		log.Fatal("DB:", err)
	}
	defer db.Close()

	repository.SeedQuizzes(ctx, db)

	// Repositories
	userRepo := repository.NewUserRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	consultationRepo := repository.NewConsultationRepository(db)
	moodRepo := repository.NewMoodRepository(db)
	quizRepo := repository.NewQuizRepository(db)
	quoteRepo := repository.NewQuoteRepository(db)

	// Services
	llmClient := llm.NewClient(cfg.LLM)
	authService := service.NewAuthService(userRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	consultationService := service.NewConsultationService(consultationRepo, llmClient)
	moodService := service.NewMoodService(moodRepo, quoteRepo, llmClient)
	quizService := service.NewQuizService(quizRepo, llmClient)

	// ✅ FIXED: Define templateDir BEFORE using it
	templateDir := "templates"
	staticDir := "static"

	// Templates (non-blocking)
	tmpl := template.New("").Funcs(template.FuncMap{"replace": strings.ReplaceAll})
	templates, _ := tmpl.ParseGlob("templates/*.html")

	// Handlers
	authHandler, err := handlers.NewAuthHandler(authService, cfg.Server.SessionSecret, templateDir)
	if err != nil {
		log.Fatal("AuthHandler:", err)
	}
	feedbackHandler, err := handlers.NewFeedbackHandler(feedbackService, authService, quizService, templateDir)
	if err != nil {
		log.Fatal("FeedbackHandler:", err)
	}

	// ✅ LINE 81 FIXED: 3 ARGUMENTS REQUIRED
	consultationHandler := handlers.NewConsultationHandler(consultationService, authService, templateDir)
	moodHandler := handlers.NewMoodHandler(moodService, authService, templates)
	quizHandler := handlers.NewQuizHandler(quizService, authService, templates)

	// Router
	router := mux.NewRouter()
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	authHandler.RegisterRoutes(router)

	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.NewAuthMiddleware(cfg.Server.SessionSecret).RequireAuth)
	feedbackHandler.RegisterRoutes(protected)
	consultationHandler.RegisterRoutes(protected)
	moodHandler.RegisterRoutes(protected)
	quizHandler.RegisterRoutes(protected)

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
	}).Methods("GET")

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Server on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShut, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctxShut)
}
