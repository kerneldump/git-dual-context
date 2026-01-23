package validator

import (
	"testing"
)

func TestValidateNumCommits(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid small", 5, false},
		{"valid medium", 50, false},
		{"valid large", 1000, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"exceeds max", 1001, true},
		{"way too large", 999999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNumCommits(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNumCommits(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateNumWorkers(t *testing.T) {
	tests := []struct {
		name    string
		input   int
		wantErr bool
	}{
		{"valid small", 1, false},
		{"valid medium", 10, false},
		{"valid large", 50, false},
		{"zero", 0, true},
		{"negative", -5, true},
		{"exceeds max", 51, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNumWorkers(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNumWorkers(%d) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty (valid - means HEAD)", "", false},
		{"simple branch", "main", false},
		{"feature branch", "feature/auth", false},
		{"with hyphens", "fix-bug-123", false},
		{"with underscores", "dev_branch", false},
		{"numeric", "v1.2.3", false},
		{"directory traversal", "../etc/passwd", true},
		{"starts with dash", "-main", true},
		{"starts with slash", "/main", true},
		{"ends with slash", "main/", true},
		{"special chars", "branch@name", true},
		{"spaces", "my branch", true},
		{"null byte", "main\x00evil", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBranchName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateBranchName(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateRepoPath(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty", "", true},
		{"current dir", ".", false},
		{"relative path", "./myrepo", false},
		{"absolute path", "/home/user/repo", false},
		{"http url", "https://github.com/user/repo.git", false},
		{"https url", "https://github.com/user/repo.git", false},
		{"git ssh", "git@github.com:user/repo.git", false},
		{"directory traversal", "../../../etc/passwd", true},
		{"etc path", "/etc/config", true},
		{"sys path", "/sys/devices", true},
		{"proc path", "/proc/cpuinfo", true},
		{"dev path", "/dev/null", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepoPath(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepoPath(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateErrorMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid message", "panic: index out of bounds", false},
		{"short message", "error", false},
		{"empty string", "", true},
		{"only spaces", "   ", true},
		{"only tabs", "\t\t", true},
		{"only newlines", "\n\n", true},
		{"mixed whitespace", " \t\n ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateErrorMessage(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateErrorMessage(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateBranchNameRegex(t *testing.T) {
	// Additional regex-specific tests
	validNames := []string{
		"main",
		"develop",
		"feature/authentication",
		"bugfix/issue-123",
		"release/v1.0.0",
		"hotfix/prod-crash",
		"user/john/experimental",
		"1234",
		"v2.0",
	}

	for _, name := range validNames {
		t.Run("valid:"+name, func(t *testing.T) {
			if !branchNameRegex.MatchString(name) {
				t.Errorf("branchNameRegex should match %q but didn't", name)
			}
		})
	}

	invalidNames := []string{
		"branch name with spaces",
		"branch@with#special",
		"branch:with:colons",
		"branch;with;semicolons",
		"branch&ampersand",
		"branch|pipe",
		"branch<>brackets",
	}

	for _, name := range invalidNames {
		t.Run("invalid:"+name, func(t *testing.T) {
			if branchNameRegex.MatchString(name) {
				t.Errorf("branchNameRegex should not match %q but did", name)
			}
		})
	}
}
