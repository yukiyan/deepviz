package app

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// ImageConfig holds image generation configuration.
type ImageConfig struct {
	Model       string // Model name (default: gemini-3-pro-image-preview)
	AspectRatio string // Aspect ratio (default: 16:9)
	ImageSize   string // Image size (default: 2K)
}

// ImageResult holds image generation result.
type ImageResult struct {
	ImagePath    string // Saved image path
	ResponsePath string // Raw response path
}

// GenaiImageClient is an image generation client.
type GenaiImageClient struct {
	config *ViperConfig
	logger Logger
}

// NewGenaiImageClient creates a new GenaiImageClient.
func NewGenaiImageClient(ctx context.Context, config *ViperConfig, logger Logger) (*GenaiImageClient, error) {
	return &GenaiImageClient{
		config: config,
		logger: logger,
	}, nil
}

// sanitizePrompt removes potentially dangerous control characters while preserving valid whitespace.
func sanitizeImagePrompt(prompt string) string {
	var builder strings.Builder
	builder.Grow(len(prompt))

	for _, r := range prompt {
		// Allow printable characters, whitespace (space, tab, newline, etc.), and non-ASCII Unicode
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			builder.WriteRune(r)
		}
		// Skip control characters (NULL, BEL, ESC, etc.)
	}

	return builder.String()
}

// BuildInfographicsPrompt builds an infographics generation prompt from Markdown content.
//
// The prompt language is controlled by ImageLang configuration (e.g., "Japanese", "English", "French").
//
// Template:
//
//	Take a good look at the content below and turn it into a single infographic image in {ImageLang}.
//	```
//	{markdown}
//	```
func (c *GenaiImageClient) BuildInfographicsPrompt(markdown string) string {
	// Sanitize markdown content
	sanitizedMarkdown := sanitizeImagePrompt(markdown)

	promptTemplate := `Take a good look at the content below and turn it into a single infographic image in %s.
` + "```" + `
%s
` + "```"

	return fmt.Sprintf(promptTemplate, c.config.ImageLang, sanitizedMarkdown)
}

// Generate generates and saves an image.
func (c *GenaiImageClient) Generate(ctx context.Context, prompt string, imgConfig ImageConfig, timestamp string) (*ImageResult, error) {
	// Sanitize prompt
	sanitizedPrompt := sanitizeImagePrompt(prompt)

	// Create request body
	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": sanitizedPrompt},
				},
			},
		},
		"tools": []map[string]interface{}{
			{"google_search": map[string]interface{}{}},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"TEXT", "IMAGE"},
			"imageConfig": map[string]interface{}{
				"aspectRatio": imgConfig.AspectRatio,
				"imageSize":   imgConfig.ImageSize,
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Get HTTP client
	httpClient := &http.Client{
		Timeout: 120 * time.Second, // Image generation takes time
	}

	// Create HTTP request
	baseURL := "https://generativelanguage.googleapis.com"
	url := baseURL + "/v1beta/models/" + imgConfig.Model + ":generateContent"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", c.config.APIKey)

	// Execute request
	c.logger.Info("Generating image", "model", imgConfig.Model, "aspect_ratio", imgConfig.AspectRatio, "size", imgConfig.ImageSize)
	c.logger.Trace("HTTP Request", "url", url, "method", "POST", "body", string(bodyBytes))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.logger.Trace("HTTP Response", "url", url, "status_code", resp.StatusCode, "body", string(body))

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse JSON
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text       string `json:"text,omitempty"`
					InlineData struct {
						Data     string `json:"data"`
						MimeType string `json:"mimeType"`
					} `json:"inlineData,omitempty"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Extract image data
	var base64ImageData string
	for _, candidate := range response.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData.Data != "" {
				base64ImageData = part.InlineData.Data
				break
			}
		}
		if base64ImageData != "" {
			break
		}
	}

	if base64ImageData == "" {
		return nil, fmt.Errorf("no image data found in response")
	}

	// Decode Base64
	imageData, err := base64.StdEncoding.DecodeString(base64ImageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 image data: %w", err)
	}

	// Build file paths
	imagePath := filepath.Join(c.config.ImagesDir(), timestamp+".png")
	responsePath := filepath.Join(c.config.ResponsesDir(), timestamp+"_image.json")

	// Save image file
	if err := WriteFile(imagePath, imageData); err != nil {
		return nil, fmt.Errorf("failed to write image file: %w", err)
	}

	c.logger.Info("Image saved", "path", imagePath)

	// Save raw response
	if err := WriteFile(responsePath, body); err != nil {
		return nil, fmt.Errorf("failed to write response file: %w", err)
	}

	c.logger.Info("Raw response saved", "path", responsePath)

	return &ImageResult{
		ImagePath:    imagePath,
		ResponsePath: responsePath,
	}, nil
}
