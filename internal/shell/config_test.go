package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfiguratorZshrc(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "bun-shell-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	mockZshrc := filepath.Join(tempDir, ".zshrc")
	initialContent := `# Initial zshrc
export PATH="/bin:/usr/bin"
ZSH_THEME="robbyrussell"
plugins=(git)
`
	if err := os.WriteFile(mockZshrc, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to write mock zshrc: %v", err)
	}

	cfg := &Configurator{
		homeDir: tempDir,
		zshrc:   mockZshrc,
	}

	// 1. Backup test
	backup, err := cfg.BackupZshrc()
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if _, err := os.Stat(backup); os.IsNotExist(err) {
		t.Fatalf("backup file does not exist: %s", backup)
	}

	// 2. Powerlevel10k configuration test
	if err := cfg.ConfigurePowerlevel10k(); err != nil {
		t.Fatalf("configure p10k failed: %v", err)
	}
	data, _ := os.ReadFile(mockZshrc)
	if !strings.Contains(string(data), `ZSH_THEME="powerlevel10k/powerlevel10k"`) {
		t.Errorf("expected p10k theme in zshrc, got:\n%s", string(data))
	}

	// Idempotent test
	if err := cfg.ConfigurePowerlevel10k(); err != nil {
		t.Fatalf("second configure p10k failed: %v", err)
	}

	// 3. Plugins configuration test
	if err := cfg.ConfigurePlugins([]string{"zsh-autosuggestions", "zsh-syntax-highlighting"}); err != nil {
		t.Fatalf("configure plugins failed: %v", err)
	}
	data, _ = os.ReadFile(mockZshrc)
	if !strings.Contains(string(data), "zsh-autosuggestions") || !strings.Contains(string(data), "zsh-syntax-highlighting") {
		t.Errorf("expected added plugins in zshrc, got:\n%s", string(data))
	}

	// Idempotent test
	if err := cfg.ConfigurePlugins([]string{"zsh-autosuggestions"}); err != nil {
		t.Fatalf("second configure plugins failed: %v", err)
	}

	// 4. Zoxide configuration test
	if err := cfg.ConfigureZoxide(); err != nil {
		t.Fatalf("configure zoxide failed: %v", err)
	}
	data, _ = os.ReadFile(mockZshrc)
	if !strings.Contains(string(data), "eval \"$(zoxide init zsh)\"") {
		t.Errorf("expected zoxide in zshrc, got:\n%s", string(data))
	}

	// Idempotent test
	if err := cfg.ConfigureZoxide(); err != nil {
		t.Fatalf("second configure zoxide failed: %v", err)
	}
}

func TestConfiguratorFreshZshrc(t *testing.T) {
	tempDir := t.TempDir()
	freshZshrc := filepath.Join(tempDir, ".zshrc")

	cfg := &Configurator{
		homeDir: tempDir,
		zshrc:   freshZshrc,
	}

	// When no zshrc exists, readOrCreateZshrc creates a new one
	if err := cfg.ConfigurePlugins([]string{"fzf"}); err != nil {
		t.Fatalf("failed to configure plugins on fresh zshrc: %v", err)
	}

	data, err := os.ReadFile(freshZshrc)
	if err != nil {
		t.Fatalf("failed to read created zshrc: %v", err)
	}
	if !strings.Contains(string(data), "plugins=(git fzf)") {
		t.Errorf("expected fresh plugins list, got:\n%s", string(data))
	}
}

func TestNewConfigurator(t *testing.T) {
	cfg := NewConfigurator()
	if cfg == nil || cfg.homeDir == "" || cfg.zshrc == "" {
		t.Fatalf("expected NewConfigurator to populate homeDir and zshrc")
	}
}
