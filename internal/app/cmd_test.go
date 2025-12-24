package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommand_Execute(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "help flag",
			args:    []string{"--help"},
			wantErr: false,
		},
		{
			name:    "version flag",
			args:    []string{"--version"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCommand()
			cmd.SetArgs(tt.args)

			// 出力をキャプチャ
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateCommand_Flags(t *testing.T) {
	cmd := NewRootCommand()

	// Verify flags are defined
	flags := []string{"prompt", "file", "output", "verbose", "no-image"}
	for _, flagName := range flags {
		flag := cmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("flag --%s should be defined", flagName)
		}
	}
}

func TestGenerateCommand_PromptRequired(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Set environment variables
	os.Setenv("GEMINI_API_KEY", "test-api-key")
	os.Setenv("GEMINI_OUTPUT_DIR", tmpDir)
	defer func() {
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("GEMINI_OUTPUT_DIR")
	}()

	cmd := NewRootCommand()
	cmd.SetArgs([]string{})

	// Capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	// Should error when neither prompt nor file is specified
	if err == nil {
		t.Error("Execute() should return error when neither prompt nor file is specified")
	}
}

func TestConfigCommand_Show(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	// Set environment variable
	os.Setenv("GEMINI_OUTPUT_DIR", tmpDir)
	defer os.Unsetenv("GEMINI_OUTPUT_DIR")

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"config", "show"})

	// Capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	output := buf.String()
	// Verify config is displayed
	if !strings.Contains(output, "output_dir") {
		t.Error("output should contain 'output_dir'")
	}
}

func TestConfigCommand_Init(t *testing.T) {
	// Temporary directory for testing
	tmpDir := t.TempDir()

	cmd := NewRootCommand()
	cmd.SetArgs([]string{"config", "init", "--config-dir", tmpDir})

	// Capture output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file should be created")
	}
}
