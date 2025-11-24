package main

import (
	"context"
	"fmt"
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
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize OpenTelemetry
	if cfg.OpenTelemetry.Enabled {
		if err := initTracing(cfg.OpenTelemetry.JaegerEndpoint); err != nil {
			log.Printf("Failed to initialize tracing: %v", err)
		} else {
			log.Println("OpenTelemetry tracing initialized")
		}
	}

	// Initialize database
	ctx := context.Background()
	log.Printf("Connecting to MongoDB: %s", maskConnectionString(cfg.Database.URI))
	db, err := database.NewDB(ctx, cfg.Database.URI)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("MongoDB connection established")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	// Temporarily disabled - MongoDB migration in progress
	// consultationRepo := repository.NewConsultationRepository(db)
	// moodRepo := repository.NewMoodRepository(db)
	// quizRepo := repository.NewQuizRepository(db)
	// quoteRepo := repository.NewQuoteRepository(db)

	// Initialize LLM client (temporarily unused - MongoDB migration in progress)
	_ = llm.NewClient(cfg.LLM)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	// Temporarily disabled - MongoDB migration in progress
	// consultationService := service.NewConsultationService(consultationRepo, llmClient)
	// moodService := service.NewMoodService(moodRepo, quoteRepo, llmClient)
	// quizService := service.NewQuizService(quizRepo, llmClient)

	// Load templates (temporarily unused - MongoDB migration in progress)
	tmpl := template.New("")
	templatePattern := "templates/*.html"
	_, err = tmpl.ParseGlob(templatePattern)
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	// Initialize handlers
	authHandler, err := handlers.NewAuthHandler(authService, cfg.Server.SessionSecret, "templates")
	if err != nil {
		log.Fatalf("Failed to initialize auth handler: %v", err)
	}
	feedbackHandler, err := handlers.NewFeedbackHandler(feedbackService, authService, "templates")
	if err != nil {
		log.Fatalf("Failed to initialize feedback handler: %v", err)
	}
	// Temporarily disabled - MongoDB migration in progress
	// consultationHandler := handlers.NewConsultationHandler(consultationService, authService)
	// moodHandler := handlers.NewMoodHandler(moodService, authService, templates)
	// quizHandler := handlers.NewQuizHandler(quizService, authService, templates)

	// Setup router
	router := mux.NewRouter()

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Public routes
	authHandler.RegisterRoutes(router)

	// Protected routes
	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.NewAuthMiddleware(cfg.Server.SessionSecret).RequireAuth)
	feedbackHandler.RegisterRoutes(protectedRouter)
	
	// Placeholder routes for features being migrated to MongoDB
	protectedRouter.HandleFunc("/mood", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Mood Tracking - Coming Soon</title>
	<link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
	<nav class="navbar">
		<div class="container">
			<a href="/dashboard" class="logo">Campus Support</a>
			<div class="nav-links">
				<a href="/dashboard">Dashboard</a>
			</div>
		</div>
	</nav>
	<main class="container">
		<div class="card">
			<h1>Mood Tracking</h1>
			<p>This feature is currently being migrated to MongoDB and will be available soon!</p>
			<a href="/dashboard" class="btn-primary">Back to Dashboard</a>
		</div>
	</main>
</body>
</html>`))
	}).Methods("GET")
	
	protectedRouter.HandleFunc("/quizzes", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>Quizzes - Coming Soon</title>
	<link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
	<nav class="navbar">
		<div class="container">
			<a href="/dashboard" class="logo">Campus Support</a>
			<div class="nav-links">
				<a href="/dashboard">Dashboard</a>
			</div>
		</div>
	</nav>
	<main class="container">
		<div class="card">
			<h1>Mental Health Quizzes</h1>
			<p>This feature is currently being migrated to MongoDB and will be available soon!</p>
			<a href="/dashboard" class="btn-primary">Back to Dashboard</a>
		</div>
	</main>
</body>
</html>`))
	}).Methods("GET")
	
	protectedRouter.HandleFunc("/consultation", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
	<title>AI Consultation - Coming Soon</title>
	<link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
	<nav class="navbar">
		<div class="container">
			<a href="/dashboard" class="logo">Campus Support</a>
			<div class="nav-links">
				<a href="/dashboard">Dashboard</a>
			</div>
		</div>
	</nav>
	<main class="container">
		<div class="card">
			<h1>AI Consultation</h1>
			<p>This feature is currently being migrated to MongoDB and will be available soon!</p>
			<a href="/dashboard" class="btn-primary">Back to Dashboard</a>
		</div>
	</main>
</body>
</html>`))
	}).Methods("GET")

	// Root redirect
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/signup", http.StatusSeeOther)
	}).Methods("GET")

	// Wrap router with OpenTelemetry middleware
	var handler http.Handler = router
	if cfg.OpenTelemetry.Enabled {
		handler = otelhttp.NewHandler(router, "feedback-sys")
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// maskConnectionString masks password in connection string for logging
func maskConnectionString(uri string) string {
	// Simple masking - replace password with ***
	// Format: mongodb+srv://user:password@host
	if idx := strings.Index(uri, "@"); idx > 0 {
		if userPassIdx := strings.Index(uri, "://"); userPassIdx > 0 {
			prefix := uri[:userPassIdx+3]
			userPass := uri[userPassIdx+3:idx]
			if colonIdx := strings.Index(userPass, ":"); colonIdx > 0 {
				user := userPass[:colonIdx]
				return prefix + user + ":***@" + uri[idx+1:]
			}
		}
	}
	return uri
}

// initTracing initializes OpenTelemetry tracing with Jaeger exporter
func initTracing(jaegerEndpoint string) error {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)))
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("feedback-sys"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(res),
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

