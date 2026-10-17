package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

// DefaultConfigCandidates returns the standard file paths checked for customization
func DefaultConfigCandidates() []string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		// System / MDM provisioned configuration
		"/etc/coffee-bun/bun.yaml",
		"/etc/coffee-bun/config.yaml",
		// User personal configuration
		filepath.Join(home, ".config", "coffee-bun", "bun.yaml"),
		filepath.Join(home, ".config", "coffee-bun", "config.yaml"),
		filepath.Join(home, ".bun.yaml"),
		filepath.Join(home, ".coffee-bun.yaml"),
		// Project / Repository workspace configuration
		"./bun.yaml",
		"./.bun.yaml",
		"./coffee-bun.yaml",
	}
	return candidates
}

// FindConfigFile locates the highest-priority configuration file, if any exists
func FindConfigFile(explicitPath string) string {
	if explicitPath != "" {
		if _, err := os.Stat(explicitPath); err == nil {
			return explicitPath
		}
		return ""
	}

	if envPath := os.Getenv("BUN_CONFIG"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			return envPath
		}
	}

	// Check candidate paths in reverse order (project takes priority over user, user over system)
	candidates := []string{
		"./bun.yaml",
		"./.bun.yaml",
		"./coffee-bun.yaml",
	}

	home, _ := os.UserHomeDir()
	if home != "" {
		candidates = append(candidates,
			filepath.Join(home, ".config", "coffee-bun", "bun.yaml"),
			filepath.Join(home, ".config", "coffee-bun", "config.yaml"),
			filepath.Join(home, ".bun.yaml"),
			filepath.Join(home, ".coffee-bun.yaml"),
		)
	}

	candidates = append(candidates,
		"/etc/coffee-bun/bun.yaml",
		"/etc/coffee-bun/config.yaml",
	)

	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path
		}
	}

	return ""
}

// LoadConfig reads and parses a YAML configuration file from disk
func LoadConfig(path string) (*model.Config, error) {
	if path == "" {
		return nil, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg model.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", path, err)
	}

	// Expand paths in packages
	home, _ := os.UserHomeDir()
	for i := range cfg.Packages {
		p := &cfg.Packages[i]
		if strings.HasPrefix(p.InstallPath, "~/") && home != "" {
			p.InstallPath = filepath.Join(home, p.InstallPath[2:])
		}
		if strings.HasPrefix(p.CheckFile, "~/") && home != "" {
			p.CheckFile = filepath.Join(home, p.CheckFile[2:])
		}
		if strings.HasPrefix(p.AppPath, "~/") && home != "" {
			p.AppPath = filepath.Join(home, p.AppPath[2:])
		}
	}

	return &cfg, nil
}

// MergeWithDefaults merges user-defined configuration with default core packages and groups
func MergeWithDefaults(
	defaultGroups []model.CustomGroup,
	defaultPkgs []model.Package,
	cfg *model.Config,
) ([]model.CustomGroup, []model.Package) {
	if cfg == nil || (len(cfg.Packages) == 0 && len(cfg.Groups) == 0) {
		return defaultGroups, defaultPkgs
	}

	mergedPkgs := make([]model.Package, len(defaultPkgs))
	copy(mergedPkgs, defaultPkgs)

	pkgIndexMap := make(map[string]int)
	for i, p := range mergedPkgs {
		pkgIndexMap[p.ID] = i
	}

	var customOrder []string

	// Merge packages: override existing or append new
	for _, customPkg := range cfg.Packages {
		if idx, exists := pkgIndexMap[customPkg.ID]; exists {
			// Override existing core package definition with custom settings
			mergedPkgs[idx] = overridePackage(mergedPkgs[idx], customPkg)
		} else {
			if customPkg.Category == "" {
				customPkg.Category = model.Category("Company & Custom Tools")
			}
			mergedPkgs = append(mergedPkgs, customPkg)
			pkgIndexMap[customPkg.ID] = len(mergedPkgs) - 1
		}
		customOrder = append(customOrder, customPkg.ID)
	}

	mergedGroups := make([]model.CustomGroup, len(defaultGroups))
	copy(mergedGroups, defaultGroups)

	groupCatMap := make(map[model.Category]int)
	for i, g := range mergedGroups {
		groupCatMap[g.Category] = i
	}

	// Merge custom groups
	for _, customGroup := range cfg.Groups {
		if idx, exists := groupCatMap[customGroup.Category]; exists {
			// Append package IDs to existing group
			for _, pid := range customGroup.PackageIDs {
				if !containsString(mergedGroups[idx].PackageIDs, pid) {
					mergedGroups[idx].PackageIDs = append(mergedGroups[idx].PackageIDs, pid)
				}
			}
			if customGroup.Title != "" {
				mergedGroups[idx].Title = customGroup.Title
			}
			if customGroup.Description != "" {
				mergedGroups[idx].Description = customGroup.Description
			}
		} else {
			// New group
			if customGroup.Icon == "" {
				customGroup.Icon = "[custom]"
			}
			if customGroup.Title == "" {
				customGroup.Title = string(customGroup.Category)
			}
			mergedGroups = append(mergedGroups, customGroup)
			groupCatMap[customGroup.Category] = len(mergedGroups) - 1
		}
	}

	// Ensure any custom package that isn't yet assigned to a group is placed into its category group
	for _, customID := range customOrder {
		inGroup := false
		for _, g := range mergedGroups {
			if containsString(g.PackageIDs, customID) {
				inGroup = true
				break
			}
		}
		if !inGroup {
			pkg := mergedPkgs[pkgIndexMap[customID]]
			targetCat := pkg.Category
			if targetCat == "" {
				targetCat = model.Category("Company & Custom Tools")
				mergedPkgs[pkgIndexMap[customID]].Category = targetCat
			}
			if idx, exists := groupCatMap[targetCat]; exists {
				if !containsString(mergedGroups[idx].PackageIDs, customID) {
					mergedGroups[idx].PackageIDs = append(mergedGroups[idx].PackageIDs, customID)
				}
			} else {
				mergedGroups = append(mergedGroups, model.CustomGroup{
					Category:    targetCat,
					Title:       string(targetCat),
					Description: "Organization-specific or project-defined developer tools",
					Icon:        "[custom]",
					PackageIDs:  []string{customID},
				})
				groupCatMap[targetCat] = len(mergedGroups) - 1
			}
		}
	}

	return mergedGroups, mergedPkgs
}

func overridePackage(base model.Package, override model.Package) model.Package {
	if override.Name != "" {
		base.Name = override.Name
	}
	if override.Description != "" {
		base.Description = override.Description
	}
	if override.Category != "" {
		base.Category = override.Category
	}
	if override.Type != "" {
		base.Type = override.Type
	}
	if len(override.SupportedOS) > 0 {
		base.SupportedOS = override.SupportedOS
	}
	if override.BrewName != "" {
		base.BrewName = override.BrewName
	}
	if override.AptName != "" {
		base.AptName = override.AptName
	}
	if override.WingetID != "" {
		base.WingetID = override.WingetID
	}
	if override.ChocoName != "" {
		base.ChocoName = override.ChocoName
	}
	if override.IsCask {
		base.IsCask = true
	}
	if override.BinaryName != "" {
		base.BinaryName = override.BinaryName
	}
	if override.AppPath != "" {
		base.AppPath = override.AppPath
	}
	if override.GitURL != "" {
		base.GitURL = override.GitURL
	}
	if override.InstallPath != "" {
		base.InstallPath = override.InstallPath
	}
	if override.CheckCmd != "" {
		base.CheckCmd = override.CheckCmd
	}
	if override.CheckFile != "" {
		base.CheckFile = override.CheckFile
	}
	if override.InstallCmd != "" {
		base.InstallCmd = override.InstallCmd
	}
	if override.UpdateCmd != "" {
		base.UpdateCmd = override.UpdateCmd
	}
	if len(override.Env) > 0 {
		base.Env = override.Env
	}
	if len(override.Dependencies) > 0 {
		base.Dependencies = override.Dependencies
	}
	if override.RequiresSudo {
		base.RequiresSudo = true
	}
	if override.SudoSnippet != "" {
		base.SudoSnippet = override.SudoSnippet
	}
	return base
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
