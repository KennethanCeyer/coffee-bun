package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func (r *Runner) installCustomTask(ctx context.Context, pkg model.Package, ch chan<- model.InstallEvent) {
	if pkg.InstallCmd != "" || pkg.Type == model.TypeScript || pkg.Type == model.TypeCommand {
		r.executeCustomScript(ctx, pkg, ch)
		return
	}

	switch pkg.ID {
	case "meslo-font":
		r.installMesloFonts(ctx, ch)
	case "iterm2-font-config":
		r.configureITermFont(ch)
	case "brogrammer-theme":
		r.configureBrogrammerTheme(ctx, ch)
	case "chrome-default":
		r.configureChromeDefault(ch)
	case "vim-runtime":
		r.installVimRuntime(ctx, ch)
	case "code-cli":
		r.verifyAndLinkVSCode(ch)
	default:
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusSuccess,
			Message:   fmt.Sprintf("%s verified", pkg.Name),
		}
	}
}

func (r *Runner) executeCustomScript(ctx context.Context, pkg model.Package, ch chan<- model.InstallEvent) {
	ch <- model.InstallEvent{
		PackageID: pkg.ID,
		Name:      pkg.Name,
		Status:    model.StatusInstalling,
		Message:   fmt.Sprintf("Executing custom task: %s...", pkg.Name),
	}

	if pkg.RequiresSudo && pkg.SudoSnippet != "" && pkg.InstallCmd == "" {
		r.waitForCustomSudoAction(ctx, pkg, ch)
		return
	}

	if pkg.InstallCmd == "" {
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusSuccess,
			Message:   fmt.Sprintf("%s ready", pkg.Name),
		}
		return
	}

	cmd := model.ShellCommandContext(ctx, pkg.InstallCmd)
	if len(pkg.Env) > 0 {
		cmd.Env = os.Environ()
		for k, v := range pkg.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusInstalling,
			Message:   fmt.Sprintf("Installing %s...", pkg.Name),
			LogLine:   line,
		}
	})
	cmd.Stdout = lw
	cmd.Stderr = lw

	err := cmd.Run()
	lw.Flush()

	if err != nil {
		if pkg.RequiresSudo && pkg.SudoSnippet != "" {
			r.waitForCustomSudoAction(ctx, pkg, ch)
			return
		}
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusFailed,
			Message:   fmt.Sprintf("Command failed: %v", err),
			Error:     err,
		}
		return
	}

	ch <- model.InstallEvent{
		PackageID: pkg.ID,
		Name:      pkg.Name,
		Status:    model.StatusSuccess,
		Message:   fmt.Sprintf("%s completed successfully", pkg.Name),
	}
}

func (r *Runner) waitForCustomSudoAction(ctx context.Context, pkg model.Package, ch chan<- model.InstallEvent) {
	ch <- model.InstallEvent{
		PackageID:     pkg.ID,
		Name:          pkg.Name,
		Status:        model.StatusWaitingAction,
		Message:       fmt.Sprintf("%s requires administrator (sudo) permission. Please run the command below in another terminal window.", pkg.Name),
		ActionSnippet: pkg.SudoSnippet,
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkCopy := pkg
			r.checker.CheckPackage(&checkCopy)
			if checkCopy.Installed {
				ch <- model.InstallEvent{
					PackageID: pkg.ID,
					Name:      pkg.Name,
					Status:    model.StatusSuccess,
					Message:   fmt.Sprintf("%s detected and ready!", pkg.Name),
				}
				return
			}
		}
	}
}

func (r *Runner) installMesloFonts(ctx context.Context, ch chan<- model.InstallEvent) {
	home := r.configurator.HomeDir()
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	fontsDir := model.DefaultFontsDir(home)
	_ = os.MkdirAll(fontsDir, 0755)

	ch <- model.InstallEvent{
		PackageID: "meslo-font",
		Name:      "MesloLGS NF Font",
		Status:    model.StatusInstalling,
		Message:   fmt.Sprintf("Downloading MesloLGS Nerd Fonts into %s...", fontsDir),
	}

	fontURLs := []string{
		"https://github.com/romkatv/powerlevel10k-media/raw/master/MesloLGS%20NF%20Regular.ttf",
		"https://github.com/romkatv/powerlevel10k-media/raw/master/MesloLGS%20NF%20Bold.ttf",
		"https://github.com/romkatv/powerlevel10k-media/raw/master/MesloLGS%20NF%20Italic.ttf",
		"https://github.com/romkatv/powerlevel10k-media/raw/master/MesloLGS%20NF%20Bold%20Italic.ttf",
	}

	var downloadErr error

	for idx, rawURL := range fontURLs {
		fileName := strings.ReplaceAll(filepath.Base(rawURL), "%20", " ")
		dest := filepath.Join(fontsDir, fileName)
		if _, err := os.Stat(dest); err == nil {
			ch <- model.InstallEvent{
				PackageID: "meslo-font",
				Name:      "MesloLGS NF Font",
				Status:    model.StatusInstalling,
				Message:   "Downloading MesloLGS Nerd Fonts...",
				LogLine:   fmt.Sprintf("Font already present: %s (%d/4)", fileName, idx+1),
			}
			continue
		}

		ch <- model.InstallEvent{
			PackageID: "meslo-font",
			Name:      "MesloLGS NF Font",
			Status:    model.StatusInstalling,
			Message:   "Downloading MesloLGS Nerd Fonts...",
			LogLine:   fmt.Sprintf("Downloading %s (%d/4)...", fileName, idx+1),
		}

		req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
		if err != nil {
			downloadErr = err
			break
		}
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			downloadErr = err
			break
		}

		out, err := os.Create(dest)
		if err != nil {
			_ = resp.Body.Close()
			downloadErr = err
			break
		}

		_, copyErr := io.Copy(out, resp.Body)
		_ = out.Close()
		_ = resp.Body.Close()
		if copyErr != nil {
			downloadErr = copyErr
			break
		}
	}

	if downloadErr != nil {
		ch <- model.InstallEvent{
			PackageID: "meslo-font",
			Name:      "MesloLGS NF Font",
			Status:    model.StatusFailed,
			Message:   fmt.Sprintf("Failed to download fonts: %v", downloadErr),
			Error:     downloadErr,
		}
	} else {
		if runtime.GOOS == "linux" {
			if _, err := exec.LookPath("fc-cache"); err == nil {
				_ = exec.CommandContext(ctx, "fc-cache", "-f", fontsDir).Run()
			}
		}
		ch <- model.InstallEvent{
			PackageID: "meslo-font",
			Name:      "MesloLGS NF Font",
			Status:    model.StatusSuccess,
			Message:   fmt.Sprintf("Installed MesloLGS NF fonts to %s", fontsDir),
		}
	}
}

func (r *Runner) configureITermFont(ch chan<- model.InstallEvent) {
	if runtime.GOOS != "darwin" {
		return
	}
	ch <- model.InstallEvent{
		PackageID: "iterm2-font-config",
		Name:      "iTerm2 Font Config",
		Status:    model.StatusInstalling,
		Message:   "Configuring iTerm2 default profile font...",
		LogLine:   "defaults write com.googlecode.iterm2 'Normal Font' MesloLGS-NF-Regular 16",
	}

	home := r.configurator.HomeDir()
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	plistPath := filepath.Join(home, "Library", "Preferences", "com.googlecode.iterm2.plist")

	_ = exec.Command("defaults", "write", "com.googlecode.iterm2", "Normal Font", "-string", "MesloLGS-NF-Regular 16").Run()
	_ = exec.Command("defaults", "write", "com.googlecode.iterm2", "Use Non-ASCII Font", "-bool", "false").Run()
	_ = exec.Command("/usr/libexec/PlistBuddy", "-c", "Set :'New Bookmarks':0:'Normal Font' 'MesloLGS-NF-Regular 16'", plistPath).Run()

	ch <- model.InstallEvent{
		PackageID: "iterm2-font-config",
		Name:      "iTerm2 Font Config",
		Status:    model.StatusSuccess,
		Message:   "iTerm2 default font set to MesloLGS-NF-Regular 16",
	}
}

func (r *Runner) configureBrogrammerTheme(ctx context.Context, ch chan<- model.InstallEvent) {
	if runtime.GOOS != "darwin" {
		return
	}
	ch <- model.InstallEvent{
		PackageID: "brogrammer-theme",
		Name:      "Brogrammer Theme",
		Status:    model.StatusInstalling,
		Message:   "Importing Brogrammer color scheme into iTerm2...",
		LogLine:   "Fetching Brogrammer.itermcolors from GitHub...",
	}

	home := r.configurator.HomeDir()
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	themeFile := filepath.Join(home, "Brogrammer.itermcolors")

	if _, err := os.Stat(themeFile); os.IsNotExist(err) {
		themeURL := "https://raw.githubusercontent.com/mbadolato/iTerm2-Color-Schemes/master/schemes/Brogrammer.itermcolors"
		req, _ := http.NewRequestWithContext(ctx, "GET", themeURL, nil)
		resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
		if err == nil && resp.StatusCode == 200 {
			defer resp.Body.Close()
			f, err := os.Create(themeFile)
			if err == nil {
				_, _ = io.Copy(f, resp.Body)
				_ = f.Close()
			}
		}
	}

	if _, err := os.Stat(themeFile); err == nil {
		ch <- model.InstallEvent{
			PackageID: "brogrammer-theme",
			Name:      "Brogrammer Theme",
			Status:    model.StatusInstalling,
			Message:   "Importing preset into iTerm2...",
			LogLine:   "open Brogrammer.itermcolors",
		}
		_ = exec.Command("open", themeFile).Run()
		ch <- model.InstallEvent{
			PackageID: "brogrammer-theme",
			Name:      "Brogrammer Theme",
			Status:    model.StatusSuccess,
			Message:   "Brogrammer theme imported into iTerm2 presets",
		}
	} else {
		ch <- model.InstallEvent{
			PackageID: "brogrammer-theme",
			Name:      "Brogrammer Theme",
			Status:    model.StatusFailed,
			Message:   "Could not locate or download Brogrammer.itermcolors",
		}
	}
}

func (r *Runner) configureChromeDefault(ch chan<- model.InstallEvent) {
	if runtime.GOOS != "darwin" {
		return
	}
	ch <- model.InstallEvent{
		PackageID: "chrome-default",
		Name:      "Set Chrome as Default",
		Status:    model.StatusInstalling,
		Message:   "Configuring Google Chrome as default browser...",
		LogLine:   "open -a 'Google Chrome' --args --make-default-browser",
	}

	_ = exec.Command("open", "-a", "Google Chrome", "--args", "--make-default-browser").Run()

	ch <- model.InstallEvent{
		PackageID: "chrome-default",
		Name:      "Set Chrome as Default",
		Status:    model.StatusSuccess,
		Message:   "Google Chrome default browser request triggered",
	}
}

func (r *Runner) installVimRuntime(ctx context.Context, ch chan<- model.InstallEvent) {
	ch <- model.InstallEvent{
		PackageID: "vim-runtime",
		Name:      "Ultimate Vim Runtime",
		Status:    model.StatusInstalling,
		Message:   "Cloning amix/vimrc repository...",
	}

	home := r.configurator.HomeDir()
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	runtimeDir := filepath.Join(home, ".vim_runtime")

	if _, err := os.Stat(runtimeDir); os.IsNotExist(err) {
		lw := newLineWriter(func(line string) {
			ch <- model.InstallEvent{
				PackageID: "vim-runtime",
				Name:      "Ultimate Vim Runtime",
				Status:    model.StatusInstalling,
				Message:   "Cloning amix/vimrc repository...",
				LogLine:   line,
			}
		})
		cmd := exec.CommandContext(ctx, "git", "clone", "--progress", "--depth=1", "https://github.com/amix/vimrc.git", runtimeDir)
		cmd.Stdout = lw
		cmd.Stderr = lw
		err := cmd.Run()
		lw.Flush()
		if err != nil {
			ch <- model.InstallEvent{
				PackageID: "vim-runtime",
				Name:      "Ultimate Vim Runtime",
				Status:    model.StatusFailed,
				Message:   fmt.Sprintf("Failed to clone vimrc: %v", err),
				LogLine:   lw.LastLine(),
				Error:     err,
			}
			return
		}
	}

	ch <- model.InstallEvent{
		PackageID: "vim-runtime",
		Name:      "Ultimate Vim Runtime",
		Status:    model.StatusInstalling,
		Message:   "Configuring ~/.vimrc...",
	}

	vimrcPath := filepath.Join(home, ".vimrc")
	awesomeVimrc := `set runtimepath+=~/.vim_runtime

source ~/.vim_runtime/vimrcs/basic.vim
source ~/.vim_runtime/vimrcs/filetypes.vim
source ~/.vim_runtime/vimrcs/plugins_config.vim
source ~/.vim_runtime/vimrcs/extended.vim

try
source ~/.vim_runtime/my_configs.vim
catch
endtry
`
	if err := os.WriteFile(vimrcPath, []byte(awesomeVimrc), 0644); err != nil {
		ch <- model.InstallEvent{
			PackageID: "vim-runtime",
			Name:      "Ultimate Vim Runtime",
			Status:    model.StatusFailed,
			Message:   "Failed to write ~/.vimrc",
			Error:     err,
		}
	} else {
		ch <- model.InstallEvent{
			PackageID: "vim-runtime",
			Name:      "Ultimate Vim Runtime",
			Status:    model.StatusSuccess,
			Message:   "Ultimate Vim runtime installed successfully",
		}
	}
}

func (r *Runner) installOhMyZsh(ctx context.Context, ch chan<- model.InstallEvent) {
	home := r.configurator.HomeDir()
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	omzPath := filepath.Join(home, ".oh-my-zsh")

	ch <- model.InstallEvent{
		PackageID: "oh-my-zsh",
		Name:      "Oh My Zsh",
		Status:    model.StatusInstalling,
		Message:   "Cloning Oh My Zsh framework...",
	}

	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: "oh-my-zsh",
			Name:      "Oh My Zsh",
			Status:    model.StatusInstalling,
			Message:   "Cloning Oh My Zsh framework...",
			LogLine:   line,
		}
	})

	cmd := exec.CommandContext(ctx, "git", "clone", "--progress", "--depth=1", "https://github.com/ohmyzsh/ohmyzsh.git", omzPath)
	cmd.Stdout = lw
	cmd.Stderr = lw
	err := cmd.Run()
	lw.Flush()

	if err != nil {
		ch <- model.InstallEvent{
			PackageID: "oh-my-zsh",
			Name:      "Oh My Zsh",
			Status:    model.StatusFailed,
			Message:   fmt.Sprintf("Failed to clone Oh My Zsh: %v", err),
			LogLine:   lw.LastLine(),
			Error:     err,
		}
		return
	}

	zshrcPath := filepath.Join(home, ".zshrc")
	if _, err := os.Stat(zshrcPath); os.IsNotExist(err) {
		templatePath := filepath.Join(omzPath, "templates", "zshrc.zsh-template")
		if content, err := os.ReadFile(templatePath); err == nil {
			_ = os.WriteFile(zshrcPath, content, 0644)
		}
	}

	ch <- model.InstallEvent{
		PackageID: "oh-my-zsh",
		Name:      "Oh My Zsh",
		Status:    model.StatusSuccess,
		Message:   "Oh My Zsh installed successfully",
	}
}

func (r *Runner) installGitRepo(ctx context.Context, pkg model.Package, ch chan<- model.InstallEvent) {
	if _, err := os.Stat(pkg.InstallPath); err == nil {
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusSuccess,
			Message:   fmt.Sprintf("%s already cloned", pkg.Name),
		}
		return
	}

	ch <- model.InstallEvent{
		PackageID: pkg.ID,
		Name:      pkg.Name,
		Status:    model.StatusInstalling,
		Message:   fmt.Sprintf("Cloning %s...", pkg.Name),
	}

	_ = os.MkdirAll(filepath.Dir(pkg.InstallPath), 0755)
	lw := newLineWriter(func(line string) {
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusInstalling,
			Message:   fmt.Sprintf("Cloning %s...", pkg.Name),
			LogLine:   line,
		}
	})
	cmd := exec.CommandContext(ctx, "git", "clone", "--progress", "--depth=1", pkg.GitURL, pkg.InstallPath)
	cmd.Stdout = lw
	cmd.Stderr = lw
	err := cmd.Run()
	lw.Flush()

	if err != nil {
		ch <- model.InstallEvent{
			PackageID: pkg.ID,
			Name:      pkg.Name,
			Status:    model.StatusFailed,
			Message:   fmt.Sprintf("Failed to clone %s: %v", pkg.Name, err),
			LogLine:   lw.LastLine(),
			Error:     err,
		}
		return
	}

	ch <- model.InstallEvent{
		PackageID: pkg.ID,
		Name:      pkg.Name,
		Status:    model.StatusSuccess,
		Message:   fmt.Sprintf("%s installed", pkg.Name),
	}
}

func (r *Runner) postConfigureShell(installed []model.Package, ch chan<- model.InstallEvent) {
	var hasP10k, hasZoxide bool
	var plugins []string

	for _, p := range installed {
		if p.ID == "powerlevel10k" {
			hasP10k = true
		}
		if p.ID == "zoxide" {
			hasZoxide = true
		}
		if p.ID == "zsh-autosuggestions" {
			plugins = append(plugins, "zsh-autosuggestions")
		}
		if p.ID == "zsh-syntax-highlighting" {
			plugins = append(plugins, "zsh-syntax-highlighting")
		}
	}

	if hasP10k || hasZoxide || len(plugins) > 0 {
		_, _ = r.configurator.BackupZshrc()

		if hasP10k {
			_ = r.configurator.ConfigurePowerlevel10k()
		}
		if hasZoxide {
			_ = r.configurator.ConfigureZoxide()
		}
		if len(plugins) > 0 {
			_ = r.configurator.ConfigurePlugins(plugins)
		}

		ch <- model.InstallEvent{
			PackageID: "shell-config",
			Name:      "Shell Integration",
			Status:    model.StatusSuccess,
			Message:   "Configured .zshrc (theme, plugins, and zoxide integration)",
		}
	}
}
