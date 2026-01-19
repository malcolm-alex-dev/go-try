package tui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tobi/try/internal/theme"
	"github.com/tobi/try/internal/workspace"
)

// State represents the current UI state.
type State int

const (
	StateSelector State = iota
	StateDeleteConfirm
)

// Action represents the result of a TUI session.
type Action struct {
	Type    ActionType
	Path    string   // For CD, Create, Clone
	URL     string   // For Clone
	Paths   []string // For Delete
	BaseDir string   // Base directory for operations
}

// ActionType represents the type of action selected.
type ActionType int

const (
	ActionNone ActionType = iota
	ActionCD
	ActionCreate
	ActionClone
	ActionDelete
	ActionCancel
)

// item implements list.Item for directory entries.
type item struct {
	entry workspace.Entry
}

func (i item) FilterValue() string { return i.entry.Name }
func (i item) Title() string       { return i.entry.Name }
func (i item) Description() string { return formatRelativeTime(i.entry.ModTime) }

// Model is the main TUI model.
type Model struct {
	// Configuration
	basePath     string
	initialQuery string
	theme        theme.Theme

	// State
	state   State
	list    list.Model
	entries []workspace.Entry
	width   int
	height  int

	// Delete confirmation
	deleteTarget  string // path of item to delete
	deleteConfirm string // user's typed confirmation

	// Result
	action *Action
	err    error
}

// itemDelegate handles rendering of list items.
type itemDelegate struct {
	styles *delegateStyles
}

type delegateStyles struct {
	normal   lipgloss.Style
	selected lipgloss.Style
	dimmed   lipgloss.Style
	desc     lipgloss.Style
}

func newDelegateStyles(t theme.Theme) *delegateStyles {
	return &delegateStyles{
		normal: lipgloss.NewStyle().
			Padding(0, 0, 0, 2),
		selected: lipgloss.NewStyle().
			Background(t.BackgroundSelected).
			Foreground(t.Text).
			Padding(0, 0, 0, 2),
		dimmed: lipgloss.NewStyle().
			Foreground(t.TextDim),
		desc: lipgloss.NewStyle().
			Foreground(t.TextMuted),
	}
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// For selected rows, don't use inner styles - just plain text
	// The row style will handle the background uniformly
	var name, meta string
	timeAgo := formatRelativeTime(i.entry.ModTime)

	if isSelected {
		// Plain text - row style handles background
		name = i.entry.Name
		meta = timeAgo
	} else {
		// Normal row - apply dim styling to date prefix and meta
		name = d.renderNameWithDim(i.entry.Name)
		meta = d.styles.desc.Render(timeAgo)
	}

	// Calculate spacing - fill entire row width
	nameWidth := lipgloss.Width(name)
	metaWidth := lipgloss.Width(meta)
	availableWidth := m.Width() - 4 // account for padding

	var line string
	if nameWidth+metaWidth+2 <= availableWidth {
		spacing := availableWidth - nameWidth - metaWidth
		if spacing < 0 {
			spacing = 0
		}
		line = fmt.Sprintf("%s%s%s", name, strings.Repeat(" ", spacing), meta)
	} else {
		// Fill remaining space to ensure full-width highlight
		spacing := availableWidth - nameWidth
		if spacing < 0 {
			spacing = 0
		}
		line = name + strings.Repeat(" ", spacing)
	}

	// Apply row style with full width
	var rowStyle lipgloss.Style
	if isSelected {
		rowStyle = d.styles.selected
	} else {
		rowStyle = d.styles.normal
	}

	fmt.Fprint(w, rowStyle.Width(m.Width()).Render(line))
}

func (d itemDelegate) renderNameWithDim(name string) string {
	// Check if name has date prefix (YYYY-MM-DD-)
	if len(name) > 11 && name[4] == '-' && name[7] == '-' && name[10] == '-' {
		dateStr := name[:11] // includes trailing dash
		rest := name[11:]
		return d.styles.dimmed.Render(dateStr) + rest
	}
	return name
}

// New creates a new TUI model.
func New(basePath string, opts ...Option) *Model {
	m := &Model{
		basePath: basePath,
		theme:    theme.Default,
		state:    StateSelector,
	}

	for _, opt := range opts {
		opt(m)
	}

	// Create delegate with theme
	delegate := itemDelegate{
		styles: newDelegateStyles(m.theme),
	}

	// Create list with empty items (will be populated in Init)
	m.list = list.New([]list.Item{}, delegate, 0, 0)
	m.list.Title = IconHome + " Try"
	m.list.SetShowStatusBar(true)
	m.list.SetFilteringEnabled(true)
	m.list.SetShowHelp(true)
	m.list.DisableQuitKeybindings()

	// Customize list styles
	m.list.Styles.Title = lipgloss.NewStyle().
		Foreground(m.theme.Accent).
		Bold(true).
		Padding(0, 1)

	m.list.Styles.FilterPrompt = lipgloss.NewStyle().
		Foreground(m.theme.Primary)

	m.list.Styles.FilterCursor = lipgloss.NewStyle().
		Foreground(m.theme.Highlight)

	// Disable default quit key
	m.list.KeyMap.Quit = key.NewBinding(key.WithDisabled())

	// Add custom key bindings to help
	m.list.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("ctrl+d"),
				key.WithHelp("ctrl+d", "delete"),
			),
			key.NewBinding(
				key.WithKeys("ctrl+n"),
				key.WithHelp("ctrl+n", "new"),
			),
		}
	}
	m.list.AdditionalFullHelpKeys = m.list.AdditionalShortHelpKeys

	return m
}

// Option is a functional option for configuring the model.
type Option func(*Model)

// WithTheme sets the color theme.
func WithTheme(t theme.Theme) Option {
	return func(m *Model) {
		m.theme = t
	}
}

// WithInitialQuery sets the initial search query.
func WithInitialQuery(q string) Option {
	return func(m *Model) {
		m.initialQuery = strings.ReplaceAll(q, " ", "-")
	}
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return m.loadEntries
}

func (m *Model) loadEntries() tea.Msg {
	entries, err := workspace.Scan(m.basePath)
	if err != nil {
		return errMsg{err}
	}
	return entriesLoadedMsg{entries}
}

type entriesLoadedMsg struct {
	entries []workspace.Entry
}

type errMsg struct {
	err error
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := lipgloss.NewStyle().Padding(1, 2).GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case entriesLoadedMsg:
		m.entries = msg.entries
		items := make([]list.Item, len(msg.entries))
		for i, e := range msg.entries {
			items[i] = item{entry: e}
		}
		m.list.SetItems(items)
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle delete confirmation state
	if m.state == StateDeleteConfirm {
		return m.handleDeleteConfirmKey(msg)
	}

	switch msg.String() {
	case "ctrl+c":
		m.action = &Action{Type: ActionCancel}
		return m, tea.Quit

	case "esc":
		// Let list handle esc (exits filter mode)
		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		m.action = &Action{Type: ActionCancel}
		return m, tea.Quit

	case "enter":
		return m.handleSelect()

	case "ctrl+d":
		return m.handleDelete()

	case "ctrl+n":
		// Create new with current filter text
		return m.handleCreateNew()
	}

	// Pass to list for filtering/navigation
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) handleSelect() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem()
	if selected == nil {
		// No selection - maybe create new?
		filterVal := m.list.FilterValue()
		if filterVal != "" {
			m.action = &Action{
				Type:    ActionCreate,
				Path:    filterVal,
				BaseDir: m.basePath,
			}
			return m, tea.Quit
		}
		return m, nil
	}

	i := selected.(item)
	m.action = &Action{
		Type:    ActionCD,
		Path:    i.entry.Path,
		BaseDir: m.basePath,
	}

	return m, tea.Quit
}

func (m *Model) handleCreateNew() (tea.Model, tea.Cmd) {
	filterValue := m.list.FilterValue()
	if filterValue == "" {
		return m, nil
	}

	m.action = &Action{
		Type:    ActionCreate,
		Path:    filterValue,
		BaseDir: m.basePath,
	}
	return m, tea.Quit
}

func (m *Model) handleDelete() (tea.Model, tea.Cmd) {
	selected := m.list.SelectedItem()
	if selected == nil {
		return m, nil
	}

	i := selected.(item)
	m.deleteTarget = i.entry.Path
	m.deleteConfirm = ""
	m.state = StateDeleteConfirm

	return m, nil
}

func (m *Model) handleDeleteConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.state = StateSelector
		m.deleteTarget = ""
		m.deleteConfirm = ""
		return m, nil

	case tea.KeyEnter:
		if m.deleteConfirm == "YES" {
			m.action = &Action{
				Type:    ActionDelete,
				Paths:   []string{m.deleteTarget},
				BaseDir: m.basePath,
			}
			return m, tea.Quit
		}
		// Wrong confirmation, go back
		m.state = StateSelector
		m.deleteTarget = ""
		m.deleteConfirm = ""
		return m, nil

	case tea.KeyBackspace:
		if len(m.deleteConfirm) > 0 {
			m.deleteConfirm = m.deleteConfirm[:len(m.deleteConfirm)-1]
		}
		return m, nil

	case tea.KeyRunes:
		m.deleteConfirm += string(msg.Runes)
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Delete confirmation bar at top
	if m.state == StateDeleteConfirm {
		bar := m.viewDeleteBar()
		return bar + "\n" + m.list.View()
	}

	return m.list.View()
}

func (m *Model) viewDeleteBar() string {
	name := filepath.Base(m.deleteTarget)

	// Build plain text content - bar style handles all formatting
	content := fmt.Sprintf("%s DELETE %s  Type YES: %s█  (esc to cancel)", IconTrash, name, m.deleteConfirm)

	// Full-width bar with danger background
	bar := lipgloss.NewStyle().
		Background(m.theme.BackgroundDanger).
		Foreground(m.theme.Text).
		Bold(true).
		Width(m.width).
		Padding(0, 1).
		Render(content)

	return bar
}

// GetAction returns the selected action after the TUI exits.
func (m *Model) GetAction() *Action {
	return m.action
}

// GetError returns any error that occurred.
func (m *Model) GetError() error {
	return m.err
}

func formatRelativeTime(t time.Time) string {
	d := time.Since(t)

	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	if d < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
	return fmt.Sprintf("%dw ago", int(d.Hours()/24/7))
}
