package packages

import "github.com/KennethanCeyer/coffee-bun/internal/model"

/*
Domain: Compilers & Runtimes
Description:
    Standard multi-language runtime environments and compilers essential for
    software engineering across macOS, Linux, and Windows.

Components:
    1. Go: High-performance compiled language with built-in concurrency.
    2. Python 3: General-purpose language interpreter with pip package installer.
    3. Java (OpenJDK): Open-source Java Platform Standard Edition development kit.
    4. GCC / Clang: GNU Compiler Collection and C/C++ toolchain.
    5. Node.js: Event-driven JavaScript runtime with npm package manager.
*/

// LanguagesGroup defines the UI presentation and grouping metadata for compilers and runtimes
var LanguagesGroup = model.CustomGroup{
	Category:    model.CategoryLanguages,
	Title:       "Compilers & Languages (Go, Python, Java, GCC, Node)",
	Description: "Go, Python 3 (+ pip), Java (OpenJDK), GCC / Clang, Node.js (+ npm)",
	Icon:        "[lang]",
	PackageIDs: []string{
		"go", "python", "java", "gcc", "node",
	},
}

// LanguagesPackages defines declarative specifications for programming languages and compilers
var LanguagesPackages = []model.Package{
	{
		ID:          "go",
		Name:        "Go",
		Description: "Open source programming language for fast and reliable software",
		Category:    model.CategoryLanguages,
		Type:        model.TypeBrewFormula,
		BrewName:    "go",
		AptName:     "golang-go",
		WingetID:    "GoLang.Go",
		ChocoName:   "golang",
		BinaryName:  "go",
	},
	{
		ID:          "python",
		Name:        "Python 3",
		Description: "Python interpreter bundled with pip",
		Category:    model.CategoryLanguages,
		Type:        model.TypeBrewFormula,
		BrewName:    "python",
		AptName:     "python3",
		WingetID:    "Python.Python.3.12",
		ChocoName:   "python",
		BinaryName:  "python3",
	},
	{
		ID:           "java",
		Name:         "Java (OpenJDK)",
		Description:  "Java Development Kit open source reference implementation",
		Category:     model.CategoryLanguages,
		Type:         model.TypeBrewFormula,
		BrewName:     "openjdk",
		AptName:      "default-jdk",
		WingetID:     "Microsoft.OpenJDK.21",
		ChocoName:    "openjdk",
		BinaryName:   "java",
		RequiresSudo: true,
		SudoSnippet:  "sudo ln -sfn /opt/homebrew/opt/openjdk/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk.jdk",
	},
	{
		ID:          "gcc",
		Name:        "GCC / Clang",
		Description: "GNU Compiler Collection & C/C++ command-line developer tools",
		Category:    model.CategoryLanguages,
		Type:        model.TypeBrewFormula,
		BrewName:    "gcc",
		AptName:     "gcc",
		WingetID:    "BrechtSanders.WinLibs.POSIX.UCRT",
		ChocoName:   "mingw",
		BinaryName:  "gcc",
	},
	{
		ID:          "node",
		Name:        "Node.js",
		Description: "JavaScript runtime built on Chrome's V8 engine with npm",
		Category:    model.CategoryLanguages,
		Type:        model.TypeBrewFormula,
		BrewName:    "node",
		AptName:     "nodejs",
		WingetID:    "OpenJS.NodeJS.LTS",
		ChocoName:   "nodejs-lts",
		BinaryName:  "node",
	},
}
