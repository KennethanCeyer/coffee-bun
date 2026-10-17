package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func TestDefaultConfigCandidates(t *testing.T) {
	candidates := DefaultConfigCandidates()
	if len(candidates) < 5 {
		t.Errorf("expected at least 5 default candidates, got %d", len(candidates))
	}
}

func TestFindConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	existingFile := filepath.Join(tempDir, "bun.yaml")
	_ = os.WriteFile(existingFile, []byte("version: '1.0'"), 0644)

	// Explicit path existing
	if found := FindConfigFile(existingFile); found != existingFile {
		t.Errorf("expected %s, got %s", existingFile, found)
	}

	// Explicit path not existing
	if found := FindConfigFile(filepath.Join(tempDir, "none.yaml")); found != "" {
		t.Errorf("expected empty string for missing explicit file, got %s", found)
	}

	// Env BUN_CONFIG
	t.Setenv("BUN_CONFIG", existingFile)
	if found := FindConfigFile(""); found != existingFile {
		t.Errorf("expected %s via BUN_CONFIG, got %s", existingFile, found)
	}

	// Nonexistent BUN_CONFIG
	t.Setenv("BUN_CONFIG", filepath.Join(tempDir, "missing-env.yaml"))
	_ = FindConfigFile("")
}

func TestLoadConfigErrors(t *testing.T) {
	// Empty path
	cfg, err := LoadConfig("")
	if cfg != nil || err != nil {
		t.Errorf("expected nil, nil for empty path")
	}

	// Nonexistent file
	_, err = LoadConfig("/path/to/definitely/nonexistent/file.yaml")
	if err == nil {
		t.Errorf("expected error for nonexistent file")
	}

	// Invalid YAML
	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.yaml")
	_ = os.WriteFile(invalidFile, []byte(":: invalid: yaml: { ["), 0644)

	_, err = LoadConfig(invalidFile)
	if err == nil {
		t.Errorf("expected error for invalid YAML")
	}
}

func TestLoadConfigPathExpansion(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "bun.yaml")

	yamlContent := `
version: "1.0"
hooks:
  pre_install: "echo pre"
  post_install: "echo post"
groups:
  - category: "Acme Internal"
    title: "Acme Developer Stack"
    description: "Internal tools for Acme Corp"
    packages:
      - acme-cli
packages:
  - id: acme-cli
    name: "Acme CLI"
    description: "Internal deployment CLI"
    category: "Acme Internal"
    type: "script"
    check_cmd: "acme --version"
    check_file: "~/acme.marker"
    install_path: "~/acme-dir"
    app_path: "~/Applications/Acme.app"
    install_cmd: "curl -fsSL https://example.com/install.sh | bash"
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", cfg.Version)
	}
	if cfg.Hooks.PreInstall != "echo pre" {
		t.Errorf("Expected pre_install 'echo pre', got %s", cfg.Hooks.PreInstall)
	}
	if len(cfg.Packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(cfg.Packages))
	}
	pkg := cfg.Packages[0]
	if filepath.IsAbs(pkg.InstallPath) == false {
		t.Errorf("expected expanded install path, got %s", pkg.InstallPath)
	}
}

func TestMergeWithDefaultsAndFullOverride(t *testing.T) {
	defaultGroups := []model.CustomGroup{
		{
			Category:   model.CategorySystem,
			Title:      "System & Toolchains",
			PackageIDs: []string{"git"},
		},
	}
	defaultPkgs := []model.Package{
		{
			ID:         "git",
			Name:       "Git",
			Category:   model.CategorySystem,
			Type:       model.TypeBrewFormula,
			BrewName:   "git",
			BinaryName: "git",
		},
	}

	// Nil config
	g1, p1 := MergeWithDefaults(defaultGroups, defaultPkgs, nil)
	if len(g1) != len(defaultGroups) || len(p1) != len(defaultPkgs) {
		t.Errorf("expected unchanged slices for nil config")
	}

	// Empty config
	g2, p2 := MergeWithDefaults(defaultGroups, defaultPkgs, &model.Config{})
	if len(g2) != len(defaultGroups) || len(p2) != len(defaultPkgs) {
		t.Errorf("expected unchanged slices for empty config")
	}

	// Full override of existing package + unassigned custom package
	cfg := &model.Config{
		Groups: []model.CustomGroup{
			{
				Category:    model.CategorySystem,
				Title:       "Custom System Title",
				Description: "Custom System Desc",
				Icon:        "[S]",
				PackageIDs:  []string{"extra-sys"},
			},
		},
		Packages: []model.Package{
			{
				ID:           "git",
				Name:         "Enterprise Git",
				Description:  "Custom Git build",
				Category:     model.CategorySystem,
				Type:         model.TypeScript,
				SupportedOS:  []string{"darwin", "linux"},
				BrewName:     "git-enterprise",
				AptName:      "git-ent-apt",
				WingetID:     "Enterprise.Git",
				ChocoName:    "git-enterprise",
				BinaryName:   "git-ent",
				AppPath:      "/Applications/Git.app",
				GitURL:       "https://example.com/git.git",
				InstallPath:  "/opt/git/bin",
				CheckCmd:     "git --version",
				CheckFile:    "/opt/git/bin/git",
				InstallCmd:   "make install",
				UpdateCmd:    "make update",
				Env:          map[string]string{"GIT_ENV": "1"},
				Dependencies: []string{"xcode-select"},
				RequiresSudo: true,
				SudoSnippet:  "sudo make install",
			},
			{
				ID:          "orphan-pkg",
				Name:        "Orphan Package",
				Description: "Package with no explicit group",
				Type:        model.TypeScript,
				InstallCmd:  "echo orphan",
			},
		},
	}

	g3, p3 := MergeWithDefaults(defaultGroups, defaultPkgs, cfg)
	if len(p3) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(p3))
	}
	if p3[0].Name != "Enterprise Git" || p3[0].InstallCmd != "make install" || !p3[0].RequiresSudo {
		t.Errorf("expected git package to be overridden, got %+v", p3[0])
	}
	if len(g3) != 2 {
		t.Fatalf("expected 2 groups (default + Company & Custom Tools), got %d", len(g3))
	}
	if g3[0].Title != "Custom System Title" {
		t.Errorf("expected overridden group title, got %s", g3[0].Title)
	}
	if g3[1].Title != "Company & Custom Tools" {
		t.Errorf("expected 'Company & Custom Tools' group, got %s", g3[1].Title)
	}
}
