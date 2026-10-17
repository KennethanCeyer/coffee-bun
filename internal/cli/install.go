package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/installer"
	"github.com/KennethanCeyer/coffee-bun/internal/model"
	"github.com/KennethanCeyer/coffee-bun/internal/packages"
	"github.com/KennethanCeyer/coffee-bun/internal/ui"
)

func newInstallCmd(configPath *string) *cobra.Command {
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "install [all|system|terminal|cloud|vim|languages|cli|apps|<pkg-id>...]",
		Short: "Install missing tools non-interactively",
		Long: `Install missing packages non-interactively from the command line.
You can target 'all', a specific category ('system', 'terminal', 'cloud',
'vim', 'languages', 'cli', 'apps'), or one or more individual package IDs.`,
		Example: `  bun install all
  bun install terminal
  bun install git fzf ripgrep`,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("\n" + ui.RenderASCIIHeader(0))

			var cfgPath string
			if configPath != nil {
				cfgPath = *configPath
			}
			_, defaultPkgs, customCfg, loadedPath := packages.GetEffectivePackages(cfgPath)
			if loadedPath != "" {
				fmt.Printf("\n  Using configuration: %s\n", loadedPath)
			}

			chk := checker.NewChecker()
			pkgs := chk.CheckAll(defaultPkgs)

			var targetPkgs []model.Package

			if allFlag || (len(args) > 0 && args[0] == "all") {
				for _, p := range pkgs {
					if !p.Installed {
						targetPkgs = append(targetPkgs, p)
					}
				}
			} else if len(args) > 0 {
				categoryOrTarget := strings.ToLower(args[0])
				switch categoryOrTarget {
				case "system", "toolchains":
					for _, p := range pkgs {
						if p.Category == model.CategorySystem && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				case "terminal", "shell":
					for _, p := range pkgs {
						if p.Category == model.CategoryTerminal && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				case "cloud", "devops":
					for _, p := range pkgs {
						if p.Category == model.CategoryCloud && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				case "vim", "editor", "editors":
					for _, p := range pkgs {
						if p.Category == model.CategoryVim && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				case "lang", "languages":
					for _, p := range pkgs {
						if p.Category == model.CategoryLanguages && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				case "cli", "tools":
					for _, p := range pkgs {
						if p.Category == model.CategoryCLITools && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				case "apps":
					for _, p := range pkgs {
						if p.Category == model.CategoryApps && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				default:
					argMap := make(map[string]bool)
					for _, a := range args {
						argMap[strings.ToLower(a)] = true
					}
					for _, p := range pkgs {
						if (argMap[strings.ToLower(p.ID)] || argMap[strings.ToLower(p.Name)]) && !p.Installed {
							targetPkgs = append(targetPkgs, p)
						}
					}
				}
			} else {
				for _, p := range pkgs {
					if !p.Installed {
						targetPkgs = append(targetPkgs, p)
					}
				}
			}

			if len(targetPkgs) == 0 {
				fmt.Println(ui.BadgeSuccess.Render("\n  [OK] Everything is already installed! No action needed.\n"))
				return nil
			}

			fmt.Printf("\n  Installing %d missing package(s)...\n\n", len(targetPkgs))
			runner := installer.NewRunner()
			if customCfg != nil && (customCfg.Hooks.PreInstall != "" || customCfg.Hooks.PostInstall != "") {
				runner.SetHooks(customCfg.Hooks)
			}
			ctx := context.Background()
			events := runner.InstallAsync(ctx, targetPkgs)

			lastStatus := make(map[string]model.InstallStatus)
			for ev := range events {
				if ev.LogLine != "" && ev.Status == model.StatusInstalling {
					continue
				}
				if lastStatus[ev.PackageID] == ev.Status && ev.Status == model.StatusInstalling {
					continue
				}
				lastStatus[ev.PackageID] = ev.Status

				var badge string
				switch ev.Status {
				case model.StatusSuccess:
					badge = ui.BadgeSuccess.Render("[OK]")
				case model.StatusFailed:
					badge = ui.BadgeError.Render("[FAIL]")
				case model.StatusWaitingAction:
					badge = ui.BadgeError.Render("[ACTION]")
				default:
					badge = ui.BadgeMissing.Render("[..]")
				}
				fmt.Printf("  %s %s: %s\n", badge, ui.Bold(ev.Name), ev.Message)
				if ev.ActionSnippet != "" {
					fmt.Printf("      %s Open another terminal tab and run:\n", ui.Bold("How:"))
					fmt.Printf("      %s %s\n", ui.Bold("Cmd:"), ui.SelectedStyle.Render(ev.ActionSnippet))
					fmt.Printf("      %s\n", ui.MutedStyle.Render("(Checking every 2s... installation will resume automatically once completed)"))
				}
			}

			fmt.Println(ui.BadgeSuccess.Render("\n  [OK] Installation finished successfully!\n"))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&allFlag, "all", "a", false, "Install all missing packages")
	return cmd
}
