package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
	"github.com/KennethanCeyer/coffee-bun/internal/packages"
	"github.com/KennethanCeyer/coffee-bun/internal/ui"
	"github.com/KennethanCeyer/coffee-bun/internal/updater"
)

func newUpdateCmd(configPath *string) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Upgrade installed Homebrew packages, casks, and git themes",
		Long:  "Update Homebrew indices, upgrade outdated formulae and casks, pull latest commits for installed Git repositories, and run custom update commands.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("\n" + ui.RenderASCIIHeader(0))
			fmt.Println("\n  Checking for updates and running maintenance...")
			u := updater.NewUpdater()
			var cfgPath string
			if configPath != nil {
				cfgPath = *configPath
			}
			_, pkgs, _, _ := packages.GetEffectivePackages(cfgPath)
			ctx := context.Background()
			events := u.UpdateAllAsync(ctx, pkgs)

			for ev := range events {
				var badge string
				switch ev.Status {
				case model.StatusSuccess:
					badge = ui.BadgeSuccess.Render("[OK]")
				case model.StatusFailed:
					badge = ui.BadgeError.Render("[FAIL]")
				default:
					badge = ui.BadgeMissing.Render("[..]")
				}
				fmt.Printf("  %s %s: %s\n", badge, ui.Bold(ev.PackageName), ev.Message)
			}

			fmt.Println(ui.BadgeSuccess.Render("\n  [OK] Update process completed!\n"))
			return nil
		},
	}
}
