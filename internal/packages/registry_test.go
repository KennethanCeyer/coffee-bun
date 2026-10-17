package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultPackages(t *testing.T) {
	pkgs := GetDefaultPackages()
	if len(pkgs) == 0 {
		t.Fatalf("expected non-empty packages list")
	}

	// Universal packages supported across darwin, linux, and windows
	universalIDs := []string{
		"git", "meslo-font", "aws", "gcloud", "azure", "vim-runtime",
		"go", "python", "java", "gcc", "node", "zoxide", "fzf",
		"ripgrep", "eza", "jq", "bat", "vscode", "chrome",
	}

	pkgMap := make(map[string]bool)
	for _, p := range pkgs {
		pkgMap[p.ID] = true
	}

	for _, id := range universalIDs {
		if !pkgMap[id] {
			t.Errorf("expected universal package %s in default packages", id)
		}
	}
}

func TestFilterPackagesForOS(t *testing.T) {
	rawGroups := []PackageGroup{
		SystemGroup,
		TerminalGroup,
		CloudGroup,
		EditorsGroup,
		LanguagesGroup,
		CLIGroup,
		AppsGroup,
	}
	rawPkgs := getAllDomainPackages()

	// 1. Test darwin filtering
	_, darwinPkgs := FilterPackagesForOS(rawGroups, rawPkgs, "darwin")
	darwinMap := make(map[string]bool)
	for _, p := range darwinPkgs {
		darwinMap[p.ID] = true
	}
	if !darwinMap["xcode-select"] || !darwinMap["iterm2"] || darwinMap["winget"] || darwinMap["build-essential"] {
		t.Errorf("darwin package filtering failed: %+v", darwinMap)
	}

	// 2. Test linux filtering
	_, linuxPkgs := FilterPackagesForOS(rawGroups, rawPkgs, "linux")
	linuxMap := make(map[string]bool)
	for _, p := range linuxPkgs {
		linuxMap[p.ID] = true
	}
	if !linuxMap["build-essential"] || !linuxMap["oh-my-zsh"] || linuxMap["xcode-select"] || linuxMap["iterm2"] || linuxMap["winget"] {
		t.Errorf("linux package filtering failed: %+v", linuxMap)
	}

	// 3. Test windows filtering
	_, winPkgs := FilterPackagesForOS(rawGroups, rawPkgs, "windows")
	winMap := make(map[string]bool)
	for _, p := range winPkgs {
		winMap[p.ID] = true
	}
	if !winMap["winget"] || !winMap["windows-terminal"] || winMap["xcode-select"] || winMap["oh-my-zsh"] || winMap["build-essential"] {
		t.Errorf("windows package filtering failed: %+v", winMap)
	}
}

func TestGetPackageGroups(t *testing.T) {
	groups := GetPackageGroups()
	if len(groups) == 0 {
		t.Fatalf("expected non-empty package groups")
	}

	totalItems := 0
	for _, g := range groups {
		if len(g.PackageIDs) == 0 {
			t.Errorf("group %s has no package IDs", g.Title)
		}
		totalItems += len(g.PackageIDs)
	}

	if totalItems < 18 {
		t.Errorf("expected at least 18 items across groups, got %d", totalItems)
	}
}

func TestGetEffectivePackages(t *testing.T) {
	// Test without config
	groups, pkgs, cfg, loadedPath := GetEffectivePackages("")
	if len(groups) == 0 || len(pkgs) == 0 {
		t.Fatalf("expected default packages when no config provided")
	}
	_ = cfg
	_ = loadedPath

	// Test with valid custom config
	tempDir := t.TempDir()
	customPath := filepath.Join(tempDir, "bun.yaml")
	yamlData := `
version: "1.0"
packages:
  - id: custom-tool
    name: "Custom Tool"
    type: "script"
    check_cmd: "echo 1.0"
`
	if err := os.WriteFile(customPath, []byte(yamlData), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}

	mergedGroups, mergedPkgs, customCfg, path := GetEffectivePackages(customPath)
	if path != customPath {
		t.Errorf("expected path to be %s, got %s", customPath, path)
	}
	if customCfg == nil {
		t.Errorf("expected customCfg to be loaded")
	}

	var foundCustom bool
	for _, p := range mergedPkgs {
		if p.ID == "custom-tool" {
			foundCustom = true
			break
		}
	}
	if !foundCustom {
		t.Errorf("expected custom-tool in merged packages")
	}
	if len(mergedGroups) < len(groups) {
		t.Errorf("expected groups to not decrease")
	}

	// Test with invalid config path fallback
	_, _, fallbackCfg, _ := GetEffectivePackages(filepath.Join(tempDir, "nonexistent.yaml"))
	if fallbackCfg != nil {
		t.Errorf("expected nil config for nonexistent path")
	}
}
