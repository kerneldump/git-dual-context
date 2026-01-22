package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/kerneldump/git-dual-context/cmd/mcp-server/internal/tools"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Redirect logs to stderr so they don't interfere with MCP JSON-RPC on stdout
	log.SetOutput(os.Stderr)

	// Create MCP server
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "git-dual-context-mcp",
		Version: "0.1.0",
	}, nil)

	// Register the analyze_root_cause tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "analyze_root_cause",
		Description: "Diagnose bugs using dual-context diff analysis. Analyzes recent commits in a git repository to identify which commit most likely caused a given error or bug. Uses LLM-powered reasoning to compare immediate changes (micro-context) with evolutionary changes to HEAD (macro-context).",
	}, handleAnalyzeRootCause)

	log.Println("Starting Git Dual-Context MCP Server...")

	// Run server over stdio transport
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// handleAnalyzeRootCause is the MCP tool handler for analyze_root_cause
func handleAnalyzeRootCause(
	ctx context.Context,
	request *mcp.CallToolRequest,
	input tools.AnalyzeInput,
) (*mcp.CallToolResult, tools.AnalyzeOutput, error) {
	log.Printf("Analyzing repository: %s for error: %q", input.RepoPath, input.ErrorMessage)

	output, err := tools.AnalyzeRootCause(ctx, input, func(msg string) {
		// Send progress logs to the client
		_ = request.Session.Log(ctx, &mcp.LoggingMessageParams{
			Level: "info",
			Data:  msg,
		})
	})
	if err != nil {
		log.Printf("Analysis failed: %v", err)
		return nil, tools.AnalyzeOutput{}, err
	}

	log.Printf("Analysis complete: %d commits analyzed, %d high, %d medium, %d low probability, %d errors",
		output.Summary.Total, output.Summary.High, output.Summary.Medium, output.Summary.Low, output.Summary.Errors)

	// Build a human-readable text summary for the Content field
	summaryText := tools.FormatResultsAsText(output)

	// Marshal structured output for debugging
	jsonBytes, _ := json.MarshalIndent(output, "", "  ")
	log.Printf("Structured output: %s", string(jsonBytes))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: summaryText,
			},
		},
	}, *output, nil
}
