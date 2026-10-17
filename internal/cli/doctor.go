package cli

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/shell"
	"github.com/KennethanCeyer/coffee-bun/internal/ui"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Verify system environment, PATH, and tool configurations",
		Long:  "Inspect operating system architecture, package manager availability, VS Code CLI symlink, and shell hooks across macOS, Linux, and Windows.",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("\n" + ui.RenderASCIIHeader(0) + "\n")

			chk := checker.NewChecker()
			cfg := shell.NewConfigurator()

			// 1. OS check
			fmt.Printf("  %s OS: %s (%s)\n", ui.BadgeSuccess.Render("[OK]"), runtime.GOOS, runtime.GOARCH)

			// 2. Package Manager check (OS-specific)
			switch runtime.GOOS {
			case "darwin":
				if chk.IsBrewInstalled() {
					fmt.Printf("  %s Homebrew: Installed and in PATH\n", ui.BadgeSuccess.Render("[OK]"))
				} else {
					fmt.Printf("  %s Homebrew: Missing! (Install via https://brew.sh)\n", ui.BadgeError.Render("[FAIL]"))
				}
			case "windows":
				if _, err := exec.LookPath("winget"); err == nil {
					fmt.Printf("  %s winget: Installed and in PATH\n", ui.BadgeSuccess.Render("[OK]"))
				} else if _, err := exec.LookPath("choco"); err == nil {
					fmt.Printf("  %s Chocolatey: Installed and in PATH\n", ui.BadgeSuccess.Render("[OK]"))
				} else {
					fmt.Printf("  %s Windows Package Manager: Neither winget nor choco found in PATH\n", ui.BadgeError.Render("[FAIL]"))
				}
			default: // linux
				foundMgr := ""
				for _, mgr := range []string{"apt-get", "dnf", "apk", "pacman", "brew"} {
					if _, err := exec.LookPath(mgr); err == nil {
						foundMgr = mgr
						break
					}
				}
				if foundMgr != "" {
					fmt.Printf("  %s Package Manager (%s): Installed and in PATH\n", ui.BadgeSuccess.Render("[OK]"), foundMgr)
				} else {
					fmt.Printf("  %s Package Manager: No standard Linux package manager found\n", ui.BadgeError.Render("[FAIL]"))
				}
			}

			// 3. Git check
			if _, err := exec.LookPath("git"); err == nil {
				fmt.Printf("  %s Git: Installed and accessible in PATH\n", ui.BadgeSuccess.Render("[OK]"))
			} else {
				fmt.Printf("  %s Git: Missing from PATH\n", ui.BadgeMissing.Render("[-]"))
			}

			// 4. VS Code CLI
			if chk.IsCodeCLILinked() {
				fmt.Printf("  %s VS Code CLI (`code`): Configured and accessible\n", ui.BadgeSuccess.Render("[OK]"))
			} else {
				appCode := chk.FindVSCodeCLI()
				if appCode != "" {
					fmt.Printf("  %s VS Code CLI: Not in PATH. Auto-linking now...\n", ui.BadgeMissing.Render("[..]"))
					if link, err := cfg.LinkVSCodeCLI(); err == nil {
						fmt.Printf("  %s VS Code CLI: Successfully linked to %s\n", ui.BadgeSuccess.Render("[OK]"), link)
					} else {
						fmt.Printf("  %s VS Code CLI: Auto-link failed (%v)\n", ui.BadgeError.Render("[FAIL]"), err)
					}
				} else {
					fmt.Printf("  %s VS Code: Application not detected\n", ui.BadgeMissing.Render("[-]"))
				}
			}

			// 5. Shell configuration checks (macOS & Linux)
			if runtime.GOOS != "windows" {
				if chk.IsP10kConfigured() {
					fmt.Printf("  %s Powerlevel10k: Active theme configured in ~/.zshrc\n", ui.BadgeSuccess.Render("[OK]"))
				} else {
					fmt.Printf("  %s Powerlevel10k: Not enabled in ~/.zshrc\n", ui.BadgeMissing.Render("[-]"))
				}

				if chk.IsZoxideConfigured() {
					fmt.Printf("  %s Zoxide: Shell hook configured in ~/.zshrc\n", ui.BadgeSuccess.Render("[OK]"))
				} else {
					fmt.Printf("  %s Zoxide: Shell hook missing in ~/.zshrc (Run 'bun install cli' to configure)\n", ui.BadgeMissing.Render("[-]"))
				}
			}

			fmt.Println()
			return nil
		},
	}
}
