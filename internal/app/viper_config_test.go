package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestViperConfig_DefaultValues(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Clear environment variables
	os.Unsetenv("DEEPVIZ_OUTPUT_DIR")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("DEEPVIZ_API_KEY")

	config, err := NewViperConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to create viper config: %v", err)
	}

	// Verify default values
	if config.OutputDir == "" {
		t.Error("OutputDir should have default value")
	}

	if config.DeepResearchAgent != "deep-research-pro-preview-12-2025" {
		t.Errorf("DeepResearchAgent = %s, want deep-research-pro-preview-12-2025", config.DeepResearchAgent)
	}

	if config.PollInterval != 10 {
		t.Errorf("PollInterval = %d, want 10", config.PollInterval)
	}

	if config.PollTimeout != 600 {
		t.Errorf("PollTimeout = %d, want 600", config.PollTimeout)
	}
}

func TestViperConfig_EnvironmentVariables(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Set environment variables
	os.Setenv("DEEPVIZ_OUTPUT_DIR", "/custom/output")
	os.Setenv("GEMINI_API_KEY", "test-api-key")
	defer func() {
		os.Unsetenv("DEEPVIZ_OUTPUT_DIR")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	config, err := NewViperConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to create viper config: %v", err)
	}

	// Verify loading from environment variables
	if config.OutputDir != "/custom/output" {
		t.Errorf("OutputDir = %s, want /custom/output", config.OutputDir)
	}

	if config.APIKey != "test-api-key" {
		t.Errorf("APIKey = %s, want test-api-key", config.APIKey)
	}
}

func TestViperConfig_ConfigFile(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Create config file
	configContent := `
output_dir: /file/output
api_key: file-api-key
deep_research_agent: custom-agent
poll_interval: 20
poll_timeout: 1200
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	config, err := NewViperConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to create viper config: %v", err)
	}

	// Verify loading from config file
	if config.OutputDir != "/file/output" {
		t.Errorf("OutputDir = %s, want /file/output", config.OutputDir)
	}

	if config.APIKey != "file-api-key" {
		t.Errorf("APIKey = %s, want file-api-key", config.APIKey)
	}

	if config.DeepResearchAgent != "custom-agent" {
		t.Errorf("DeepResearchAgent = %s, want custom-agent", config.DeepResearchAgent)
	}

	if config.PollInterval != 20 {
		t.Errorf("PollInterval = %d, want 20", config.PollInterval)
	}

	if config.PollTimeout != 1200 {
		t.Errorf("PollTimeout = %d, want 1200", config.PollTimeout)
	}
}

func TestViperConfig_Priority(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Create config file
	configContent := `
output_dir: /file/output
api_key: file-api-key
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Set environment variables (env should take precedence)
	os.Setenv("DEEPVIZ_OUTPUT_DIR", "/env/output")
	os.Setenv("GEMINI_API_KEY", "env-api-key")
	defer func() {
		os.Unsetenv("DEEPVIZ_OUTPUT_DIR")
		os.Unsetenv("GEMINI_API_KEY")
	}()

	config, err := NewViperConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to create viper config: %v", err)
	}

	// Verify that environment variables take precedence
	if config.OutputDir != "/env/output" {
		t.Errorf("OutputDir = %s, want /env/output (env should override file)", config.OutputDir)
	}

	if config.APIKey != "env-api-key" {
		t.Errorf("APIKey = %s, want env-api-key (env should override file)", config.APIKey)
	}
}

func TestViperConfig_Save(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Clear environment variables
	os.Unsetenv("DEEPVIZ_OUTPUT_DIR")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("DEEPVIZ_API_KEY")

	config, err := NewViperConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to create viper config: %v", err)
	}

	// Modify config
	config.Set("api_key", "new-api-key")
	config.Set("output_dir", "/new/output")

	// Save config
	if err := config.Save(); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should be created")
	}

	// Load new config
	newConfig, err := NewViperConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}

	// Verify saved values are loaded
	if newConfig.APIKey != "new-api-key" {
		t.Errorf("APIKey = %s, want new-api-key", newConfig.APIKey)
	}

	if newConfig.OutputDir != "/new/output" {
		t.Errorf("OutputDir = %s, want /new/output", newConfig.OutputDir)
	}
}
