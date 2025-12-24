package app

import (
	"context"
	"os"
	"testing"
)

func TestNewGenaiResearchClient(t *testing.T) {
	// Skip if API key is not set
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	config := &ViperConfig{
		DeepResearchAgent: "deep-research-pro-preview-12-2025",
		PollInterval:      10,
		PollTimeout:       600,
	}
	logger := NewNullLogger()

	client, err := NewGenaiResearchClient(ctx, config, logger)
	if err != nil {
		t.Fatalf("failed to create genai research client: %v", err)
	}

	if client == nil {
		t.Error("client should not be nil")
	}
}

func TestGenaiResearchClient_Execute(t *testing.T) {
	// Skip if API key is not set
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()
	config := &ViperConfig{
		OutputDir:         tmpDir,
		APIKey:            apiKey,
		DeepResearchAgent: "deep-research-pro-preview-12-2025",
		PollInterval:      2,
		PollTimeout:       60,
	}
	logger := NewNullLogger()

	client, err := NewGenaiResearchClient(ctx, config, logger)
	if err != nil {
		t.Fatalf("failed to create genai research client: %v", err)
	}

	// Test with simple prompt
	prompt := "Goプログラミング言語の特徴を3つ教えてください"
	result, err := client.Execute(ctx, prompt, "test-timestamp")
	if err != nil {
		t.Fatalf("failed to execute research: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Content == "" {
		t.Error("content should not be empty")
	}

	if result.MarkdownPath == "" {
		t.Error("markdown path should not be empty")
	}

	// Verify file was created
	if _, err := os.Stat(result.MarkdownPath); os.IsNotExist(err) {
		t.Error("markdown file should be created")
	}
}

func TestGenaiResearchClient_Cancel(t *testing.T) {
	// Skip if API key is not set
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()
	config := &ViperConfig{
		OutputDir:         tmpDir,
		APIKey:            apiKey,
		DeepResearchAgent: "deep-research-pro-preview-12-2025",
		PollInterval:      2,
		PollTimeout:       60,
	}
	logger := NewNullLogger()

	client, err := NewGenaiResearchClient(ctx, config, logger)
	if err != nil {
		t.Fatalf("failed to create genai research client: %v", err)
	}

	// Cancel context
	ctx, cancel := context.WithCancel(ctx)
	cancel()

	// Execute with cancelled context
	prompt := "長いリサーチタスク"
	_, err = client.Execute(ctx, prompt, "test-timestamp")
	if err == nil {
		t.Error("should return error when context is cancelled")
	}
}
