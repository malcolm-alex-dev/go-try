package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tobi/try/internal/shell"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Output shell function for integration",
	Long: `Output a shell function definition that wraps 'try exec'.

Add to your shell config:

  # bash/zsh (~/.bashrc or ~/.zshrc)
  eval "$(try init)"

  # fish (~/.config/fish/config.fish)
  eval (try init | string collect)

Optionally specify a custom tries directory:

  eval "$(try init ~/code/experiments)"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get the path to the try binary
	scriptPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	scriptPath, err = filepath.EvalSymlinks(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Get tries path from arg or flag
	tryPath := ""
	if len(args) > 0 {
		tryPath = args[0]
	} else if triesPath != "" && triesPath != getTriesPath() {
		// Only include if explicitly set via flag
		tryPath = triesPath
	}

	// Detect shell
	shellType := detectShell()

	var script string
	if shellType == "fish" {
		script = shell.InitFish(scriptPath, tryPath)
	} else {
		script = shell.InitBash(scriptPath, tryPath)
	}

	fmt.Print(script)
	return nil
}

func detectShell() string {
	// Check SHELL env var first
	shellEnv := os.Getenv("SHELL")
	if strings.Contains(shellEnv, "fish") {
		return "fish"
	}

	// Could also check parent process, but SHELL is usually sufficient
	return "bash"
}
