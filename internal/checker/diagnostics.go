package checker

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
)

// IsCodeCLILinked checks if `code` binary is available in PATH
func (c *Checker) IsCodeCLILinked() bool {
	_, err := exec.LookPath("code")
	return err == nil
}

// FindVSCodeCLI finds the bundled code CLI path across macOS, Linux, and Windows
func (c *Checker) FindVSCodeCLI() string {
	var candidates []string
	switch runtime.GOOS {
	case "darwin":
		candidates = []string{
			"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code",
			filepath.Join(c.getHomeDir(), "Applications/Visual Studio Code.app/Contents/Resources/app/bin/code"),
		}
	case "windows":
		candidates = []string{
			filepath.Join(c.getHomeDir(), "AppData", "Local", "Programs", "Microsoft VS Code", "bin", "code.cmd"),
			`C:\Program Files\Microsoft VS Code\bin\code.cmd`,
		}
	default: // linux
		candidates = []string{
			"/usr/share/code/bin/code",
			"/snap/bin/code",
		}
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// IsP10kConfigured checks if .zshrc contains powerlevel10k theme
func (c *Checker) IsP10kConfigured() bool {
	zshrc := filepath.Join(c.getHomeDir(), ".zshrc")
	content, err := os.ReadFile(zshrc)
	if err != nil {
		return false
	}
	re := regexp.MustCompile(`(?m)^[^#]*ZSH_THEME=.*powerlevel10k`)
	return re.Match(content)
}

// IsZoxideConfigured checks if .zshrc contains zoxide init zsh
func (c *Checker) IsZoxideConfigured() bool {
	zshrc := filepath.Join(c.getHomeDir(), ".zshrc")
	content, err := os.ReadFile(zshrc)
	if err != nil {
		return false
	}
	return bytes.Contains(content, []byte("zoxide init zsh"))
}
