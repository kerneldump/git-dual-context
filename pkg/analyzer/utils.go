package analyzer

import "strings"

// TruncateCommitMessage truncates a commit message to the first line
// and ensures it doesn't exceed maxLength characters.
// If truncation occurs, "..." is appended.
func TruncateCommitMessage(message string, maxLength int) string {
	// Get first line only
	firstLine := message
	if idx := strings.Index(message, "\n"); idx != -1 {
		firstLine = message[:idx]
	}

	// Truncate if too long
	if len(firstLine) > maxLength {
		if maxLength <= 3 {
			return firstLine[:maxLength]
		}
		return firstLine[:maxLength-3] + "..."
	}

	return firstLine
}
