// Package cli implements the command-line interface for try.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tobi/try/internal/theme"
	"github.com/tobi/try/internal/workspace"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"

	// Global flags
	triesPath  string
	themeName  string
	noColors   bool
)

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "try",
	Short: "Ephemeral workspace manager",
	Long: `try - fresh directories for every vibe

An ephemeral workspace manager that helps organize project directories
with date-prefixed naming and instant fuzzy navigation.

To use try, add to your shell config:

  # bash/zsh (~/.bashrc or ~/.zshrc)
  eval "$(try init)"

  # fish (~/.config/fish/config.fish)
  eval (try init | string collect)`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// No args: show help
		cmd.Help()
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&triesPath, "path", "", 
		fmt.Sprintf("tries directory (default: %s)", workspace.DefaultPath()))
	rootCmd.PersistentFlags().StringVar(&themeName, "theme", "default",
		fmt.Sprintf("color theme (%v)", theme.Names()))
	rootCmd.PersistentFlags().BoolVar(&noColors, "no-colors", false,
		"disable colors")

	// Hide help command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

func initConfig() {
	// Set tries path from flag or default
	if triesPath == "" {
		triesPath = workspace.DefaultPath()
	}

	// Handle NO_COLOR env var
	if os.Getenv("NO_COLOR") != "" {
		noColors = true
	}
}

// getTriesPath returns the configured tries path.
func getTriesPath() string {
	return triesPath
}

// getTheme returns the configured theme.
func getTheme() theme.Theme {
	return theme.Get(themeName)
}
