package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/nahue/pr-toolbox-go/internal/app"
	"github.com/nahue/pr-toolbox-go/internal/database"
	"github.com/nahue/pr-toolbox-go/internal/github"
	"github.com/nahue/pr-toolbox-go/internal/openai"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize database
	db, err := database.NewDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize OpenAI service
	openaiService, err := openai.NewService()
	if err != nil {
		log.Fatalf("Failed to initialize OpenAI service: %v", err)
	}

	// Initialize GitHub service
	githubService := github.NewService()

	// Create application with all dependencies
	application := app.NewApplication(db, openaiService, githubService)

	// Start server
	log.Fatal(application.Start("9090"))
}
