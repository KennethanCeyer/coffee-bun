package packages

import (
	"os"
	"path/filepath"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

/*
Domain: Terminal & Shell
Description:
    Complete interactive command-line environment configured for developer
    productivity across macOS, Linux, and Windows.

Components:
    1. iTerm2 (macOS): Advanced terminal emulator with profile and color support.
    2. Windows Terminal (Windows): Modern tabbed terminal application for Windows.
    3. MesloLGS NF Font (All OS): Nerd Font patched with glyphs for Powerlevel10k.
    4. iTerm2 Font Config (macOS): Sets default font in iTerm2 preferences to MesloLGS NF.
    5. Brogrammer Theme (macOS): High-contrast color scheme for iTerm2.
    6. Oh My Zsh (macOS / Linux): Extensible zsh configuration framework.
    7. Powerlevel10k (macOS / Linux): Fast and responsive zsh theme.
    8. Zsh Autosuggestions (macOS / Linux): Fish shell-like history-based autosuggestions.
    9. Zsh Syntax Highlighting (macOS / Linux): Live command syntax validation on the CLI.
*/

// TerminalGroup defines the UI presentation and grouping metadata for terminal tools
var TerminalGroup = model.CustomGroup{
	Category:    model.CategoryTerminal,
	Title:       "Terminal & Shell",
	Description: "Terminal emulators, MesloLGS Nerd Font, Oh My Zsh, Powerlevel10k & plugins",
	Icon:        "[shell]",
	PackageIDs: []string{
		"iterm2", "windows-terminal", "meslo-font", "iterm2-font-config", "brogrammer-theme",
		"oh-my-zsh", "powerlevel10k", "zsh-autosuggestions", "zsh-syntax-highlighting",
	},
}

// GetTerminalPackages returns declarative specifications for terminal and shell tools
func GetTerminalPackages() []model.Package {
	home, _ := os.UserHomeDir()
	omzDir := filepath.Join(home, ".oh-my-zsh")
	fontsDir := model.DefaultFontsDir(home)

	return []model.Package{
		{
			ID:          "iterm2",
			Name:        "iTerm2",
			Description: "Terminal emulator for macOS with rich customization and tabs",
			Category:    model.CategoryTerminal,
			Type:        model.TypeBrewCask,
			SupportedOS: []string{"darwin"},
			BrewName:    "iterm2",
			IsCask:      true,
			AppPath:     "/Applications/iTerm.app",
		},
		{
			ID:          "windows-terminal",
			Name:        "Windows Terminal",
			Description: "Modern, fast, tabbed terminal application for Windows",
			Category:    model.CategoryTerminal,
			Type:        model.TypeBrewCask,
			SupportedOS: []string{"windows"},
			WingetID:    "Microsoft.WindowsTerminal",
			ChocoName:   "microsoft-windows-terminal",
			BinaryName:  "wt",
		},
		{
			ID:          "meslo-font",
			Name:        "MesloLGS NF Font",
			Description: "Nerd Font with icons for Powerlevel10k and modern terminal prompts",
			Category:    model.CategoryTerminal,
			Type:        model.TypeFont,
			InstallPath: filepath.Join(fontsDir, "MesloLGS NF Regular.ttf"),
		},
		{
			ID:          "iterm2-font-config",
			Name:        "iTerm2 Font Config",
			Description: "Configure iTerm2 default profile to use MesloLGS NF font",
			Category:    model.CategoryTerminal,
			Type:        model.TypeITermConfig,
			SupportedOS: []string{"darwin"},
		},
		{
			ID:          "brogrammer-theme",
			Name:        "Brogrammer Theme",
			Description: "Brogrammer high-contrast aesthetic color theme for iTerm2",
			Category:    model.CategoryTerminal,
			Type:        model.TypeCustom,
			SupportedOS: []string{"darwin"},
		},
		{
			ID:          "oh-my-zsh",
			Name:        "Oh My Zsh",
			Description: "Community-driven framework for managing zsh configuration",
			Category:    model.CategoryTerminal,
			Type:        model.TypeCustom,
			SupportedOS: []string{"darwin", "linux"},
			InstallPath: omzDir,
		},
		{
			ID:           "powerlevel10k",
			Name:         "Powerlevel10k",
			Description:  "Fast, flexible, and visually stunning theme for Zsh",
			Category:     model.CategoryTerminal,
			Type:         model.TypeGitRepo,
			SupportedOS:  []string{"darwin", "linux"},
			GitURL:       "https://github.com/romkatv/powerlevel10k.git",
			InstallPath:  filepath.Join(omzDir, "custom", "themes", "powerlevel10k"),
			Dependencies: []string{"oh-my-zsh"},
		},
		{
			ID:           "zsh-autosuggestions",
			Name:         "Zsh Autosuggestions",
			Description:  "Fish-like fast autosuggestions for zsh",
			Category:     model.CategoryTerminal,
			Type:         model.TypeGitRepo,
			SupportedOS:  []string{"darwin", "linux"},
			GitURL:       "https://github.com/zsh-users/zsh-autosuggestions.git",
			InstallPath:  filepath.Join(omzDir, "custom", "plugins", "zsh-autosuggestions"),
			Dependencies: []string{"oh-my-zsh"},
		},
		{
			ID:           "zsh-syntax-highlighting",
			Name:         "Zsh Syntax Highlighting",
			Description:  "Fish shell-like syntax highlighting for zsh",
			Category:     model.CategoryTerminal,
			Type:         model.TypeGitRepo,
			SupportedOS:  []string{"darwin", "linux"},
			GitURL:       "https://github.com/zsh-users/zsh-syntax-highlighting.git",
			InstallPath:  filepath.Join(omzDir, "custom", "plugins", "zsh-syntax-highlighting"),
			Dependencies: []string{"oh-my-zsh"},
		},
	}
}
