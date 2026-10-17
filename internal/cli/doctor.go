package cli

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/model"
	"github.com/KennethanCeyer/coffee-bun/internal/packages"
	"github.com/KennethanCeyer/coffee-bun/internal/shell"
	"github.com/KennethanCeyer/coffee-bun/internal/ui"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run comprehensive diagnostics on system, toolchains, shell, fonts, and updates",
		Long:  "Inspect operating system health, package managers, PATH entries, Git identity, SSH keys, Nerd Fonts, shell plugins, language runtimes, and available package updates across macOS, Linux (including WSL), and Windows (PowerShell).",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("\n" + ui.RenderASCIIHeader(0) + "\n")

			chk := checker.NewChecker()
			cfg := shell.NewConfigurator()
			homeDir, _ := os.UserHomeDir()

			var passed, warnings, failed int
			okBadge := ui.BadgeSuccess.Render("  [OK]   ")
			warnBadge := ui.BadgeMissing.Render("  [WARN] ")
			failBadge := ui.BadgeError.Render("  [FAIL] ")

			printSection := func(title string) {
				fmt.Println(ui.CategoryHeaderStyle.Render(fmt.Sprintf("  [%s]", title)))
				fmt.Println(ui.MutedStyle.Render("  " + strings.Repeat("-", 68)))
			}

			printRow := func(status string, label string, detail string) {
				padded := fmt.Sprintf("%-26s", label)
				fmt.Printf("%s %s %s\n", status, ui.NormalStyle.Render(padded), ui.MutedStyle.Render(detail))
			}

			// =========================================================================
			// 1. System & Platform Environment
			// =========================================================================
			printSection("System & Platform")

			osVersion := detectOSVersionString()
			printRow(okBadge, "Operating System", fmt.Sprintf("%s %s (%s/%s)", model.PlatformDisplayName(), osVersion, runtime.GOOS, runtime.GOARCH))
			passed++

			// Locale & UTF-8 check
			lang := os.Getenv("LANG")
			if lang == "" {
				lang = os.Getenv("LC_ALL")
			}
			if runtime.GOOS == "windows" || strings.Contains(strings.ToUpper(lang), "UTF-8") || strings.Contains(strings.ToUpper(lang), "UTF8") {
				if lang == "" {
					lang = "UTF-8 (PowerShell)"
				}
				printRow(okBadge, "Terminal Locale", fmt.Sprintf("%s (UTF-8 ready)", lang))
				passed++
			} else {
				printRow(warnBadge, "Terminal Locale", fmt.Sprintf("%q (Recommend export LANG=\"en_US.UTF-8\")", lang))
				warnings++
			}

			// ~/.local/bin in PATH check
			localBin := filepath.Join(homeDir, ".local", "bin")
			pathEnv := os.Getenv("PATH")
			if strings.Contains(pathEnv, ".local/bin") || strings.Contains(pathEnv, `.local\bin`) {
				printRow(okBadge, "Local Bin PATH", fmt.Sprintf("%s is in PATH", localBin))
				passed++
			} else {
				printRow(warnBadge, "Local Bin PATH", fmt.Sprintf("%s is not in current PATH", localBin))
				warnings++
			}

			// Network connectivity to GitHub
			client := &http.Client{Timeout: 3 * time.Second}
			if resp, err := client.Head("https://github.com"); err == nil {
				_ = resp.Body.Close()
				printRow(okBadge, "Network (github.com)", "Reachable (HTTPS 200 OK)")
				passed++
			} else {
				printRow(warnBadge, "Network (github.com)", "Offline or unreachable (downloads may fail)")
				warnings++
			}
			fmt.Println()

			// =========================================================================
			// 2. Package Managers & Core Toolchains
			// =========================================================================
			printSection("Package Managers & Toolchains")

			switch runtime.GOOS {
			case "darwin":
				if out, err := exec.Command("xcode-select", "-p").Output(); err == nil && len(out) > 0 {
					printRow(okBadge, "Xcode CLI Tools", strings.TrimSpace(string(out)))
					passed++
				} else {
					printRow(failBadge, "Xcode CLI Tools", "Missing (Run 'xcode-select --install')")
					failed++
				}

				if chk.IsBrewInstalled() {
					brewVer := firstLine(runQuickCmd("brew", "--version"))
					printRow(okBadge, "Homebrew", brewVer)
					passed++
				} else {
					printRow(failBadge, "Homebrew", "Missing (Install via https://brew.sh)")
					failed++
				}

			case "windows":
				psVer := firstLine(runQuickCmd("powershell", "-NoProfile", "-Command", "$PSVersionTable.PSVersion.ToString()"))
				if psVer != "" {
					printRow(okBadge, "PowerShell Runtime", "v"+psVer)
					passed++
				}
				hasWinMgr := false
				if _, err := exec.LookPath("winget"); err == nil {
					printRow(okBadge, "WinGet Package Manager", firstLine(runQuickCmd("winget", "--version")))
					passed++
					hasWinMgr = true
				}
				if _, err := exec.LookPath("choco"); err == nil {
					printRow(okBadge, "Chocolatey", firstLine(runQuickCmd("choco", "--version")))
					passed++
					hasWinMgr = true
				}
				if !hasWinMgr {
					printRow(failBadge, "Windows Package Manager", "Neither winget nor choco found in PATH")
					failed++
				}

			default: // linux & WSL
				if model.IsWSL() {
					distro := os.Getenv("WSL_DISTRO_NAME")
					if distro == "" {
						distro = "WSL"
					}
					printRow(okBadge, "WSL Linux Subsystem", fmt.Sprintf("Active (%s grouped under Linux)", distro))
					passed++
				}
				foundMgr := ""
				for _, mgr := range []string{"apt-get", "dnf", "apk", "pacman", "brew"} {
					if path, err := exec.LookPath(mgr); err == nil {
						foundMgr = fmt.Sprintf("%s (%s)", mgr, path)
						break
					}
				}
				if foundMgr != "" {
					printRow(okBadge, "Linux Package Manager", foundMgr)
					passed++
				} else {
					printRow(failBadge, "Linux Package Manager", "No supported package manager found")
					failed++
				}
			}

			// Git installation & identity
			if _, err := exec.LookPath("git"); err == nil {
				gitVer := firstLine(runQuickCmd("git", "--version"))
				printRow(okBadge, "Git", gitVer)
				passed++

				gitName := firstLine(runQuickCmd("git", "config", "--global", "user.name"))
				gitEmail := firstLine(runQuickCmd("git", "config", "--global", "user.email"))
				if gitName != "" && gitEmail != "" {
					printRow(okBadge, "Git Global Identity", fmt.Sprintf("%s <%s>", gitName, gitEmail))
					passed++
				} else {
					printRow(warnBadge, "Git Global Identity", "user.name / user.email not set in global gitconfig")
					warnings++
				}
			} else {
				printRow(failBadge, "Git", "Missing from PATH")
				failed++
			}

			// SSH Keys check
			sshEd := filepath.Join(homeDir, ".ssh", "id_ed25519")
			sshRsa := filepath.Join(homeDir, ".ssh", "id_rsa")
			if _, err := os.Stat(sshEd); err == nil {
				printRow(okBadge, "SSH Key", "~/.ssh/id_ed25519 present")
				passed++
			} else if _, err := os.Stat(sshRsa); err == nil {
				printRow(okBadge, "SSH Key", "~/.ssh/id_rsa present")
				passed++
			} else {
				printRow(warnBadge, "SSH Key", "No default SSH key found in ~/.ssh")
				warnings++
			}
			fmt.Println()

			// =========================================================================
			// 3. Terminal, Shell & Fonts
			// =========================================================================
			printSection("Terminal, Shell & Fonts")

			fontFile := filepath.Join(model.DefaultFontsDir(homeDir), "MesloLGS NF Regular.ttf")
			if _, err := os.Stat(fontFile); err == nil {
				printRow(okBadge, "MesloLGS NF Nerd Font", fmt.Sprintf("Installed (%s)", model.DefaultFontsDir(homeDir)))
				passed++
			} else {
				printRow(warnBadge, "MesloLGS NF Nerd Font", "Missing (Run 'bun install meslo-font')")
				warnings++
			}

			if runtime.GOOS != "windows" {
				shellEnv := os.Getenv("SHELL")
				if shellEnv == "" {
					shellEnv = "/bin/zsh"
				}
				printRow(okBadge, "Login Shell", shellEnv)
				passed++

				omzDir := filepath.Join(homeDir, ".oh-my-zsh")
				if info, err := os.Stat(omzDir); err == nil && info.IsDir() {
					rev := firstLine(runQuickCmd("git", "-C", omzDir, "rev-parse", "--short", "HEAD"))
					printRow(okBadge, "Oh My Zsh", fmt.Sprintf("Installed (%s)", rev))
					passed++
				} else {
					printRow(warnBadge, "Oh My Zsh", "Not installed in ~/.oh-my-zsh")
					warnings++
				}

				if chk.IsP10kConfigured() {
					printRow(okBadge, "Powerlevel10k Theme", "Active in ~/.zshrc")
					passed++
				} else {
					printRow(warnBadge, "Powerlevel10k Theme", "Not enabled in ~/.zshrc")
					warnings++
				}

				if chk.IsZoxideConfigured() {
					printRow(okBadge, "Zoxide Shell Hook", "Configured in ~/.zshrc")
					passed++
				} else {
					printRow(warnBadge, "Zoxide Shell Hook", "Not configured in ~/.zshrc")
					warnings++
				}
			}
			fmt.Println()

			// =========================================================================
			// 4. Runtimes, Compilers & Editors
			// =========================================================================
			printSection("Runtimes, Compilers & Editors")

			runtimes := []struct {
				label string
				bin   string
				args  []string
			}{
				{"Go Compiler", "go", []string{"version"}},
				{"Python 3", "python3", []string{"--version"}},
				{"Node.js", "node", []string{"--version"}},
				{"C/C++ Compiler", "clang", []string{"--version"}},
			}
			for _, rt := range runtimes {
				if _, err := exec.LookPath(rt.bin); err == nil {
					ver := firstLine(runQuickCmd(rt.bin, rt.args...))
					printRow(okBadge, rt.label, ver)
					passed++
				} else {
					printRow(warnBadge, rt.label, fmt.Sprintf("%s not found in PATH", rt.bin))
					warnings++
				}
			}

			// Vim runtime
			vimDir := filepath.Join(homeDir, ".vim_runtime")
			if info, err := os.Stat(vimDir); err == nil && info.IsDir() {
				printRow(okBadge, "Ultimate Vim Runtime", "Installed in ~/.vim_runtime")
				passed++
			} else {
				printRow(warnBadge, "Ultimate Vim Runtime", "Not installed (Run 'bun install vim-runtime')")
				warnings++
			}

			// VS Code CLI with auto-heal
			if chk.IsCodeCLILinked() {
				codeVer := firstLine(runQuickCmd("code", "--version"))
				printRow(okBadge, "VS Code CLI (`code`)", fmt.Sprintf("Linked in PATH (%s)", codeVer))
				passed++
			} else {
				appCode := chk.FindVSCodeCLI()
				if appCode != "" {
					if link, err := cfg.LinkVSCodeCLI(); err == nil {
						printRow(okBadge, "VS Code CLI (`code`)", fmt.Sprintf("Auto-healed symlink -> %s", link))
						passed++
					} else {
						printRow(failBadge, "VS Code CLI (`code`)", fmt.Sprintf("Auto-link failed: %v", err))
						failed++
					}
				} else {
					printRow(warnBadge, "VS Code CLI (`code`)", "VS Code application not detected")
					warnings++
				}
			}
			fmt.Println()

			// =========================================================================
			// 5. Bundle & Update Check Summary
			// =========================================================================
			printSection("Bundle & Update Status")

			_, pkgs, _, _ := packages.GetEffectivePackages("")
			checked := chk.CheckAllWithUpdates(pkgs)

			var installedCount, outdatedCount, missingCount int
			var outdatedNames, missingNames []string
			for _, p := range checked {
				if p.Installed {
					installedCount++
					if p.Outdated {
						outdatedCount++
						outdatedNames = append(outdatedNames, p.Name)
					}
				} else {
					missingCount++
					missingNames = append(missingNames, p.Name)
				}
			}

			if missingCount == 0 {
				printRow(okBadge, "Bundled Packages", fmt.Sprintf("All %d/%d supported packages are installed", installedCount, len(checked)))
				passed++
			} else {
				printRow(warnBadge, "Bundled Packages", fmt.Sprintf("%d/%d installed (%d missing: %s)", installedCount, len(checked), missingCount, strings.Join(missingNames, ", ")))
				warnings++
			}

			if outdatedCount == 0 {
				printRow(okBadge, "Package Updates", "All installed packages are at their latest versions (0 outdated)")
				passed++
			} else {
				printRow(warnBadge, "Package Updates", fmt.Sprintf("%d update(s) available (%s) -> Run 'bun' or 'bun update'", outdatedCount, strings.Join(outdatedNames, ", ")))
				warnings++
			}

			fmt.Println()
			fmt.Printf("  Diagnostic Summary: %s passed, %s warning(s), %s issue(s)\n",
				ui.BadgeSuccess.Render(fmt.Sprintf("%d", passed)),
				ui.BadgeMissing.Render(fmt.Sprintf("%d", warnings)),
				ui.BadgeError.Render(fmt.Sprintf("%d", failed)),
			)
			if failed == 0 && missingCount == 0 && outdatedCount == 0 {
				fmt.Printf("  %s %s\n\n",
					ui.BadgeSuccess.Render("[OK]"),
					ui.NormalStyle.Render("Your workstation is healthy and all packages are up to date!"),
				)
			} else {
				fmt.Println()
			}

			return nil
		},
	}
}

func runQuickCmd(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func firstLine(s string) string {
	if s == "" {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[0])
}

func detectOSVersionString() string {
	switch runtime.GOOS {
	case "darwin":
		return firstLine(runQuickCmd("sw_vers", "-productVersion"))
	case "linux":
		return firstLine(runQuickCmd("uname", "-r"))
	default:
		return ""
	}
}

