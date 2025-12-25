package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ViperConfig holds application configuration using Viper.
type ViperConfig struct {
	// OutputDir is the base path for output directory
	OutputDir string
	// APIKey is the Gemini API key
	APIKey string
	// DeepResearchAgent is the Deep Research API agent name
	DeepResearchAgent string
	// PollInterval is the polling interval in seconds
	PollInterval int
	// PollTimeout is the polling timeout in seconds
	PollTimeout int
	// Model is the image generation model name
	Model string
	// AspectRatio is the aspect ratio for image generation
	AspectRatio string
	// ImageSize is the image size for generation
	ImageSize string
	// ImageLang is the language for image generation (e.g., "Japanese", "English", "French")
	ImageLang string
	// AutoOpen enables automatic opening of generated images
	AutoOpen bool

	configDir string
	v         *viper.Viper
}

// NewViperConfig creates a new ViperConfig by loading configuration from environment variables and config file.
//
// Priority (high to low):
//  1. Environment variables
//  2. Config file
//  3. Default values
//
// If configDir is empty, XDG_CONFIG_HOME is used.
func NewViperConfig(configDir string) (*ViperConfig, error) {
	// Create a new Viper instance (avoid global state)
	v := viper.New()

	// Set default output directory (XDG Base Directory compliant)
	defaultOutputDir := "/tmp/deepviz-output"
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			xdgDataHome = filepath.Join(home, ".local", "share")
		}
	}
	if xdgDataHome != "" {
		defaultOutputDir = filepath.Join(xdgDataHome, "deepviz")
	}

	// Set default values
	v.SetDefault("output_dir", defaultOutputDir)
	v.SetDefault("deep_research_agent", "deep-research-pro-preview-12-2025")
	v.SetDefault("poll_interval", 10)
	v.SetDefault("poll_timeout", 600)
	v.SetDefault("model", "gemini-3-pro-image-preview")
	v.SetDefault("aspect_ratio", "16:9")
	v.SetDefault("image_size", "2K")
	v.SetDefault("image_lang", "Japanese")
	v.SetDefault("auto_open", true)

	// Set environment variable prefix
	v.SetEnvPrefix("DEEPVIZ")
	v.AutomaticEnv()

	// Determine config file directory (XDG Base Directory compliant)
	if configDir == "" {
		// Use XDG_CONFIG_HOME if set, otherwise default to ~/.config
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			xdgConfigHome = filepath.Join(home, ".config")
		}
		configDir = filepath.Join(xdgConfigHome, "deepviz")
	}

	// Load config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(configDir)

	// Read config file if it exists (don't error if it doesn't)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Map configuration to struct
	// Priority: DEEPVIZ_API_KEY (env) > GEMINI_API_KEY (env) > config file
	apiKey := os.Getenv("DEEPVIZ_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		apiKey = v.GetString("api_key")
	}

	// Priority: DEEPVIZ_MODEL (env) > GEMINI_MODEL (env) > config file
	model := os.Getenv("DEEPVIZ_MODEL")
	if model == "" {
		model = os.Getenv("GEMINI_MODEL")
	}
	if model == "" {
		model = v.GetString("model")
	}

	// Priority: DEEPVIZ_DEEP_RESEARCH_AGENT (env) > GEMINI_DEEP_RESEARCH_AGENT (env) > config file
	deepResearchAgent := os.Getenv("DEEPVIZ_DEEP_RESEARCH_AGENT")
	if deepResearchAgent == "" {
		deepResearchAgent = os.Getenv("GEMINI_DEEP_RESEARCH_AGENT")
	}
	if deepResearchAgent == "" {
		deepResearchAgent = v.GetString("deep_research_agent")
	}

	config := &ViperConfig{
		OutputDir:         v.GetString("output_dir"),
		APIKey:            apiKey,
		DeepResearchAgent: deepResearchAgent,
		PollInterval:      v.GetInt("poll_interval"),
		PollTimeout:       v.GetInt("poll_timeout"),
		Model:             model,
		AspectRatio:       v.GetString("aspect_ratio"),
		ImageSize:         v.GetString("image_size"),
		ImageLang:         v.GetString("image_lang"),
		AutoOpen:          v.GetBool("auto_open"),
		configDir:         configDir,
		v:                 v,
	}

	return config, nil
}

// ResearchDir returns the output directory for research results.
func (c *ViperConfig) ResearchDir() string {
	return filepath.Join(c.OutputDir, "research")
}

// ImagesDir returns the output directory for images.
func (c *ViperConfig) ImagesDir() string {
	return filepath.Join(c.OutputDir, "images")
}

// ResponsesDir returns the output directory for raw responses.
func (c *ViperConfig) ResponsesDir() string {
	return filepath.Join(c.OutputDir, "responses")
}

// LogsDir returns the output directory for logs.
func (c *ViperConfig) LogsDir() string {
	return filepath.Join(c.OutputDir, "logs")
}

// EnsureDirectories ensures all output directories exist.
func (c *ViperConfig) EnsureDirectories() error {
	dirs := []string{
		c.ResearchDir(),
		c.ImagesDir(),
		c.ResponsesDir(),
		c.LogsDir(),
	}

	for _, dir := range dirs {
		if err := EnsureDir(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// Set sets a configuration value.
func (c *ViperConfig) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// Save saves the current configuration to the config file.
func (c *ViperConfig) Save() error {
	// Ensure config directory exists
	if err := os.MkdirAll(c.configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(c.configDir, "config.yaml")
	if err := c.v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}
