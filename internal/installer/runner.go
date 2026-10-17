package installer

import (
	"context"
	"runtime"
	"sync"

	"github.com/KennethanCeyer/coffee-bun/internal/checker"
	"github.com/KennethanCeyer/coffee-bun/internal/model"
	"github.com/KennethanCeyer/coffee-bun/internal/shell"
)

// Runner manages concurrent execution of software installations across macOS, Linux, and Windows
type Runner struct {
	checker      *checker.Checker
	configurator *shell.Configurator
	hooks        model.Hooks
}

// NewRunner creates a new installer runner
func NewRunner() *Runner {
	return &Runner{
		checker:      checker.NewChecker(),
		configurator: shell.NewConfigurator(),
	}
}

// NewRunnerWithHome creates an installer runner scoped to a specific home directory
func NewRunnerWithHome(homeDir string) *Runner {
	return &Runner{
		checker:      checker.NewCheckerWithHome(homeDir),
		configurator: shell.NewConfiguratorWithDir(homeDir),
	}
}

// SetHooks configures lifecycle pre-install and post-install hooks
func (r *Runner) SetHooks(hooks model.Hooks) {
	r.hooks = hooks
}

// InstallAsync executes installations concurrently and streams events over a channel
func (r *Runner) InstallAsync(ctx context.Context, pkgs []model.Package) <-chan model.InstallEvent {
	eventChan := make(chan model.InstallEvent, 50)

	go func() {
		defer close(eventChan)

		if len(pkgs) == 0 {
			eventChan <- model.InstallEvent{
				Name:    "System Check",
				Status:  model.StatusSuccess,
				Message: "No packages selected for installation.",
			}
			return
		}

		// Partition packages into sequential prerequisites vs parallel tasks
		var xcodeTask, brewTask *model.Package
		var needOMZ bool
		var gitRepos []model.Package
		var brewFormulae []model.Package
		var brewCasks []model.Package
		var customTasks []model.Package

		for _, p := range pkgs {
			if p.ID == "xcode-select" {
				xcodeTask = &p
			} else if p.ID == "brew" {
				brewTask = &p
			} else if p.ID == "oh-my-zsh" {
				needOMZ = true
			} else if p.Type == model.TypeGitRepo {
				gitRepos = append(gitRepos, p)
			} else if p.Type == model.TypeBrewFormula {
				brewFormulae = append(brewFormulae, p)
			} else if p.Type == model.TypeBrewCask {
				brewCasks = append(brewCasks, p)
			} else {
				customTasks = append(customTasks, p)
			}
		}

		// Execute PreInstall hook if configured
		if r.hooks.PreInstall != "" {
			eventChan <- model.InstallEvent{
				PackageID: "pre-install-hook",
				Name:      "Pre-Install Hook",
				Status:    model.StatusInstalling,
				Message:   "Executing pre-install hook...",
			}
			cmd := model.ShellCommandContext(ctx, r.hooks.PreInstall)
			_ = cmd.Run()
		}

		// 1. Prerequisites first: Xcode Command Line Tools (macOS only)
		if runtime.GOOS == "darwin" && xcodeTask != nil && !xcodeTask.Installed {
			r.installXcodeSelectWithWatcher(ctx, eventChan)
		}

		// 2. Prerequisites: Homebrew (macOS or when explicitly requested)
		if brewTask != nil && !brewTask.Installed {
			r.installHomebrewWithWatcher(ctx, eventChan)
		}

		var mainWg sync.WaitGroup

		// --- Worker 1: Custom tasks & Fonts & Vim runtime (goroutine) ---
		mainWg.Add(1)
		go func() {
			defer mainWg.Done()
			var customWg sync.WaitGroup
			for _, task := range customTasks {
				customWg.Add(1)
				go func(t model.Package) {
					defer customWg.Done()
					r.installCustomTask(ctx, t, eventChan)
				}(task)
			}
			customWg.Wait()
		}()

		// --- Worker 2: Shell & Git Repos (goroutine) ---
		mainWg.Add(1)
		go func() {
			defer mainWg.Done()

			if needOMZ && runtime.GOOS != "windows" {
				r.installOhMyZsh(ctx, eventChan)
			}

			if len(gitRepos) > 0 {
				var gitWg sync.WaitGroup
				for _, repoPkg := range gitRepos {
					gitWg.Add(1)
					go func(pkg model.Package) {
						defer gitWg.Done()
						r.installGitRepo(ctx, pkg, eventChan)
					}(repoPkg)
				}
				gitWg.Wait()
			}

			if runtime.GOOS != "windows" {
				r.postConfigureShell(pkgs, eventChan)
			}
		}()

		// --- Worker 3: Package Manager Formulae & Casks (goroutine) ---
		mainWg.Add(1)
		go func() {
			defer mainWg.Done()

			if len(brewFormulae) == 0 && len(brewCasks) == 0 {
				return
			}

			switch runtime.GOOS {
			case "darwin":
				if !r.checker.IsBrewInstalled() {
					eventChan <- model.InstallEvent{
						PackageID: "brew",
						Name:      "Homebrew",
						Status:    model.StatusFailed,
						Message:   "Homebrew is not installed. Please install Homebrew first (brew.sh)",
					}
					return
				}

				if len(brewFormulae) > 0 {
					r.installBrewFormulae(ctx, brewFormulae, eventChan)
				}

				var javaWg sync.WaitGroup
				for _, p := range brewFormulae {
					if p.ID == "java" {
						javaWg.Add(1)
						go func() {
							defer javaWg.Done()
							r.verifyAndLinkJavaWithWatcher(ctx, eventChan)
						}()
						break
					}
				}

				if len(brewCasks) > 0 {
					r.installBrewCasks(ctx, brewCasks, eventChan)
				}

				javaWg.Wait()

			case "linux":
				allLinuxPkgs := append([]model.Package{}, brewFormulae...)
				allLinuxPkgs = append(allLinuxPkgs, brewCasks...)
				r.installLinuxPackages(ctx, allLinuxPkgs, eventChan)

			case "windows":
				allWinPkgs := append([]model.Package{}, brewFormulae...)
				allWinPkgs = append(allWinPkgs, brewCasks...)
				r.installWindowsPackages(ctx, allWinPkgs, eventChan)
			}
		}()

		mainWg.Wait()

		// Final check: if VS Code was targeted, verify code CLI
		for _, p := range pkgs {
			if p.ID == "vscode" || p.ID == "code-cli" {
				r.verifyAndLinkVSCode(eventChan)
				break
			}
		}

		// Execute PostInstall hook if configured
		if r.hooks.PostInstall != "" {
			eventChan <- model.InstallEvent{
				PackageID: "post-install-hook",
				Name:      "Post-Install Hook",
				Status:    model.StatusInstalling,
				Message:   "Executing post-install hook...",
			}
			cmd := model.ShellCommandContext(ctx, r.hooks.PostInstall)
			_ = cmd.Run()
		}
	}()

	return eventChan
}
