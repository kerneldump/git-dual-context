// Package analyzer provides the core reasoning engine for Git Dual-Context Analysis.
//
// It uses LLMs (specifically Google Gemini) to analyze git commits by comparing
// a standard diff (micro-context) with a full evolutionary diff to HEAD (macro-context).
// This dual-context approach helps identify both immediate bugs and "sleeper" bugs
// that only manifest as the codebase evolves.
package analyzer
