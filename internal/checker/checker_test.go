package checker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func TestCheckerCheckAll(t *testing.T) {
	c := NewChecker()
	pkgs := []model.Package{
		{
			ID:         "git",
			Name:       "Git",
			Type:       model.TypeBrewFormula,
			BinaryName: "git",
		},
		{
			ID:         "nonexistent-binary-12345",
			Name:       "Nonexistent",
			Type:       model.TypeBrewFormula,
			BinaryName: "nonexistent-binary-12345",
		},
	}

	results := c.CheckAll(pkgs)
	if len(results) != len(pkgs) {
		t.Fatalf("expected %d results, got %d", len(pkgs), len(results))
	}

	if !results[0].Installed {
		t.Errorf("expected git to be installed on system")
	}

	if results[1].Installed {
		t.Errorf("expected nonexistent binary to be reported as not installed")
	}
}

func TestCheckerIsBrewInstalled(t *testing.T) {
	c := NewChecker()
	if !c.IsBrewInstalled() {
		t.Logf("brew is not installed or not in PATH")
	}
}

func TestCheckerCustomCommandAndFile(t *testing.T) {
	c := NewChecker()
	pkg := model.Package{
		ID:       "custom-echo",
		Name:     "Custom Echo",
		Type:     model.TypeScript,
		CheckCmd: "echo 'v2.4.0'",
	}
	c.CheckPackage(&pkg)
	if !pkg.Installed {
		t.Errorf("expected custom check command to mark package installed")
	}
	if pkg.Version != "v2.4.0" {
		t.Errorf("expected version 'v2.4.0', got '%s'", pkg.Version)
	}

	failingPkg := model.Package{
		ID:       "custom-fail",
		Name:     "Custom Fail",
		Type:     model.TypeScript,
		CheckCmd: "false",
	}
	c.CheckPackage(&failingPkg)
	if failingPkg.Installed {
		t.Errorf("expected failing check command to mark package not installed")
	}

	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "installed.marker")
	_ = os.WriteFile(existingFile, []byte("ok"), 0644)

	filePkg := model.Package{
		ID:        "file-pkg",
		CheckFile: existingFile,
	}
	c.CheckPackage(&filePkg)
	if !filePkg.Installed {
		t.Errorf("expected filePkg to be installed via CheckFile")
	}
}

func TestCheckerCheckDirOrGit(t *testing.T) {
	c := NewChecker()
	tempDir := t.TempDir()

	pkg := model.Package{
		Type:        model.TypeGitRepo,
		InstallPath: tempDir,
	}
	c.checkDirOrGit(&pkg)
	if !pkg.Installed {
		t.Errorf("expected tempDir to be marked as installed directory")
	}

	nonexistentPkg := model.Package{
		Type:        model.TypeGitRepo,
		InstallPath: filepath.Join(tempDir, "does-not-exist"),
	}
	c.checkDirOrGit(&nonexistentPkg)
	if nonexistentPkg.Installed {
		t.Errorf("expected nonexistent directory to be marked as not installed")
	}
}

func TestCheckerCheckFont(t *testing.T) {
	c := NewChecker()
	tempDir := t.TempDir()
	fontFile := filepath.Join(tempDir, "TestFont.ttf")
	_ = os.WriteFile(fontFile, []byte("ttf"), 0644)

	pkg := model.Package{
		Type:        model.TypeFont,
		InstallPath: fontFile,
	}
	c.checkFont(&pkg)
	if !pkg.Installed {
		t.Errorf("expected font to be installed")
	}

	missingFontPkg := model.Package{
		Type:        model.TypeFont,
		InstallPath: filepath.Join(tempDir, "MissingFont.ttf"),
	}
	c.checkFont(&missingFontPkg)
	if missingFontPkg.Installed {
		t.Errorf("expected missing font to not be installed")
	}
}

func TestCheckerDetectBinaryVersion(t *testing.T) {
	c := NewChecker()

	bins := []string{"go", "git", "python3", "node", "gcc"}
	for _, bin := range bins {
		ver := c.detectBinaryVersion(bin)
		if ver == "" {
			t.Logf("binary %s not present or returned empty version", bin)
		} else {
			t.Logf("binary %s detected version: %s", bin, ver)
		}
	}

	nonVer := c.detectBinaryVersion("nonexistent-tool-9999")
	if nonVer != "" {
		t.Errorf("expected empty version for nonexistent tool, got %q", nonVer)
	}
}

func TestCheckerCheckCustom(t *testing.T) {
	c := NewChecker()

	for _, id := range []string{"xcode-select", "brew", "chrome-default", "code-cli", "brogrammer-theme"} {
		pkg := model.Package{
			ID:   id,
			Type: model.TypeCustom,
		}
		c.checkCustom(&pkg)
	}
}

func TestCheckerEnvironmentDiagnosticsIsolated(t *testing.T) {
	tmpHome := t.TempDir()
	c := NewCheckerWithHome(tmpHome)

	// Without .zshrc
	if c.IsP10kConfigured() {
		t.Errorf("expected IsP10kConfigured to be false when .zshrc does not exist")
	}
	if c.IsZoxideConfigured() {
		t.Errorf("expected IsZoxideConfigured to be false when .zshrc does not exist")
	}

	// Write .zshrc with p10k and zoxide
	zshrcPath := filepath.Join(tmpHome, ".zshrc")
	content := "ZSH_THEME=\"powerlevel10k/powerlevel10k\"\neval \"$(zoxide init zsh)\"\n"
	_ = os.WriteFile(zshrcPath, []byte(content), 0644)

	if !c.IsP10kConfigured() {
		t.Errorf("expected IsP10kConfigured to be true")
	}
	if !c.IsZoxideConfigured() {
		t.Errorf("expected IsZoxideConfigured to be true")
	}

	// Check VimRuntime isolated
	vimPkg := model.Package{
		ID:          "vim-runtime",
		Type:        model.TypeVimRuntime,
		InstallPath: filepath.Join(tmpHome, ".vim_runtime"),
	}
	c.checkVimRuntime(&vimPkg)
	if vimPkg.Installed {
		t.Errorf("expected vim-runtime to be missing in empty temp home")
	}

	// Create .vim_runtime and .vimrc with vim_runtime reference
	_ = os.MkdirAll(filepath.Join(tmpHome, ".vim_runtime"), 0755)
	_ = os.WriteFile(filepath.Join(tmpHome, ".vimrc"), []byte("source ~/.vim_runtime/vimrcs/basic.vim\n"), 0644)
	c.checkVimRuntime(&vimPkg)
	if !vimPkg.Installed {
		t.Errorf("expected vim-runtime to be detected as installed")
	}

	// Check ShellIntegration isolated
	shellPkg := model.Package{
		ID:   "shell-integration",
		Type: model.TypeCustom,
	}
	c.CheckPackage(&shellPkg)
	if !shellPkg.Installed {
		t.Errorf("expected shell-integration to be installed when both p10k and zoxide are configured")
	}

	// Check ITermConfig
	itermPkg := model.Package{
		ID:   "iterm-font-config",
		Type: model.TypeITermConfig,
	}
	c.CheckPackage(&itermPkg)
}

func TestApplyOutdatedJSON(t *testing.T) {
	c := NewCheckerWithHome(t.TempDir())
	pkgs := []model.Package{
		{
			ID:       "git",
			BrewName: "git",
		},
		{
			ID:       "iterm2",
			BrewName: "iterm2",
			IsCask:   true,
		},
	}

	rawJSON := []byte(`{
		"formulae": [
			{
				"name": "git",
				"installed_versions": ["2.42.0"],
				"current_version": "2.43.0"
			}
		],
		"casks": [
			{
				"name": "iterm2",
				"installed_versions": ["3.4.20"],
				"current_version": "3.4.23"
			}
		]
	}`)

	c.applyOutdatedJSON(pkgs, rawJSON)

	if !pkgs[0].Outdated || pkgs[0].LatestVersion != "2.43.0" {
		t.Errorf("expected git to be marked outdated with latest 2.43.0, got outdated=%v latest=%s", pkgs[0].Outdated, pkgs[0].LatestVersion)
	}
	if !pkgs[1].Outdated || pkgs[1].LatestVersion != "3.4.23" {
		t.Errorf("expected iterm2 to be marked outdated with latest 3.4.23, got outdated=%v latest=%s", pkgs[1].Outdated, pkgs[1].LatestVersion)
	}

	// Invalid JSON should not panic or mutate packages
	c.applyOutdatedJSON(pkgs, []byte("invalid json"))
}
