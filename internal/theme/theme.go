// Package theme provides configurable color themes for the TUI.
package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme defines the color scheme for the TUI.
type Theme struct {
	// Primary colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color

	// Text colors
	Text       lipgloss.Color
	TextDim    lipgloss.Color
	TextMuted  lipgloss.Color
	Highlight  lipgloss.Color

	// Background colors
	Background         lipgloss.Color
	BackgroundSelected lipgloss.Color
	BackgroundDanger   lipgloss.Color

	// Status colors
	Success lipgloss.Color
	Warning lipgloss.Color
	Error   lipgloss.Color
}

// Default is the default theme, inspired by Catppuccin Mocha.
var Default = Theme{
	Primary:   lipgloss.Color("183"), // Lavender
	Secondary: lipgloss.Color("117"), // Sky
	Accent:    lipgloss.Color("214"), // Peach/Orange

	Text:       lipgloss.Color("255"), // White
	TextDim:    lipgloss.Color("245"), // Gray
	TextMuted:  lipgloss.Color("240"), // Darker gray
	Highlight:  lipgloss.Color("226"), // Yellow

	Background:         lipgloss.Color(""),    // Terminal default
	BackgroundSelected: lipgloss.Color("238"), // Dark gray
	BackgroundDanger:   lipgloss.Color("52"),  // Dark red

	Success: lipgloss.Color("114"), // Green
	Warning: lipgloss.Color("214"), // Orange
	Error:   lipgloss.Color("203"), // Red
}

// Dracula theme - purple/pink accent with darker backgrounds
var Dracula = Theme{
	Primary:   lipgloss.Color("141"), // Purple
	Secondary: lipgloss.Color("139"), // Cyan
	Accent:    lipgloss.Color("212"), // Pink

	Text:       lipgloss.Color("255"),
	TextDim:    lipgloss.Color("103"), // Comment color
	TextMuted:  lipgloss.Color("60"),
	Highlight:  lipgloss.Color("228"), // Yellow

	Background:         lipgloss.Color(""),
	BackgroundSelected: lipgloss.Color("53"),  // Purple-ish selection
	BackgroundDanger:   lipgloss.Color("124"), // Brighter red

	Success: lipgloss.Color("84"),  // Green
	Warning: lipgloss.Color("215"), // Orange
	Error:   lipgloss.Color("203"), // Red
}

// Nord theme - blue frost accent
var Nord = Theme{
	Primary:   lipgloss.Color("111"), // Frost blue
	Secondary: lipgloss.Color("109"), // Frost teal
	Accent:    lipgloss.Color("110"), // Frost light blue

	Text:       lipgloss.Color("255"),
	TextDim:    lipgloss.Color("246"),
	TextMuted:  lipgloss.Color("242"),
	Highlight:  lipgloss.Color("229"), // Aurora yellow

	Background:         lipgloss.Color(""),
	BackgroundSelected: lipgloss.Color("24"),  // Nord blue selection
	BackgroundDanger:   lipgloss.Color("131"), // Aurora red bg

	Success: lipgloss.Color("108"), // Aurora green
	Warning: lipgloss.Color("173"), // Aurora orange
	Error:   lipgloss.Color("167"), // Aurora red
}

// Monochrome theme for minimal distraction
var Monochrome = Theme{
	Primary:   lipgloss.Color("250"),
	Secondary: lipgloss.Color("245"),
	Accent:    lipgloss.Color("255"),

	Text:       lipgloss.Color("255"),
	TextDim:    lipgloss.Color("245"),
	TextMuted:  lipgloss.Color("240"),
	Highlight:  lipgloss.Color("255"),

	Background:         lipgloss.Color(""),
	BackgroundSelected: lipgloss.Color("238"),
	BackgroundDanger:   lipgloss.Color("52"),

	Success: lipgloss.Color("255"),
	Warning: lipgloss.Color("250"),
	Error:   lipgloss.Color("245"),
}

// Available themes by name
var Themes = map[string]Theme{
	"default":    Default,
	"dracula":    Dracula,
	"nord":       Nord,
	"monochrome": Monochrome,
}

// Get returns a theme by name, falling back to Default if not found.
func Get(name string) Theme {
	if t, ok := Themes[name]; ok {
		return t
	}
	return Default
}

// Names returns all available theme names.
func Names() []string {
	names := make([]string, 0, len(Themes))
	for name := range Themes {
		names = append(names, name)
	}
	return names
}
