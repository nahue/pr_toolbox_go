package github

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v62/github"
)

type Service struct {
	client *github.Client
}

type PRData struct {
	Title        string                `json:"title"`
	Body         string                `json:"body"`
	User         *github.User          `json:"user"`
	Assignees    []*github.User        `json:"assignees"`
	Labels       []*github.Label       `json:"labels"`
	State        string                `json:"state"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	ChangedFiles []*github.CommitFile  `json:"changed_files"`
	Additions    int                   `json:"additions"`
	Deletions    int                   `json:"deletions"`
	Repository   string                `json:"repository"`
	PRNumber     int                   `json:"pr_number"`
	Contributors []*github.Contributor `json:"contributors"`
}

func NewService() *Service {
	githubToken := os.Getenv("GITHUB_TOKEN")

	var client *github.Client
	if githubToken != "" {
		client = github.NewClient(nil).WithAuthToken(githubToken)
	} else {
		client = github.NewClient(nil)
	}

	return &Service{client: client}
}

func (s *Service) FetchPRData(owner, repo string, prNumber int) (*PRData, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	if githubToken == "" {
		// Fallback to mock data if no token is provided
		log.Println("Warning: GITHUB_TOKEN not provided, using mock data")
		return getMockPRData(owner, repo, prNumber), nil
	}

	ctx := context.Background()

	// Fetch PR data
	pr, _, err := s.client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		log.Printf("Error fetching PR data: %v", err)
		return getMockPRData(owner, repo, prNumber), nil
	}

	// Fetch additional data
	labels, _, err := s.client.Issues.ListLabelsByIssue(ctx, owner, repo, prNumber, nil)
	if err != nil {
		log.Printf("Error fetching labels: %v", err)
		labels = []*github.Label{}
	}

	files, _, err := s.client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	if err != nil {
		log.Printf("Error fetching files: %v", err)
		files = []*github.CommitFile{}
	}

	// Note: GitHub API doesn't have a direct endpoint for PR contributors
	// We'll use the PR author and assignees as contributors
	contributors := []*github.Contributor{}
	if pr.User != nil {
		contributors = append(contributors, &github.Contributor{
			Login:     pr.User.Login,
			ID:        pr.User.ID,
			AvatarURL: pr.User.AvatarURL,
		})
	}

	return &PRData{
		Title:        pr.GetTitle(),
		Body:         pr.GetBody(),
		User:         pr.User,
		Assignees:    pr.Assignees,
		Labels:       labels,
		State:        pr.GetState(),
		CreatedAt:    pr.GetCreatedAt().Time,
		UpdatedAt:    pr.GetUpdatedAt().Time,
		ChangedFiles: files,
		Additions:    pr.GetAdditions(),
		Deletions:    pr.GetDeletions(),
		Repository:   fmt.Sprintf("%s/%s", owner, repo),
		PRNumber:     prNumber,
		Contributors: contributors,
	}, nil
}

func getMockPRData(owner, repo string, prNumber int) *PRData {
	return &PRData{
		Title: "Sample Pull Request",
		Body:  "This is a sample pull request description for testing purposes.",
		User: &github.User{
			Login:     github.String("sample-user"),
			ID:        github.Int64(12345),
			AvatarURL: github.String("https://github.com/images/error/octocat_happy.gif"),
		},
		Assignees: []*github.User{
			{
				Login:     github.String("reviewer1"),
				ID:        github.Int64(67890),
				AvatarURL: github.String("https://github.com/images/error/octocat_happy.gif"),
			},
		},
		Labels: []*github.Label{
			{
				Name:  github.String("enhancement"),
				Color: github.String("a2eeef"),
			},
			{
				Name:  github.String("documentation"),
				Color: github.String("0075ca"),
			},
		},
		State:     "open",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
		ChangedFiles: []*github.CommitFile{
			{
				Filename:  github.String("main.go"),
				Additions: github.Int(10),
				Deletions: github.Int(2),
				Changes:   github.Int(12),
			},
			{
				Filename:  github.String("README.md"),
				Additions: github.Int(5),
				Deletions: github.Int(0),
				Changes:   github.Int(5),
			},
		},
		Additions:    15,
		Deletions:    2,
		Repository:   fmt.Sprintf("%s/%s", owner, repo),
		PRNumber:     prNumber,
		Contributors: []*github.Contributor{},
	}
}

// ParseGitHubURL extracts owner, repo, and PR number from a GitHub PR URL
func (s *Service) ParseGitHubURL(url string) (string, string, int, error) {
	// Expected format: https://github.com/owner/repo/pull/123
	parts := strings.Split(url, "/")
	if len(parts) < 7 || parts[2] != "github.com" || parts[5] != "pull" {
		return "", "", 0, fmt.Errorf("invalid GitHub PR URL format")
	}

	owner := parts[3]
	repo := parts[4]
	prNumberStr := parts[6]

	// Remove any query parameters or fragments
	prNumberStr = strings.Split(prNumberStr, "?")[0]
	prNumberStr = strings.Split(prNumberStr, "#")[0]

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid PR number: %s", prNumberStr)
	}

	return owner, repo, prNumber, nil
}

// Helper functions for formatting GitHub data
func GetLabelsString(labels []*github.Label) string {
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

func GetUserString(user *github.User) string {
	if user == nil || user.Login == nil {
		return "unknown"
	}
	return *user.Login
}

func GetAssigneesString(assignees []*github.User) string {
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
