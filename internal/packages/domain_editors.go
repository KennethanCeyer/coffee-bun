package packages

import (
	"os"
	"path/filepath"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

/*
Domain: Editors & Vim Runtime
Description:
    Primary code editing environments for terminal and GUI workflows across
    macOS, Linux, and Windows. Includes amix/vimrc Ultimate Vim Runtime and
    Visual Studio Code with the `code` CLI.

Components:
    1. Ultimate Vim Runtime: amix/vimrc configuration with syntax, plugins, and mappings.
    2. Visual Studio Code: Extensible GUI code editor for modern development.
    3. VS Code CLI: Symlink `code` into PATH to launch VS Code from terminal.
*/

// EditorsGroup defines the UI presentation and grouping metadata for editor tools
var EditorsGroup = model.CustomGroup{
	Category:    model.CategoryVim,
	Title:       "Editors & Vim Runtime",
	Description: "Ultimate Vim runtime (amix/vimrc), Visual Studio Code & `code` CLI linking",
	Icon:        "[editor]",
	PackageIDs: []string{
		"vim-runtime", "vscode", "code-cli",
	},
}

// GetEditorsPackages returns declarative specifications for editors and vim runtime
func GetEditorsPackages() []model.Package {
	home, _ := os.UserHomeDir()

	return []model.Package{
		{
			ID:          "vim-runtime",
			Name:        "Ultimate Vim Runtime",
			Description: "The Ultimate Vim configuration (amix/vimrc) with plugins and syntax",
			Category:    model.CategoryVim,
			Type:        model.TypeVimRuntime,
			GitURL:      "https://github.com/amix/vimrc.git",
			InstallPath: filepath.Join(home, ".vim_runtime"),
		},
		{
			ID:          "vscode",
			Name:        "Visual Studio Code",
			Description: "Code editor redefined and optimized for building modern apps",
			Category:    model.CategoryVim,
			Type:        model.TypeBrewCask,
			BrewName:    "visual-studio-code",
			WingetID:    "Microsoft.VisualStudioCode",
			ChocoName:   "vscode",
			IsCask:      true,
			BinaryName:  "code",
			AppPath:     "/Applications/Visual Studio Code.app",
		},
		{
			ID:          "code-cli",
			Name:        "VS Code CLI (`code`)",
			Description: "Symlink `code` binary into PATH for instant terminal access",
			Category:    model.CategoryVim,
			Type:        model.TypeCustom,
			BinaryName:  "code",
		},
	}
}
