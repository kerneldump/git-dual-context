// Package gitdiff provides utilities for extracting and filtering git diffs.
//
// It includes logic for generating standard commit diffs and evolutionary diffs
// between a commit and the current HEAD. It also features smart filtering to
// exclude irrelevant files like lockfiles, tests, and documentation, ensuring
// that only functional code changes are passed to the reasoning engine.
package gitdiff
