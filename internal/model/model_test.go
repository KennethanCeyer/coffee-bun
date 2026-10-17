package model

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPackageString(t *testing.T) {
	pkgInstalledWithVer := Package{
		Name:      "Go",
		Installed: true,
		Version:   "1.26.0",
	}
	if str := pkgInstalledWithVer.String(); str != "Go (1.26.0)" {
		t.Errorf("expected 'Go (1.26.0)', got %q", str)
	}

	pkgInstalledNoVer := Package{
		Name:      "Git",
		Installed: true,
	}
	if str := pkgInstalledNoVer.String(); str != "Git (installed)" {
		t.Errorf("expected 'Git (installed)', got %q", str)
	}

	pkgMissing := Package{
		Name:      "Azure CLI",
		Installed: false,
	}
	if str := pkgMissing.String(); str != "Azure CLI (missing)" {
		t.Errorf("expected 'Azure CLI (missing)', got %q", str)
	}
}

func TestIsSupportedOnAndPlatformHelpers(t *testing.T) {
	allOSPkg := Package{ID: "git"}
	if !allOSPkg.IsSupportedOn("darwin") || !allOSPkg.IsSupportedOn("linux") || !allOSPkg.IsSupportedOn("windows") {
		t.Errorf("expected package with empty SupportedOS to support all OSes")
	}

	darwinOnlyPkg := Package{ID: "iterm2", SupportedOS: []string{"darwin"}}
	if !darwinOnlyPkg.IsSupportedOn("darwin") || darwinOnlyPkg.IsSupportedOn("linux") || darwinOnlyPkg.IsSupportedOn("windows") {
		t.Errorf("expected darwinOnlyPkg to only support darwin")
	}

	cmd := ShellCommand("echo hello")
	if out, err := cmd.Output(); err != nil || !strings.Contains(string(out), "hello") {
		t.Errorf("ShellCommand failed: err=%v out=%q", err, string(out))
	}

	cmdCtx := ShellCommandContext(context.Background(), "echo world")
	if out, err := cmdCtx.Output(); err != nil || !strings.Contains(string(out), "world") {
		t.Errorf("ShellCommandContext failed: err=%v out=%q", err, string(out))
	}

	fontsDir := DefaultFontsDir("/tmp/home")
	if fontsDir == "" {
		t.Errorf("expected non-empty DefaultFontsDir")
	}
}

func TestModelJSONAndYAML(t *testing.T) {
	pkg := Package{
		ID:          "docker",
		Name:        "Docker CLI",
		Description: "Container tool",
		Category:    CategoryCLITools,
		Type:        TypeBrewFormula,
		BrewName:    "docker",
		AptName:     "docker.io",
		WingetID:    "Docker.DockerCLI",
		ChocoName:   "docker-cli",
		BinaryName:  "docker",
		CheckCmd:    "docker --version",
		Installed:   true,
		Version:     "26.0.0",
	}

	// Test JSON
	jsonData, err := json.Marshal(pkg)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var unmarshaledJSON Package
	if err := json.Unmarshal(jsonData, &unmarshaledJSON); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}
	if unmarshaledJSON.ID != "docker" || unmarshaledJSON.Category != CategoryCLITools {
		t.Errorf("JSON mismatch: %+v", unmarshaledJSON)
	}

	// Test YAML Config
	cfg := Config{
		Version: "1.0",
		Hooks: Hooks{
			PreInstall:  "echo pre",
			PostInstall: "echo post",
		},
		Groups: []CustomGroup{
			{
				Category:    Category("Acme Tools"),
				Title:       "Acme Tools",
				Description: "Internal suite",
				PackageIDs:  []string{"docker"},
			},
		},
		Packages: []Package{pkg},
	}

	yamlData, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}
	if !strings.Contains(string(yamlData), "Acme Tools") {
		t.Errorf("YAML missing content: %s", string(yamlData))
	}

	var unmarshaledYAML Config
	if err := yaml.Unmarshal(yamlData, &unmarshaledYAML); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}
	if unmarshaledYAML.Version != "1.0" || len(unmarshaledYAML.Packages) != 1 {
		t.Errorf("YAML mismatch: %+v", unmarshaledYAML)
	}
}

func TestCategoryConstants(t *testing.T) {
	cats := []Category{
		CategorySystem,
		CategoryTerminal,
		CategoryCloud,
		CategoryVim,
		CategoryLanguages,
		CategoryCLITools,
		CategoryApps,
	}
	for _, c := range cats {
		if string(c) == "" {
			t.Errorf("category constant is empty: %v", c)
		}
	}
}

func TestStatusConstants(t *testing.T) {
	statuses := []InstallStatus{
		StatusPending,
		StatusInstalling,
		StatusWaitingAction,
		StatusSuccess,
		StatusFailed,
		StatusSkipped,
	}
	for _, s := range statuses {
		if string(s) == "" {
			t.Errorf("status constant is empty: %v", s)
		}
	}
}
