# go-try

A Go port of [tobi/try](https://github.com/tobi/try) - the ephemeral workspace manager for developers who constantly create little experiments.

## Why a Go port?

The original `try` is a single Ruby file that just works. This port exists as an experiment itself - reimplementing the core features using Go and the [Charm](https://charm.sh) ecosystem (Bubble Tea, Lip Gloss, Bubbles).

Key differences from the original:
- Single static binary (no Ruby runtime needed)
- Built with Bubble Tea TUI framework
- Theme support (default, dracula, nord, monochrome)

Features intentionally omitted:
- Worktree support
- Rename functionality
- Fuzzy match score display

## Installation

### Build from source

Requires Go 1.21+

```bash
go install github.com/tobi/try@latest
```

Or clone and build:

```bash
git clone https://github.com/tobi/try
cd try
go build -o go-try .
```

### Shell integration

Add to your `.zshrc` or `.bashrc`:

```bash
eval "$(go-try init)"
```

For Fish shell, add to `config.fish`:

```fish
go-try init | source
```

This creates a `try` shell function that wraps the TUI.

## Usage

```bash
try                    # Browse all experiment directories
try redis              # Filter to "redis" or create new
try --path ~/projects  # Use a different base directory
try --theme dracula    # Use dracula color theme
```

### Keyboard shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate |
| `Enter` | Select directory (or create if typing new name) |
| `Ctrl+N` | Create new directory with current filter text |
| `Ctrl+D` | Delete selected directory (with confirmation) |
| `/` | Start filtering |
| `Esc` | Cancel / exit filter mode |
| `?` | Toggle help |

### Creating directories

Type a name and press Enter (or Ctrl+N). New directories are automatically prefixed with today's date:

```
typing: redis-test
creates: 2025-01-19-redis-test
```

### Cloning repositories

Paste a Git SSH URL to clone directly:

```bash
try git@github.com:user/repo.git
# Creates: 2025-01-19-user-repo
```

### Deleting directories

Press `Ctrl+D` on any directory. A confirmation bar appears at the top - type `YES` and press Enter to confirm.

## Configuration

### Environment variables

- `TRY_PATH` - Base directory for experiments (default: `~/src/tries`)

### Command-line flags

```
--path, -p     Base directory for experiments
--theme, -t    Color theme: default, dracula, nord, monochrome
--no-colors    Disable colors
--version      Show version
--help         Show help
```

## Themes

Use `--theme` to change the color scheme:

```bash
try --theme dracula
try --theme nord
try --theme monochrome
```

Or set a default in your shell config:

```bash
eval "$(go-try init --theme dracula)"
```

## How it works

The `try` shell function captures the TUI's stdout, which outputs shell commands to execute (cd, mkdir, git clone, rm). The TUI itself renders to `/dev/tty` directly, allowing it to work even when stdout is captured.

## Credits

Original [try](https://github.com/tobi/try) by Tobi Lutke - a single-file Ruby script that inspired this port.

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components

## License

MIT
