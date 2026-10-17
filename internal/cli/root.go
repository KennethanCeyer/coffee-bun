package cli

import (
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/KennethanCeyer/coffee-bun/internal/ui"
)

// Execute runs the root CLI command
func Execute() error {
	return NewRootCmd().Execute()
}

// NewRootCmd constructs and returns the cobra root command for Coffee Bun
func NewRootCmd() *cobra.Command {
	var checkFlag bool
	var versionFlag bool
	var configFile string

	longDesc := ui.RenderASCIIHeader(0) + `

  Setting up and maintaining a developer laptop should be effortless.
  Coffee Bun bundles essential toolchains, terminal themes, cloud CLIs,
  editors, language runtimes, and modern utilities into a single workflow.

  INTERACTIVE TUI SHORTCUTS (VIM & ERGONOMIC NAVIGATION):
    j / Down, k / Up     Move cursor down / up one item (or scroll logs)
    g / Home             Jump to the very first item (top of list)
    G / End              Jump to the very last item (bottom of list)
    Ctrl+d / J / PgDn    Jump half-page down
    Ctrl+u / K / PgUp    Jump half-page up
    H / M / L            Jump to Top / Middle / Bottom of current viewport
    h / Left, l / Right  Fold / unfold category group (or switch buttons)
    f / z                Toggle fold/unfold for all category groups
    o / c                Open (unfold) all / Close (fold) all groups
    Space                Toggle selection of item or entire category group
    a / m                Select all packages / Select only missing packages
    r                    Refresh and re-scan environment status
    ?                    Open in-app keyboard shortcuts help screen
    q / Esc / Ctrl+c     Quit Coffee Bun

  CUSTOMIZATION (YAML):
    Coffee Bun automatically discovers custom YAML files in this order:
      1. --config / -C <path>
      2. $BUN_CONFIG environment variable
      3. ./bun.yaml or ./.bun.yaml
      4. ~/.config/coffee-bun/bun.yaml
      5. /etc/coffee-bun/bun.yaml`

	examples := `  # Launch the interactive TUI (default)
  bun

  # Inspect installed versions and missing components
  bun check

  # Install all missing packages or a specific category/package
  bun install all
  bun install terminal
  bun install git ripgrep eza

  # Upgrade installed Homebrew formulae, casks, and Git plugins
  bun update

  # Run environment diagnostics and auto-heal PATH links
  bun doctor

  # Use a custom company or team YAML configuration
  bun -C ./company-bun.yaml check`

	rootCmd := &cobra.Command{
		Use:     "bun",
		Aliases: []string{"coffee-bun"},
		Short:   "Coffee Bun",
		Long:    longDesc,
		Example: examples,
		RunE: func(cmd *cobra.Command, args []string) error {
			if versionFlag {
				fmt.Println("\n" + ui.RenderASCIIHeader(0) + "\n")
				fmt.Printf("  Coffee Bun v%s (%s/%s)\n\n", Version, runtime.GOOS, runtime.GOARCH)
				return nil
			}

			if checkFlag {
				return runCheck(configFile)
			}

			// If running in a real interactive terminal, launch TUI
			if term.IsTerminal(int(os.Stdin.Fd())) {
				p := tea.NewProgram(ui.NewModelWithConfig(configFile), tea.WithAltScreen())
				_, err := p.Run()
				return err
			}

			// Fallback to check overview if non-terminal
			return runCheck(configFile)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "C", "", "Path to custom YAML configuration file (e.g. bun.yaml)")
	rootCmd.Flags().BoolVarP(&checkFlag, "check", "c", false, "Check and display status of all monitored packages")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show Coffee Bun version")

	rootCmd.AddCommand(newCheckCmd(&configFile))
	rootCmd.AddCommand(newInstallCmd(&configFile))
	rootCmd.AddCommand(newUpdateCmd(&configFile))
	rootCmd.AddCommand(newDoctorCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}
