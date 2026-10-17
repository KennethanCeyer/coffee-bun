package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Configurator manages shell configuration files safely
type Configurator struct {
	homeDir string
	zshrc   string
}

// NewConfigurator creates a new Configurator instance
func NewConfigurator() *Configurator {
	home, _ := os.UserHomeDir()
	return NewConfiguratorWithDir(home)
}

// NewConfiguratorWithDir creates a Configurator rooted at a custom directory
func NewConfiguratorWithDir(homeDir string) *Configurator {
	return &Configurator{
		homeDir: homeDir,
		zshrc:   filepath.Join(homeDir, ".zshrc"),
	}
}

// HomeDir returns the base home directory for this Configurator
func (c *Configurator) HomeDir() string {
	return c.homeDir
}

// BackupZshrc creates a timestamped backup of ~/.zshrc
func (c *Configurator) BackupZshrc() (string, error) {
	if _, err := os.Stat(c.zshrc); os.IsNotExist(err) {
		return "", nil
	}
	content, err := os.ReadFile(c.zshrc)
	if err != nil {
		return "", err
	}
	backupPath := fmt.Sprintf("%s.bun.%d.bak", c.zshrc, time.Now().Unix())
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return "", err
	}
	return backupPath, nil
}

// ConfigurePowerlevel10k sets ZSH_THEME="powerlevel10k/powerlevel10k" in ~/.zshrc
func (c *Configurator) ConfigurePowerlevel10k() error {
	content, err := c.readOrCreateZshrc()
	if err != nil {
		return err
	}

	reTheme := regexp.MustCompile(`(?m)^#?\s*ZSH_THEME=.*$`)
	p10kTheme := `ZSH_THEME="powerlevel10k/powerlevel10k"`

	var newContent string
	if reTheme.MatchString(content) {
		newContent = reTheme.ReplaceAllString(content, p10kTheme)
	} else {
		newContent = p10kTheme + "\n" + content
	}

	return os.WriteFile(c.zshrc, []byte(newContent), 0644)
}

// ConfigureZoxide ensures `eval "$(zoxide init zsh)"` is configured in ~/.zshrc
func (c *Configurator) ConfigureZoxide() error {
	content, err := c.readOrCreateZshrc()
	if err != nil {
		return err
	}

	if strings.Contains(content, "zoxide init zsh") {
		return nil
	}

	block := "\n# Added by Coffee Bun: zoxide integration\nif command -v zoxide >/dev/null 2>&1; then\n  eval \"$(zoxide init zsh)\"\nfi\n"
	newContent := content + block
	return os.WriteFile(c.zshrc, []byte(newContent), 0644)
}

// ConfigurePlugins ensures the given plugins are enabled in ~/.zshrc plugins=(...)
func (c *Configurator) ConfigurePlugins(neededPlugins []string) error {
	content, err := c.readOrCreateZshrc()
	if err != nil {
		return err
	}

	rePlugins := regexp.MustCompile(`(?m)^plugins=\(([^)]*)\)`)
	matches := rePlugins.FindStringSubmatch(content)

	if len(matches) > 1 {
		currentPlugins := strings.Fields(matches[1])
		pluginMap := make(map[string]bool)
		for _, p := range currentPlugins {
			pluginMap[p] = true
		}
		var added []string
		for _, p := range neededPlugins {
			if !pluginMap[p] {
				currentPlugins = append(currentPlugins, p)
				added = append(added, p)
			}
		}
		if len(added) == 0 {
			return nil
		}
		newPluginsStr := fmt.Sprintf("plugins=(%s)", strings.Join(currentPlugins, " "))
		newContent := rePlugins.ReplaceAllString(content, newPluginsStr)
		return os.WriteFile(c.zshrc, []byte(newContent), 0644)
	}

	// If no plugins=(...) line found, append one
	block := fmt.Sprintf("\n# Added by Coffee Bun\nplugins=(git %s)\n", strings.Join(neededPlugins, " "))
	return os.WriteFile(c.zshrc, []byte(content+block), 0644)
}

// LinkVSCodeCLI finds Visual Studio Code's code binary and symlinks it to ~/.local/bin or /usr/local/bin
func (c *Configurator) LinkVSCodeCLI() (string, error) {
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
			filepath.Join(c.homeDir, "Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"),
		}
	case "windows":
		candidates = []string{
			filepath.Join(c.homeDir, "AppData", "Local", "Programs", "Microsoft VS Code", "bin", "code.cmd"),
			`C:\Program Files\Microsoft VS Code\bin\code.cmd`,
		}
	default: // linux
		candidates = []string{
			"/usr/share/code/bin/code",
			"/snap/bin/code",
		}
	}

	var appPath string
	for _, cand := range candidates {
		if _, err := os.Stat(cand); err == nil {
			appPath = cand
			break
		}
	}
	if appPath == "" {
		return "", fmt.Errorf("Visual Studio Code application not found")
	}

	if runtime.GOOS == "windows" {
		return appPath, nil
	}

	targetDirs := []string{
		filepath.Join(c.homeDir, ".local", "bin"),
		"/usr/local/bin",
	}

	for _, dir := range targetDirs {
		if err := os.MkdirAll(dir, 0755); err == nil {
			targetSymlink := filepath.Join(dir, "code")
			if dest, err := os.Readlink(targetSymlink); err == nil && dest == appPath {
				return targetSymlink, nil
			}
			_ = os.Remove(targetSymlink)
			if err := os.Symlink(appPath, targetSymlink); err == nil {
				return targetSymlink, nil
			}
		}
	}

	return "", fmt.Errorf("could not create symlink for code CLI in standard PATH directories")
}

func (c *Configurator) readOrCreateZshrc() (string, error) {
	if _, err := os.Stat(c.zshrc); os.IsNotExist(err) {
		defaultContent := "# Created by Coffee Bun\nexport PATH=\"$HOME/.local/bin:/opt/homebrew/bin:$PATH\"\n"
		if err := os.WriteFile(c.zshrc, []byte(defaultContent), 0644); err != nil {
			return "", err
		}
		return defaultContent, nil
	}
	contentBytes, err := os.ReadFile(c.zshrc)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}
