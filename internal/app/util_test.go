package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestGenerateTimestamp tests timestamp generation.
func TestGenerateTimestamp(t *testing.T) {
	timestamp := GenerateTimestamp()

	// Verify format (YYYYMMDD_HHMMSS = 15 characters)
	if len(timestamp) != 15 {
		t.Errorf("expected timestamp length 15, got %d", len(timestamp))
	}

	// Verify underscore position
	if timestamp[8] != '_' {
		t.Errorf("expected underscore at position 8, got %c", timestamp[8])
	}

	// Validate with time.Parse
	_, err := time.Parse("20060102_150405", timestamp)
	if err != nil {
		t.Errorf("failed to parse timestamp %s: %v", timestamp, err)
	}
}

// TestEnsureDir tests directory creation.
func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test", "nested", "dir")

	// Create directory
	if err := EnsureDir(testDir); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Verify existence
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("directory does not exist: %v", err)
	}

	if !info.IsDir() {
		t.Error("expected directory, got file")
	}

	// Verify no error when ensuring existing directory
	if err := EnsureDir(testDir); err != nil {
		t.Errorf("failed to ensure existing directory: %v", err)
	}
}

// TestWriteFile_Success tests successful file writing.
func TestWriteFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test", "file.txt")
	testData := []byte("test content")

	// Write file
	if err := WriteFile(testFile, testData); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

// TestWriteFile_InvalidPath tests error with invalid path.
func TestWriteFile_InvalidPath(t *testing.T) {
	// Invalid path (non-existent device)
	invalidPath := "/dev/null/invalid/path/file.txt"
	err := WriteFile(invalidPath, []byte("test"))
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

// TestReadFile_Success tests successful file reading.
func TestReadFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("test content")

	// Create test file
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Read file
	data, err := ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("expected %s, got %s", testData, data)
	}
}

// TestReadFile_NotFound tests error with non-existent file.
func TestReadFile_NotFound(t *testing.T) {
	_, err := ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
