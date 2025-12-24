package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// GenerateTimestamp generates a timestamp string from the current time.
//
// Format: YYYYMMDD_HHMMSS
func GenerateTimestamp() string {
	return time.Now().Format("20060102_150405")
}

// EnsureDir ensures that a directory exists.
//
// Creates the directory if it doesn't exist.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

// WriteFile writes data to a file.
//
// Automatically creates the directory if it doesn't exist.
func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(path, data, 0644)
}

// ReadFile reads data from a file.
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
