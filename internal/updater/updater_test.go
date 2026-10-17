package updater

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func TestNewUpdater(t *testing.T) {
	u := NewUpdater()
	if u == nil || u.checker == nil {
		t.Fatalf("expected NewUpdater to return initialized updater")
	}
}

func TestUpdateAllAsyncWithCanceledContext(t *testing.T) {
	u := NewUpdaterWithHome(t.TempDir(), true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	ch := u.UpdateAllAsync(ctx, nil)

	select {
	case _, ok := <-ch:
		if ok {
			for range ch {
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatalf("UpdateAllAsync channel did not close within timeout")
	}
}

func TestUpdateAllAsyncCustomAndGit(t *testing.T) {
	tmpHome := t.TempDir()
	u := NewUpdaterWithHome(tmpHome, true)

	// Create a local git repo to test updateGitRepo
	repoDir := t.TempDir()
	_ = exec.Command("git", "init", repoDir).Run()

	pkgs := []model.Package{
		{
			ID:        "custom-success",
			Name:      "Custom Success",
			Installed: true,
			UpdateCmd: "echo updated",
		},
		{
			ID:        "custom-fail",
			Name:      "Custom Fail",
			Installed: true,
			UpdateCmd: "exit 1",
		},
		{
			ID:          "custom-git",
			Name:        "Custom Git Repo",
			Type:        model.TypeGitRepo,
			InstallPath: repoDir,
			Installed:   true,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch := u.UpdateAllAsync(ctx, pkgs)
	var events []UpdateEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) == 0 {
		t.Fatalf("expected update events to be emitted")
	}
}
