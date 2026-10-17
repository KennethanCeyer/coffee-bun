package updater

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

// UpdateEvent carries real-time status from concurrent package update workers
type UpdateEvent struct {
	PackageName string
	Status      model.InstallStatus
	Message     string
	Error       error
}

// Updater coordinates concurrent upgrades of Homebrew formulae, casks, git repositories, and custom update commands
type Updater struct {
	checker  *checker.Checker
	homeDir  string
	skipBrew bool
}

// NewUpdater creates a new Updater instance
func NewUpdater() *Updater {
	home, _ := os.UserHomeDir()
	return &Updater{
		checker: checker.NewCheckerWithHome(home),
		homeDir: home,
	}
}

// NewUpdaterWithHome creates an Updater scoped to a specific home directory (useful for isolated testing)
func NewUpdaterWithHome(homeDir string, skipBrew bool) *Updater {
	return &Updater{
		checker:  checker.NewCheckerWithHome(homeDir),
		homeDir:  homeDir,
		skipBrew: skipBrew,
	}
}

// UpdateAllAsync runs update operations concurrently and streams events
func (u *Updater) UpdateAllAsync(ctx context.Context, pkgs []model.Package) <-chan UpdateEvent {
	eventChan := make(chan UpdateEvent, 40)

	go func() {
		defer close(eventChan)

		var wg sync.WaitGroup

		// 1. Homebrew Worker
		if !u.skipBrew && u.checker.IsBrewInstalled() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				u.updateHomebrew(ctx, eventChan)
			}()
		}

		// 2. Discover Git Repositories from packages + default shell directories
		gitDirs := make(map[string]string)
		if u.homeDir != "" {
			defaultDirs := map[string]string{
				"Oh My Zsh":               filepath.Join(u.homeDir, ".oh-my-zsh"),
				"Powerlevel10k":           filepath.Join(u.homeDir, ".oh-my-zsh", "custom", "themes", "powerlevel10k"),
				"Zsh Autosuggestions":     filepath.Join(u.homeDir, ".oh-my-zsh", "custom", "plugins", "zsh-autosuggestions"),
				"Zsh Syntax Highlighting": filepath.Join(u.homeDir, ".oh-my-zsh", "custom", "plugins", "zsh-syntax-highlighting"),
			}
			for name, dir := range defaultDirs {
				gitDirs[name] = dir
			}
		}

		var customUpdatePkgs []model.Package
		for _, p := range pkgs {
			if p.InstallPath != "" && (p.Type == model.TypeGitRepo || p.Type == model.TypeVimRuntime || p.ID == "oh-my-zsh") {
				gitDirs[p.Name] = p.InstallPath
			}
			if p.UpdateCmd != "" {
				customUpdatePkgs = append(customUpdatePkgs, p)
			}
		}

		for name, dir := range gitDirs {
			if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
				wg.Add(1)
				go func(repoName, repoDir string) {
					defer wg.Done()
					u.updateGitRepo(ctx, repoName, repoDir, eventChan)
				}(name, dir)
			}
		}

		// 3. Custom UpdateCmd Workers
		for _, pkg := range customUpdatePkgs {
			wg.Add(1)
			go func(p model.Package) {
				defer wg.Done()
				u.runCustomUpdateCmd(ctx, p, eventChan)
			}(pkg)
		}

		wg.Wait()
	}()

	return eventChan
}

func (u *Updater) updateHomebrew(ctx context.Context, eventChan chan<- UpdateEvent) {
	eventChan <- UpdateEvent{
		PackageName: "Homebrew",
		Status:      model.StatusInstalling,
		Message:     "Updating Homebrew package index...",
	}

	if err := exec.CommandContext(ctx, "brew", "update").Run(); err != nil {
		eventChan <- UpdateEvent{
			PackageName: "Homebrew",
			Status:      model.StatusFailed,
			Message:     "Failed to run brew update",
			Error:       err,
		}
		return
	}

	eventChan <- UpdateEvent{
		PackageName: "Homebrew Packages",
		Status:      model.StatusInstalling,
		Message:     "Upgrading outdated formulae and casks...",
	}

	if out, err := exec.CommandContext(ctx, "brew", "upgrade").CombinedOutput(); err != nil {
		eventChan <- UpdateEvent{
			PackageName: "Homebrew Packages",
			Status:      model.StatusFailed,
			Message:     fmt.Sprintf("Brew upgrade error: %s", strings.TrimSpace(string(out))),
			Error:       err,
		}
	} else {
		eventChan <- UpdateEvent{
			PackageName: "Homebrew Packages",
			Status:      model.StatusSuccess,
			Message:     "All Homebrew packages and casks upgraded to latest version",
		}
	}
}

func (u *Updater) updateGitRepo(ctx context.Context, repoName, repoDir string, eventChan chan<- UpdateEvent) {
	eventChan <- UpdateEvent{
		PackageName: repoName,
		Status:      model.StatusInstalling,
		Message:     fmt.Sprintf("Pulling latest changes for %s...", repoName),
	}

	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "pull", "--rebase")
	out, err := cmd.CombinedOutput()
	if err != nil {
		eventChan <- UpdateEvent{
			PackageName: repoName,
			Status:      model.StatusFailed,
			Message:     fmt.Sprintf("Failed to update: %s", strings.TrimSpace(string(out))),
			Error:       err,
		}
		return
	}

	outputStr := strings.TrimSpace(string(out))
	msg := fmt.Sprintf("%s updated (%s)", repoName, outputStr)
	if strings.Contains(outputStr, "Already up to date") {
		msg = fmt.Sprintf("%s is already up to date", repoName)
	}
	eventChan <- UpdateEvent{
		PackageName: repoName,
		Status:      model.StatusSuccess,
		Message:     msg,
	}
}

func (u *Updater) runCustomUpdateCmd(ctx context.Context, p model.Package, eventChan chan<- UpdateEvent) {
	eventChan <- UpdateEvent{
		PackageName: p.Name,
		Status:      model.StatusInstalling,
		Message:     fmt.Sprintf("Updating %s...", p.Name),
	}

	cmd := model.ShellCommandContext(ctx, p.UpdateCmd)
	if len(p.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range p.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		eventChan <- UpdateEvent{
			PackageName: p.Name,
			Status:      model.StatusFailed,
			Message:     fmt.Sprintf("Update failed: %s", strings.TrimSpace(string(out))),
			Error:       err,
		}
		return
	}

	eventChan <- UpdateEvent{
		PackageName: p.Name,
		Status:      model.StatusSuccess,
		Message:     fmt.Sprintf("%s updated successfully", p.Name),
	}
}
