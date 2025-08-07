# People Management System

A Go REST API built with the Chi library featuring a modern web interface using Alpine.js for dynamic interactions.

## Features

- **REST API**: Built with Chi router for clean, fast routing
- **Health Check**: `/health` endpoint for monitoring system status
- **People API**: `/api/people` endpoint serving mocked people data
- **PR Descriptions**: `/pr_descriptions` page for generating GitHub pull request descriptions using OpenAI
- **Modern UI**: Beautiful, responsive interface with Alpine.js
- **Interactive Features**: 
  - Load people data dynamically
  - Health status checking
  - Statistics dashboard
  - Responsive table display
  - AI-powered PR description generation

## API Endpoints

### Health Check
```
GET /health
```
Returns the system health status and timestamp.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2025-08-06T19:58:16.706922-03:00"
}
```

### People API
```
GET /api/people
```
Returns a list of people with their details.

**Response:**
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john.doe@example.com",
    "age": 32,
    "department": "Engineering",
    "position": "Senior Developer"
  }
]
```

### Generate PR Description
```
POST /api/generate-pr-description
```
Generates a professional GitHub pull request description using OpenAI and GitHub API data.

**Request:**
```json
{
  "prUrl": "https://github.com/owner/repo/pull/123"
}
```

**Response:**
```json
{
  "description": "## Description\n\nThis pull request implements..."
}
```

**Features:**
- Fetches real PR data from GitHub API (if GITHUB_TOKEN provided)
- Falls back to mock data if GitHub token not available
- Uses PR title, description, labels, assignees, and file changes
- Generates context-aware descriptions based on actual PR content
- Alpine AJAX integration for seamless frontend-backend communication
- Supports both GET (Alpine AJAX) and POST (regular API) requests

## Pages

### Home Page
```
GET /
```
Serves the main application interface with people management features.

### PR Descriptions
```
GET /pr_descriptions
```
Serves a form for generating GitHub pull request descriptions.

## Frontend Features

The web interface includes:

- **Modern Design**: Gradient backgrounds, card layouts, and smooth animations
- **Interactive Controls**: Buttons to load data and check health status
- **Statistics Dashboard**: Shows total people, average age, and department count
- **Responsive Table**: Displays people data in a clean, sortable format
- **Real-time Updates**: Uses Alpine.js for reactive UI updates
- **Error Handling**: Graceful error display and loading states

## Getting Started

### Prerequisites
- Go 1.24.4 or higher
- OpenAI API key (for PR description generation)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd pr-toolbox-go
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables (optional, for PR description feature):
   
   **Option A: Using .env file (recommended)**
   ```bash
   cp .env.example .env
   # Edit .env file and add your API keys:
   # - OPENAI_API_KEY for PR description generation
   # - GITHUB_TOKEN for fetching real PR data (optional)
   ```
   
   **Option B: Using environment variables**
   ```bash
   export OPENAI_API_KEY="your-openai-api-key-here"
   export GITHUB_TOKEN="your-github-token-here"  # optional
   ```

4. Run the application:

   **Option A: Using go run**
   ```bash
   go run .
   ```

   **Option B: Using Makefile (recommended)**
   ```bash
   make run
   ```

   **Option C: Development mode (with templ generation)**
   ```bash
   make dev
   ```

   **Option D: Hot reload development with Air**
   ```bash
   make air
   ```
   This will automatically reload the application when you modify any Go or templ files.

5. Open your browser and navigate to:
```
http://localhost:8080
```

## Usage

1. **Load People**: Click the "Load People" button to fetch and display the people data
2. **Check Health**: Click the "Check Health" button to verify the API is running properly
3. **View Statistics**: Once people are loaded, view statistics including total count, average age, and department count
4. **Browse Data**: Scroll through the table to view all people details
5. **PR Descriptions**: Navigate to the PR Descriptions page to generate GitHub pull request descriptions

## Project Structure

```
pr-toolbox-go/
├── main.go              # Main application file with server setup
├── github_service.go     # GitHub API integration service
├── go.mod               # Go module dependencies
├── Makefile             # Build and run commands
├── .env.example         # Example environment variables file
├── templates/           # Templ templates
│   ├── layout.templ      # Base layout template
│   ├── index.templ       # Main application page
│   └── pr_descriptions.templ  # PR descriptions template
└── README.md            # This documentation
```

## Technologies Used

- **Backend**: Go with Chi router
- **Frontend**: Alpine.js for reactive UI
- **Templating**: Templ for type-safe HTML templates
- **AI Integration**: OpenAI GPT-4 for PR description generation
- **GitHub Integration**: Go GitHub API client for fetching PR data
- **Configuration**: Godotenv for environment variable management
- **HTTP Client**: Axios for API calls
- **Styling**: Tailwind CSS with modern design patterns
- **Middleware**: CORS support, logging, and error recovery

## Development

The application is structured for easy extension:

- Add new API endpoints in the main router
- Extend the Person struct for additional fields
- Create new templ templates for additional pages using the base layout
- Add new Alpine.js functions for additional frontend functionality

## Environment Variables

The application supports loading environment variables from a `.env` file. Copy `.env.example` to `.env` and configure the following variables:

### Required Variables
- `OPENAI_API_KEY`: Your OpenAI API key for PR description generation

### Optional Variables
- `GITHUB_TOKEN`: Your GitHub API token for fetching PR data (falls back to mock data if not provided)
- `PORT`: Server port (default: 8080)
- `LOG_LEVEL`: Logging level (default: info)

## Running the Application

The application has multiple Go files, so you need to run it using:

```bash
go run .
```

This will compile and run all Go files in the project.

## API Testing

Test the API endpoints using curl:

```bash
# Health check
curl http://localhost:8080/health

# Get people data
curl http://localhost:8080/api/people

# Generate PR description (requires OpenAI API key)
curl -X POST http://localhost:8080/api/generate-pr-description \
  -H "Content-Type: application/json" \
  -d '{"prUrl": "https://github.com/owner/repo/pull/123"}'
```

## License

This project is open source and available under the MIT License. 