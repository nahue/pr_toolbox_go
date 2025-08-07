package openai

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/nahue/pr-toolbox-go/internal/github"
	"github.com/sashabaranov/go-openai"
)

type Service struct {
	client *openai.Client
}

func NewService() (*Service, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key not configured")
	}

	client := openai.NewClient(apiKey)
	return &Service{client: client}, nil
}

func (s *Service) GeneratePRDescription(prData *github.PRData) (string, error) {
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
		github.GetLabelsString(prData.Labels),
		github.GetUserString(prData.User),
		github.GetAssigneesString(prData.Assignees),
	)

	// Make OpenAI API call
	resp, err := s.client.CreateChatCompletion(
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
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	// Extract the generated description
	return resp.Choices[0].Message.Content, nil
}
