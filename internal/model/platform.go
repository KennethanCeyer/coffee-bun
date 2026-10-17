package model

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// ShellCommand constructs an OS-appropriate shell command for executing a command string.
// Windows uses PowerShell exclusively; macOS and Linux (including WSL) use bash/sh.
func ShellCommand(cmdStr string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		psBin := "powershell"
		if _, err := exec.LookPath("pwsh"); err == nil {
			psBin = "pwsh"
		}
		cmd = exec.Command(psBin, "-NoProfile", "-NonInteractive", "-Command", cmdStr)
	} else if _, err := exec.LookPath("bash"); err == nil {
		cmd = exec.Command("bash", "-c", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	ensureCommandEnv(cmd)
	return cmd
}

// ShellCommandContext constructs an OS-appropriate context-aware shell command.
// Windows uses PowerShell exclusively; macOS and Linux (including WSL) use bash/sh.
func ShellCommandContext(ctx context.Context, cmdStr string) *exec.Cmd {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		psBin := "powershell"
		if _, err := exec.LookPath("pwsh"); err == nil {
			psBin = "pwsh"
		}
		cmd = exec.CommandContext(ctx, psBin, "-NoProfile", "-NonInteractive", "-Command", cmdStr)
	} else if _, err := exec.LookPath("bash"); err == nil {
		cmd = exec.CommandContext(ctx, "bash", "-c", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	ensureCommandEnv(cmd)
	return cmd
}

func ensureCommandEnv(cmd *exec.Cmd) {
	env := os.Environ()
	if runtime.GOOS == "windows" && os.Getenv("HOME") == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			env = append(env, "HOME="+filepath.ToSlash(home))
		}
	}
	env = append(env, "LANG=en_US.UTF-8", "LC_ALL=en_US.UTF-8", "LC_MESSAGES=en_US.UTF-8")
	cmd.Env = env
}

// IsWSL returns true when running on Linux inside Windows Subsystem for Linux (WSL).
// WSL is grouped under the "linux" platform for all package definitions and installers.
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	if data, err := os.ReadFile("/proc/version"); err == nil {
		lower := strings.ToLower(string(data))
		if strings.Contains(lower, "microsoft") || strings.Contains(lower, "wsl") {
			return true
		}
	}
	return false
}

// PlatformDisplayName returns a human-friendly OS label (e.g. "macOS", "Linux (WSL: Ubuntu)", "Linux", "Windows (PowerShell)").
func PlatformDisplayName() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows (PowerShell)"
	case "linux":
		if IsWSL() {
			if distro := os.Getenv("WSL_DISTRO_NAME"); distro != "" {
				return "Linux (WSL: " + distro + ")"
			}
			return "Linux (WSL)"
		}
		return "Linux"
	default:
		return runtime.GOOS
	}
}

// DefaultFontsDir returns the OS-specific user fonts directory
func DefaultFontsDir(homeDir string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Fonts")
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "Microsoft", "Windows", "Fonts")
	default: // linux (including WSL), freebsd, etc.
		return filepath.Join(homeDir, ".local", "share", "fonts")
	}
}

