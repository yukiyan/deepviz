package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

// Options holds CLI options.
type Options struct {
	Prompt       string
	File         string
	ResearchOnly bool
	ImageOnly    bool
	Model        string
	AspectRatio  string
	ImageSize    string
	Output       string
	Verbose      bool
	NoOpen       bool
}

// NewRootCommand creates the root command.
//
// The root command executes research and image generation.
func NewRootCommand() *cobra.Command {
	var (
		prompt       string
		file         string
		output       string
		verbose      bool
		researchOnly bool
		imageOnly    bool
		model        string
		aspectRatio  string
		imageSize    string
		noOpen       bool
	)

	rootCmd := &cobra.Command{
		Use:     "deepviz",
		Short:   "Research and image generation tool using Gemini API",
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Error if neither prompt nor file is specified
			if prompt == "" && file == "" {
				return fmt.Errorf("either --prompt or --file must be specified")
			}

			// Load configuration
			config, err := NewViperConfig("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Override with flags if explicitly set
			if output != "" {
				config.OutputDir = output
			}
			if cmd.Flags().Changed("model") {
				config.Model = model
			}
			if cmd.Flags().Changed("aspect-ratio") {
				config.AspectRatio = aspectRatio
			}
			if cmd.Flags().Changed("image-size") {
				config.ImageSize = imageSize
			}

			// Create options
			opts := &Options{
				Prompt:       prompt,
				File:         file,
				Output:       config.OutputDir,
				Verbose:      verbose,
				ResearchOnly: researchOnly,
				ImageOnly:    imageOnly,
				Model:        config.Model,
				AspectRatio:  config.AspectRatio,
				ImageSize:    config.ImageSize,
				NoOpen:       noOpen,
			}

			// Execute Run function (existing logic)
			return RunWithConfig(opts, config)
		},
	}

	// Define flags
	rootCmd.Flags().StringVarP(&prompt, "prompt", "p", "", "Generation prompt")
	rootCmd.Flags().StringVarP(&file, "file", "f", "", "Prompt file path")
	rootCmd.Flags().StringVarP(&output, "output", "o", "", "Output directory")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging (DEBUG level)")
	rootCmd.Flags().BoolVar(&researchOnly, "research-only", false, "Execute research only")
	rootCmd.Flags().BoolVar(&imageOnly, "image-only", false, "Execute image generation only")
	rootCmd.Flags().StringVar(&model, "model", "gemini-3-pro-image-preview", "Image generation model name")
	rootCmd.Flags().StringVar(&aspectRatio, "aspect-ratio", "16:9", "Aspect ratio")
	rootCmd.Flags().StringVar(&imageSize, "image-size", "2K", "Image size")
	rootCmd.Flags().BoolVar(&noOpen, "no-open", false, "Disable auto-open after image generation")

	// --no-image is an alias for --research-only
	rootCmd.Flags().BoolVar(&researchOnly, "no-image", false, "Skip image generation (same as --research-only)")

	// Register completion functions for flags
	rootCmd.RegisterFlagCompletionFunc("file", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterFileExt
	})
	rootCmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveFilterDirs
	})
	rootCmd.RegisterFlagCompletionFunc("model", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"gemini-3-pro-image-preview\tGemini 3 Pro Image Preview",
			"gemini-2.0-flash-exp\tGemini 2.0 Flash Experimental",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.RegisterFlagCompletionFunc("aspect-ratio", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"16:9\tWidescreen",
			"4:3\tStandard",
			"1:1\tSquare",
			"9:16\tPortrait",
			"3:4\tPortrait standard",
		}, cobra.ShellCompDirectiveNoFileComp
	})
	rootCmd.RegisterFlagCompletionFunc("image-size", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			"2K\t2048x1152",
			"4K\t3840x2160",
		}, cobra.ShellCompDirectiveNoFileComp
	})

	// Add subcommands
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newCompletionCommand())

	return rootCmd
}

// newConfigCommand creates the configuration management command.
func newConfigCommand() *cobra.Command {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	// config show command
	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := NewViperConfig("")
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Display configuration
			fmt.Fprintf(cmd.OutOrStdout(), "Current Configuration:\n")
			fmt.Fprintf(cmd.OutOrStdout(), "  output_dir: %s\n", config.OutputDir)
			fmt.Fprintf(cmd.OutOrStdout(), "  api_key: %s\n", maskAPIKey(config.APIKey))
			fmt.Fprintf(cmd.OutOrStdout(), "  deep_research_agent: %s\n", config.DeepResearchAgent)
			fmt.Fprintf(cmd.OutOrStdout(), "  poll_interval: %d\n", config.PollInterval)
			fmt.Fprintf(cmd.OutOrStdout(), "  poll_timeout: %d\n", config.PollTimeout)
			fmt.Fprintf(cmd.OutOrStdout(), "  model: %s\n", config.Model)
			fmt.Fprintf(cmd.OutOrStdout(), "  aspect_ratio: %s\n", config.AspectRatio)
			fmt.Fprintf(cmd.OutOrStdout(), "  image_size: %s\n", config.ImageSize)
			fmt.Fprintf(cmd.OutOrStdout(), "  image_lang: %s\n", config.ImageLang)

			return nil
		},
	}

	// config init command
	var configDir string
	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine config file directory (XDG Base Directory compliant)
			if configDir == "" {
				xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
				if xdgConfigHome == "" {
					home, err := os.UserHomeDir()
					if err != nil {
						return fmt.Errorf("failed to get home directory: %w", err)
					}
					xdgConfigHome = filepath.Join(home, ".config")
				}
				configDir = filepath.Join(xdgConfigHome, "deepviz")
			}

			// Create new configuration
			config, err := NewViperConfig(configDir)
			if err != nil {
				return fmt.Errorf("failed to create config: %w", err)
			}

			// Set default values (XDG Base Directory compliant)
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

			config.Set("output_dir", defaultOutputDir)
			config.Set("api_key", "")
			config.Set("deep_research_agent", "deep-research-pro-preview-12-2025")
			config.Set("poll_interval", 10)
			config.Set("poll_timeout", 600)
			config.Set("model", "gemini-3-pro-image-preview")
			config.Set("aspect_ratio", "16:9")
			config.Set("image_size", "2K")
			config.Set("image_lang", "Japanese")
			config.Set("auto_open", true)

			// Save config file
			if err := config.Save(); err != nil {
				return fmt.Errorf("failed to save config file: %w", err)
			}

			configPath := filepath.Join(configDir, "config.yaml")
			fmt.Fprintf(cmd.OutOrStdout(), "Config file created: %s\n", configPath)
			return nil
		},
	}
	configInitCmd.Flags().StringVar(&configDir, "config-dir", "", "Configuration file directory")

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)

	return configCmd
}

// newCompletionCommand creates the shell completion command.
func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: `To load completions:

Bash:
  $ source <(deepviz completion bash)
  $ echo "source <(deepviz completion bash)" >> ~/.bashrc

Zsh:
  $ source <(deepviz completion zsh)
  $ echo "source <(deepviz completion zsh)" >> ~/.zshrc

Fish:
  $ deepviz completion fish | source
  $ deepviz completion fish > ~/.config/fish/completions/deepviz.fish

PowerShell:
  PS> deepviz completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	}
}

// maskAPIKey masks the API key.
func maskAPIKey(apiKey string) string {
	if apiKey == "" {
		return "(not set)"
	}
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
}

// RunWithConfig executes the main processing using the configuration.
func RunWithConfig(opts *Options, config *ViperConfig) error {
	// Create context
	ctx := context.Background()

	// Generate timestamp
	timestamp := GenerateTimestamp()

	// Ensure output directories exist
	if err := config.EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Create log file path with timestamp
	logFilePath := filepath.Join(config.LogsDir(), timestamp+".log")

	// Create logger
	logger := NewSlogLogger(opts.Verbose, logFilePath)

	// Get prompt (from file or direct)
	prompt := opts.Prompt
	if opts.File != "" {
		data, err := ReadFile(opts.File)
		if err != nil {
			return fmt.Errorf("failed to read prompt file: %w", err)
		}
		prompt = string(data)
		if prompt == "" {
			return fmt.Errorf("prompt file is empty: %s", opts.File)
		}
		logger.Info("Loaded prompt from file", "file", opts.File)
	}

	logger.Info("Pipeline started")
	logger.Info("Configuration", "timestamp", timestamp, "output_dir", config.OutputDir)

	var researchResult *ResearchResult
	var imageResult *ImageResult

	// Execute research (except ImageOnly mode)
	if !opts.ImageOnly {
		logger.Info("Starting Deep Research")

		researchClient, err := NewGenaiResearchClient(ctx, config, logger)
		if err != nil {
			return fmt.Errorf("failed to create research client: %w", err)
		}

		researchResult, err = researchClient.Execute(ctx, prompt, timestamp)
		if err != nil {
			return fmt.Errorf("failed to execute research: %w", err)
		}
		logger.Info("Deep Research completed")
	}

	// Execute image generation (except ResearchOnly mode)
	if !opts.ResearchOnly {
		logger.Info("Starting image generation")

		imageClient, err := NewGenaiImageClient(ctx, config, logger)
		if err != nil {
			return fmt.Errorf("failed to create image client: %w", err)
		}

		// Build prompt for image generation
		var imagePrompt string
		if researchResult != nil {
			// Generate infographics from research results
			imagePrompt = imageClient.BuildInfographicsPrompt(researchResult.Content)
		} else {
			// Use prompt template in ImageOnly mode
			imagePrompt = imageClient.BuildInfographicsPrompt(prompt)
		}

		// Image generation configuration
		imgConfig := ImageConfig{
			Model:       opts.Model,
			AspectRatio: opts.AspectRatio,
			ImageSize:   opts.ImageSize,
		}

		imageResult, err = imageClient.Generate(ctx, imagePrompt, imgConfig, timestamp)
		if err != nil {
			return fmt.Errorf("failed to generate image: %w", err)
		}
		logger.Info("Image generation completed", "image_path", imageResult.ImagePath)

		// Auto-open image if enabled (flag takes priority, then config)
		if !opts.NoOpen && config.AutoOpen {
			if err := OpenFile(imageResult.ImagePath); err != nil {
				logger.Info("Failed to open image", "error", err)
			}
		}
	}

	// Output results summary
	logger.Info("Pipeline completed")
	fmt.Println("\n=== Pipeline Completed ===")
	fmt.Printf("Timestamp: %s\n", timestamp)
	if researchResult != nil {
		fmt.Printf("Research: %s\n", researchResult.MarkdownPath)
	}
	if imageResult != nil {
		fmt.Printf("Image: %s\n", imageResult.ImagePath)
	}
	fmt.Printf("Output directory: %s\n", config.OutputDir)

	return nil
}
