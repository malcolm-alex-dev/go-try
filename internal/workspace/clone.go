package workspace

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ParsedURL contains extracted info from a git URL.
type ParsedURL struct {
	User string
	Repo string
	Host string
}

// ParseGitURL extracts user and repo from various git URL formats.
// Supports:
//   - git@github.com:user/repo.git (SSH)
//   - https://github.com/user/repo.git (HTTPS)
//   - git@host.com:user/repo.git (SSH other hosts)
//   - https://host.com/user/repo.git (HTTPS other hosts)
func ParseGitURL(url string) (*ParsedURL, error) {
	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// SSH format: git@host:user/repo
	sshPattern := regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+)$`)
	if matches := sshPattern.FindStringSubmatch(url); matches != nil {
		return &ParsedURL{
			Host: matches[1],
			User: matches[2],
			Repo: matches[3],
		}, nil
	}

	// HTTPS format: https://host/user/repo
	httpsPattern := regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/]+)$`)
	if matches := httpsPattern.FindStringSubmatch(url); matches != nil {
		return &ParsedURL{
			Host: matches[1],
			User: matches[2],
			Repo: matches[3],
		}, nil
	}

	return nil, fmt.Errorf("unable to parse git URL: %s", url)
}

// IsGitURL returns true if the string looks like a git URL.
func IsGitURL(s string) bool {
	if strings.HasPrefix(s, "git@") {
		return true
	}
	if strings.HasPrefix(s, "https://") || strings.HasPrefix(s, "http://") {
		return true
	}
	if strings.HasSuffix(s, ".git") {
		return true
	}
	if strings.Contains(s, "github.com") || strings.Contains(s, "gitlab.com") {
		return true
	}
	return false
}

// CloneDirName generates a directory name for a cloned repo.
// Format: YYYY-MM-DD-user-repo
func CloneDirName(url string) (string, error) {
	parsed, err := ParseGitURL(url)
	if err != nil {
		return "", err
	}

	datePrefix := time.Now().Format("2006-01-02")
	return fmt.Sprintf("%s-%s-%s", datePrefix, parsed.User, parsed.Repo), nil
}

// Clone clones a git repository into basePath.
// Returns the full path to the cloned directory.
func Clone(basePath, url string) (string, error) {
	dirName, err := CloneDirName(url)
	if err != nil {
		return "", err
	}

	// Ensure unique name
	dirName = uniqueName(basePath, dirName)
	fullPath := basePath + "/" + dirName

	// Run git clone
	cmd := exec.Command("git", "clone", url, fullPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("git clone failed: %s\n%s", err, output)
	}

	return fullPath, nil
}

// CloneScript returns the shell commands to clone a repo (for exec mode).
// This is used when we want the shell to perform the clone.
func CloneScript(basePath, url string) (string, string, error) {
	dirName, err := CloneDirName(url)
	if err != nil {
		return "", "", err
	}

	dirName = uniqueName(basePath, dirName)
	fullPath := basePath + "/" + dirName

	return fullPath, url, nil
}
