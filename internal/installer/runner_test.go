package installer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func TestCleanLogLine(t *testing.T) {
	// Test ANSI code removal
	input := "\x1b[32m==> Pouring openjdk--21.0.2.arm64_sonoma.bottle.tar.gz\x1b[0m"
	expected := "==> Pouring openjdk--21.0.2.arm64_sonoma.bottle.tar.gz"
	if got := cleanLogLine(input); got != expected {
		t.Errorf("cleanLogLine() = %q, want %q", got, expected)
	}

	// Test emoji removal (e.g. brew beer emoji \U0001F37A)
	inputWithEmoji := "\U0001F37A  /opt/homebrew/Cellar/openjdk/21.0.2: 641 files, 331.4MB"
	got := cleanLogLine(inputWithEmoji)
	if strings.Contains(got, "\U0001F37A") {
		t.Errorf("cleanLogLine() still contains emoji: %q", got)
	}
	if !strings.Contains(got, "/opt/homebrew/Cellar/openjdk/21.0.2") {
		t.Errorf("cleanLogLine() lost text: %q", got)
	}

	// Test long string truncation
	longInput := strings.Repeat("A", 120)
	cleanedLong := cleanLogLine(longInput)
	if len(cleanedLong) > 90 || !strings.HasSuffix(cleanedLong, "...") {
		t.Errorf("expected long string to be truncated to <=90 with ..., got length %d", len(cleanedLong))
	}
}

func TestLineWriterStreaming(t *testing.T) {
	var captured []string
	lw := newLineWriter(func(line string) {
		captured = append(captured, line)
	})

	// Simulate streaming writes with \r (progress updates) and \n
	_, _ = lw.Write([]byte("Receiving objects:  10%\rReceiving objects:  50%\rReceiving objects: 100%\n"))
	lw.Flush()

	if len(captured) < 3 {
		t.Fatalf("Expected at least 3 captured lines, got %d: %v", len(captured), captured)
	}
	if captured[len(captured)-1] != "Receiving objects: 100%" {
		t.Errorf("Last captured line = %q, want %q", captured[len(captured)-1], "Receiving objects: 100%")
	}
	if lw.LastLine() != "Receiving objects: 100%" {
		t.Errorf("lw.LastLine() = %q, want %q", lw.LastLine(), "Receiving objects: 100%")
	}
}

func TestCustomScriptExecution(t *testing.T) {
	if r := NewRunner(); r == nil {
		t.Fatalf("expected NewRunner to return non-nil runner")
	}
	runner := NewRunnerWithHome(t.TempDir())
	pkg := model.Package{
		ID:         "test-echo",
		Name:       "Test Echo",
		Type:       model.TypeScript,
		InstallCmd: "echo 'hello custom runner'",
		Env:        map[string]string{"TEST_ENV": "123"},
	}

	ch := make(chan model.InstallEvent, 10)
	runner.installCustomTask(t.Context(), pkg, ch)
	close(ch)

	var successFound bool
	for ev := range ch {
		if ev.Status == model.StatusSuccess {
			successFound = true
		}
	}

	if !successFound {
		t.Errorf("expected custom script to produce success status")
	}

	// Test failing custom script
	failingPkg := model.Package{
		ID:         "test-fail",
		Name:       "Test Fail",
		Type:       model.TypeScript,
		InstallCmd: "exit 1",
	}
	chFail := make(chan model.InstallEvent, 10)
	runner.installCustomTask(t.Context(), failingPkg, chFail)
	close(chFail)

	var failFound bool
	for ev := range chFail {
		if ev.Status == model.StatusFailed {
			failFound = true
		}
	}
	if !failFound {
		t.Errorf("expected failing script to produce failed status")
	}
}

func TestInstallAsyncEmpty(t *testing.T) {
	runner := NewRunnerWithHome(t.TempDir())
	ch := runner.InstallAsync(t.Context(), nil)

	var gotEvent bool
	for ev := range ch {
		gotEvent = true
		if ev.Status != model.StatusSuccess {
			t.Errorf("expected success for empty install list, got %v", ev.Status)
		}
	}
	if !gotEvent {
		t.Errorf("expected at least one event from InstallAsync on empty slice")
	}
}

func TestInstallAsyncWithHooks(t *testing.T) {
	runner := NewRunnerWithHome(t.TempDir())
	runner.SetHooks(model.Hooks{
		PreInstall:  "echo 'pre hook running'",
		PostInstall: "echo 'post hook running'",
	})

	pkgs := []model.Package{
		{
			ID:         "mock-task",
			Name:       "Mock Task",
			Type:       model.TypeScript,
			InstallCmd: "echo 'mock install'",
		},
	}

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()

	ch := runner.InstallAsync(ctx, pkgs)
	var preFound, postFound, pkgSuccess bool

	for ev := range ch {
		if ev.PackageID == "pre-install-hook" {
			preFound = true
		}
		if ev.PackageID == "post-install-hook" {
			postFound = true
		}
		if ev.PackageID == "mock-task" && ev.Status == model.StatusSuccess {
			pkgSuccess = true
		}
	}

	if !preFound {
		t.Errorf("expected pre-install hook event")
	}
	if !postFound {
		t.Errorf("expected post-install hook event")
	}
	if !pkgSuccess {
		t.Errorf("expected mock package success event")
	}
}

func TestPostConfigureShell(t *testing.T) {
	runner := NewRunnerWithHome(t.TempDir())
	ch := make(chan model.InstallEvent, 10)

	pkgs := []model.Package{
		{ID: "powerlevel10k"},
		{ID: "zoxide"},
		{ID: "zsh-autosuggestions"},
		{ID: "zsh-syntax-highlighting"},
	}

	runner.postConfigureShell(pkgs, ch)
	close(ch)

	var shellConfigFound bool
	for ev := range ch {
		if ev.PackageID == "shell-config" && ev.Status == model.StatusSuccess {
			shellConfigFound = true
		}
	}

	if !shellConfigFound {
		t.Errorf("expected shell-config success event")
	}
}

func TestVerifyAndLinkVSCode(t *testing.T) {
	runner := NewRunnerWithHome(t.TempDir())
	ch := make(chan model.InstallEvent, 10)
	runner.verifyAndLinkVSCode(ch)
	close(ch)

	for ev := range ch {
		t.Logf("verifyAndLinkVSCode event: %+v", ev)
	}
}
