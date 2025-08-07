package main

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

func fetchGitHubPRData(owner, repo string, prNumber int) (*PRData, error) {
	githubToken := os.Getenv("GITHUB_TOKEN")

	if githubToken == "" {
		// Fallback to mock data if no token is provided
		log.Println("Warning: GITHUB_TOKEN not provided, using mock data")
		return getMockPRData(owner, repo, prNumber), nil
	}

	// Create GitHub client
	client := github.NewClient(nil).WithAuthToken(githubToken)

	ctx := context.Background()

	// Fetch PR data
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		log.Printf("Error fetching PR data: %v", err)
		return getMockPRData(owner, repo, prNumber), nil
	}

	// Fetch additional data
	labels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, prNumber, nil)
	if err != nil {
		log.Printf("Error fetching labels: %v", err)
		labels = []*github.Label{}
	}

	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
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

// parseGitHubURL extracts owner, repo, and PR number from a GitHub PR URL
func parseGitHubURL(url string) (string, string, int, error) {
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
