package analyzer

import "testing"

func TestTruncateCommitMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		maxLength int
		expected  string
	}{
		{
			name:      "single line under limit",
			message:   "Fix bug",
			maxLength: 80,
			expected:  "Fix bug",
		},
		{
			name:      "single line at limit",
			message:   "Fix bug in authentication module that caused crashes",
			maxLength: 52,
			expected:  "Fix bug in authentication module that caused crashes",
		},
		{
			name:      "single line over limit",
			message:   "Fix bug in authentication module that caused crashes when users logged in",
			maxLength: 50,
			expected:  "Fix bug in authentication module that caused cr...",
		},
		{
			name:      "multiline message",
			message:   "Fix important bug\n\nThis is a detailed description\nwith multiple lines",
			maxLength: 80,
			expected:  "Fix important bug",
		},
		{
			name:      "multiline message over limit",
			message:   "This is an extremely long commit message that exceeds the limit\nDetailed description here",
			maxLength: 50,
			expected:  "This is an extremely long commit message that e...",
		},
		{
			name:      "very short limit",
			message:   "Fix bug",
			maxLength: 3,
			expected:  "Fix",
		},
		{
			name:      "very short limit with long message",
			message:   "Fix bug in authentication",
			maxLength: 5,
			expected:  "Fi...",
		},
		{
			name:      "empty message",
			message:   "",
			maxLength: 80,
			expected:  "",
		},
		{
			name:      "newline only",
			message:   "\n",
			maxLength: 80,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateCommitMessage(tt.message, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateCommitMessage(%q, %d) = %q, expected %q",
					tt.message, tt.maxLength, result, tt.expected)
			}
			if len(result) > tt.maxLength {
				t.Errorf("Result length %d exceeds maxLength %d", len(result), tt.maxLength)
			}
		})
	}
}
