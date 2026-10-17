package packages

import "github.com/KennethanCeyer/coffee-bun/internal/model"

/*
Domain: Modern CLI Utilities
Description:
    High-performance modern command-line utilities that dramatically improve
    terminal navigation, process inspection, fuzzy searching, and text processing
    across macOS, Linux, and Windows.

Components:
    1. Zoxide: Smarter `cd` navigation command that learns your habits.
    2. Htop (macOS / Linux): Interactive and visually rich process viewer.
    3. Fzf: Blazing fast command-line fuzzy finder.
    4. Ripgrep (rg): Extremely fast recursively search utility that respects .gitignore.
    5. Eza: Modern, feature-rich replacement for `ls` with colors and git status.
    6. Tree: Recursive directory listing program producing depth-indented listings.
    7. Jq: Lightweight and flexible command-line JSON processor.
    8. Bat: A `cat` clone with syntax highlighting and Git integration.
*/

// CLIGroup defines the UI presentation and grouping metadata for CLI utilities
var CLIGroup = model.CustomGroup{
	Category:    model.CategoryCLITools,
	Title:       "Modern CLI Power Utilities",
	Description: "zoxide, htop, fzf, ripgrep, eza, tree, jq, bat",
	Icon:        "[cli]",
	PackageIDs: []string{
		"zoxide", "htop", "fzf", "ripgrep", "eza", "tree", "jq", "bat",
	},
}

// CLIPackages defines declarative specifications for modern CLI tools
var CLIPackages = []model.Package{
	{
		ID:          "zoxide",
		Name:        "Zoxide",
		Description: "A smarter cd command that learns your habits for fast jumping",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		BrewName:    "zoxide",
		AptName:     "zoxide",
		WingetID:    "ajeetdsouza.zoxide",
		ChocoName:   "zoxide",
		BinaryName:  "zoxide",
	},
	{
		ID:          "htop",
		Name:        "Htop",
		Description: "Interactive real-time process viewer and system monitor",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		SupportedOS: []string{"darwin", "linux"},
		BrewName:    "htop",
		AptName:     "htop",
		BinaryName:  "htop",
	},
	{
		ID:          "fzf",
		Name:        "Fzf",
		Description: "General-purpose command-line fuzzy finder with preview support",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		BrewName:    "fzf",
		AptName:     "fzf",
		WingetID:    "junegunn.fzf",
		ChocoName:   "fzf",
		BinaryName:  "fzf",
	},
	{
		ID:          "ripgrep",
		Name:        "Ripgrep",
		Description: "Ultra-fast line-oriented search tool respecting .gitignore",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		BrewName:    "ripgrep",
		AptName:     "ripgrep",
		WingetID:    "BurntSushi.ripgrep.MSVC",
		ChocoName:   "ripgrep",
		BinaryName:  "rg",
	},
	{
		ID:          "eza",
		Name:        "Eza",
		Description: "Modern, maintained replacement for ls with colors and git status",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		BrewName:    "eza",
		AptName:     "eza",
		WingetID:    "eza-community.eza",
		ChocoName:   "eza",
		BinaryName:  "eza",
	},
	{
		ID:          "tree",
		Name:        "Tree",
		Description: "Recursive directory listing program producing depth-indented tree",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		SupportedOS: []string{"darwin", "linux"},
		BrewName:    "tree",
		AptName:     "tree",
		BinaryName:  "tree",
	},
	{
		ID:          "jq",
		Name:        "Jq",
		Description: "Lightweight and flexible command-line JSON processor",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		BrewName:    "jq",
		AptName:     "jq",
		WingetID:    "jqlang.jq",
		ChocoName:   "jq",
		BinaryName:  "jq",
	},
	{
		ID:          "bat",
		Name:        "Bat",
		Description: "A cat clone with syntax highlighting and Git integration",
		Category:    model.CategoryCLITools,
		Type:        model.TypeBrewFormula,
		BrewName:    "bat",
		AptName:     "bat",
		WingetID:    "sharkdp.bat",
		ChocoName:   "bat",
		BinaryName:  "bat",
	},
}
