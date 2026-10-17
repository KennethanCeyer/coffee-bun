package installer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func (r *Runner) installBrewFormulae(ctx context.Context, pkgs []model.Package, ch chan<- model.InstallEvent) {
	for _, p := range pkgs {
		r.installSingleBrewFormula(ctx, p, ch)
	}
}

func (r *Runner) installSingleBrewFormula(ctx context.Context, p model.Package, ch chan<- model.InstallEvent) {
	ch <- model.InstallEvent{
		PackageID: p.ID,
		Name:      p.Name,
		Status:    model.StatusInstalling,
		Message:   "Installing formula via Homebrew...",
	}

	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusInstalling,
			Message:   "Installing formula via Homebrew...",
			LogLine:   line,
		}
	})

	cmd := exec.CommandContext(ctx, "brew", "install", p.BrewName)
	cmd.Stdout = lw
	cmd.Stderr = lw
	err := cmd.Run()
	lw.Flush()

	if err != nil {
		if _, lookErr := exec.LookPath(p.BinaryName); lookErr == nil {
			ch <- model.InstallEvent{
				PackageID: p.ID,
				Name:      p.Name,
				Status:    model.StatusSuccess,
				Message:   "Installed successfully",
			}
			return
		}
		last := lw.LastLine()
		if last == "" {
			last = err.Error()
		}
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusFailed,
			Message:   "Homebrew install failed",
			LogLine:   last,
			Error:     err,
		}
	} else {
		if p.ID == "java" && exec.Command("java", "-version").Run() != nil {
			// verifyAndLinkJavaWithWatcher will transition Java to StatusWaitingAction and then StatusSuccess
			return
		}
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusSuccess,
			Message:   "Installed successfully via Homebrew",
		}
	}
}

func (r *Runner) installBrewCasks(ctx context.Context, pkgs []model.Package, ch chan<- model.InstallEvent) {
	for _, p := range pkgs {
		r.installSingleBrewCask(ctx, p, ch)
	}
}

func (r *Runner) installSingleBrewCask(ctx context.Context, p model.Package, ch chan<- model.InstallEvent) {
	ch <- model.InstallEvent{
		PackageID: p.ID,
		Name:      p.Name,
		Status:    model.StatusInstalling,
		Message:   fmt.Sprintf("Downloading and installing %s cask...", p.Name),
	}

	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusInstalling,
			Message:   fmt.Sprintf("Installing %s cask...", p.Name),
			LogLine:   line,
		}
	})

	cmd := exec.CommandContext(ctx, "brew", "install", "--cask", p.BrewName)
	cmd.Stdout = lw
	cmd.Stderr = lw
	err := cmd.Run()
	lw.Flush()

	if err != nil {
		if p.AppPath != "" {
			if _, statErr := os.Stat(p.AppPath); statErr == nil {
				ch <- model.InstallEvent{
					PackageID: p.ID,
					Name:      p.Name,
					Status:    model.StatusSuccess,
					Message:   "Detected installed application",
				}
				return
			}
		}
		last := lw.LastLine()
		if last == "" {
			last = err.Error()
		}
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusFailed,
			Message:   fmt.Sprintf("Cask install failed for %s", p.Name),
			LogLine:   last,
			Error:     err,
		}
	} else {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusSuccess,
			Message:   fmt.Sprintf("%s installed successfully", p.Name),
		}
	}
}

func (r *Runner) installLinuxPackages(ctx context.Context, pkgs []model.Package, ch chan<- model.InstallEvent) {
	hasApt := false
	if _, err := exec.LookPath("apt-get"); err == nil {
		hasApt = true
		// Update apt cache once before installing packages
		if os.Geteuid() == 0 {
			_ = exec.CommandContext(ctx, "apt-get", "update", "-y").Run()
		} else {
			_ = exec.CommandContext(ctx, "sudo", "-n", "apt-get", "update", "-y").Run()
		}
	}

	for _, p := range pkgs {
		r.installSingleLinuxPackage(ctx, p, hasApt, ch)
	}
}

func (r *Runner) installSingleLinuxPackage(ctx context.Context, p model.Package, hasApt bool, ch chan<- model.InstallEvent) {
	pkgName := p.AptName
	if pkgName == "" {
		pkgName = p.BrewName
	}
	if pkgName == "" {
		pkgName = p.BinaryName
	}

	ch <- model.InstallEvent{
		PackageID: p.ID,
		Name:      p.Name,
		Status:    model.StatusInstalling,
		Message:   fmt.Sprintf("Installing %s via Linux package manager...", p.Name),
	}

	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusInstalling,
			Message:   fmt.Sprintf("Installing %s...", p.Name),
			LogLine:   line,
		}
	})

	var cmd *exec.Cmd
	if hasApt && pkgName != "" {
		if os.Geteuid() == 0 {
			cmd = exec.CommandContext(ctx, "apt-get", "install", "-y", pkgName)
		} else {
			cmd = exec.CommandContext(ctx, "sudo", "-n", "apt-get", "install", "-y", pkgName)
		}
		cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	} else if r.checker.IsBrewInstalled() && p.BrewName != "" {
		cmd = exec.CommandContext(ctx, "brew", "install", p.BrewName)
	} else if _, err := exec.LookPath("dnf"); err == nil && pkgName != "" {
		if os.Geteuid() == 0 {
			cmd = exec.CommandContext(ctx, "dnf", "install", "-y", pkgName)
		} else {
			cmd = exec.CommandContext(ctx, "sudo", "-n", "dnf", "install", "-y", pkgName)
		}
	} else if _, err := exec.LookPath("apk"); err == nil && pkgName != "" {
		if os.Geteuid() == 0 {
			cmd = exec.CommandContext(ctx, "apk", "add", "--no-cache", pkgName)
		} else {
			cmd = exec.CommandContext(ctx, "sudo", "-n", "apk", "add", "--no-cache", pkgName)
		}
	}

	if cmd == nil {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusFailed,
			Message:   "No supported Linux package manager found (apt-get, dnf, apk, or brew)",
		}
		return
	}

	cmd.Stdout = lw
	cmd.Stderr = lw
	err := cmd.Run()
	lw.Flush()

	// If apt Failed and Homebrew is installed on Linux, try Homebrew as a fallback
	if err != nil && hasApt && r.checker.IsBrewInstalled() && p.BrewName != "" {
		brewCmd := exec.CommandContext(ctx, "brew", "install", p.BrewName)
		brewCmd.Stdout = lw
		brewCmd.Stderr = lw
		err = brewCmd.Run()
		lw.Flush()
	}

	if err != nil {
		if p.BinaryName != "" {
			if _, lookErr := exec.LookPath(p.BinaryName); lookErr == nil {
				ch <- model.InstallEvent{
					PackageID: p.ID,
					Name:      p.Name,
					Status:    model.StatusSuccess,
					Message:   "Installed successfully",
				}
				return
			}
		}
		last := lw.LastLine()
		if last == "" {
			last = err.Error()
		}
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusFailed,
			Message:   fmt.Sprintf("Linux package install failed for %s", p.Name),
			LogLine:   last,
			Error:     err,
		}
		return
	}

	// On Debian/Ubuntu, `bat` installs as `/usr/bin/batcat`; symlink `~/.local/bin/bat`
	if p.ID == "bat" {
		if batcatPath, lookErr := exec.LookPath("batcat"); lookErr == nil {
			if _, batErr := exec.LookPath("bat"); batErr != nil {
				localBinDir := filepath.Join(r.configurator.HomeDir(), ".local", "bin")
				_ = os.MkdirAll(localBinDir, 0755)
				_ = os.Symlink(batcatPath, filepath.Join(localBinDir, "bat"))
			}
		}
	}

	ch <- model.InstallEvent{
		PackageID: p.ID,
		Name:      p.Name,
		Status:    model.StatusSuccess,
		Message:   fmt.Sprintf("%s installed successfully", p.Name),
	}
}

func (r *Runner) installWindowsPackages(ctx context.Context, pkgs []model.Package, ch chan<- model.InstallEvent) {
	hasWinget := false
	if _, err := exec.LookPath("winget"); err == nil {
		hasWinget = true
	}
	hasChoco := false
	if _, err := exec.LookPath("choco"); err == nil {
		hasChoco = true
	}

	for _, p := range pkgs {
		r.installSingleWindowsPackage(ctx, p, hasWinget, hasChoco, ch)
	}
}

func (r *Runner) installSingleWindowsPackage(ctx context.Context, p model.Package, hasWinget, hasChoco bool, ch chan<- model.InstallEvent) {
	ch <- model.InstallEvent{
		PackageID: p.ID,
		Name:      p.Name,
		Status:    model.StatusInstalling,
		Message:   fmt.Sprintf("Installing %s on Windows...", p.Name),
	}

	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusInstalling,
			Message:   fmt.Sprintf("Installing %s...", p.Name),
			LogLine:   line,
		}
	})

	var err error
	installed := false

	// Prefer Chocolatey when available in non-interactive environments or try winget first
	if hasWinget && p.WingetID != "" {
		cmd := exec.CommandContext(ctx, "winget", "install", "--id", p.WingetID, "-e", "--accept-source-agreements", "--accept-package-agreements", "--disable-interactivity")
		cmd.Stdout = lw
		cmd.Stderr = lw
		err = cmd.Run()
		lw.Flush()
		if err == nil {
			installed = true
		}
	}

	if !installed && hasChoco && p.ChocoName != "" {
		cmd := exec.CommandContext(ctx, "choco", "install", p.ChocoName, "-y", "--no-progress")
		cmd.Stdout = lw
		cmd.Stderr = lw
		err = cmd.Run()
		lw.Flush()
		if err == nil {
			installed = true
		}
	}

	if installed {
		ch <- model.InstallEvent{
			PackageID: p.ID,
			Name:      p.Name,
			Status:    model.StatusSuccess,
			Message:   fmt.Sprintf("%s installed successfully", p.Name),
		}
		return
	}

	if p.BinaryName != "" {
		if _, lookErr := exec.LookPath(p.BinaryName); lookErr == nil {
			ch <- model.InstallEvent{
				PackageID: p.ID,
				Name:      p.Name,
				Status:    model.StatusSuccess,
				Message:   "Installed successfully",
			}
			return
		}
	}

	last := lw.LastLine()
	if last == "" && err != nil {
		last = err.Error()
	}
	ch <- model.InstallEvent{
		PackageID: p.ID,
		Name:      p.Name,
		Status:    model.StatusFailed,
		Message:   fmt.Sprintf("Windows package install failed for %s", p.Name),
		LogLine:   last,
		Error:     err,
	}
}
