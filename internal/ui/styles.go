package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

var (
	// Soft Coffee Pastel Palette
	ColorPrimary   = lipgloss.Color("#A67B5B") // Warm Café au Lait / Mocha Pastel
	ColorSecondary = lipgloss.Color("#D4A373") // Soft Caramel / Latte Pastel
	ColorSuccess   = lipgloss.Color("#84A98C") // Soft Matcha / Sage Pastel Green
	ColorWarning   = lipgloss.Color("#E9C46A") // Warm Honey / Biscuit Pastel
	ColorDanger    = lipgloss.Color("#E07A5F") // Soft Terracotta / Rose Pastel
	ColorMuted     = lipgloss.Color("#9C8F84") // Warm Taupe / Coffee Gray
	ColorHighlight = lipgloss.Color("#E6CCB2") // Soft Vanilla / Oat Milk Cream

	// ASCII Logo Styles
	LogoBunStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C68B59")).
			Bold(true)

	LogoTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EDE0D4")).
			Bold(true)

	// Text & Container Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F5EBE0")).
			Background(lipgloss.Color("#4A3525")).
			Padding(0, 1)

	SubTitleStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	CategoryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorSecondary)

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EDE0D4"))

	MutedStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	BadgeSuccess = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	BadgeMissing = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	BadgeError = lipgloss.NewStyle().
			Foreground(ColorDanger).
			Bold(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	// Interactive Button Styles
	ButtonNormal = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D6CCC2")).
			Background(lipgloss.Color("#2C221E")).
			Padding(0, 1).
			MarginRight(1)

	ButtonFocused = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF8F0")).
			Background(ColorPrimary).
			Padding(0, 1).
			MarginRight(1)

	ButtonSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFF8F0")).
			Background(lipgloss.Color("#6B9080")).
			Padding(0, 1).
			MarginRight(1)

	// Live Stream / Output Line Style (Warm Muted Taupe)
	LogStreamStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A7968"))
)

var asciiBunLines = []struct {
	icon string
	text string
}{
	{"      ▄▄██████▄▄         ", "▄▀▀▀ ▄▀▀▄ █▀▀▀ █▀▀▀ █▀▀▀ █▀▀▀"},
	{"   ▄██▀▀▀▀▀▀▀▀▀▀██▄      ", "█    █  █ █▀▀  █▀▀  █▀▀  █▀▀ "},
	{"  ██▀ ▄▄██████▄▄ ▀██     ", "▀▄▄▄ ▀▄▄▀ ▀    ▀    ▀▄▄▄ ▀▄▄▄"},
	{" ██▀ ██▀▀▀▀▀▀▀▀██ ▀██    ", "█▀▀▄ █  █ █▄ █"},
	{"▄██▄▄████████████▄▄██▄   ", "█▀▀▄ █  █ █ ▀█"},
	{" ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀    ", "▀▀▀   ▀▀  ▀  ▀"},
}

// ASCIIHeaderRaw returns the unstyled ASCII art banner for CLI help output
func ASCIIHeaderRaw() string {
	var b strings.Builder
	for _, l := range asciiBunLines {
		b.WriteString("  " + l.icon + l.text + "\n")
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// RenderASCIIHeader renders the colored 6-line ASCII art banner truncated to width
func RenderASCIIHeader(width int) string {
	var b strings.Builder
	for i, l := range asciiBunLines {
		line := "  " + LogoBunStyle.Render(l.icon) + LogoTextStyle.Render(l.text)
		if width > 0 {
			line = ansi.Truncate(line, width-1, "")
		}
		b.WriteString(line)
		if i < len(asciiBunLines)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}
