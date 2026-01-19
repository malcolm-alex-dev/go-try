package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/tobi/try/internal/shell"
	"github.com/tobi/try/internal/tui"
	"github.com/tobi/try/internal/workspace"
)

var execCmd = &cobra.Command{
	Use:   "exec [query]",
	Short: "Run selector and output shell script",
	Long: `Run the interactive selector and output a shell script to stdout.

This command is typically called via the shell wrapper function created by 'try init'.
The output is meant to be eval'd by the shell.

If a git URL is provided instead of a query, it will clone the repository.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExec,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExec(cmd *cobra.Command, args []string) error {
	basePath := getTriesPath()

	// Ensure tries directory exists
	if err := workspace.EnsureDir(basePath); err != nil {
		return fmt.Errorf("failed to create tries directory: %w", err)
	}

	// Check if arg is a git URL
	if len(args) > 0 && workspace.IsGitURL(args[0]) {
		return handleClone(basePath, args[0])
	}

	// Run interactive selector
	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	return runSelector(basePath, query)
}

func runSelector(basePath, query string) error {
	// Create TUI model
	opts := []tui.Option{
		tui.WithTheme(getTheme()),
	}
	if query != "" {
		opts = append(opts, tui.WithInitialQuery(query))
	}

	m := tui.New(basePath, opts...)

	// Run Bubble Tea program
	// Open /dev/tty directly for TUI rendering to ensure it works
	// even when stdout is captured by the shell wrapper
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer tty.Close()

	// Force lipgloss to use colors since stdout may not be a TTY
	// when run through the shell wrapper (stdout is captured)
	lipgloss.DefaultRenderer().SetColorProfile(termenv.TrueColor)

	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithInput(tty),
		tea.WithOutput(tty),
	)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Get the action from the final model
	model, ok := finalModel.(*tui.Model)
	if !ok || model == nil {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		os.Exit(1)
	}
	if model.GetError() != nil {
		return model.GetError()
	}

	action := model.GetAction()
	if action == nil {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		os.Exit(1)
	}

	// Output the appropriate shell script
	return outputScript(action, basePath)
}

func outputScript(action *tui.Action, basePath string) error {
	var script string

	switch action.Type {
	case tui.ActionCD:
		// Touch to update mtime, then cd
		script = shell.CD(action.Path)

	case tui.ActionCreate:
		// Create new directory with date prefix
		path, err := workspace.Create(basePath, action.Path)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		script = shell.MkdirCD(path)

	case tui.ActionClone:
		script = shell.Clone(action.Path, action.URL)

	case tui.ActionDelete:
		script = shell.Delete(action.Paths, basePath)

	case tui.ActionCancel:
		fmt.Fprintln(os.Stderr, "Cancelled.")
		os.Exit(1)

	default:
		fmt.Fprintln(os.Stderr, "Cancelled.")
		os.Exit(1)
	}

	fmt.Print(script)
	return nil
}

func handleClone(basePath, url string) error {
	path, cloneURL, err := workspace.CloneScript(basePath, url)
	if err != nil {
		return fmt.Errorf("failed to parse git URL: %w", err)
	}

	script := shell.Clone(path, cloneURL)
	fmt.Print(script)
	return nil
}
