package workspace

import (
	"strings"
	"testing"
	"time"
)

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantUser string
		wantRepo string
		wantHost string
		wantErr  bool
	}{
		{
			name:     "SSH GitHub",
			url:      "git@github.com:tobi/try.git",
			wantUser: "tobi",
			wantRepo: "try",
			wantHost: "github.com",
		},
		{
			name:     "SSH GitHub without .git",
			url:      "git@github.com:tobi/try",
			wantUser: "tobi",
			wantRepo: "try",
			wantHost: "github.com",
		},
		{
			name:     "HTTPS GitHub",
			url:      "https://github.com/tobi/try.git",
			wantUser: "tobi",
			wantRepo: "try",
			wantHost: "github.com",
		},
		{
			name:     "HTTPS GitHub without .git",
			url:      "https://github.com/tobi/try",
			wantUser: "tobi",
			wantRepo: "try",
			wantHost: "github.com",
		},
		{
			name:     "SSH GitLab",
			url:      "git@gitlab.com:user/project.git",
			wantUser: "user",
			wantRepo: "project",
			wantHost: "gitlab.com",
		},
		{
			name:     "HTTPS GitLab",
			url:      "https://gitlab.com/user/project.git",
			wantUser: "user",
			wantRepo: "project",
			wantHost: "gitlab.com",
		},
		{
			name:     "SSH custom host",
			url:      "git@git.company.com:team/repo.git",
			wantUser: "team",
			wantRepo: "repo",
			wantHost: "git.company.com",
		},
		{
			name:    "invalid URL",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "file path",
			url:     "/path/to/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseGitURL(tt.url)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.User != tt.wantUser {
				t.Errorf("user: got %s, want %s", parsed.User, tt.wantUser)
			}
			if parsed.Repo != tt.wantRepo {
				t.Errorf("repo: got %s, want %s", parsed.Repo, tt.wantRepo)
			}
			if parsed.Host != tt.wantHost {
				t.Errorf("host: got %s, want %s", parsed.Host, tt.wantHost)
			}
		})
	}
}

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"git@github.com:user/repo.git", true},
		{"https://github.com/user/repo.git", true},
		{"http://github.com/user/repo", true},
		{"git@gitlab.com:user/repo", true},
		{"something.git", true},
		{"github.com/user/repo", true},
		{"gitlab.com/user/repo", true},
		{"my-project", false},
		{"test", false},
		{"/path/to/dir", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsGitURL(tt.input)
			if got != tt.want {
				t.Errorf("IsGitURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCloneDirName(t *testing.T) {
	tests := []struct {
		url      string
		contains string
	}{
		{"git@github.com:tobi/try.git", "tobi-try"},
		{"https://github.com/user/project.git", "user-project"},
	}

	datePrefix := time.Now().Format("2006-01-02")

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			name, err := CloneDirName(tt.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasPrefix(name, datePrefix) {
				t.Errorf("expected date prefix %s, got %s", datePrefix, name)
			}
			if !strings.Contains(name, tt.contains) {
				t.Errorf("expected to contain %s, got %s", tt.contains, name)
			}
		})
	}
}
