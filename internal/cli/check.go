package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/packages"
	"github.com/KennethanCeyer/coffee-bun/internal/ui"
)

func newCheckCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Inspect installed tools, versions, and missing components",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg string
			if configPath != nil {
				cfg = *configPath
			}
			return runCheck(cfg)
		},
	}
}

func runCheck(configFile string) error {
	chk := checker.NewChecker()
	_, pkgs, _, loadedPath := packages.GetEffectivePackages(configFile)
	if loadedPath != "" {
		fmt.Printf("\n  Loaded configuration: %s\n", loadedPath)
	}
	fmt.Print("\n  Scanning macOS developer environment...")
	checked := chk.CheckAll(pkgs)
	fmt.Print("\r\033[K")
	output := ui.PrintStatusSummary(checked)
	fmt.Print(output)
	return nil
}
