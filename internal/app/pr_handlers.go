package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/nahue/pr-toolbox-go/templates"
)

// servePrDescriptions handles GET /
func (app *Application) servePrDescriptions(w http.ResponseWriter, r *http.Request) {
	component := templates.PrDescriptions()
	component.Render(r.Context(), w)
}

// generatePRDescription handles POST /api/generate-pr-description
func (app *Application) generatePRDescription(w http.ResponseWriter, r *http.Request) {
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
	owner, repo, prNumber, err := app.githubService.ParseGitHubURL(prUrl)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid GitHub PR URL: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch GitHub PR data
	prData, err := app.githubService.FetchPRData(owner, repo, prNumber)
	if err != nil {
		log.Printf("Error fetching GitHub PR data: %v", err)
		http.Error(w, "Failed to fetch PR data", http.StatusInternalServerError)
		return
	}

	// Generate description using OpenAI service
	description, err := app.openaiService.GeneratePRDescription(prData)
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
