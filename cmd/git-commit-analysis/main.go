package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/kerneldump/git-dual-context/pkg/analyzer"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// commitResult holds the analysis result for ordered streaming output
type commitResult struct {
	index  int
	result *analyzer.AnalysisResult
	err    error
	commit *object.Commit
}

// orderedPrinter handles streaming results in commit order
type orderedPrinter struct {
	encoder     *json.Encoder
	mu          sync.Mutex
	results     map[int]*commitResult // buffered results waiting to print
	nextToPrint int                   // next index we're waiting to print
	total       int                   // total number of commits

	// Summary counters
	high    int
	medium  int
	low     int
	skipped int
	errors  int
}

func newOrderedPrinter(encoder *json.Encoder, total int) *orderedPrinter {
	return &orderedPrinter{
		encoder:     encoder,
		results:     make(map[int]*commitResult),
		nextToPrint: 0,
		total:       total,
	}
}

// submit adds a result and prints any results that are ready (in order)
func (p *orderedPrinter) submit(r *commitResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Store the result
	p.results[r.index] = r

	// Print all consecutive results starting from nextToPrint
	for {
		result, ok := p.results[p.nextToPrint]
		if !ok {
			break // Next result not ready yet
		}

		p.printResult(result)
		delete(p.results, p.nextToPrint)
		p.nextToPrint++
	}
}

// printResult outputs a single result and updates counters
func (p *orderedPrinter) printResult(r *commitResult) {
	if r.err != nil {
		p.encoder.Encode(analyzer.NewLogEntry("ERROR", fmt.Sprintf("Failed to analyze commit %s: %v", r.commit.Hash.String(), r.err)))
		p.errors++
		return
	}
	if r.result == nil {
		p.errors++
		return
	}
	if r.result.Skipped {
		p.encoder.Encode(analyzer.NewLogEntry("INFO", fmt.Sprintf("Commit: %s | [Skipped - No relevant code changes]", r.commit.Hash.String()[:8])))
		p.skipped++
		return
	}

	// Count by probability
	switch r.result.Probability {
	case analyzer.ProbHigh:
		p.high++
	case analyzer.ProbMedium:
		p.medium++
	case analyzer.ProbLow:
		p.low++
	}

	// Encode and print as JSON with commit message
	jr := r.result.ToJSONResult(r.commit.Hash.String()[:8], r.commit.Message)
	p.encoder.Encode(jr)
}

// summary returns the final summary
func (p *orderedPrinter) summary() analyzer.Summary {
	p.mu.Lock()
	defer p.mu.Unlock()

	return analyzer.Summary{
		Type:    "summary",
		Total:   p.total,
		High:    p.high,
		Medium:  p.medium,
		Low:     p.low,
		Skipped: p.skipped,
		Errors:  p.errors,
	}
}

// Global temp directory for cleanup on fatal exit
var tempDir string

func main() {
	// Set up signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Parse flags first
	repoPath := flag.String("repo", ".", "Path to the git repository or remote URL")
	branch := flag.String("branch", "", "Branch to analyze (default: current HEAD)")
	errorMsg := flag.String("error", "", "The error message or bug description to analyze")
	numCommits := flag.Int("n", 5, "Number of commits to analyze")
	numWorkers := flag.Int("j", 3, "Number of concurrent workers")
	modelName := flag.String("model", "models/gemini-3-pro-preview", "Gemini model to use")
	timeout := flag.Duration("timeout", 5*time.Minute, "Timeout per commit analysis")
	outputFile := flag.String("o", "", "Output file path (default: stdout)")
	apiKey := flag.String("apikey", "", "Google Gemini API Key (prefer GEMINI_API_KEY env var)")
	verbose := flag.Bool("v", false, "Verbose output (show additional debug info)")
	flag.Parse()

	// Set up output writer
	var output io.Writer = os.Stdout
	if *outputFile != "" {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"type":"log","level":"ERROR","msg":"Failed to create output file: %s"}`, err.Error())
			os.Exit(1)
		}
		defer f.Close()
		output = f
	}

	encoder := json.NewEncoder(output)
	var logMutex sync.Mutex

	logJSON := func(level, msg string) {
		logMutex.Lock()
		defer logMutex.Unlock()
		encoder.Encode(analyzer.NewLogEntry(level, msg))
	}

	fatalJSON := func(msg string) {
		logJSON("ERROR", msg)
		// Clean up temp directory on fatal exit
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
		os.Exit(1)
	}

	if *errorMsg == "" {
		fatalJSON("Please provide an error message using -error")
	}

	key := *apiKey
	if key != "" {
		logJSON("WARN", "API key passed via command line may be visible in process list. Consider using GEMINI_API_KEY environment variable instead.")
	} else {
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
		tempDir, err = os.MkdirTemp("", "git-analysis-*")
		if err != nil {
			fatalJSON(err.Error())
		}
		defer os.RemoveAll(tempDir) // Clean up on normal exit

		logJSON("INFO", "Cloning "+*repoPath+" into temporary directory...")
		r, err = git.PlainClone(tempDir, false, &git.CloneOptions{
			URL: *repoPath,
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

	// Get HEAD reference (or specified branch)
	var headRef *plumbing.Reference
	if *branch != "" {
		refName := plumbing.NewBranchReferenceName(*branch)
		headRef, err = r.Reference(refName, true)
		if err != nil {
			fatalJSON(fmt.Sprintf("Failed to find branch %s: %v", *branch, err))
		}
		logJSON("INFO", fmt.Sprintf("Analyzing branch: %s", *branch))
	} else {
		headRef, err = r.Head()
		if err != nil {
			fatalJSON("Failed to get HEAD: " + err.Error())
		}
	}

	// Get HEAD commit once for all goroutines (performance optimization)
	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		fatalJSON("Failed to get HEAD commit: " + err.Error())
	}

	// Initialize Gemini
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		fatalJSON("Failed to create Gemini client: " + err.Error())
	}
	defer client.Close()

	model := client.GenerativeModel(*modelName)
	model.SetTemperature(0.1)

	if *verbose {
		logJSON("DEBUG", fmt.Sprintf("Using model: %s, timeout: %v", *modelName, *timeout))
	}

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

	// Parallel Processing with ordered streaming output
	printer := newOrderedPrinter(encoder, len(commits))
	var wg sync.WaitGroup
	if *numWorkers < 1 {
		*numWorkers = 1
	}
	sem := make(chan struct{}, *numWorkers) // Limit to N concurrent requests

	for i, c := range commits {
		wg.Add(1)
		sem <- struct{}{}

		go func(idx int, commit *object.Commit) {
			defer wg.Done()
			defer func() { <-sem }()

			// Check for cancellation before starting
			select {
			case <-ctx.Done():
				printer.submit(&commitResult{index: idx, err: ctx.Err(), commit: commit})
				return
			default:
			}

			// Create a context with timeout for each request
			reqCtx, cancel := context.WithTimeout(ctx, *timeout)
			defer cancel()

			if *verbose {
				logJSON("DEBUG", fmt.Sprintf("Starting analysis of commit %s", commit.Hash.String()[:8]))
			}

			// Use retry logic for transient failures
			var res *analyzer.AnalysisResult
			err := analyzer.WithRetry(reqCtx, analyzer.DefaultRetryConfig(), func() error {
				var analyzeErr error
				res, analyzeErr = analyzer.AnalyzeCommit(reqCtx, r, commit, headCommit, *errorMsg, model)
				return analyzeErr
			})

			// Submit result for ordered streaming output
			printer.submit(&commitResult{index: idx, result: res, err: err, commit: commit})
		}(i, c)
	}

	// Wait for completion or cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Normal completion
	case <-ctx.Done():
		logJSON("WARN", "Received interrupt signal, shutting down...")
		// Wait briefly for goroutines to finish
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			logJSON("WARN", "Timeout waiting for goroutines, forcing exit")
		}
	}

	// Output summary
	encoder.Encode(printer.summary())
}
