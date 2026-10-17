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

var (
	shimmerBaseStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#A67B5B"))
	shimmerEdgeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D4A373")).Bold(true)
	shimmerMidStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5EBE0")).Bold(true)
	shimmerPeakStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFDF9")).Bold(true)
)

// RenderShimmerText renders text with an animated left-to-right repeating highlight window.
// The window width is computed as max(minFixedWidth, proportionalWidth) so it scales cleanly.
func RenderShimmerText(text string, tick int) string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return ""
	}

	const minFixedWidth = 6
	propWidth := n / 4
	winWidth := minFixedWidth
	if propWidth > winWidth {
		winWidth = propWidth
	}

	if tick < 0 {
		tick = -tick
	}

	totalSpan := n + winWidth
	winStart := (tick % totalSpan) - (winWidth / 2)
	winEnd := winStart + winWidth
	winCenter := winStart + (winWidth / 2)
	coreRadius := winWidth / 6
	if coreRadius < 1 {
		coreRadius = 1
	}
	midRadius := winWidth / 3
	if midRadius < 2 {
		midRadius = 2
	}

	zoneFor := func(i int) int {
		if i >= winStart && i < winEnd {
			dist := i - winCenter
			if dist < 0 {
				dist = -dist
			}
			if dist <= coreRadius {
				return 3 // Peak
			}
			if dist <= midRadius {
				return 2 // Mid
			}
			return 1 // Edge
		}
		return 0 // Base
	}

	renderChunk := func(zone int, chunk string) string {
		switch zone {
		case 3:
			return shimmerPeakStyle.Render(chunk)
		case 2:
			return shimmerMidStyle.Render(chunk)
		case 1:
			return shimmerEdgeStyle.Render(chunk)
		default:
			return shimmerBaseStyle.Render(chunk)
		}
	}

	var b strings.Builder
	curZone := zoneFor(0)
	startIdx := 0
	for i := 1; i < n; i++ {
		z := zoneFor(i)
		if z != curZone {
			b.WriteString(renderChunk(curZone, string(runes[startIdx:i])))
			curZone = z
			startIdx = i
		}
	}
	b.WriteString(renderChunk(curZone, string(runes[startIdx:n])))
	return b.String()
}


