package model

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ShellCommand constructs an OS-appropriate shell command for executing a command string
func ShellCommand(cmdStr string) *exec.Cmd {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("bash"); err == nil {
		cmd = exec.Command("bash", "-c", cmdStr)
	} else if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("powershell"); err == nil {
			cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", cmdStr)
		} else {
			cmd = exec.Command("cmd", "/C", cmdStr)
		}
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}
	ensureHomeEnv(cmd)
	return cmd
}

// ShellCommandContext constructs an OS-appropriate context-aware shell command
func ShellCommandContext(ctx context.Context, cmdStr string) *exec.Cmd {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("bash"); err == nil {
		cmd = exec.CommandContext(ctx, "bash", "-c", cmdStr)
	} else if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("powershell"); err == nil {
			cmd = exec.CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command", cmdStr)
		} else {
			cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
		}
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	ensureHomeEnv(cmd)
	return cmd
}

func ensureHomeEnv(cmd *exec.Cmd) {
	if runtime.GOOS == "windows" && os.Getenv("HOME") == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			cmd.Env = append(os.Environ(), "HOME="+filepath.ToSlash(home))
		}
	}
}

// DefaultFontsDir returns the OS-specific user fonts directory
func DefaultFontsDir(homeDir string) string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Fonts")
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "Microsoft", "Windows", "Fonts")
	default: // linux, freebsd, etc.
		return filepath.Join(homeDir, ".local", "share", "fonts")
	}
}
