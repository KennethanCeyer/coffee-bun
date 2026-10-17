package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func (m Model) View() string {
	if m.quitting {
		return "\n  Thanks for using Coffee Bun!\n\n"
	}

	if m.width <= 0 {
		m.width = 80
	}
	if m.height <= 0 {
		m.height = 24
	}

	if m.showHelp {
		return m.renderHelpView()
	}

	switch m.state {
	case stateScanning:
		return m.renderScanningView()
	case stateTree:
		return m.renderTreeView()
	case stateInstalling:
		return m.renderInstallingView()
	case stateComplete:
		return m.renderCompleteView()
	default:
		return ""
	}
}

func renderDivider(width int, indicator string) string {
	if width <= 0 {
		width = 80
	}
	if indicator == "" {
		return MutedStyle.Render(strings.Repeat("-", width-1))
	}
	indLen := lipgloss.Width(indicator)
	dashCount := width - indLen - 4
	if dashCount < 2 {
		dashCount = 2
	}
	left := strings.Repeat("-", 2)
	right := strings.Repeat("-", dashCount)
	return MutedStyle.Render(left) + SelectedStyle.Render(indicator) + MutedStyle.Render(right)
}

func (m Model) renderScanningView() string {
	var b strings.Builder
	if m.height >= 14 {
		b.WriteString(RenderASCIIHeader(m.width) + "\n\n")
	} else {
		title := TitleStyle.Render(" Coffee Bun ")
		b.WriteString(ansi.Truncate(title, m.width-1, "") + "\n\n")
	}

	spinner := spinnerFrames[m.spinnerIdx]
	line := fmt.Sprintf("  [%s] %s",
		SelectedStyle.Render(spinner),
		NormalStyle.Render("Checking environment and installed tools..."),
	)
	b.WriteString(ansi.Truncate(line, m.width-1, "") + "\n\n")
	b.WriteString(ansi.Truncate(MutedStyle.Render("  Please wait while components are verified..."), m.width-1, "") + "\n")
	return b.String()
}

func (m Model) renderHelpView() string {
	var b strings.Builder

	if m.height >= 24 {
		b.WriteString(RenderASCIIHeader(m.width) + "\n")
	} else {
		title := TitleStyle.Render(" Coffee Bun ")
		b.WriteString(ansi.Truncate(title, m.width-1, "") + "\n")
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, "[ Keyboard Shortcuts ]"), m.width-1, "") + "\n")

	helpRows := []struct {
		keys string
		desc string
	}{
		{"j / Down, k / Up", "Move cursor down / up one item (or scroll in install view)"},
		{"g / Home", "Jump to the very first item (top of list)"},
		{"G / End", "Jump to the very last item (bottom of list)"},
		{"Ctrl+d / J / PgDn", "Jump half-page down"},
		{"Ctrl+u / K / PgUp", "Jump half-page up"},
		{"H / M / L", "Jump to Top / Middle / Bottom of current viewport"},
		{"h / Left", "Fold category group (or collapse parent group from item)"},
		{"l / Right", "Unfold category group"},
		{"f / z", "Toggle fold/unfold for all category groups"},
		{"o / c", "Open (unfold) all groups / Close (fold) all groups"},
		{"Space", "Toggle selection of current item or entire category group"},
		{"Enter", "Toggle item selection, fold/unfold group, or activate button"},
		{"a", "Toggle select all / deselect all packages"},
		{"m", "Select only missing (uninstalled) packages"},
		{"r", "Refresh and re-scan environment status"},
		{"Tab / Shift+Tab", "Move focus between package tree and bottom action buttons"},
		{"?", "Open / close this keyboard shortcuts guide"},
		{"q / Esc / Ctrl+c", "Quit Coffee Bun"},
	}

	headerUsed := 2
	if m.height >= 24 {
		headerUsed = 7
	}
	availRows := m.height - headerUsed - 2
	if availRows < 4 {
		availRows = 4
	}

	for i := 0; i < len(helpRows) && i < availRows; i++ {
		hr := helpRows[i]
		line := fmt.Sprintf("  %s  %s",
			SelectedStyle.Render(fmt.Sprintf("%-19s", hr.keys)),
			NormalStyle.Render(hr.desc),
		)
		b.WriteString(ansi.Truncate(line, m.width-1, "") + "\n")
	}

	b.WriteString(ansi.Truncate(renderDivider(m.width, ""), m.width-1, "") + "\n")
	footer := MutedStyle.Render("  Press [?], [Esc], [Enter], or [q] to return")
	b.WriteString(ansi.Truncate(footer, m.width-1, ""))
	return b.String()
}

func (m *Model) renderTreeView() string {
	var b strings.Builder

	// Render ASCII banner when terminal height has enough room, otherwise compact "Coffee Bun"
	if m.height >= 22 {
		b.WriteString(RenderASCIIHeader(m.width) + "\n")
	} else {
		title := TitleStyle.Render(" Coffee Bun ")
		b.WriteString(ansi.Truncate(title, m.width-1, "") + "\n")
	}

	// Status Summary
	var totalReady, totalMissing, totalSelected int
	for _, p := range m.packages {
		if p.Installed {
			totalReady++
		} else {
			totalMissing++
		}
		if m.selected[p.ID] {
			totalSelected++
		}
	}
	summaryText := fmt.Sprintf("  Status: %s ready, %s missing | %s selected to install",
		BadgeSuccess.Render(fmt.Sprintf("%d", totalReady)),
		BadgeMissing.Render(fmt.Sprintf("%d", totalMissing)),
		SelectedStyle.Render(fmt.Sprintf("%d", totalSelected)),
	)
	b.WriteString(ansi.Truncate(summaryText, m.width-1, "") + "\n")

	rows := m.getVisibleRows()
	totalRows := len(rows)

	viewportHeight := m.treeViewportHeight(totalRows)

	// Adjust cursor
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= totalRows && totalRows > 0 {
		m.cursor = totalRows - 1
	}

	// Adjust scrollOffset to keep cursor inside viewport
	if m.cursor < m.scrollOffset {
		m.scrollOffset = m.cursor
	}
	if m.cursor >= m.scrollOffset+viewportHeight {
		m.scrollOffset = m.cursor - viewportHeight + 1
	}
	maxScroll := totalRows - viewportHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	// Top divider with scroll up indicator
	var topIndicator string
	if m.scrollOffset > 0 {
		topIndicator = fmt.Sprintf("[ ▲ %d more above ]", m.scrollOffset)
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, topIndicator), m.width-1, "") + "\n")

	// Viewport Content
	end := m.scrollOffset + viewportHeight
	if end > totalRows {
		end = totalRows
	}

	pkgMap := make(map[string]model.Package, len(m.packages))
	for _, p := range m.packages {
		pkgMap[p.ID] = p
	}

	for i := m.scrollOffset; i < end; i++ {
		row := rows[i]
		isCursor := (m.focus == focusTree && m.cursor == i)
		cursorPrefix := "  "
		if isCursor {
			cursorPrefix = SelectedStyle.Render("> ")
		}

		if row.rowType == rowGroup {
			foldIcon := "[-]"
			if m.folded[row.groupCat] {
				foldIcon = "[+]"
			}

			var totalInGroup, installedInGroup, selectedInGroup int
			for _, id := range row.group.PackageIDs {
				totalInGroup++
				if p, ok := pkgMap[id]; ok {
					if p.Installed {
						installedInGroup++
					}
					if m.selected[id] {
						selectedInGroup++
					}
				}
			}

			var countBadge string
			if installedInGroup == totalInGroup {
				countBadge = BadgeSuccess.Render(fmt.Sprintf("(%d/%d ready)", installedInGroup, totalInGroup))
			} else {
				countBadge = BadgeMissing.Render(fmt.Sprintf("(%d/%d ready, %d missing)", installedInGroup, totalInGroup, totalInGroup-installedInGroup))
			}

			selBadge := MutedStyle.Render("[ ]")
			if selectedInGroup == totalInGroup && totalInGroup > 0 {
				selBadge = BadgeSuccess.Render("[All]")
			} else if selectedInGroup > 0 {
				selBadge = SelectedStyle.Render(fmt.Sprintf("[%d sel]", selectedInGroup))
			}

			groupLine := fmt.Sprintf("%s%s %s %s %s %s",
				cursorPrefix,
				Bold(foldIcon),
				row.group.Icon,
				Bold(row.group.Title),
				countBadge,
				selBadge,
			)
			if isCursor {
				groupLine = SelectedStyle.Render(groupLine)
			} else {
				groupLine = CategoryHeaderStyle.Render(groupLine)
			}
			b.WriteString(ansi.Truncate(groupLine, m.width-1, "") + "\n")

		} else {
			// Item row
			checkMark := "[ ]"
			if row.pkg.Installed {
				if m.selected[row.pkg.ID] {
					checkMark = BadgeSuccess.Render("[x]")
				} else {
					checkMark = MutedStyle.Render("[x]")
				}
			} else {
				if m.selected[row.pkg.ID] {
					checkMark = SelectedStyle.Render("[+]")
				} else {
					checkMark = MutedStyle.Render("[ ]")
				}
			}

			nameStyle := NormalStyle
			if isCursor {
				nameStyle = SelectedStyle
			}

			nameWidth := 22
			if m.width >= 90 {
				nameWidth = 26
			}
			nameCol := fmt.Sprintf("%-*s", nameWidth, ansi.Truncate(row.pkg.Name, nameWidth, ""))

			var detail string
			if row.pkg.Installed {
				ver := row.pkg.Version
				if ver == "" {
					ver = "installed"
				}
				detail = BadgeSuccess.Render(fmt.Sprintf("%-13s", ansi.Truncate(ver, 13, "")))
			} else {
				detail = BadgeMissing.Render(fmt.Sprintf("%-13s", "missing"))
			}

			var desc string
			fixedWidth := 2 + 2 + 3 + 1 + nameWidth + 1 + 13 + 1 // ~43
			availDesc := m.width - fixedWidth
			if availDesc > 12 {
				desc = MutedStyle.Render("- " + ansi.Truncate(row.pkg.Description, availDesc-2, "..."))
			}

			itemLine := fmt.Sprintf("%s    %s %s %s %s",
				cursorPrefix,
				checkMark,
				nameStyle.Render(nameCol),
				detail,
				desc,
			)
			b.WriteString(ansi.Truncate(itemLine, m.width-1, "") + "\n")
		}
	}

	// Pad remaining lines in viewport if needed
	renderedCount := end - m.scrollOffset
	for i := renderedCount; i < viewportHeight; i++ {
		b.WriteString("\n")
	}

	// Bottom divider with scroll down indicator
	remainingBelow := totalRows - end
	var bottomIndicator string
	if remainingBelow > 0 {
		bottomIndicator = fmt.Sprintf("[ ▼ %d more below ]", remainingBelow)
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, bottomIndicator), m.width-1, "") + "\n")

	// 2 Clean, intuitive buttons on Left and Right
	btnLeft := renderButton("< Exit", m.focus == focusBtnPrev, false)
	btnRight := renderButton(fmt.Sprintf("Start Install (%d) >", totalSelected), m.focus == focusBtnNext, true)

	leftW := lipgloss.Width(btnLeft)
	rightW := lipgloss.Width(btnRight)
	gap := m.width - leftW - rightW - 4
	if gap < 2 {
		gap = 2
	}
	buttonsLine := fmt.Sprintf("  %s%s%s", btnLeft, strings.Repeat(" ", gap), btnRight)
	b.WriteString(ansi.Truncate(buttonsLine, m.width-1, "") + "\n")

	// Hint line
	hintLine := MutedStyle.Render("  [j/k/g/G] Move | [Space] Toggle | [a/m] All/Missing | [?] Shortcuts | [q] Quit")
	b.WriteString(ansi.Truncate(hintLine, m.width-1, ""))

	return b.String()
}

func (m *Model) renderInstallingView() string {
	var b strings.Builder

	spinner := spinnerFrames[m.spinnerIdx]
	titleLine := fmt.Sprintf(" %s  %s",
		SelectedStyle.Render(fmt.Sprintf("[%s]", spinner)),
		Bold("Coffee Bun"),
	)
	b.WriteString(ansi.Truncate(titleLine, m.width-1, "") + "\n")

	fixedLines := 4 // Title (1) + TopDivider (1) + BottomDivider (1) + Summary (1)
	if m.actionSnippet != "" {
		targetName := m.actionTitle
		if targetName == "" {
			targetName = "System Component"
		}
		whyText := m.actionReason
		if whyText == "" {
			whyText = "Administrator (sudo) permission is required to complete this setup step."
		}

		line1 := BadgeError.Render(fmt.Sprintf("  [!] Manual Step Required for %s", targetName))
		line2 := fmt.Sprintf("      %s %s", Bold("Why:"), NormalStyle.Render(whyText))
		line3 := fmt.Sprintf("      %s %s", Bold("How:"), MutedStyle.Render("Open a new terminal tab, run the command below, and wait (auto-resumes every 2s):"))
		line4 := fmt.Sprintf("      %s %s", Bold("Cmd:"), SelectedStyle.Render(m.actionSnippet))

		b.WriteString(ansi.Truncate(line1, m.width-1, "") + "\n")
		b.WriteString(ansi.Truncate(line2, m.width-1, "...") + "\n")
		b.WriteString(ansi.Truncate(line3, m.width-1, "...") + "\n")
		b.WriteString(ansi.Truncate(line4, m.width-1, "") + "\n")
		fixedLines += 4
	}

	viewportHeight := m.height - fixedLines
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	var inProgressCount, completedCount, failedCount, pendingCount int
	var activeIndices []int
	for idx, id := range m.installOrder {
		item := m.installItems[id]
		if item == nil {
			continue
		}
		switch item.Status {
		case model.StatusInstalling, model.StatusWaitingAction:
			inProgressCount++
			activeIndices = append(activeIndices, idx)
		case model.StatusSuccess:
			completedCount++
		case model.StatusFailed:
			failedCount++
		default:
			pendingCount++
		}
	}

	type visualLine struct {
		isHeader bool
		text     string
	}
	var allVisualLines []visualLine

	for _, id := range m.installOrder {
		item := m.installItems[id]
		if item == nil {
			continue
		}

		switch item.Status {
		case model.StatusInstalling:
			spin := SelectedStyle.Render(fmt.Sprintf("[%s]", spinnerFrames[m.spinnerIdx]))
			nameStr := Bold(ansi.Truncate(item.Name, 22, ""))
			msgStr := NormalStyle.Render(item.Message)
			header := fmt.Sprintf("   %s %-22s: %s", spin, nameStr, msgStr)
			allVisualLines = append(allVisualLines, visualLine{isHeader: true, text: header})

			if item.LogLine != "" {
				availLog := m.width - 10
				if availLog < 20 {
					availLog = 20
				}
				logText := ansi.Truncate(item.LogLine, availLog, "...")
				logLine := fmt.Sprintf("       %s %s", MutedStyle.Render(">"), LogStreamStyle.Render(logText))
				allVisualLines = append(allVisualLines, visualLine{isHeader: false, text: logLine})
			}

		case model.StatusSuccess:
			prefix := BadgeSuccess.Render("[OK]")
			nameStr := ansi.Truncate(item.Name, 22, "")
			msgStr := MutedStyle.Render(item.Message)
			header := fmt.Sprintf("   %s %-22s: %s", prefix, nameStr, msgStr)
			allVisualLines = append(allVisualLines, visualLine{isHeader: true, text: header})

		case model.StatusFailed:
			prefix := BadgeError.Render("[FAIL]")
			nameStr := Bold(ansi.Truncate(item.Name, 22, ""))
			msgStr := BadgeError.Render(item.Message)
			header := fmt.Sprintf("   %s %-22s: %s", prefix, nameStr, msgStr)
			allVisualLines = append(allVisualLines, visualLine{isHeader: true, text: header})
			if item.LogLine != "" {
				availLog := m.width - 10
				logText := ansi.Truncate(item.LogLine, availLog, "...")
				logLine := fmt.Sprintf("       %s %s", BadgeError.Render("!"), MutedStyle.Render(logText))
				allVisualLines = append(allVisualLines, visualLine{isHeader: false, text: logLine})
			}

		case model.StatusWaitingAction:
			prefix := BadgeError.Render("[ACTION]")
			nameStr := Bold(ansi.Truncate(item.Name, 22, ""))
			reason := item.Message
			if reason == "" {
				reason = "Waiting for administrator command (see instructions above)"
			}
			header := fmt.Sprintf("   %s %-22s: %s", prefix, nameStr, BadgeMissing.Render(reason))
			allVisualLines = append(allVisualLines, visualLine{isHeader: true, text: header})
			if item.ActionSnippet != "" {
				cmdHint := fmt.Sprintf("       %s %s", MutedStyle.Render("$"), SelectedStyle.Render(item.ActionSnippet))
				allVisualLines = append(allVisualLines, visualLine{isHeader: false, text: cmdHint})
			}

		default: // StatusPending
			prefix := MutedStyle.Render("[..]")
			nameStr := MutedStyle.Render(ansi.Truncate(item.Name, 22, ""))
			msgStr := MutedStyle.Render(item.Message)
			header := fmt.Sprintf("   %s %-22s: %s", prefix, nameStr, msgStr)
			allVisualLines = append(allVisualLines, visualLine{isHeader: true, text: header})
		}
	}

	totalVisual := len(allVisualLines)

	if totalVisual <= viewportHeight {
		m.installScrollOffset = 0
	} else {
		if !m.manualInstallScroll && len(activeIndices) > 0 {
			targetLine := 0
			count := 0
			for idx, id := range m.installOrder {
				if idx == activeIndices[0] {
					targetLine = count
					break
				}
				item := m.installItems[id]
				count++
				if item != nil && ((item.Status == model.StatusInstalling && item.LogLine != "") || (item.Status == model.StatusWaitingAction && item.ActionSnippet != "")) {
					count++
				}
			}
			if targetLine < m.installScrollOffset {
				m.installScrollOffset = targetLine
			}
			if targetLine >= m.installScrollOffset+viewportHeight {
				m.installScrollOffset = targetLine - viewportHeight + 2
			}
		}
		maxScroll := totalVisual - viewportHeight
		if m.installScrollOffset > maxScroll {
			m.installScrollOffset = maxScroll
		}
		if m.installScrollOffset < 0 {
			m.installScrollOffset = 0
		}
	}

	var topIndicator string
	if m.installScrollOffset > 0 {
		topIndicator = fmt.Sprintf("[ ▲ %d more above ]", m.installScrollOffset)
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, topIndicator), m.width-1, "") + "\n")

	end := m.installScrollOffset + viewportHeight
	if end > totalVisual {
		end = totalVisual
	}
	for i := m.installScrollOffset; i < end; i++ {
		b.WriteString(ansi.Truncate(allVisualLines[i].text, m.width-1, "") + "\n")
	}

	for i := end - m.installScrollOffset; i < viewportHeight; i++ {
		b.WriteString("\n")
	}

	remainingBelow := totalVisual - end
	var bottomIndicator string
	if remainingBelow > 0 {
		bottomIndicator = fmt.Sprintf("[ ▼ %d more below ]", remainingBelow)
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, bottomIndicator), m.width-1, "") + "\n")

	summary := fmt.Sprintf("  Progress: %s completed, %s active, %s queued (Total: %d)  %s",
		BadgeSuccess.Render(fmt.Sprintf("%d", completedCount)),
		SelectedStyle.Render(fmt.Sprintf("%d", inProgressCount)),
		MutedStyle.Render(fmt.Sprintf("%d", pendingCount)),
		len(m.installOrder),
		MutedStyle.Render("[j/k/g/G] Scroll"),
	)
	b.WriteString(ansi.Truncate(summary, m.width-1, ""))

	return b.String()
}

func (m *Model) renderCompleteView() string {
	var b strings.Builder

	headerLines := 2
	if m.height >= 22 {
		b.WriteString(RenderASCIIHeader(m.width) + "\n")
		headerLines = 7
	} else {
		title := BadgeSuccess.Render(" Coffee Bun ")
		b.WriteString(ansi.Truncate(title, m.width-1, "") + "\n")
	}

	var successCount, failCount int
	for _, id := range m.installOrder {
		item := m.installItems[id]
		if item == nil {
			continue
		}
		if item.Status == model.StatusSuccess {
			successCount++
		} else if item.Status == model.StatusFailed {
			failCount++
		}
	}

	summaryText := fmt.Sprintf("  Completed: %s succeeded, %s failed (Total: %d)",
		BadgeSuccess.Render(fmt.Sprintf("%d", successCount)),
		BadgeError.Render(fmt.Sprintf("%d", failCount)),
		len(m.installOrder),
	)
	b.WriteString(ansi.Truncate(summaryText, m.width-1, "") + "\n")

	viewportHeight := m.height - (headerLines + 4)
	if viewportHeight < 3 {
		viewportHeight = 3
	}

	type completeRow struct {
		statusBadge string
		name        string
		msg         string
	}
	var compRows []completeRow
	for _, id := range m.installOrder {
		item := m.installItems[id]
		if item == nil {
			continue
		}
		if item.Status == model.StatusSuccess {
			compRows = append(compRows, completeRow{
				statusBadge: BadgeSuccess.Render("[OK]"),
				name:        item.Name,
				msg:         item.Message,
			})
		} else if item.Status == model.StatusFailed {
			compRows = append(compRows, completeRow{
				statusBadge: BadgeError.Render("[FAIL]"),
				name:        item.Name,
				msg:         item.Message,
			})
		}
	}

	totalRows := len(compRows)
	if totalRows < viewportHeight && totalRows > 0 {
		viewportHeight = totalRows
	}

	maxScroll := totalRows - viewportHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.completeScrollOffset > maxScroll {
		m.completeScrollOffset = maxScroll
	}
	if m.completeScrollOffset < 0 {
		m.completeScrollOffset = 0
	}

	var topIndicator string
	if m.completeScrollOffset > 0 {
		topIndicator = fmt.Sprintf("[ ▲ %d more above ]", m.completeScrollOffset)
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, topIndicator), m.width-1, "") + "\n")

	end := m.completeScrollOffset + viewportHeight
	if end > totalRows {
		end = totalRows
	}

	for i := m.completeScrollOffset; i < end; i++ {
		r := compRows[i]
		availMsg := m.width - 32
		if availMsg < 10 {
			availMsg = 10
		}
		line := fmt.Sprintf("   %s %-22s: %s", r.statusBadge, ansi.Truncate(r.name, 22, ""), ansi.Truncate(r.msg, availMsg, "..."))
		b.WriteString(ansi.Truncate(line, m.width-1, "") + "\n")
	}

	remainingBelow := totalRows - end
	var bottomIndicator string
	if remainingBelow > 0 {
		bottomIndicator = fmt.Sprintf("[ ▼ %d more below ]", remainingBelow)
	}
	b.WriteString(ansi.Truncate(renderDivider(m.width, bottomIndicator), m.width-1, "") + "\n")

	btnBack := renderButton("< Back to Selection", m.focus == focusBtnPrev, false)
	btnDone := renderButton("Done / Exit >", m.focus == focusBtnNext, true)

	leftW := lipgloss.Width(btnBack)
	rightW := lipgloss.Width(btnDone)
	gap := m.width - leftW - rightW - 4
	if gap < 2 {
		gap = 2
	}
	buttonsLine := fmt.Sprintf("  %s%s%s", btnBack, strings.Repeat(" ", gap), btnDone)
	b.WriteString(ansi.Truncate(buttonsLine, m.width-1, "") + "\n")

	hintsLine := MutedStyle.Render("  [Tab/←/→] Buttons | [j/k/g/G] Scroll | [Enter] Execute | [q] Quit")
	b.WriteString(ansi.Truncate(hintsLine, m.width-1, ""))

	return b.String()
}

func renderButton(label string, focused bool, isPrimary bool) string {
	if focused {
		if isPrimary {
			return ButtonSuccess.Render(fmt.Sprintf("[ %s ]", label))
		}
		return ButtonFocused.Render(fmt.Sprintf("[ %s ]", label))
	}
	return ButtonNormal.Render(fmt.Sprintf("  %s  ", label))
}

func Bold(s string) string {
	return lipgloss.NewStyle().Bold(true).Render(s)
}
