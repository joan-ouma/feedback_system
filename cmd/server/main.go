package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
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

	// Seed quizzes if needed
	if err := repository.SeedQuizzes(ctx, db); err != nil {
		log.Printf("Warning: Failed to seed quizzes: %v", err)
	} else {
		log.Println("Quizzes seeded successfully")
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	feedbackRepo := repository.NewFeedbackRepository(db)
	consultationRepo := repository.NewConsultationRepository(db)
	moodRepo := repository.NewMoodRepository(db)
	quizRepo := repository.NewQuizRepository(db)
	quoteRepo := repository.NewQuoteRepository(db)

	// Initialize LLM client
	llmClient := llm.NewClient(cfg.LLM)

	// Initialize services
	authService := service.NewAuthService(userRepo)
	feedbackService := service.NewFeedbackService(feedbackRepo)
	consultationService := service.NewConsultationService(consultationRepo, llmClient)
	moodService := service.NewMoodService(moodRepo, quoteRepo, llmClient)
	quizService := service.NewQuizService(quizRepo, llmClient)

	// Get the project root directory
	// Try multiple methods to find the project root
	var templateDir, staticDir string

	// Method 1: Check environment variable (useful for Docker/CI)
	if envTemplateDir := os.Getenv("TEMPLATE_DIR"); envTemplateDir != "" {
		templateDir = envTemplateDir
		staticDir = os.Getenv("STATIC_DIR")
		if staticDir == "" {
			staticDir = filepath.Join(filepath.Dir(envTemplateDir), "static")
		}
		log.Printf("Using template directory from environment: %s", templateDir)
	} else {
		// Method 2: Try to find relative to current working directory
		cwd, err := os.Getwd()
		if err == nil {
			// Check if we're in project root
			if _, err := os.Stat(filepath.Join(cwd, "templates")); err == nil {
				templateDir = filepath.Join(cwd, "templates")
				staticDir = filepath.Join(cwd, "static")
				log.Printf("Found templates relative to working directory: %s", cwd)
			} else {
				// Check if we're in cmd/server/
				if _, err := os.Stat(filepath.Join(cwd, "..", "..", "templates")); err == nil {
					templateDir = filepath.Join(cwd, "..", "..", "templates")
					staticDir = filepath.Join(cwd, "..", "..", "static")
					log.Printf("Found templates two levels up from working directory")
				} else {
					// Method 3: Use runtime.Caller to find the source file location
					_, execPath, _, ok := runtime.Caller(0)
					if ok {
						projectRoot := filepath.Join(filepath.Dir(execPath), "..", "..")
						templateDir = filepath.Join(projectRoot, "templates")
						staticDir = filepath.Join(projectRoot, "static")
						log.Printf("Using runtime.Caller to find templates: %s", projectRoot)
					} else {
						// Fallback: use relative paths
						templateDir = "templates"
						staticDir = "static"
						log.Printf("Using fallback relative paths: templates/ and static/")
					}
				}
			}
		} else {
			// Fallback: use relative paths
			templateDir = "templates"
			staticDir = "static"
			log.Printf("Could not get working directory, using fallback relative paths")
		}
	}

	// Convert to absolute paths for reliability
	if absTemplateDir, err := filepath.Abs(templateDir); err == nil {
		templateDir = absTemplateDir
	}
	if absStaticDir, err := filepath.Abs(staticDir); err == nil {
		staticDir = absStaticDir
	}

	log.Printf("Template directory: %s", templateDir)
	log.Printf("Static directory: %s", staticDir)

	// Verify directories exist
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		log.Fatalf("Template directory does not exist: %s", templateDir)
	}
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		log.Printf("Warning: Static directory does not exist: %s", staticDir)
	}

	// Load templates with helper functions
	tmpl := template.New("").Funcs(template.FuncMap{
		"replace": func(s, old, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
	})
	templatePattern := filepath.Join(templateDir, "*.html")
	templates, err := tmpl.ParseGlob(templatePattern)
	if err != nil {
		log.Fatalf("Failed to load templates from %s: %v", templatePattern, err)
	}
	log.Printf("Successfully loaded templates from %s", templateDir)

	// Initialize handlers
	authHandler, err := handlers.NewAuthHandler(authService, cfg.Server.SessionSecret, templateDir)
	if err != nil {
		log.Fatalf("Failed to initialize auth handler: %v", err)
	}
	feedbackHandler, err := handlers.NewFeedbackHandler(feedbackService, authService, quizService, templateDir)
	if err != nil {
		log.Fatalf("Failed to initialize feedback handler: %v", err)
	}
	consultationHandler := handlers.NewConsultationHandler(consultationService, authService, templateDir)
	moodHandler := handlers.NewMoodHandler(moodService, authService, templates)
	quizHandler := handlers.NewQuizHandler(quizService, authService, templates)

	// Setup router
	router := mux.NewRouter()

	// Static files
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Public routes
	authHandler.RegisterRoutes(router)

	// Protected routes
	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(middleware.NewAuthMiddleware(cfg.Server.SessionSecret).RequireAuth)
	feedbackHandler.RegisterRoutes(protectedRouter)
	consultationHandler.RegisterRoutes(protectedRouter)
	moodHandler.RegisterRoutes(protectedRouter)
	quizHandler.RegisterRoutes(protectedRouter)

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
			userPass := uri[userPassIdx+3 : idx]
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
