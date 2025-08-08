package app

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/nahue/pr-toolbox-go/internal/database"
	githubsvc "github.com/nahue/pr-toolbox-go/internal/github"
	"github.com/nahue/pr-toolbox-go/internal/openai"
)

// Application holds all the services and dependencies
type Application struct {
	db            *database.Database
	openaiService *openai.Service
	githubService *githubsvc.Service
	router        *chi.Mux
	useAuth       bool
}

type GeneratePRDescriptionRequest struct {
	PRUrl string `json:"prUrl"`
}

// NewApplication creates a new application instance with all dependencies
func NewApplication(db *database.Database, openaiService *openai.Service, githubService *githubsvc.Service) *Application {
	// Check if authentication is enabled via environment variable
	useAuth := true // default to true for security
	if useAuthStr := os.Getenv("USE_AUTH"); useAuthStr != "" {
		if parsed, err := strconv.ParseBool(useAuthStr); err == nil {
			useAuth = parsed
		} else {
			log.Printf("Warning: Invalid USE_AUTH value '%s', defaulting to true", useAuthStr)
		}
	}

	if !useAuth {
		log.Println("Warning: Authentication is DISABLED. This should only be used for development.")
	}

	app := &Application{
		db:            db,
		openaiService: openaiService,
		githubService: githubService,
		router:        chi.NewRouter(),
		useAuth:       useAuth,
	}

	app.setupMiddleware()
	app.setupRoutes()

	return app
}

// setupMiddleware configures all middleware for the application
func (app *Application) setupMiddleware() {
	app.router.Use(middleware.Logger)
	app.router.Use(middleware.Recoverer)
	app.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
}

// setupRoutes configures all routes for the application
func (app *Application) setupRoutes() {
	// Health check routes (public)
	app.router.Get("/health", app.handleHealth)
	app.router.Get("/ready", app.handleReadiness)
	app.router.Get("/live", app.handleLiveness)

	// Authentication routes (public)
	app.router.Get("/auth/login", app.handleLogin)
	app.router.Post("/auth/magic-link", app.handleMagicLinkRequest)
	app.router.Get("/auth/verify", app.handleMagicLinkVerification)
	app.router.Post("/auth/logout", app.handleLogout)
	app.router.Get("/auth/me", app.handleCurrentUser)

	// Apply auth middleware to protected routes
	app.router.Group(func(r chi.Router) {
		r.Use(app.authMiddleware)

		// Protected routes
		r.Get("/", app.servePrDescriptions)
		r.Post("/api/generate-pr-description", app.generatePRDescription)
	})
}

// Start starts the HTTP server on the specified port
func (app *Application) Start(port string) error {
	log.Printf("Server starting on port %s", port)
	return http.ListenAndServe(":"+port, app.router)
}

// HTTP Handlers moved to separate files:
// - auth_handlers.go for authentication routes
// - pr_handlers.go for PR description routes
// - health_handlers.go for health check routes
