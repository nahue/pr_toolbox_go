package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/go-github/v62/github"
	"github.com/joho/godotenv"
	"github.com/nahue/pr-toolbox-go/templates"
)

type Person struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Age        int    `json:"age"`
	Department string `json:"department"`
	Position   string `json:"position"`
}

type GeneratePRDescriptionRequest struct {
	PRUrl string `json:"prUrl"`
}

type GeneratePRDescriptionResponse struct {
	Description string `json:"description"`
	Error       string `json:"error,omitempty"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/", servePrDescriptions)
	r.Post("/api/generate-pr-description", generatePRDescription)

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func servePrDescriptions(w http.ResponseWriter, r *http.Request) {
	component := templates.PrDescriptions()
	component.Render(r.Context(), w)
}

func generatePRDescription(w http.ResponseWriter, r *http.Request) {
	var prUrl string

	// Handle POST requests - try to parse form data first, then JSON
	if err := r.ParseForm(); err == nil {
		// Try to get from form data
		prUrl = r.FormValue("prUrl")
	}

	// If no form data found, try JSON
	if prUrl == "" {
		var req GeneratePRDescriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err == nil {
			prUrl = req.PRUrl
		}
	}

	if prUrl == "" {
		http.Error(w, "PR URL is required", http.StatusBadRequest)
		return
	}

	// Parse GitHub URL to extract owner, repo, and PR number
	owner, repo, prNumber, err := parseGitHubURL(prUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid GitHub PR URL: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch GitHub PR data
	prData, err := fetchGitHubPRData(owner, repo, prNumber)
	if err != nil {
		log.Printf("Error fetching GitHub PR data: %v", err)
		http.Error(w, "Failed to fetch PR data", http.StatusInternalServerError)
		return
	}

	// Create OpenAI service
	openaiService, err := NewOpenAIService()
	if err != nil {
		log.Printf("Error creating OpenAI service: %v", err)
		http.Error(w, "Failed to initialize OpenAI service", http.StatusInternalServerError)
		return
	}

	// Generate description using OpenAI service
	description, err := openaiService.GeneratePRDescription(prData)
	if err != nil {
		log.Printf("Error generating description: %v", err)
		http.Error(w, "Failed to generate description", http.StatusInternalServerError)
		return
	}

	// Return HTML for Alpine AJAX
	w.Header().Set("Content-Type", "text/html")
	component := templates.PrDescriptionResult(description)
	component.Render(r.Context(), w)
}

// Helper functions for formatting GitHub data
func getLabelsString(labels []*github.Label) string {
	if len(labels) == 0 {
		return "none"
	}

	var labelNames []string
	for _, label := range labels {
		if label.Name != nil {
			labelNames = append(labelNames, *label.Name)
		}
	}
	return strings.Join(labelNames, ", ")
}

func getUserString(user *github.User) string {
	if user == nil || user.Login == nil {
		return "unknown"
	}
	return *user.Login
}

func getAssigneesString(assignees []*github.User) string {
	if len(assignees) == 0 {
		return "none"
	}

	var assigneeNames []string
	for _, assignee := range assignees {
		if assignee.Login != nil {
			assigneeNames = append(assigneeNames, *assignee.Login)
		}
	}
	return strings.Join(assigneeNames, ", ")
}
