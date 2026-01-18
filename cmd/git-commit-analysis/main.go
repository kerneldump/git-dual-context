package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"git-commit-analysis/internal/analyzer"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func main() {
	repoPath := flag.String("repo", ".", "Path to the git repository")
	errorMsg := flag.String("error", "", "The error message or bug description to analyze")
	numCommits := flag.Int("n", 5, "Number of commits to analyze")
	numWorkers := flag.Int("j", 3, "Number of concurrent workers")
	apiKey := flag.String("apikey", "", "Google Gemini API Key (optional, defaults to env var GEMINI_API_KEY)")
	flag.Parse()

	if *errorMsg == "" {
		log.Fatal("Please provide an error message using -error")
	}

	key := *apiKey
	if key == "" {
		key = os.Getenv("GEMINI_API_KEY")
	}
	if key == "" {
		log.Fatal("Error: No API key provided. Please use -apikey flag or set GEMINI_API_KEY environment variable.")
	}

	// Initialize Git
	var r *git.Repository
	var err error

	// Check if it's a remote URL
	if strings.HasPrefix(*repoPath, "http") || strings.HasPrefix(*repoPath, "git@") {
		// Create temp dir
		tempDir, err := os.MkdirTemp("", "git-analysis-*")
		if err != nil {
			log.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir) // Clean up

		fmt.Printf("Cloning %s into temporary directory...\n", *repoPath)
		r, err = git.PlainClone(tempDir, false, &git.CloneOptions{
			URL:      *repoPath,
			Progress: os.Stdout,
		})
		if err != nil {
			log.Fatalf("Failed to clone repo: %v", err)
		}
	} else {
		// Local repo
		r, err = git.PlainOpen(*repoPath)
		if err != nil {
			log.Fatalf("Failed to open git repo at %s: %v", *repoPath, err)
		}
	}

	headRef, err := r.Head()
	if err != nil {
		log.Fatalf("Failed to get HEAD: %v", err)
	}

	// Initialize Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("models/gemini-3-pro-preview")
	// model.ResponseMIMEType = "application/json" // Removed as it causes hangs with this model

	// Iterate Commits
	cIter, err := r.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		log.Fatalf("Failed to get commit log: %v", err)
	}

	fmt.Printf("Analyzing last %d commits for error: %q\n", *numCommits, *errorMsg)
	fmt.Println("---------------------------------------------------")

	// Collect commits first
	var commits []*object.Commit
	count := 0
	for {
		if count >= *numCommits {
			break
		}
		c, err := cIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating commits: %v", err)
		}

		// Skip merge commits
		if len(c.ParentHashes) > 1 {
			continue
		}

		commits = append(commits, c)
		count++
	}

	// Parallel Processing
	var wg sync.WaitGroup
	if *numWorkers < 1 {
		*numWorkers = 1
	}
	sem := make(chan struct{}, *numWorkers) // Limit to N concurrent requests
	var printMutex sync.Mutex

	for _, c := range commits {
		wg.Add(1)
		sem <- struct{}{}

		go func(c *object.Commit) {
			defer wg.Done()
			defer func() { <-sem }()

			// Create a context with timeout for each request
			reqCtx, cancel := context.WithTimeout(ctx, 300*time.Second)
			defer cancel()

			// Analyze
			res, err := analyzer.AnalyzeCommit(reqCtx, r, c, headRef.Hash(), *errorMsg, model)

			printMutex.Lock()
			defer printMutex.Unlock()

			if err != nil {
				log.Printf("Failed to analyze commit %s: %v", c.Hash.String(), err)
				return
			}
			if res.Skipped {
				fmt.Printf("Commit: %s | [Skipped - No relevant code changes]\n", c.Hash.String()[:8])
				fmt.Println("---------------------------------------------------")
				return
			}

			color := ""
			label := string(res.Probability)
			switch res.Probability {
			case analyzer.ProbHigh:
				color = "\033[31m" // Red
			case analyzer.ProbMedium:
				color = "\033[33m" // Yellow
			case analyzer.ProbLow:
				color = "\033[32m" // Green
			}

			fmt.Printf("Commit: %s | Prob: %s%s\033[0m\n", c.Hash.String()[:8], color, label)
			fmt.Printf("Reason: %s\n", res.Reasoning)
			fmt.Println("---------------------------------------------------")
		}(c)
	}

	wg.Wait()
}