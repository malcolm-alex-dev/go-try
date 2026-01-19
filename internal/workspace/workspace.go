// Package workspace handles directory operations for try workspaces.
package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Entry represents a directory in the tries folder.
type Entry struct {
	Name      string    // Directory name (basename)
	Path      string    // Full path
	ModTime   time.Time // Last modification time
	BaseScore float64   // Pre-computed score based on recency
}

// DefaultPath returns the default tries directory path.
func DefaultPath() string {
	if p := os.Getenv("TRY_PATH"); p != "" {
		return expandPath(p)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "src", "tries")
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

// EnsureDir creates the directory if it doesn't exist.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// Scan reads all directories in basePath and returns them sorted by recency.
func Scan(basePath string) ([]Entry, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}

	now := time.Now()
	datePrefix := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

	var result []Entry
	for _, e := range entries {
		// Skip hidden directories
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}

		// Only include directories
		if !e.IsDir() {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		mtime := info.ModTime()
		hoursSinceAccess := now.Sub(mtime).Hours()

		// Base score from recency: 3.0 / sqrt(hours + 1)
		baseScore := 3.0 / sqrt(hoursSinceAccess+1)

		// Bonus for date-prefixed directories
		if datePrefix.MatchString(e.Name()) {
			baseScore += 2.0
		}

		result = append(result, Entry{
			Name:      e.Name(),
			Path:      filepath.Join(basePath, e.Name()),
			ModTime:   mtime,
			BaseScore: baseScore,
		})
	}

	// Sort by modification time (most recent first)
	sort.Slice(result, func(i, j int) bool {
		return result[i].ModTime.After(result[j].ModTime)
	})

	return result, nil
}

// sqrt is a simple square root approximation using Newton's method.
func sqrt(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 10; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

// Touch updates the modification time of a directory.
func Touch(path string) error {
	now := time.Now()
	return os.Chtimes(path, now, now)
}

// Create creates a new date-prefixed directory and returns its path.
func Create(basePath, name string) (string, error) {
	// Sanitize name: replace spaces with hyphens
	name = strings.ReplaceAll(strings.TrimSpace(name), " ", "-")

	// Create date prefix
	datePrefix := time.Now().Format("2006-01-02")
	dirName := fmt.Sprintf("%s-%s", datePrefix, name)

	// Ensure unique name
	dirName = uniqueName(basePath, dirName)

	fullPath := filepath.Join(basePath, dirName)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		return "", err
	}

	return fullPath, nil
}

// uniqueName returns a unique directory name by appending -2, -3, etc. if needed.
func uniqueName(basePath, name string) string {
	candidate := name
	i := 2
	for {
		path := filepath.Join(basePath, candidate)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d", name, i)
		i++
	}
}

// DatePrefix returns today's date in YYYY-MM-DD format.
func DatePrefix() string {
	return time.Now().Format("2006-01-02")
}

// Delete removes a directory and all its contents.
// It validates that the path is inside basePath for safety.
func Delete(basePath, path string) error {
	// Resolve to absolute paths
	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return fmt.Errorf("failed to resolve base path: %w", err)
	}

	absTarget, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve target path: %w", err)
	}

	// Resolve symlinks
	realBase, err := filepath.EvalSymlinks(absBase)
	if err != nil {
		return fmt.Errorf("failed to resolve base symlinks: %w", err)
	}

	realTarget, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		return fmt.Errorf("failed to resolve target symlinks: %w", err)
	}

	// Safety check: target must be inside base
	if !strings.HasPrefix(realTarget, realBase+string(filepath.Separator)) {
		return fmt.Errorf("safety check failed: %s is not inside %s", realTarget, realBase)
	}

	return os.RemoveAll(realTarget)
}
