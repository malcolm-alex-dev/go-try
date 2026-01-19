package workspace

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScan(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create some test directories
	dirs := []string{"2024-01-15-project-a", "2024-01-20-project-b", "other-dir"}
	for _, d := range dirs {
		if err := os.Mkdir(filepath.Join(tmpDir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create a hidden directory (should be skipped)
	if err := os.Mkdir(filepath.Join(tmpDir, ".hidden"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a file (should be skipped)
	if err := os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	entries, err := Scan(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}

	// Check that entries have expected fields
	for _, e := range entries {
		if e.Name == "" {
			t.Error("entry name should not be empty")
		}
		if e.Path == "" {
			t.Error("entry path should not be empty")
		}
		if e.ModTime.IsZero() {
			t.Error("entry mod time should not be zero")
		}
	}
}

func TestScanEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	entries, err := Scan(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestScanNonExistent(t *testing.T) {
	entries, err := Scan("/nonexistent/path")
	if err != nil {
		t.Fatal("expected no error for non-existent path")
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestCreate(t *testing.T) {
	tmpDir := t.TempDir()

	path, err := Create(tmpDir, "test project")
	if err != nil {
		t.Fatal(err)
	}

	// Check path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("created directory should exist")
	}

	// Check name format (YYYY-MM-DD-test-project)
	base := filepath.Base(path)
	datePrefix := time.Now().Format("2006-01-02")
	expected := datePrefix + "-test-project"
	if base != expected {
		t.Errorf("expected %s, got %s", expected, base)
	}
}

func TestCreateUnique(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first directory
	path1, err := Create(tmpDir, "test")
	if err != nil {
		t.Fatal(err)
	}

	// Create second directory with same name (should get -2 suffix)
	path2, err := Create(tmpDir, "test")
	if err != nil {
		t.Fatal(err)
	}

	if path1 == path2 {
		t.Error("second path should be different")
	}

	base2 := filepath.Base(path2)
	if base2[len(base2)-2:] != "-2" {
		t.Errorf("expected -2 suffix, got %s", base2)
	}
}

func TestTouch(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test")
	os.Mkdir(testDir, 0755)

	// Get original mtime
	info1, _ := os.Stat(testDir)
	time.Sleep(10 * time.Millisecond)

	// Touch
	if err := Touch(testDir); err != nil {
		t.Fatal(err)
	}

	// Get new mtime
	info2, _ := os.Stat(testDir)

	if !info2.ModTime().After(info1.ModTime()) {
		t.Error("mtime should be updated after touch")
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "to-delete")
	os.Mkdir(testDir, 0755)

	// Create a file inside
	os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("test"), 0644)

	err := Delete(tmpDir, testDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("directory should be deleted")
	}
}

func TestDeleteSafety(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := t.TempDir()

	// Try to delete directory outside base path
	err := Delete(tmpDir, outsideDir)
	if err == nil {
		t.Error("expected error when deleting outside base path")
	}
}

func TestDatePrefix(t *testing.T) {
	prefix := DatePrefix()
	expected := time.Now().Format("2006-01-02")
	if prefix != expected {
		t.Errorf("expected %s, got %s", expected, prefix)
	}
}
