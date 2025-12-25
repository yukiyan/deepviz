package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// OpenFile opens a file with the system's default application.
//
// Supports cross-platform file opening:
// - macOS: open command
// - Linux: xdg-open command
// - Windows: cmd /c start command
func OpenFile(path string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{path}
	case "linux":
		cmd = "xdg-open"
		args = []string{path}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", "", path}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
