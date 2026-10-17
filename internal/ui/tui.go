package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/installer"
	"github.com/KennethanCeyer/coffee-bun/internal/model"
	"github.com/KennethanCeyer/coffee-bun/internal/packages"
)

type viewState int

const (
	stateScanning viewState = iota
	stateTree
	stateInstalling
	stateComplete
)

type focusTarget int

const (
	focusTree focusTarget = iota
	focusBtnPrev          // Left: < Exit / < Back
	focusBtnNext          // Right: Start Install > / Done >
)

type treeRowType int

const (
	rowGroup treeRowType = iota
	rowItem
)

type treeRow struct {
	rowType  treeRowType
	group    packages.PackageGroup
	pkg      model.Package
	groupCat model.Category
}

type scanDoneMsg struct {
	pkgs []model.Package
}

type bgScanDoneMsg struct {
	pkgs []model.Package
}

type tickMsg time.Time
type installDoneMsg struct{}

type installItemProgress struct {
	ID            string
	Name          string
	Category      model.Category
	Status        model.InstallStatus
	Message       string
	LogLine       string
	ActionSnippet string
	UpdatedAt     time.Time
}

// Model represents the Bubble Tea TUI state
type Model struct {
	state                viewState
	focus                focusTarget
	cursor               int
	scrollOffset         int
	installScrollOffset  int
	completeScrollOffset int
	showHelp             bool
	manualInstallScroll  bool
	width                int
	height               int
	packages             []model.Package
	groups               []packages.PackageGroup
	selected             map[string]bool
	folded               map[model.Category]bool
	checker              *checker.Checker
	runner               *installer.Runner
	installOrder         []string
	installItems         map[string]*installItemProgress
	activeStatus         string
	actionSnippet        string
	actionTitle          string
	actionReason         string
	lastBgPoll           time.Time
	spinnerIdx           int
	installChan          <-chan model.InstallEvent
	quitting             bool
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewModel creates an interactive TUI model using auto-discovered configuration
func NewModel() Model {
	return NewModelWithConfig("")
}

// NewModelWithConfig creates an interactive TUI model with an optional explicit config path
func NewModelWithConfig(explicitConfigPath string) Model {
	groups, pkgs, cfg, configPath := packages.GetEffectivePackages(explicitConfigPath)

	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 || h <= 0 {
		w = 80
		h = 24
	}

	r := installer.NewRunner()
	if cfg != nil && (cfg.Hooks.PreInstall != "" || cfg.Hooks.PostInstall != "") {
		r.SetHooks(cfg.Hooks)
	}

	initialStatus := "Inspecting macOS developer environment..."
	if configPath != "" {
		initialStatus = fmt.Sprintf("Using config: %s", filepath.Base(configPath))
	}

	return Model{
		state:        stateScanning,
		focus:        focusTree,
		cursor:       0,
		width:        w,
		height:       h,
		packages:     pkgs,
		groups:       groups,
		selected:     make(map[string]bool),
		folded:       make(map[model.Category]bool),
		checker:      checker.NewChecker(),
		runner:       r,
		installOrder: make([]string, 0),
		installItems: make(map[string]*installItemProgress),
		lastBgPoll:   time.Now(),
		activeStatus: initialStatus,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		scanPackagesCmd(m.checker, m.packages),
		tickCmd(),
	)
}

func scanPackagesCmd(c *checker.Checker, pkgs []model.Package) tea.Cmd {
	return func() tea.Msg {
		checked := c.CheckAll(pkgs)
		return scanDoneMsg{pkgs: checked}
	}
}

func bgScanCmd(c *checker.Checker, pkgs []model.Package) tea.Cmd {
	return func() tea.Msg {
		checked := c.CheckAll(pkgs)
		return bgScanDoneMsg{pkgs: checked}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func waitForInstallEvent(ch <-chan model.InstallEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return installDoneMsg{}
		}
		return ev
	}
}

func (m Model) getVisibleRows() []treeRow {
	var rows []treeRow
	pkgMap := make(map[string]model.Package, len(m.packages))
	for _, p := range m.packages {
		pkgMap[p.ID] = p
	}

	for _, g := range m.groups {
		rows = append(rows, treeRow{
			rowType:  rowGroup,
			group:    g,
			groupCat: g.Category,
		})

		if !m.folded[g.Category] {
			for _, id := range g.PackageIDs {
				if p, ok := pkgMap[id]; ok {
					rows = append(rows, treeRow{
						rowType:  rowItem,
						pkg:      p,
						groupCat: g.Category,
					})
				}
			}
		}
	}
	return rows
}

func (m Model) treeViewportHeight(totalRows int) int {
	headerLines := 2 // Compact title + summary
	if m.height >= 22 {
		headerLines = 7 // 6-line ASCII banner + summary
	}
	// Total fixed lines = headerLines + 4 (TopDivider 1, BottomDivider 1, Buttons 1, Hint 1)
	vh := m.height - (headerLines + 4)
	if vh < 3 {
		vh = 3
	}
	if vh > totalRows && totalRows > 0 {
		vh = totalRows
	}
	return vh
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		key := msg.String()

		if key == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}

		// Help modal toggle
		if key == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			if key == "esc" || key == "q" || key == "enter" || key == " " {
				m.showHelp = false
			}
			return m, nil
		}

		switch key {
		case "q":
			if m.state != stateInstalling {
				m.quitting = true
				return m, tea.Quit
			}

		case "esc":
			if m.state == stateTree {
				if m.focus != focusTree {
					m.focus = focusTree
				} else {
					m.quitting = true
					return m, tea.Quit
				}
			} else if m.state == stateComplete {
				m.quitting = true
				return m, tea.Quit
			}

		case "m":
			if m.state == stateTree {
				m.selectAllMissing()
			}

		case "a":
			if m.state == stateTree {
				m.toggleSelectAll()
			}

		case "f", "z":
			if m.state == stateTree {
				m.toggleAllFold()
			}

		case "o":
			if m.state == stateTree {
				for _, g := range m.groups {
					m.folded[g.Category] = false
				}
			}

		case "c":
			if m.state == stateTree {
				for _, g := range m.groups {
					m.folded[g.Category] = true
				}
			}

		case "r":
			if m.state == stateTree {
				m.state = stateScanning
				return m, scanPackagesCmd(m.checker, m.packages)
			}

		case "g", "home":
			if m.state == stateTree {
				m.focus = focusTree
				m.cursor = 0
				m.scrollOffset = 0
			} else if m.state == stateInstalling {
				m.manualInstallScroll = true
				m.installScrollOffset = 0
			} else if m.state == stateComplete {
				m.completeScrollOffset = 0
			}

		case "G", "end":
			if m.state == stateTree {
				rows := m.getVisibleRows()
				if len(rows) > 0 {
					m.focus = focusTree
					m.cursor = len(rows) - 1
				}
			} else if m.state == stateInstalling {
				m.manualInstallScroll = true
				m.installScrollOffset = 9999
			} else if m.state == stateComplete {
				m.completeScrollOffset = 9999
			}

		case "ctrl+d", "pgdown", "J":
			step := 8
			if m.state == stateTree {
				rows := m.getVisibleRows()
				m.focus = focusTree
				m.cursor += step
				if m.cursor >= len(rows) && len(rows) > 0 {
					m.cursor = len(rows) - 1
				}
			} else if m.state == stateInstalling {
				m.manualInstallScroll = true
				m.installScrollOffset += step
			} else if m.state == stateComplete {
				m.completeScrollOffset += step
			}

		case "ctrl+u", "pgup", "K":
			step := 8
			if m.state == stateTree {
				m.focus = focusTree
				m.cursor -= step
				if m.cursor < 0 {
					m.cursor = 0
				}
			} else if m.state == stateInstalling {
				m.manualInstallScroll = true
				m.installScrollOffset -= step
				if m.installScrollOffset < 0 {
					m.installScrollOffset = 0
				}
			} else if m.state == stateComplete {
				m.completeScrollOffset -= step
				if m.completeScrollOffset < 0 {
					m.completeScrollOffset = 0
				}
			}

		case "H":
			if m.state == stateTree {
				m.focus = focusTree
				m.cursor = m.scrollOffset
			}

		case "M":
			if m.state == stateTree {
				rows := m.getVisibleRows()
				vh := m.treeViewportHeight(len(rows))
				m.focus = focusTree
				mid := m.scrollOffset + vh/2
				if mid >= len(rows) && len(rows) > 0 {
					mid = len(rows) - 1
				}
				m.cursor = mid
			}

		case "L":
			if m.state == stateTree {
				rows := m.getVisibleRows()
				vh := m.treeViewportHeight(len(rows))
				m.focus = focusTree
				bot := m.scrollOffset + vh - 1
				if bot >= len(rows) && len(rows) > 0 {
					bot = len(rows) - 1
				}
				m.cursor = bot
			}

		case "tab":
			if m.focus == focusTree {
				m.focus = focusBtnNext
			} else if m.focus == focusBtnNext {
				m.focus = focusBtnPrev
			} else {
				m.focus = focusTree
			}

		case "shift+tab":
			if m.focus == focusTree {
				m.focus = focusBtnPrev
			} else if m.focus == focusBtnPrev {
				m.focus = focusBtnNext
			} else {
				m.focus = focusTree
			}

		case "up", "k":
			if m.state == stateTree {
				if m.focus != focusTree {
					m.focus = focusTree
				} else if m.cursor > 0 {
					m.cursor--
				}
			} else if m.state == stateInstalling {
				m.manualInstallScroll = true
				if m.installScrollOffset > 0 {
					m.installScrollOffset--
				}
			} else if m.state == stateComplete {
				if m.completeScrollOffset > 0 {
					m.completeScrollOffset--
				}
			}

		case "down", "j":
			if m.state == stateTree {
				rows := m.getVisibleRows()
				if m.focus == focusTree {
					if m.cursor < len(rows)-1 {
						m.cursor++
					} else {
						m.focus = focusBtnNext
					}
				}
			} else if m.state == stateInstalling {
				m.manualInstallScroll = true
				m.installScrollOffset++
			} else if m.state == stateComplete {
				m.completeScrollOffset++
			}

		case "left", "h":
			if m.focus != focusTree {
				m.focus = focusBtnPrev
			} else if m.state == stateTree {
				rows := m.getVisibleRows()
				if m.cursor < len(rows) {
					row := rows[m.cursor]
					if row.rowType == rowGroup {
						m.folded[row.groupCat] = true
					} else {
						// On an item row, pressing h folds its parent group and moves cursor to the group header
						m.folded[row.groupCat] = true
						newRows := m.getVisibleRows()
						for idx, nr := range newRows {
							if nr.rowType == rowGroup && nr.groupCat == row.groupCat {
								m.cursor = idx
								break
							}
						}
					}
				}
			}

		case "right", "l":
			if m.focus != focusTree {
				m.focus = focusBtnNext
			} else if m.state == stateTree {
				rows := m.getVisibleRows()
				if m.cursor < len(rows) && rows[m.cursor].rowType == rowGroup {
					m.folded[rows[m.cursor].groupCat] = false
				}
			}

		case " ":
			if m.state == stateTree && m.focus == focusTree {
				rows := m.getVisibleRows()
				if m.cursor < len(rows) {
					row := rows[m.cursor]
					if row.rowType == rowGroup {
						m.toggleGroupSelection(row.group)
					} else {
						m.selected[row.pkg.ID] = !m.selected[row.pkg.ID]
					}
				}
			}

		case "enter":
			if m.state == stateTree {
				if m.focus == focusTree {
					rows := m.getVisibleRows()
					if m.cursor < len(rows) {
						row := rows[m.cursor]
						if row.rowType == rowGroup {
							m.folded[row.groupCat] = !m.folded[row.groupCat]
						} else {
							m.selected[row.pkg.ID] = !m.selected[row.pkg.ID]
						}
					}
				} else if m.focus == focusBtnPrev {
					m.quitting = true
					return m, tea.Quit
				} else if m.focus == focusBtnNext {
					return m.startInstallation()
				}
			} else if m.state == stateComplete {
				if m.focus == focusBtnPrev {
					m.state = stateTree
					m.focus = focusTree
				} else {
					m.quitting = true
					return m, tea.Quit
				}
			}
		}

	case tickMsg:
		m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
		// Periodic background polling in tree mode
		if m.state == stateTree && time.Since(m.lastBgPoll) > 3*time.Second {
			m.lastBgPoll = time.Now()
			return m, tea.Batch(tickCmd(), bgScanCmd(m.checker, m.packages))
		}
		return m, tickCmd()

	case scanDoneMsg:
		m.packages = msg.pkgs
		for _, p := range m.packages {
			if !p.Installed {
				m.selected[p.ID] = true
			}
		}
		m.state = stateTree
		return m, nil

	case bgScanDoneMsg:
		m.packages = msg.pkgs
		return m, nil

	case model.InstallEvent:
		item, exists := m.installItems[msg.PackageID]
		if !exists {
			item = &installItemProgress{
				ID:            msg.PackageID,
				Name:          msg.Name,
				Status:        msg.Status,
				Message:       msg.Message,
				LogLine:       msg.LogLine,
				ActionSnippet: msg.ActionSnippet,
				UpdatedAt:     time.Now(),
			}
			m.installItems[msg.PackageID] = item
			m.installOrder = append(m.installOrder, msg.PackageID)
		} else {
			if msg.Name != "" {
				item.Name = msg.Name
			}
			item.Status = msg.Status
			if msg.Message != "" {
				item.Message = msg.Message
			}
			if msg.LogLine != "" {
				item.LogLine = msg.LogLine
			}
			if msg.ActionSnippet != "" {
				item.ActionSnippet = msg.ActionSnippet
			}
			item.UpdatedAt = time.Now()
		}

		m.refreshActiveActionBanner()

		if msg.Message != "" {
			m.activeStatus = msg.Message
		}
		return m, waitForInstallEvent(m.installChan)

	case installDoneMsg:
		m.packages = m.checker.CheckAll(m.packages)
		m.actionSnippet = ""
		m.actionTitle = ""
		m.actionReason = ""
		m.state = stateComplete
		m.focus = focusBtnNext
		m.activeStatus = "Installation and configuration completed!"
		return m, nil
	}

	return m, nil
}

func (m *Model) refreshActiveActionBanner() {
	for _, id := range m.installOrder {
		item := m.installItems[id]
		if item != nil && item.Status == model.StatusWaitingAction && item.ActionSnippet != "" {
			m.actionTitle = item.Name
			m.actionReason = item.Message
			m.actionSnippet = item.ActionSnippet
			return
		}
	}
	m.actionTitle = ""
	m.actionReason = ""
	m.actionSnippet = ""
}

func (m *Model) toggleGroupSelection(group packages.PackageGroup) {
	allSelected := true
	for _, id := range group.PackageIDs {
		if !m.selected[id] {
			allSelected = false
			break
		}
	}
	newState := !allSelected
	for _, id := range group.PackageIDs {
		m.selected[id] = newState
	}
}

func (m *Model) toggleAllFold() {
	var anyUnfolded bool
	for _, g := range m.groups {
		if !m.folded[g.Category] {
			anyUnfolded = true
			break
		}
	}
	for _, g := range m.groups {
		m.folded[g.Category] = anyUnfolded
	}
}

func (m *Model) toggleSelectAll() {
	allSelected := true
	for _, p := range m.packages {
		if !m.selected[p.ID] {
			allSelected = false
			break
		}
	}
	newState := !allSelected
	for _, p := range m.packages {
		m.selected[p.ID] = newState
	}
}

func (m *Model) selectAllMissing() {
	for _, p := range m.packages {
		if !p.Installed {
			m.selected[p.ID] = true
		} else {
			m.selected[p.ID] = false
		}
	}
}

func (m Model) startInstallation() (tea.Model, tea.Cmd) {
	var targetPkgs []model.Package
	for _, p := range m.packages {
		if m.selected[p.ID] {
			targetPkgs = append(targetPkgs, p)
		}
	}

	if len(targetPkgs) == 0 {
		m.state = stateComplete
		m.focus = focusBtnNext
		return m, nil
	}

	m.state = stateInstalling
	m.manualInstallScroll = false
	m.installScrollOffset = 0
	m.installOrder = make([]string, 0, len(targetPkgs))
	m.installItems = make(map[string]*installItemProgress)

	for _, p := range targetPkgs {
		m.installOrder = append(m.installOrder, p.ID)
		m.installItems[p.ID] = &installItemProgress{
			ID:        p.ID,
			Name:      p.Name,
			Category:  p.Category,
			Status:    model.StatusPending,
			Message:   "Queued for installation",
			UpdatedAt: time.Now(),
		}
	}

	m.activeStatus = "Installing selected packages..."
	ctx := context.Background()
	m.installChan = m.runner.InstallAsync(ctx, targetPkgs)
	return m, waitForInstallEvent(m.installChan)
}
