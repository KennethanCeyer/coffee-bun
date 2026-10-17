package packages

import "github.com/KennethanCeyer/coffee-bun/internal/model"

/*
Domain: Essential Desktop Apps & Settings
Description:
    Core desktop applications and initial operating system preferences needed
    for a modern developer workstation across macOS, Linux, and Windows.

Components:
    1. Google Chrome: Fast, multi-profile desktop web browser with developer tools.
    2. Set Chrome as Default (macOS): Automates default HTTP/HTTPS protocol handler setup.
*/

// AppsGroup defines the UI presentation and grouping metadata for desktop applications
var AppsGroup = model.CustomGroup{
	Category:    model.CategoryApps,
	Title:       "Essential Desktop Apps & Settings",
	Description: "Google Chrome browser and default browser configuration",
	Icon:        "[app]",
	PackageIDs: []string{
		"chrome", "chrome-default",
	},
}

// AppsPackages defines declarative specifications for desktop applications and settings
var AppsPackages = []model.Package{
	{
		ID:          "chrome",
		Name:        "Google Chrome",
		Description: "Fast, reliable web browser built for modern web development",
		Category:    model.CategoryApps,
		Type:        model.TypeBrewCask,
		BrewName:    "google-chrome",
		WingetID:    "Google.Chrome",
		ChocoName:   "googlechrome",
		IsCask:      true,
		BinaryName:  "google-chrome",
		AppPath:     "/Applications/Google Chrome.app",
	},
	{
		ID:           "chrome-default",
		Name:         "Set Chrome as Default",
		Description:  "Configure Google Chrome as macOS default web browser",
		Category:     model.CategoryApps,
		Type:         model.TypeCustom,
		SupportedOS:  []string{"darwin"},
		Dependencies: []string{"chrome"},
	},
}
