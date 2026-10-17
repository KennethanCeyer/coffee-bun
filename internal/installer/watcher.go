package installer

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func (r *Runner) installXcodeSelectWithWatcher(ctx context.Context, ch chan<- model.InstallEvent) {
	if exec.Command("xcode-select", "-p").Run() == nil {
		return
	}

	snippet := "xcode-select --install"
	ch <- model.InstallEvent{
		PackageID:     "xcode-select",
		Name:          "Xcode Command Line Tools",
		Status:        model.StatusWaitingAction,
		Message:       "macOS system compilers (clang/git) require Xcode Command Line Tools. Click 'Install' on the macOS popup or run the command below.",
		ActionSnippet: snippet,
	}

	_ = exec.CommandContext(ctx, "xcode-select", "--install").Run()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if exec.Command("xcode-select", "-p").Run() == nil {
				ch <- model.InstallEvent{
					PackageID: "xcode-select",
					Name:      "Xcode Command Line Tools",
					Status:    model.StatusSuccess,
					Message:   "Xcode Command Line Tools detected and ready",
				}
				return
			}
		}
	}
}

func (r *Runner) installHomebrewWithWatcher(ctx context.Context, ch chan<- model.InstallEvent) {
	if r.checker.IsBrewInstalled() {
		return
	}

	snippet := `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`
	ch <- model.InstallEvent{
		PackageID:     "brew",
		Name:          "Homebrew",
		Status:        model.StatusWaitingAction,
		Message:       "Homebrew installation requires administrator (sudo) password entry. Please run the command below in another terminal window.",
		ActionSnippet: snippet,
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if r.checker.IsBrewInstalled() {
				ch <- model.InstallEvent{
					PackageID: "brew",
					Name:      "Homebrew",
					Status:    model.StatusSuccess,
					Message:   "Homebrew detected! Resuming installation...",
				}
				return
			}
		}
	}
}

func (r *Runner) verifyAndLinkJavaWithWatcher(ctx context.Context, ch chan<- model.InstallEvent) {
	if exec.Command("java", "-version").Run() == nil {
		ch <- model.InstallEvent{
			PackageID: "java",
			Name:      "Java (OpenJDK)",
			Status:    model.StatusSuccess,
			Message:   "Installed and registered in system JavaVirtualMachines",
		}
		return
	}

	snippet := "sudo ln -sfn /opt/homebrew/opt/openjdk/libexec/openjdk.jdk /Library/Java/JavaVirtualMachines/openjdk.jdk"

	ch <- model.InstallEvent{
		PackageID:     "java",
		Name:          "Java (OpenJDK)",
		Status:        model.StatusWaitingAction,
		Message:       "OpenJDK is installed via Homebrew, but macOS requires administrator (sudo) permission to register the JDK in /Library/Java/JavaVirtualMachines.",
		ActionSnippet: snippet,
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if exec.Command("java", "-version").Run() == nil {
				ch <- model.InstallEvent{
					PackageID: "java",
					Name:      "Java (OpenJDK)",
					Status:    model.StatusSuccess,
					Message:   "OpenJDK registered! Java is ready.",
				}
				return
			}
		}
	}
}

func (r *Runner) verifyAndLinkVSCode(ch chan<- model.InstallEvent) {
	if !r.checker.IsCodeCLILinked() {
		target, err := r.configurator.LinkVSCodeCLI()
		if err == nil {
			ch <- model.InstallEvent{
				PackageID: "code-cli",
				Name:      "VS Code CLI (`code`)",
				Status:    model.StatusSuccess,
				Message:   fmt.Sprintf("Linked `code` CLI to %s", target),
			}
		} else {
			ch <- model.InstallEvent{
				PackageID: "code-cli",
				Name:      "VS Code CLI (`code`)",
				Status:    model.StatusFailed,
				Message:   fmt.Sprintf("Could not auto-link `code` CLI: %v", err),
			}
		}
	}
}
