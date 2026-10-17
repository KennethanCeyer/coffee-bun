package packages

import (
	"runtime"

	"github.com/KennethanCeyer/coffee-bun/internal/config"
	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

// PackageGroup represents a folding category of packages in the tree UI
type PackageGroup = model.CustomGroup

// GetPackageGroups returns the categorized group structure filtered for the current OS
func GetPackageGroups() []PackageGroup {
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
	groups, _ := FilterPackagesForOS(rawGroups, rawPkgs, runtime.GOOS)
	return groups
}

// GetDefaultPackages returns the consolidated list of developer packages filtered for the current OS
func GetDefaultPackages() []model.Package {
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
	_, pkgs := FilterPackagesForOS(rawGroups, rawPkgs, runtime.GOOS)
	return pkgs
}

func getAllDomainPackages() []model.Package {
	var pkgs []model.Package
	pkgs = append(pkgs, SystemPackages...)
	pkgs = append(pkgs, GetTerminalPackages()...)
	pkgs = append(pkgs, CloudPackages...)
	pkgs = append(pkgs, GetEditorsPackages()...)
	pkgs = append(pkgs, LanguagesPackages...)
	pkgs = append(pkgs, CLIPackages...)
	pkgs = append(pkgs, AppsPackages...)
	return pkgs
}

// FilterPackagesForOS filters packages and group PackageIDs to only those supported on goos
func FilterPackagesForOS(groups []PackageGroup, pkgs []model.Package, goos string) ([]PackageGroup, []model.Package) {
	supportedMap := make(map[string]bool, len(pkgs))
	var filteredPkgs []model.Package

	for _, p := range pkgs {
		if p.IsSupportedOn(goos) {
			supportedMap[p.ID] = true
			filteredPkgs = append(filteredPkgs, p)
		}
	}

	var filteredGroups []PackageGroup
	for _, g := range groups {
		var validIDs []string
		for _, id := range g.PackageIDs {
			if supportedMap[id] {
				validIDs = append(validIDs, id)
			}
		}
		if len(validIDs) > 0 {
			gCopy := g
			gCopy.PackageIDs = validIDs
			filteredGroups = append(filteredGroups, gCopy)
		}
	}

	return filteredGroups, filteredPkgs
}

// GetEffectivePackages merges core domain packages with any discovered custom configuration
func GetEffectivePackages(explicitConfigPath string) ([]PackageGroup, []model.Package, *model.Config, string) {
	defaultGroups := GetPackageGroups()
	defaultPkgs := GetDefaultPackages()

	configPath := config.FindConfigFile(explicitConfigPath)
	if configPath == "" {
		return defaultGroups, defaultPkgs, nil, ""
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		// If error reading config, fallback to defaults
		return defaultGroups, defaultPkgs, nil, configPath
	}

	mergedGroups, mergedPkgs := config.MergeWithDefaults(defaultGroups, defaultPkgs, cfg)
	filteredGroups, filteredPkgs := FilterPackagesForOS(mergedGroups, mergedPkgs, runtime.GOOS)
	return filteredGroups, filteredPkgs, cfg, configPath
}
