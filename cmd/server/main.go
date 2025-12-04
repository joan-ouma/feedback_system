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
	// 1. Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config:", err)
	}

	// 2. Database
	ctx := context.Background()
	db, err := database.NewDB(ctx, cfg.Database.URI)
	if err != nil {
		log.Fatal("DB:", err)
	}
	defer db.Close()

	// 3. Seed data
	repository.SeedQuizzes(ctx, db)

	// 4. Repositories
	userRepo := repository.NewUserRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	consultationRepo := repository.NewConsultationRepository(db)
	moodRepo := repository.NewMoodRepository(db)
	quizRepo := repository.NewQuizRepository(db)
	quoteRepo := repository.NewQuoteRepository(db)

	// 5. Services
	llmClient := llm.NewClient(cfg.LLM)
	authService := service.NewAuthService(userRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	consultationService := service.NewConsultationService(consultationRepo, llmClient)
	moodService := service.NewMoodService(moodRepo, quoteRepo, llmClient)
	quizService := service.NewQuizService(quizRepo, llmClient)

	// 6. Paths - SIMPLE
	templateDir := "templates"
	staticDir := "static"

	// 7. Load templates (non-fatal)
	tmpl := template.New("").Funcs(template.FuncMap{
		"replace": strings.ReplaceAll,
	})
	templates, _ := tmpl.ParseGlob("templates/*.html")

	// 8. Handlers - PERFECT MATCH FOR YOUR SIGNATURE
	authHandler, err := handlers.NewAuthHandler(authService, cfg.Server.SessionSecret, templateDir)
	if err != nil {
		log.Fatal("AuthHandler:", err)
	}

	feedbackHandler, err := handlers.NewFeedbackHandler(feedbackService, authService, quizService, templateDir)
	if err != nil {
		log.Fatal("FeedbackHandler:", err)
	}

	// ✅ EXACTLY 3 ARGS - matches your line 41 signature
	consultationHandler := handlers.NewConsultationHandler(consultationService, authService)
	moodHandler := handlers.NewMoodHandler(moodService, authService, templates)
	quizHandler := handlers.NewQuizHandler(quizService, authService, templates)

	// 9. Router
	router := mux.NewRouter()

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Public routes
	authHandler.RegisterRoutes(router)

	// Protected routes
	protected := router.PathPrefix("").Subrouter()
	protected.Use(middleware.NewAuthMiddleware(cfg.Server.SessionSecret).RequireAuth)
	feedbackHandler.RegisterRoutes(protected)
	consultationHandler.RegisterRoutes(protected)
	moodHandler.RegisterRoutes(protected)
	quizHandler.RegisterRoutes(protected)

	// Root redirect
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
	}).Methods("GET")

	// 10. Server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("🚀 Starting server on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctxShut, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShut); err != nil {
		log.Fatal("Shutdown:", err)
	}
	log.Println("✅ Server stopped")
}
