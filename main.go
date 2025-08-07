package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/go-github/v62/github"
	"github.com/joho/godotenv"
	"github.com/nahue/pr-toolbox-go/templates"
	"github.com/sashabaranov/go-openai"
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

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		http.Error(w, "OpenAI API key not configured", http.StatusInternalServerError)
		return
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	// Create detailed prompt with GitHub data
	prompt := fmt.Sprintf(`You are a helpful assistant that generates professional GitHub pull request descriptions.

Given the following GitHub pull request data:

Repository: %s
PR Number: %d
Title: %s
Current Description: %s
State: %s
Created: %s
Updated: %s
Additions: %d lines
Deletions: %d lines
Changed Files: %d files
Labels: %s
Author: %s
Assignees: %s

Please structure the description with the following sections:
1. **Summary** - A brief, high-level overview of the purpose of this pull request.
2. **Changes Made** - A clear and itemized list of the specific modifications made in this PR.
3. **Motivation/Context:** Explain *why* these changes were necessary (e.g., bug fix, new feature, refactoring, performance improvement).
4. **How to Test (Optional but Recommended):** Provide instructions for how a reviewer can verify the changes.
5. **Potential Impacts/Considerations:** Mention any known side effects, performance implications, or areas that require particular attention during review.
6. **Relevant Links (Optional):** Include links to related issues, design documents, or external resources.
7. **Contributors** - List of contributors to the PR with their contribution counts

Ensure the description is easy to read, uses clear language, and is formatted for readability (e.g., bullet points, headings).
Make the description clear, professional, and helpful for code reviewers. Focus on the "why" and "what" of the changes. Include the contributors section to acknowledge all team members who contributed to this PR.`,
		prData.Repository,
		prData.PRNumber,
		prData.Title,
		prData.Body,
		prData.State,
		prData.CreatedAt.Format("2006-01-02 15:04:05"),
		prData.UpdatedAt.Format("2006-01-02 15:04:05"),
		prData.Additions,
		prData.Deletions,
		len(prData.ChangedFiles),
		getLabelsString(prData.Labels),
		getUserString(prData.User),
		getAssigneesString(prData.Assignees),
	)

	// Make OpenAI API call
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an expert software developer and technical writer. Please create comprehensive, professional pull request descriptions based on GitHub PR data. Focus on clarity, technical accuracy, and helpfulness for reviewers.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   1000,
			Temperature: 0.7,
		},
	)

	if err != nil {
		log.Printf("OpenAI API error: %v", err)
		http.Error(w, "Failed to generate description", http.StatusInternalServerError)
		return
	}

	if len(resp.Choices) == 0 {
		http.Error(w, "No response from OpenAI", http.StatusInternalServerError)
		return
	}

	// Extract the generated description
	description := resp.Choices[0].Message.Content

	// Return HTML for Alpine AJAX
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<div id="pr-result">
			<div class="bg-green-50 border border-green-200 rounded-lg p-6">
				<h3 class="text-lg font-semibold text-green-800 mb-4">Generated Description</h3>
				<div class="bg-white border border-green-200 rounded-lg p-4">
					<pre class="whitespace-pre-wrap text-sm text-gray-800">%s</pre>
				</div>
				<div class="mt-4 flex gap-2">
					<button
						onclick="navigator.clipboard.writeText('%s')"
						class="px-4 py-2 bg-green-500 text-white rounded-lg text-sm font-medium hover:bg-green-600 transition-colors"
					>
						Copy to Clipboard
					</button>
				</div>
			</div>
		</div>
	`, description, strings.ReplaceAll(description, "'", "\\'"))
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
