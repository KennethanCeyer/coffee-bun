package checker

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

// Checker handles system state inspection across macOS, Linux, and Windows
type Checker struct {
	homeDir string
}

// NewChecker creates a new Checker instance using the user's home directory
func NewChecker() *Checker {
	home, _ := os.UserHomeDir()
	return &Checker{homeDir: home}
}

// NewCheckerWithHome creates a Checker scoped to a specific home directory (useful for isolated testing)
func NewCheckerWithHome(homeDir string) *Checker {
	return &Checker{homeDir: homeDir}
}

func (c *Checker) getHomeDir() string {
	if c.homeDir != "" {
		return c.homeDir
	}
	home, _ := os.UserHomeDir()
	return home
}

// IsBrewInstalled returns true if homebrew binary is found
func (c *Checker) IsBrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

// CheckPackage inspects a single package and updates Installed and Version fields
func (c *Checker) CheckPackage(pkg *model.Package) {
	// 1. If explicit check command is specified, execute it via OS-appropriate shell
	if pkg.CheckCmd != "" {
		cmd := model.ShellCommand(pkg.CheckCmd)
		out, err := cmd.CombinedOutput()
		if err == nil {
			pkg.Installed = true
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			if len(lines) > 0 && lines[0] != "" {
				ver := strings.TrimSpace(lines[0])
				if len(ver) > 28 {
					ver = ver[:25] + "..."
				}
				pkg.Version = ver
			} else {
				pkg.Version = "ready"
			}
			return
		}
		pkg.Installed = false
		return
	}

	// 2. If explicit check file or directory is specified, verify existence
	if pkg.CheckFile != "" {
		if _, err := os.Stat(pkg.CheckFile); err == nil {
			pkg.Installed = true
			pkg.Version = "ready"
			return
		}
		pkg.Installed = false
		return
	}

	switch pkg.Type {
	case model.TypeBrewFormula:
		c.checkFormulaOrBinary(pkg)
	case model.TypeBrewCask:
		c.checkCaskOrApp(pkg)
	case model.TypeGitRepo:
		c.checkDirOrGit(pkg)
	case model.TypeFont:
		c.checkFont(pkg)
	case model.TypeITermConfig:
		c.checkITermFontConfig(pkg)
	case model.TypeVimRuntime:
		c.checkVimRuntime(pkg)
	case model.TypeScript, model.TypeCommand:
		c.checkScriptOrCommand(pkg)
	case model.TypeCustom:
		c.checkCustom(pkg)
	default:
		c.checkFormulaOrBinary(pkg)
	}
}

func (c *Checker) checkScriptOrCommand(pkg *model.Package) {
	if pkg.BinaryName != "" {
		if path, err := exec.LookPath(pkg.BinaryName); err == nil && path != "" {
			pkg.Installed = true
			pkg.Version = c.detectBinaryVersion(pkg.BinaryName)
			return
		}
	}
	if pkg.AppPath != "" {
		if _, err := os.Stat(pkg.AppPath); err == nil {
			pkg.Installed = true
			pkg.Version = c.detectAppVersion(pkg.AppPath)
			return
		}
	}
	pkg.Installed = false
}

// CheckAll checks all packages concurrently using goroutines
func (c *Checker) CheckAll(pkgs []model.Package) []model.Package {
	var wg sync.WaitGroup
	result := make([]model.Package, len(pkgs))
	copy(result, pkgs)

	for i := range result {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c.CheckPackage(&result[idx])
		}(i)
	}

	wg.Wait()
	return result
}

// CheckAllWithUpdates inspects installation state and checks for available package updates concurrently
func (c *Checker) CheckAllWithUpdates(pkgs []model.Package) []model.Package {
	var wg sync.WaitGroup
	var result []model.Package
	var outdatedMap map[string]string

	wg.Add(2)
	go func() {
		defer wg.Done()
		result = c.CheckAll(pkgs)
	}()
	go func() {
		defer wg.Done()
		outdatedMap = c.fetchOutdatedMap()
	}()
	wg.Wait()

	for i := range result {
		if !result[i].Installed {
			continue
		}
		if result[i].BrewName != "" {
			if latest, ok := outdatedMap[result[i].BrewName]; ok {
				result[i].Outdated = true
				result[i].LatestVersion = latest
				continue
			}
			// Also check base formula name (e.g. tap/repo/name -> name)
			parts := strings.Split(result[i].BrewName, "/")
			if base := parts[len(parts)-1]; base != "" {
				if latest, ok := outdatedMap[base]; ok {
					result[i].Outdated = true
					result[i].LatestVersion = latest
					continue
				}
			}
		}
		if result[i].AptName != "" {
			if latest, ok := outdatedMap[result[i].AptName]; ok {
				result[i].Outdated = true
				result[i].LatestVersion = latest
				continue
			}
		}
		// Check git repo behind-upstream status locally
		if result[i].Type == model.TypeGitRepo || result[i].Type == model.TypeVimRuntime {
			dir := result[i].InstallPath
			if dir == "" && result[i].Type == model.TypeVimRuntime {
				dir = filepath.Join(c.getHomeDir(), ".vim_runtime")
			}
			if dir != "" && c.isGitRepoBehind(dir) {
				result[i].Outdated = true
				result[i].LatestVersion = "new commits"
			}
		}
	}

	return result
}

type brewOutdatedJSON struct {
	Formulae []struct {
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
	} `json:"formulae"`
	Casks []struct {
		Name           string `json:"name"`
		CurrentVersion string `json:"current_version"`
	} `json:"casks"`
}

func (c *Checker) fetchOutdatedMap() map[string]string {
	outdated := make(map[string]string)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	if c.IsBrewInstalled() {
		cmd := exec.CommandContext(ctx, "brew", "outdated", "--json=v2")
		cmd.Env = append(os.Environ(), "HOMEBREW_NO_AUTO_UPDATE=1")
		if out, err := cmd.Output(); err == nil && len(out) > 0 {
			var parsed brewOutdatedJSON
			if json.Unmarshal(out, &parsed) == nil {
				for _, f := range parsed.Formulae {
					outdated[f.Name] = f.CurrentVersion
				}
				for _, k := range parsed.Casks {
					outdated[k.Name] = k.CurrentVersion
				}
			}
		}
	}

	if runtime.GOOS == "linux" {
		if _, err := exec.LookPath("apt"); err == nil {
			cmd := exec.CommandContext(ctx, "apt", "list", "--upgradable")
			if out, err := cmd.Output(); err == nil {
				for _, line := range strings.Split(string(out), "\n") {
					if idx := strings.Index(line, "/"); idx > 0 {
						pkgName := strings.TrimSpace(line[:idx])
						fields := strings.Fields(line)
						if len(fields) >= 2 {
							outdated[pkgName] = fields[1]
						}
					}
				}
			}
		}
	}

	return outdated
}

func (c *Checker) isGitRepoBehind(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return false
	}
	out, err := exec.Command("git", "-C", dir, "status", "-uno", "--porcelain=v2", "--branch").Output()
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "# branch.ab ") {
			// Format: # branch.ab +0 -2
			fields := strings.Fields(line)
			if len(fields) >= 4 && strings.HasPrefix(fields[3], "-") && fields[3] != "-0" {
				return true
			}
		}
	}
	return false
}


func (c *Checker) checkFormulaOrBinary(pkg *model.Package) {
	bin := pkg.BinaryName
	if bin == "" {
		bin = pkg.BrewName
	}

	// Special case for Java: macOS includes a stub /usr/bin/java that fails if no JDK is installed
	if bin == "java" {
		cmd := exec.Command("java", "-version")
		if err := cmd.Run(); err == nil {
			pkg.Installed = true
			pkg.Version = c.detectBinaryVersion("java")
			return
		}
		if runtime.GOOS == "darwin" && c.IsBrewInstalled() {
			if exec.Command("brew", "list", "--formula", "openjdk").Run() == nil {
				pkg.Installed = true
				pkg.Version = "OpenJDK (brew)"
				return
			}
		}
		pkg.Installed = false
		return
	}

	// Special case for Python on Windows where `python` is used or `python3` is a store stub
	if bin == "python3" && runtime.GOOS == "windows" {
		if out, err := exec.Command("python", "--version").CombinedOutput(); err == nil && len(out) > 0 {
			pkg.Installed = true
			pkg.Version = strings.TrimSpace(strings.Split(string(out), "\n")[0])
			return
		}
	}

	// Special case for bat on Debian/Ubuntu Linux where binary is installed as `batcat`
	if bin == "bat" && runtime.GOOS == "linux" {
		if _, err := exec.LookPath("bat"); err != nil {
			if _, errCat := exec.LookPath("batcat"); errCat == nil {
				pkg.Installed = true
				pkg.Version = c.detectBinaryVersion("batcat")
				return
			}
		}
	}

	// 1. Check if binary is in PATH
	if path, err := exec.LookPath(bin); err == nil && path != "" {
		// On Windows, verify python3 isn't the Microsoft Store alias stub
		if bin == "python3" {
			if err := exec.Command(bin, "--version").Run(); err != nil {
				pkg.Installed = false
				return
			}
		}
		pkg.Installed = true
		pkg.Version = c.detectBinaryVersion(bin)
		return
	}

	// 2. Check local bin (~/.local/bin) explicitly in case PATH hasn't been reloaded in current process
	localBin := filepath.Join(c.getHomeDir(), ".local", "bin", bin)
	if _, err := os.Stat(localBin); err == nil {
		pkg.Installed = true
		pkg.Version = c.detectBinaryVersion(localBin)
		return
	}

	// 2b. On Windows, check WinGet Links and Chocolatey bin directories in case PATH hasn't refreshed in current session
	if runtime.GOOS == "windows" {
		exeName := bin + ".exe"
		winCandidates := []string{
			filepath.Join(c.getHomeDir(), ".local", "bin", exeName),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "WinGet", "Links", exeName),
			filepath.Join(`C:\ProgramData\chocolatey\bin`, exeName),
		}
		for _, candidate := range winCandidates {
			if _, err := os.Stat(candidate); err == nil {
				pkg.Installed = true
				pkg.Version = c.detectBinaryVersion(candidate)
				return
			}
		}
	}

	// 3. Check if installed via Homebrew if brew is available
	if pkg.BrewName != "" && c.IsBrewInstalled() {
		cmd := exec.Command("brew", "list", "--formula", pkg.BrewName)
		if err := cmd.Run(); err == nil {
			pkg.Installed = true
			pkg.Version = c.detectBinaryVersion(bin)
			return
		}
	}

	pkg.Installed = false
}

func (c *Checker) checkCaskOrApp(pkg *model.Package) {
	// 1. Check AppPath if provided (macOS .app bundle)
	if pkg.AppPath != "" && runtime.GOOS == "darwin" {
		if _, err := os.Stat(pkg.AppPath); err == nil {
			pkg.Installed = true
			pkg.Version = c.detectAppVersion(pkg.AppPath)
			return
		}
		userApp := filepath.Join(c.getHomeDir(), "Applications", filepath.Base(pkg.AppPath))
		if _, err := os.Stat(userApp); err == nil {
			pkg.Installed = true
			pkg.Version = c.detectAppVersion(userApp)
			return
		}
	}

	// 2. Check OS-specific desktop app paths on Windows and Linux
	if runtime.GOOS == "windows" {
		for _, winPath := range c.windowsAppCandidates(pkg.ID) {
			if _, err := os.Stat(winPath); err == nil {
				pkg.Installed = true
				pkg.Version = "Installed"
				return
			}
		}
	} else if runtime.GOOS == "linux" && pkg.ID == "chrome" {
		for _, bin := range []string{"google-chrome", "google-chrome-stable", "chromium-browser", "chromium"} {
			if _, err := exec.LookPath(bin); err == nil {
				pkg.Installed = true
				pkg.Version = c.detectBinaryVersion(bin)
				return
			}
		}
	}

	// 3. Check binary name (e.g. `gcloud`, `code`, `wt`)
	if pkg.BinaryName != "" {
		if _, err := exec.LookPath(pkg.BinaryName); err == nil {
			pkg.Installed = true
			pkg.Version = c.detectBinaryVersion(pkg.BinaryName)
			return
		}
		if pkg.BinaryName == "gcloud" {
			gcloudBin := filepath.Join(c.getHomeDir(), "google-cloud-sdk", "bin", "gcloud")
			if _, err := os.Stat(gcloudBin); err == nil {
				pkg.Installed = true
				pkg.Version = c.detectBinaryVersion(gcloudBin)
				return
			}
		}
	}

	// 4. Check brew cask list on macOS
	if runtime.GOOS == "darwin" && pkg.BrewName != "" && c.IsBrewInstalled() {
		cmd := exec.Command("brew", "list", "--cask", pkg.BrewName)
		if err := cmd.Run(); err == nil {
			pkg.Installed = true
			return
		}
	}

	pkg.Installed = false
}

func (c *Checker) windowsAppCandidates(pkgID string) []string {
	home := c.getHomeDir()
	switch pkgID {
	case "chrome":
		return []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			filepath.Join(home, "AppData", "Local", "Google", "Chrome", "Application", "chrome.exe"),
		}
	case "vscode":
		return []string{
			`C:\Program Files\Microsoft VS Code\Code.exe`,
			filepath.Join(home, "AppData", "Local", "Programs", "Microsoft VS Code", "Code.exe"),
		}
	case "windows-terminal":
		return []string{
			filepath.Join(home, "AppData", "Local", "Microsoft", "WindowsApps", "wt.exe"),
		}
	}
	return nil
}

func (c *Checker) checkDirOrGit(pkg *model.Package) {
	if pkg.InstallPath == "" {
		return
	}

	info, err := os.Stat(pkg.InstallPath)
	if err == nil && info.IsDir() {
		pkg.Installed = true

		gitDir := filepath.Join(pkg.InstallPath, ".git")
		if _, err := os.Stat(gitDir); err == nil {
			out, err := exec.Command("git", "-C", pkg.InstallPath, "describe", "--tags", "--always").Output()
			if err == nil {
				pkg.Version = strings.TrimSpace(string(out))
			} else {
				out, err = exec.Command("git", "-C", pkg.InstallPath, "rev-parse", "--short", "HEAD").Output()
				if err == nil {
					pkg.Version = strings.TrimSpace(string(out))
				}
			}
		}
	} else {
		pkg.Installed = false
	}
}

func (c *Checker) checkFont(pkg *model.Package) {
	fontPath := pkg.InstallPath
	if fontPath == "" {
		fontPath = filepath.Join(model.DefaultFontsDir(c.getHomeDir()), "MesloLGS NF Regular.ttf")
	}
	if _, err := os.Stat(fontPath); err == nil {
		pkg.Installed = true
		pkg.Version = "Installed"
		return
	}
	pkg.Installed = false
}

func (c *Checker) checkITermFontConfig(pkg *model.Package) {
	if runtime.GOOS != "darwin" {
		pkg.Installed = false
		return
	}
	plistPath := filepath.Join(c.getHomeDir(), "Library", "Preferences", "com.googlecode.iterm2.plist")

	out, err := exec.Command("/usr/libexec/PlistBuddy", "-c", "Print :'New Bookmarks':0:'Normal Font'", plistPath).Output()
	if err == nil && strings.Contains(string(out), "Meslo") {
		pkg.Installed = true
		pkg.Version = strings.TrimSpace(string(out))
		return
	}

	out, err = exec.Command("defaults", "read", "com.googlecode.iterm2").Output()
	if err == nil && strings.Contains(string(out), "Meslo") {
		pkg.Installed = true
		pkg.Version = "MesloLGS NF configured"
		return
	}

	pkg.Installed = false
}

func (c *Checker) checkVimRuntime(pkg *model.Package) {
	runtimeDir := pkg.InstallPath
	if runtimeDir == "" {
		runtimeDir = filepath.Join(c.getHomeDir(), ".vim_runtime")
	}
	if info, err := os.Stat(runtimeDir); err == nil && info.IsDir() {
		pkg.Installed = true
		pkg.Version = "Awesome vimrc"
	} else {
		pkg.Installed = false
	}
}

func (c *Checker) checkCustom(pkg *model.Package) {
	switch pkg.ID {
	case "xcode-select":
		if runtime.GOOS != "darwin" {
			pkg.Installed = false
			return
		}
		out, err := exec.Command("xcode-select", "-p").Output()
		if err == nil && len(strings.TrimSpace(string(out))) > 0 {
			pkg.Installed = true
			pkg.Version = strings.TrimSpace(string(out))
		} else {
			pkg.Installed = false
		}
	case "brew":
		if path, err := exec.LookPath("brew"); err == nil && path != "" {
			pkg.Installed = true
			out, err := exec.Command("brew", "--version").Output()
			if err == nil {
				fields := strings.Fields(string(out))
				if len(fields) >= 2 {
					pkg.Version = fields[1]
					return
				}
			}
			pkg.Version = "ready"
		} else {
			pkg.Installed = false
		}
	case "winget":
		if path, err := exec.LookPath("winget"); err == nil && path != "" {
			pkg.Installed = true
			pkg.Version = c.detectBinaryVersion("winget")
			if pkg.Version == "" {
				pkg.Version = "ready"
			}
		} else {
			pkg.Installed = false
		}
	case "chrome-default":
		if runtime.GOOS != "darwin" {
			pkg.Installed = false
			return
		}
		out, err := exec.Command("defaults", "read", "com.apple.LaunchServices/com.apple.launchservices.secure", "LSHandlers").Output()
		if err == nil && strings.Contains(string(out), "com.google.chrome") {
			pkg.Installed = true
			pkg.Version = "Default Browser"
		} else {
			pkg.Installed = false
		}
	case "code-cli":
		if path, err := exec.LookPath("code"); err == nil && path != "" {
			pkg.Installed = true
			pkg.Version = c.detectBinaryVersion("code")
		} else {
			pkg.Installed = false
		}
	case "brogrammer-theme":
		if runtime.GOOS != "darwin" {
			pkg.Installed = false
			return
		}
		out, err := exec.Command("defaults", "read", "com.googlecode.iterm2", "Custom Color Presets").Output()
		if err == nil && strings.Contains(string(out), "Brogrammer") {
			pkg.Installed = true
			pkg.Version = "Preset Ready"
		} else {
			if _, err := os.Stat(filepath.Join(c.getHomeDir(), "Brogrammer.itermcolors")); err == nil {
				pkg.Installed = true
				pkg.Version = "File Available"
			} else {
				pkg.Installed = false
			}
		}
	case "oh-my-zsh":
		c.checkDirOrGit(pkg)
	case "shell-integration":
		if c.IsP10kConfigured() && c.IsZoxideConfigured() {
			pkg.Installed = true
			pkg.Version = ".zshrc configured"
		} else {
			pkg.Installed = false
		}
	default:
		c.checkFormulaOrBinary(pkg)
	}
}
