package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"deepviz/internal/genai/interactions"
)

// ResearchResult holds research result.
type ResearchResult struct {
	InteractionID string // Research ID
	Status        string // Completion status
	Content       string // Markdown content
	MarkdownPath  string // Save destination path
	ResponsePath  string // Raw response save destination
}

// GenaiResearchClient is a Deep Research API client.
type GenaiResearchClient struct {
	config *ViperConfig
	logger Logger
	client *interactions.ClientWithResponses
}

// NewGenaiResearchClient creates a new GenaiResearchClient.
func NewGenaiResearchClient(ctx context.Context, config *ViperConfig, logger Logger) (*GenaiResearchClient, error) {
	baseURL := "https://generativelanguage.googleapis.com"

	client, err := interactions.NewClientWithResponses(baseURL,
		interactions.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("x-goog-api-key", config.APIKey)
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create interactions client: %w", err)
	}

	return &GenaiResearchClient{
		config: config,
		logger: logger,
		client: client,
	}, nil
}

// sanitizePrompt removes potentially dangerous control characters while preserving valid whitespace.
func sanitizePrompt(prompt string) string {
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

// Execute executes Deep Research.
func (c *GenaiResearchClient) Execute(ctx context.Context, prompt string, timestamp string) (*ResearchResult, error) {
	// Start research
	interactionID, err := c.startResearch(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to start research: %w", err)
	}

	c.logger.Info("Research started", "interaction_id", interactionID)

	// Cancel research on failure (defer runs even if ctx is cancelled)
	var success bool
	defer func() {
		if !success {
			if cancelErr := c.cancelResearch(interactionID); cancelErr != nil {
				c.logger.Error("Failed to cancel research", "error", cancelErr)
			}
		}
	}()

	// Wait for completion by polling
	result, err := c.pollUntilComplete(ctx, interactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to poll research: %w", err)
	}

	// Save result
	if err := c.saveResult(result, timestamp); err != nil {
		return nil, fmt.Errorf("failed to save result: %w", err)
	}

	success = true
	return result, nil
}

// startResearch starts a research.
func (c *GenaiResearchClient) startResearch(ctx context.Context, prompt string) (string, error) {
	// Sanitize prompt to remove potentially dangerous control characters
	sanitizedPrompt := sanitizePrompt(prompt)

	// Create request body manually to avoid generated code issues with agent_config type
	// The generated code sets type="deep_research" but API expects "deep-research"
	requestBodyMap := map[string]interface{}{
		"input":      sanitizedPrompt,
		"agent":      c.config.DeepResearchAgent,
		"background": true,
		"store":      true,
		"agent_config": map[string]interface{}{
			"type":               "deep-research", // API expects hyphen, not underscore
			"thinking_summaries": "auto",
		},
		"tools": []map[string]interface{}{
			{"type": "google_search"},
			{"type": "url_context"},
		},
	}

	c.logger.Debug("Sending request", "agent", c.config.DeepResearchAgent)

	// Marshal request body to JSON
	bodyJSON, err := json.Marshal(requestBodyMap)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Trace log request body
	c.logger.Trace("HTTP Request", "method", "POST", "body", string(bodyJSON))

	// Execute request using WithBody variant to avoid union type issues
	resp, err := c.client.CreateInteractionWithBodyWithResponse(ctx, "v1beta", "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create interaction: %w", err)
	}

	// Trace log response (raw body)
	c.logger.Trace("HTTP Response", "status_code", resp.StatusCode(), "body", string(resp.Body))

	c.logger.Debug("Response received", "status_code", resp.StatusCode())

	// Check status code
	if resp.StatusCode() != http.StatusOK {
		// Log error details from JSONDefault if available
		var errorMsg string
		if resp.JSONDefault != nil && resp.JSONDefault.Error != nil {
			if resp.JSONDefault.Error.Message != nil {
				errorMsg = *resp.JSONDefault.Error.Message
			}
			if resp.JSONDefault.Error.Code != nil {
				errorMsg = fmt.Sprintf("code=%s, message=%s", *resp.JSONDefault.Error.Code, errorMsg)
			}
		} else {
			errorMsg = string(resp.Body)
		}
		c.logger.Error("API request failed", "status_code", resp.StatusCode(), "error", errorMsg)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode(), errorMsg)
	}

	// Parse response
	if resp.JSON200 == nil {
		return "", fmt.Errorf("empty response body")
	}

	interaction := resp.JSON200
	if interaction.Id == nil || *interaction.Id == "" {
		return "", fmt.Errorf("empty interaction ID in response")
	}

	return *interaction.Id, nil
}

// pollUntilComplete polls until research completes.
func (c *GenaiResearchClient) pollUntilComplete(ctx context.Context, interactionID string) (*ResearchResult, error) {
	ticker := time.NewTicker(time.Duration(c.config.PollInterval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(c.config.PollTimeout) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("polling timeout after %d seconds", c.config.PollTimeout)
		case <-ticker.C:
			// Check status
			result, err := c.checkStatus(ctx, interactionID)
			if err != nil {
				return nil, err
			}

			// Return result if completed
			if result.Status == "completed" {
				c.logger.Info("Research completed", "interaction_id", interactionID)
				return result, nil
			}

			// Return error if failed
			if result.Status == "failed" {
				return nil, fmt.Errorf("research failed. Interaction ID: %s", interactionID)
			}

			c.logger.Info("Research in progress", "status", result.Status)
		}
	}
}

// checkStatus checks research status.
func (c *GenaiResearchClient) checkStatus(ctx context.Context, interactionID string) (*ResearchResult, error) {
	resp, err := c.client.GetInteractionByIdWithResponse(ctx, "v1beta", interactionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get interaction: %w", err)
	}

	// Trace log response (raw body)
	c.logger.Trace("HTTP Response", "status_code", resp.StatusCode(), "body", string(resp.Body))

	// Check status code
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), string(resp.Body))
	}

	if resp.JSON200 == nil {
		return nil, fmt.Errorf("empty response body")
	}

	interaction := resp.JSON200

	// Extract status
	var status string
	if interaction.Status != nil {
		status = string(*interaction.Status)
	}

	// Extract text content from outputs
	var content string
	if interaction.Outputs != nil {
		for _, output := range *interaction.Outputs {
			// Content is a union type, try to extract as TextContent
			textContent, err := output.AsTextContent()
			if err == nil && textContent.Text != nil {
				content = *textContent.Text
				break
			}
		}
	}

	return &ResearchResult{
		InteractionID: interactionID,
		Status:        status,
		Content:       content,
	}, nil
}

// cancelResearch cancels a research interaction.
func (c *GenaiResearchClient) cancelResearch(interactionID string) error {
	// Use background context since the original context may be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.client.CancelInteractionByIdWithResponse(ctx, "v1beta", interactionID)
	if err != nil {
		return fmt.Errorf("failed to cancel research: %w", err)
	}

	// Trace log response (raw body)
	c.logger.Trace("HTTP Response", "status_code", resp.StatusCode(), "body", string(resp.Body))

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("cancel failed with status %d: %s", resp.StatusCode(), string(resp.Body))
	}

	c.logger.Info("Research cancelled", "interaction_id", interactionID)
	return nil
}

// saveResult saves the research result.
func (c *GenaiResearchClient) saveResult(result *ResearchResult, timestamp string) error {
	// Build file path
	markdownPath := filepath.Join(c.config.ResearchDir(), timestamp+".md")

	// Save markdown file
	if err := WriteFile(markdownPath, []byte(result.Content)); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	c.logger.Info("Research saved", "path", markdownPath)

	// Set path to result
	result.MarkdownPath = markdownPath

	return nil
}
