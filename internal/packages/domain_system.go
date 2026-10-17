package packages

import "github.com/KennethanCeyer/coffee-bun/internal/model"

/*
Domain: System & Toolchains
Description:
    Foundational toolchains and package managers required on any developer
    workstation across macOS, Linux, and Windows.

Components:
    1. Xcode Command Line Tools (macOS): Provides Apple clang, git, make, and SDK headers.
    2. Build Essential (Linux): Provides GNU gcc, g++, make, and libc headers.
    3. Homebrew (macOS / Linux): Package manager for formulae and casks.
    4. Windows Package Manager (Windows): Native winget CLI for Windows software management.
    5. Git (All OS): Distributed version control system.
*/

// SystemGroup defines the UI presentation and grouping metadata for system tools
var SystemGroup = model.CustomGroup{
	Category:    model.CategorySystem,
	Title:       "System & Toolchains",
	Description: "Core compilers, system package managers, and Git",
	Icon:        "[sys]",
	PackageIDs:  []string{"xcode-select", "build-essential", "brew", "winget", "git"},
}

// SystemPackages defines declarative specifications for core system tools
var SystemPackages = []model.Package{
	{
		ID:           "xcode-select",
		Name:         "Xcode Command Line Tools",
		Description:  "Core compilers, make, headers, and git toolchain for macOS",
		Category:     model.CategorySystem,
		Type:         model.TypeCustom,
		SupportedOS:  []string{"darwin"},
		RequiresSudo: false,
		SudoSnippet:  "xcode-select --install",
	},
	{
		ID:          "build-essential",
		Name:        "Build Essential (C/C++ Make)",
		Description: "GNU C/C++ compilers, make, and development headers for Linux",
		Category:    model.CategorySystem,
		Type:        model.TypeBrewFormula,
		SupportedOS: []string{"linux"},
		AptName:     "build-essential",
		BinaryName:  "make",
	},
	{
		ID:           "brew",
		Name:         "Homebrew",
		Description:  "Package manager for macOS and Linux",
		Category:     model.CategorySystem,
		Type:         model.TypeCustom,
		SupportedOS:  []string{"darwin", "linux"},
		BinaryName:   "brew",
		RequiresSudo: true,
		SudoSnippet:  `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
	},
	{
		ID:          "winget",
		Name:        "Windows Package Manager (winget)",
		Description: "Official Microsoft command-line package manager for Windows",
		Category:    model.CategorySystem,
		Type:        model.TypeCustom,
		SupportedOS: []string{"windows"},
		BinaryName:  "winget",
	},
	{
		ID:         "git",
		Name:       "Git",
		Description: "Fast distributed version control system",
		Category:   model.CategorySystem,
		Type:       model.TypeBrewFormula,
		BrewName:   "git",
		AptName:    "git",
		WingetID:   "Git.Git",
		ChocoName:  "git",
		BinaryName: "git",
	},
}
