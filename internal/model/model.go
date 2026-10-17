package model

import "fmt"

type Category string

const (
	CategorySystem    Category = "System & Toolchains"
	CategoryTerminal  Category = "Terminal & Shell"
	CategoryCloud     Category = "Cloud & DevOps CLIs"
	CategoryVim       Category = "Editors & Vim Runtime"
	CategoryLanguages Category = "Compilers & Runtimes"
	CategoryCLITools  Category = "Modern CLI Utilities"
	CategoryApps      Category = "Essential Desktop Apps"
)

type PackageType string

const (
	TypeBrewFormula PackageType = "brew_formula"
	TypeBrewCask    PackageType = "brew_cask"
	TypeGitRepo     PackageType = "git_repo"
	TypeFont        PackageType = "font"
	TypeITermConfig PackageType = "iterm_config"
	TypeVimRuntime  PackageType = "vim_runtime"
	TypeScript      PackageType = "script"
	TypeCommand     PackageType = "command"
	TypeCustom      PackageType = "custom"
)

type InstallStatus string

const (
	StatusPending       InstallStatus = "pending"
	StatusInstalling    InstallStatus = "installing"
	StatusWaitingAction InstallStatus = "waiting_action"
	StatusSuccess       InstallStatus = "success"
	StatusFailed        InstallStatus = "failed"
	StatusSkipped       InstallStatus = "skipped"
)

// Package represents an installable software component across macOS, Linux, and Windows
type Package struct {
	ID            string            `json:"id" yaml:"id"`
	Name          string            `json:"name" yaml:"name"`
	Description   string            `json:"description" yaml:"description"`
	Category      Category          `json:"category" yaml:"category"`
	Type          PackageType       `json:"type" yaml:"type"`
	SupportedOS   []string          `json:"supported_os,omitempty" yaml:"supported_os,omitempty"`
	BrewName      string            `json:"brew_name,omitempty" yaml:"brew_name,omitempty"`
	AptName       string            `json:"apt_name,omitempty" yaml:"apt_name,omitempty"`
	WingetID      string            `json:"winget_id,omitempty" yaml:"winget_id,omitempty"`
	ChocoName     string            `json:"choco_name,omitempty" yaml:"choco_name,omitempty"`
	IsCask        bool              `json:"is_cask,omitempty" yaml:"is_cask,omitempty"`
	BinaryName    string            `json:"binary_name,omitempty" yaml:"binary_name,omitempty"`
	AppPath       string            `json:"app_path,omitempty" yaml:"app_path,omitempty"`
	GitURL        string            `json:"git_url,omitempty" yaml:"git_url,omitempty"`
	InstallPath   string            `json:"install_path,omitempty" yaml:"install_path,omitempty"`
	CheckCmd      string            `json:"check_cmd,omitempty" yaml:"check_cmd,omitempty"`
	CheckFile     string            `json:"check_file,omitempty" yaml:"check_file,omitempty"`
	InstallCmd    string            `json:"install_cmd,omitempty" yaml:"install_cmd,omitempty"`
	UpdateCmd     string            `json:"update_cmd,omitempty" yaml:"update_cmd,omitempty"`
	Env           map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Dependencies  []string          `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	RequiresSudo  bool              `json:"requires_sudo,omitempty" yaml:"requires_sudo,omitempty"`
	SudoSnippet   string            `json:"sudo_snippet,omitempty" yaml:"sudo_snippet,omitempty"`
	Installed     bool              `json:"installed" yaml:"installed"`
	Version       string            `json:"version,omitempty" yaml:"version,omitempty"`
	Outdated      bool              `json:"outdated" yaml:"outdated"`
	LatestVersion string            `json:"latest_version,omitempty" yaml:"latest_version,omitempty"`
}

// IsSupportedOn returns true if the package supports the given GOOS ("darwin", "linux", "windows")
func (p Package) IsSupportedOn(goos string) bool {
	if len(p.SupportedOS) == 0 {
		return true
	}
	for _, osName := range p.SupportedOS {
		if osName == goos {
			return true
		}
	}
	return false
}

// Config represents enterprise or user customization configuration
type Config struct {
	Version  string        `json:"version" yaml:"version"`
	Groups   []CustomGroup `json:"groups,omitempty" yaml:"groups,omitempty"`
	Packages []Package     `json:"packages,omitempty" yaml:"packages,omitempty"`
	Hooks    Hooks         `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// CustomGroup represents a custom category group defined in configuration
type CustomGroup struct {
	Category    Category `json:"category" yaml:"category"`
	Title       string   `json:"title" yaml:"title"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
	Icon        string   `json:"icon,omitempty" yaml:"icon,omitempty"`
	PackageIDs  []string `json:"packages,omitempty" yaml:"packages,omitempty"`
}

// Hooks specifies shell commands to execute before and after the installation pipeline
type Hooks struct {
	PreInstall  string `json:"pre_install,omitempty" yaml:"pre_install,omitempty"`
	PostInstall string `json:"post_install,omitempty" yaml:"post_install,omitempty"`
}

func (p Package) String() string {
	if p.Installed {
		if p.Version != "" {
			return fmt.Sprintf("%s (%s)", p.Name, p.Version)
		}
		return fmt.Sprintf("%s (installed)", p.Name)
	}
	return fmt.Sprintf("%s (missing)", p.Name)
}

// InstallEvent carries real-time progress events from the installer engine
type InstallEvent struct {
	PackageID     string
	Name          string
	Status        InstallStatus
	Message       string
	LogLine       string
	ActionSnippet string
	Error         error
}
