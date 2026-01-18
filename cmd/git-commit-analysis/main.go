package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	// Redirect all logs to stdout via our encoder
	encoder := json.NewEncoder(os.Stdout)
	var printMutex sync.Mutex

	logJSON := func(level, msg string) {
		printMutex.Lock()
		defer printMutex.Unlock()
		encoder.Encode(analyzer.NewLogEntry(level, msg))
	}

	fatalJSON := func(msg string) {
		logJSON("ERROR", msg)
		os.Exit(1)
	}

	repoPath := flag.String("repo", ".", "Path to the git repository")
	errorMsg := flag.String("error", "", "The error message or bug description to analyze")
	numCommits := flag.Int("n", 5, "Number of commits to analyze")
	numWorkers := flag.Int("j", 3, "Number of concurrent workers")
	apiKey := flag.String("apikey", "", "Google Gemini API Key (optional, defaults to env var GEMINI_API_KEY)")
	flag.Parse()

	if *errorMsg == "" {
		fatalJSON("Please provide an error message using -error")
	}

	key := *apiKey
	if key == "" {
		key = os.Getenv("GEMINI_API_KEY")
	}
	if key == "" {
		fatalJSON("Error: No API key provided. Please use -apikey flag or set GEMINI_API_KEY environment variable.")
	}

	// Initialize Git
	var r *git.Repository
	var err error

	// Check if it's a remote URL
	if strings.HasPrefix(*repoPath, "http") || strings.HasPrefix(*repoPath, "git@") {
		// Create temp dir
		tempDir, err := os.MkdirTemp("", "git-analysis-*")
		if err != nil {
			fatalJSON(err.Error())
		}
		defer os.RemoveAll(tempDir) // Clean up

		logJSON("INFO", "Cloning "+*repoPath+" into temporary directory...")
		r, err = git.PlainClone(tempDir, false, &git.CloneOptions{
			URL: *repoPath,
			// Progress: os.Stderr, // Removing progress as it's hard to make JSON
		})
		if err != nil {
			fatalJSON("Failed to clone repo: " + err.Error())
		}
	} else {
		// Local repo
		r, err = git.PlainOpen(*repoPath)
		if err != nil {
			fatalJSON("Failed to open git repo at " + *repoPath + ": " + err.Error())
		}
	}

	headRef, err := r.Head()
	if err != nil {
		fatalJSON("Failed to get HEAD: " + err.Error())
	}

	// Initialize Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		fatalJSON("Failed to create Gemini client: " + err.Error())
	}
	defer client.Close()

	model := client.GenerativeModel("models/gemini-3-pro-preview")
	model.SetTemperature(0.1)

	// Iterate Commits
	cIter, err := r.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		fatalJSON("Failed to get commit log: " + err.Error())
	}

	logJSON("INFO", fmt.Sprintf("Analyzing last %d commits for error: %q", *numCommits, *errorMsg))

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
			fatalJSON("Error iterating commits: " + err.Error())
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
				encoder.Encode(analyzer.NewLogEntry("ERROR", fmt.Sprintf("Failed to analyze commit %s: %v", c.Hash.String(), err)))
				return
			}
			if res.Skipped {
				encoder.Encode(analyzer.NewLogEntry("INFO", fmt.Sprintf("Commit: %s | [Skipped - No relevant code changes]", c.Hash.String()[:8])))
				return
			}

			// Encode and print as JSON
			jr := res.ToJSONResult(c.Hash.String()[:8])
			encoder.Encode(jr)
		}(c)
	}

	wg.Wait()
}
