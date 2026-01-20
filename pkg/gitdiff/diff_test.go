package gitdiff

import (
	"strings"
	"testing"
)

func TestShouldIgnoreFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Lock files
		{"go.sum", "go.sum", true},
		{"nested go.sum", "pkg/go.sum", true},
		{"package-lock.json", "package-lock.json", true},
		{"yarn.lock", "yarn.lock", true},
		{"poetry.lock", "poetry.lock", true},
		{"pnpm-lock.yaml", "pnpm-lock.yaml", true},
		{"Cargo.lock", "Cargo.lock", true},
		{"composer.lock", "composer.lock", true},

		// Test files
		{"Go test file", "handler_test.go", true},
		{"JS test file", "handler.test.js", true},
		{"TS test file", "handler.test.ts", true},
		{"JS spec file", "handler.spec.js", true},
		{"TS spec file", "handler.spec.ts", true},
		{"Python test file suffix", "handler_test.py", true},
		{"Python test file prefix", "test_handler.py", true},
		{"Ruby spec file", "handler_spec.rb", true},

		// Vendor and node_modules
		{"vendor file", "vendor/github.com/pkg/errors/errors.go", true},
		{"node_modules file", "node_modules/lodash/index.js", true},

		// Build directories
		{"dist file", "dist/bundle.js", true},
		{"build file", "build/output.js", true},
		{"out file", "out/main.js", true},

		// IDE directories
		{".idea file", ".idea/workspace.xml", true},
		{".vscode file", ".vscode/settings.json", true},

		// CI/CD files
		{"GitHub workflow", ".github/workflows/ci.yml", true},
		{"GitLab CI", ".gitlab-ci.yml", true},
		{"Travis CI", ".travis.yml", true},
		{"CircleCI", ".circleci/config.yml", true},

		// Python cache
		{"pycache file", "__pycache__/module.cpython-39.pyc", true},
		{"pytest cache", ".pytest_cache/v/cache/nodeids", true},

		// Should NOT ignore
		{"Go source", "main.go", false},
		{"Go source in pkg", "pkg/analyzer/engine.go", false},
		{"JS source", "src/index.js", false},
		{"TS source", "src/index.ts", false},
		{"Config file", "config.yaml", false},
		{"Python source", "main.py", false},

		// Edge cases
		{"empty path", "", false},
		{"path with test in name", "testing/utils.go", false},
		{"file named vendor", "vendor.go", false},
		{"file named test", "test.go", false}, // Not matching _test.go pattern
		{"testdata directory", "testdata/fixture.json", false},
		{"Windows path separator", "vendor\\github.com\\pkg\\errors.go", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldIgnoreFile(tt.path)
			if result != tt.expected {
				t.Errorf("ShouldIgnoreFile(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestTruncateDiff(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		maxSize         int
		shouldTruncate  bool
		containsMarker  bool
	}{
		{
			name:           "no truncation needed",
			input:          "short diff",
			maxSize:        100,
			shouldTruncate: false,
			containsMarker: false,
		},
		{
			name:           "exact size",
			input:          "exact",
			maxSize:        5,
			shouldTruncate: false,
			containsMarker: false,
		},
		{
			name:           "truncation needed",
			input:          "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n",
			maxSize:        60,
			shouldTruncate: true,
			containsMarker: true,
		},
		{
			name:           "truncation at line boundary",
			input:          strings.Repeat("a", 100) + "\n" + strings.Repeat("b", 100),
			maxSize:        150,
			shouldTruncate: true,
			containsMarker: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateDiff(tt.input, tt.maxSize)

			if tt.shouldTruncate {
				if len(result) > tt.maxSize+len(TruncationMarker) {
					t.Errorf("TruncateDiff result too long: got %d, max %d", len(result), tt.maxSize)
				}
			} else {
				if result != tt.input {
					t.Errorf("TruncateDiff should not modify input: got %q, want %q", result, tt.input)
				}
			}

			hasMarker := strings.Contains(result, TruncationMarker)
			if hasMarker != tt.containsMarker {
				t.Errorf("TruncateDiff marker presence: got %v, want %v", hasMarker, tt.containsMarker)
			}
		})
	}
}

func TestTruncateDiffPreservesLineBreaks(t *testing.T) {
	// Create a diff with clear line boundaries
	lines := []string{
		"--- a/file.go",
		"+func newFunc() {",
		"+    return nil",
		"+}",
		"-func oldFunc() {",
		"-    return err",
		"-}",
	}
	input := strings.Join(lines, "\n") + "\n"

	// Truncate to roughly half
	result := TruncateDiff(input, 50)

	// Should end with a newline (line boundary) before the marker
	// or the marker itself
	if !strings.HasSuffix(result, "\n") && !strings.HasSuffix(result, TruncationMarker) {
		t.Errorf("Truncated diff should end at line boundary, got: %q", result)
	}
}
