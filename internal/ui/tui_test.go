package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func TestASCIIHeaderRendering(t *testing.T) {
	raw := ASCIIHeaderRaw()
	if !strings.Contains(raw, "▄▄██████▄▄") {
		t.Errorf("expected ASCIIHeaderRaw to contain coffee bun art, got %q", raw)
	}
	lines := strings.Split(raw, "\n")
	if len(lines) != 6 {
		t.Errorf("expected 6 lines in ASCIIHeaderRaw, got %d", len(lines))
	}

	styled := RenderASCIIHeader(80)
	styledLines := strings.Split(styled, "\n")
	if len(styledLines) != 6 {
		t.Errorf("expected 6 lines in RenderASCIIHeader, got %d", len(styledLines))
	}
	for i, line := range styledLines {
		if lipgloss.Width(line) > 80 {
			t.Errorf("line %d of RenderASCIIHeader(80) exceeded width 80: %d", i+1, lipgloss.Width(line))
		}
	}
}

func TestViewportLineBounds(t *testing.T) {
	testSizes := []struct {
		width  int
		height int
	}{
		{width: 80, height: 24},
		{width: 60, height: 15},
		{width: 100, height: 35},
		{width: 70, height: 20},
	}

	for _, sz := range testSizes {
		m := NewModel()
		m.state = stateTree
		m.width = sz.width
		m.height = sz.height

		view := m.View()
		lines := strings.Split(view, "\n")

		// View height must never exceed terminal height
		if len(lines) > sz.height {
			t.Errorf("At size %dx%d: rendered %d lines, exceeding height %d", sz.width, sz.height, len(lines), sz.height)
		}

		// No line should visually exceed terminal width
		for idx, line := range lines {
			w := lipgloss.Width(line)
			if w > sz.width {
				t.Errorf("At size %dx%d, line %d has visual width %d > %d: %q", sz.width, sz.height, idx+1, w, sz.width, line)
			}
		}
	}
}

func TestCursorScrollingBounds(t *testing.T) {
	m := NewModel()
	m.state = stateTree
	m.width = 80
	m.height = 20

	rows := m.getVisibleRows()
	if len(rows) < 10 {
		t.Fatalf("Expected at least 10 rows, got %d", len(rows))
	}

	// Move cursor down past initial viewport
	for i := 0; i < len(rows)-1; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(Model)
	}

	if m.cursor != len(rows)-1 {
		t.Errorf("Cursor expected at bottom (%d), got %d", len(rows)-1, m.cursor)
	}

	// Render view at bottom: should still strictly satisfy bounds
	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > m.height {
		t.Errorf("Rendered %d lines at bottom, exceeding height %d", len(lines), m.height)
	}
}

func TestVimNavigationShortcuts(t *testing.T) {
	m := NewModel()
	m.state = stateTree
	m.width = 80
	m.height = 24

	rows := m.getVisibleRows()
	if len(rows) < 10 {
		t.Fatalf("expected at least 10 rows, got %d", len(rows))
	}

	// Press 'G' -> jump to very last item
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = updated.(Model)
	if m.cursor != len(rows)-1 {
		t.Errorf("expected 'G' to move cursor to %d, got %d", len(rows)-1, m.cursor)
	}

	// Press 'g' -> jump to very first item
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("expected 'g' to move cursor to 0, got %d", m.cursor)
	}

	// Press 'J' (half-page down)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'J'}})
	m = updated.(Model)
	if m.cursor != 8 {
		t.Errorf("expected 'J' to jump cursor to 8, got %d", m.cursor)
	}

	// Press 'K' (half-page up)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'K'}})
	m = updated.(Model)
	if m.cursor != 0 {
		t.Errorf("expected 'K' to jump cursor back to 0, got %d", m.cursor)
	}

	// Press 'L', 'M', 'H' (viewport bottom, middle, top)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	m = updated.(Model)
	if m.cursor <= 0 {
		t.Errorf("expected 'L' to move cursor to bottom of viewport, got %d", m.cursor)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'M'}})
	m = updated.(Model)
	if m.cursor <= 0 {
		t.Errorf("expected 'M' to move cursor to middle of viewport, got %d", m.cursor)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'H'}})
	m = updated.(Model)
	if m.cursor != m.scrollOffset {
		t.Errorf("expected 'H' to move cursor to scrollOffset (%d), got %d", m.scrollOffset, m.cursor)
	}

	// Press 'c' (close all folds) and 'o' (open all folds)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = updated.(Model)
	for _, g := range m.groups {
		if !m.folded[g.Category] {
			t.Errorf("expected group %s to be folded after 'c'", g.Category)
		}
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	m = updated.(Model)
	for _, g := range m.groups {
		if m.folded[g.Category] {
			t.Errorf("expected group %s to be unfolded after 'o'", g.Category)
		}
	}

	// Press '?' to open help view, verify output, and close with '?'
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if !m.showHelp {
		t.Errorf("expected showHelp to be true after '?'")
	}
	helpOut := m.View()
	if !strings.Contains(helpOut, "Keyboard Shortcuts") {
		t.Errorf("expected help view title in output")
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = updated.(Model)
	if m.showHelp {
		t.Errorf("expected showHelp to be false after second '?'")
	}
}

func TestInstallingViewportLineBounds(t *testing.T) {
	testSizes := []struct {
		width  int
		height int
	}{
		{width: 80, height: 24},
		{width: 60, height: 16},
		{width: 100, height: 35},
	}

	for _, sz := range testSizes {
		m := NewModel()
		m.state = stateInstalling
		m.width = sz.width
		m.height = sz.height
		m.actionTitle = "Java (OpenJDK)"
		m.actionReason = "OpenJDK is installed via Homebrew, but macOS requires administrator (sudo) permission to register the JDK."
		m.actionSnippet = "sudo ln -sfn /opt/homebrew/opt/openjdk/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk.jdk"

		// Populate items with various statuses
		statuses := []model.InstallStatus{
			model.StatusInstalling,
			model.StatusSuccess,
			model.StatusFailed,
			model.StatusWaitingAction,
			model.StatusPending,
		}
		for i := 0; i < 10; i++ {
			id := string(rune('a' + i))
			m.installOrder = append(m.installOrder, id)
			m.installItems[id] = &installItemProgress{
				ID:            id,
				Name:          "Package " + id,
				Status:        statuses[i%len(statuses)],
				Message:       "Processing...",
				LogLine:       "==> Downloading files...",
				ActionSnippet: "sudo echo test",
			}
		}

		view := m.View()
		lines := strings.Split(view, "\n")

		if len(lines) > sz.height {
			t.Errorf("Installing view at %dx%d rendered %d lines > %d", sz.width, sz.height, len(lines), sz.height)
		}

		for idx, line := range lines {
			w := lipgloss.Width(line)
			if w > sz.width {
				t.Errorf("Installing view line %d width %d > %d: %q", idx+1, w, sz.width, line)
			}
		}
	}
}

func TestTwoButtonNavigation(t *testing.T) {
	m := NewModel()
	m.state = stateTree
	m.width = 80
	m.height = 24

	// Initially in tree
	if m.focus != focusTree {
		t.Fatalf("Expected focusTree, got %v", m.focus)
	}

	// Press Tab -> focus jumps to focusBtnNext (Right: Start Install)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focus != focusBtnNext {
		t.Fatalf("Expected focusBtnNext on Tab, got %v", m.focus)
	}

	// Press Shift+Tab -> focus moves to previous button focusBtnPrev (Left: Exit)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.focus != focusBtnPrev {
		t.Fatalf("Expected focusBtnPrev on Shift+Tab from focusBtnNext, got %v", m.focus)
	}

	// Press Shift+Tab again -> focus returns to focusTree
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.focus != focusTree {
		t.Fatalf("Expected focusTree on Shift+Tab from focusBtnPrev, got %v", m.focus)
	}

	// Press Tab -> focusBtnNext, then Left -> focusBtnPrev, then Right -> focusBtnNext
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(Model)
	if m.focus != focusBtnPrev {
		t.Fatalf("Expected focusBtnPrev on Left, got %v", m.focus)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(Model)
	if m.focus != focusBtnNext {
		t.Fatalf("Expected focusBtnNext on Right, got %v", m.focus)
	}

	// Press Up -> focus returns to focusTree
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(Model)
	if m.focus != focusTree {
		t.Fatalf("Expected focusTree on Up, got %v", m.focus)
	}

	// In stateComplete, Shift+Tab should toggle cleanly between focusBtnNext and focusBtnPrev
	m.state = stateComplete
	m.focus = focusBtnNext
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = updated.(Model)
	if m.focus != focusBtnPrev {
		t.Fatalf("Expected focusBtnPrev on Shift+Tab in stateComplete, got %v", m.focus)
	}
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = updated.(Model)
	if m.focus != focusBtnNext {
		t.Fatalf("Expected focusBtnNext on Tab in stateComplete, got %v", m.focus)
	}
}

func TestSelectionAndFoldingShortcuts(t *testing.T) {
	m := NewModel()
	m.state = stateTree
	m.width = 80
	m.height = 24

	// Mark first package as installed, second as installed+outdated, others missing
	if len(m.packages) > 1 {
		m.packages[0].Installed = true
		m.packages[0].Outdated = false
		m.packages[1].Installed = true
		m.packages[1].Outdated = true
		m.packages[1].LatestVersion = "2.0.0"
	}

	// Test 'm' (select missing and outdated)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	m = updated.(Model)
	if len(m.packages) > 1 {
		if m.selected[m.packages[0].ID] {
			t.Errorf("expected up-to-date installed package to not be selected by 'm'")
		}
		if !m.selected[m.packages[1].ID] {
			t.Errorf("expected outdated installed package to be selected by 'm'")
		}
	}

	// Test 'u' (select updates only)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = updated.(Model)
	if len(m.packages) > 1 && !m.selected[m.packages[1].ID] {
		t.Errorf("expected outdated package to be selected by 'u'")
	}

	// Test 'a' (toggle select all)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)
	for _, p := range m.packages {
		if !m.selected[p.ID] {
			t.Errorf("expected package %s to be selected after first 'a'", p.ID)
		}
	}

	// Second 'a' should deselect all
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(Model)
	for _, p := range m.packages {
		if m.selected[p.ID] {
			t.Errorf("expected package %s to be deselected after second 'a'", p.ID)
		}
	}

	// Test 'f' (toggle fold all)
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
	m = updated.(Model)
	for _, g := range m.groups {
		if !m.folded[g.Category] {
			t.Errorf("expected group %s to be folded after 'f'", g.Category)
		}
	}

	// Space on group row toggles group selection
	m.cursor = 0
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = updated.(Model)
	for _, id := range m.groups[0].PackageIDs {
		if !m.selected[id] {
			t.Errorf("expected package %s in group 0 to be selected on Space", id)
		}
	}
}

func TestScanningAndCompleteViews(t *testing.T) {
	m := NewModel()
	m.width = 80
	m.height = 24

	// Shimmer text across ticks
	for tick := 0; tick < 25; tick++ {
		m.shimmerTick = tick
		shimmer := RenderShimmerText("Installing selected packages...", tick)
		if shimmer == "" {
			t.Errorf("expected non-empty shimmer output at tick %d", tick)
		}
	}

	// Scanning view
	m.state = stateScanning
	scanOut := m.View()
	if !strings.Contains(scanOut, "available updates") {
		t.Errorf("expected scanning view text, got %q", scanOut)
	}

	// Complete view when 0 items were run (all up to date)
	m.state = stateComplete
	m.installOrder = nil
	zeroCompOut := m.View()
	if !strings.Contains(zeroCompOut, "Everything is already installed and up to date") {
		t.Errorf("expected friendly up-to-date message when 0 items run, got %q", zeroCompOut)
	}

	// Complete view with items
	m.installOrder = []string{"git", "fail-pkg"}
	m.installItems["git"] = &installItemProgress{
		ID:      "git",
		Name:    "Git",
		Status:  model.StatusSuccess,
		Message: "Installed",
	}
	m.installItems["fail-pkg"] = &installItemProgress{
		ID:      "fail-pkg",
		Name:    "Fail Pkg",
		Status:  model.StatusFailed,
		Message: "Error occurred",
	}
	compOut := m.View()
	if !strings.Contains(compOut, "Completed:") {
		t.Errorf("expected complete view summary, got %q", compOut)
	}

	// Quitting view
	m.quitting = true
	quitOut := m.View()
	if !strings.Contains(quitOut, "Thanks for using Coffee Bun") {
		t.Errorf("expected quitting message, got %q", quitOut)
	}
}

