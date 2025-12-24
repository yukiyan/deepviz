package app

import (
	"context"
	"os"
	"testing"
)

func TestNewGenaiImageClient(t *testing.T) {
	// Skip if API key is not set
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	config := &ViperConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	}
	logger := NewNullLogger()

	client, err := NewGenaiImageClient(ctx, config, logger)
	if err != nil {
		t.Fatalf("failed to create genai image client: %v", err)
	}

	if client == nil {
		t.Error("client should not be nil")
	}
}

func TestGenaiImageClient_Generate(t *testing.T) {
	// Skip if API key is not set
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	tmpDir := t.TempDir()
	config := &ViperConfig{
		OutputDir: tmpDir,
		APIKey:    apiKey,
	}
	logger := NewNullLogger()

	client, err := NewGenaiImageClient(ctx, config, logger)
	if err != nil {
		t.Fatalf("failed to create genai image client: %v", err)
	}

	// Test with simple prompt
	prompt := "A beautiful sunset over mountains"
	imageConfig := ImageConfig{
		Model:       "gemini-3-pro-image-preview",
		AspectRatio: "16:9",
		ImageSize:   "2K",
	}

	result, err := client.Generate(ctx, prompt, imageConfig, "test-timestamp")
	if err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	// Verify result
	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.ImagePath == "" {
		t.Error("image path should not be empty")
	}

	// Verify file was created
	if _, err := os.Stat(result.ImagePath); os.IsNotExist(err) {
		t.Error("image file should be created")
	}
}

func TestGenaiImageClient_BuildInfographicsPrompt(t *testing.T) {
	ctx := context.Background()
	config := &ViperConfig{
		APIKey: "dummy-api-key",
	}
	logger := NewNullLogger()

	client, err := NewGenaiImageClient(ctx, config, logger)
	if err != nil {
		t.Fatalf("failed to create genai image client: %v", err)
	}

	markdown := "# Test\nThis is a test markdown."
	prompt := client.BuildInfographicsPrompt(markdown)

	if prompt == "" {
		t.Error("prompt should not be empty")
	}

	if len(prompt) <= len(markdown) {
		t.Error("prompt should be longer than markdown (contains template)")
	}
}
