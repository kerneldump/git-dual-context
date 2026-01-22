package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/kerneldump/git-dual-context/pkg/analyzer"

	"github.com/go-git/go-git/v5"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	// 1. Initialize Gemini Client
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("models/gemini-1.5-pro")
	model.SetTemperature(0.1)

	// 2. Open Git Repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatalf("Failed to open git repo: %v", err)
	}

	headRef, err := repo.Head()
	if err != nil {
		log.Fatalf("Failed to get HEAD: %v", err)
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		log.Fatalf("Failed to get HEAD commit: %v", err)
	}

	// 3. Analyze a Commit (in this example, we analyze HEAD)
	errorMsg := "The system is returning a 500 error on the /login endpoint"

	result, err := analyzer.AnalyzeCommit(ctx, repo, headCommit, headCommit, errorMsg, model)
	if err != nil {
		log.Fatalf("Analysis failed: %v", err)
	}

	if result.Skipped {
		fmt.Println("Commit skipped (no relevant code changes)")
		return
	}

	// 4. Print Results
	fmt.Printf("Probability: %s\n", result.Probability)
	fmt.Printf("Reasoning: %s\n", result.Reasoning)
}
