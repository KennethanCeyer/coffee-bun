package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRootCmdFlags(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"--version"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing --version: %v", err)
	}

	cmd = NewRootCmd()
	cmd.SetArgs([]string{"--help"})
	buf.Reset()
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error executing --help: %v", err)
	}
}

func TestVersionCmd(t *testing.T) {
	cmd := newVersionCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestDoctorCmd(t *testing.T) {
	cmd := newDoctorCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("doctor command failed: %v", err)
	}
}

func TestCheckCmd(t *testing.T) {
	cfg := ""
	cmd := newCheckCmd(&cfg)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("check command failed: %v", err)
	}
}

func TestCheckCmdWithConfig(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "custom.yaml")
	yamlContent := `
version: "1.0"
packages:
  - id: custom-check
    name: "Custom Check Tool"
    type: "script"
    check_cmd: "echo 1.2.3"
`
	if err := os.WriteFile(cfgPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cmd := newCheckCmd(&cfgPath)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("check command with config failed: %v", err)
	}
}

func TestInstallCmdNonexistent(t *testing.T) {
	cfg := ""
	cmd := newInstallCmd(&cfg)
	cmd.SetArgs([]string{"nonexistent-package-id-9999"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Since nothing matches, it should finish saying nothing to install
	if err := cmd.Execute(); err != nil {
		t.Fatalf("install command with nonexistent target failed: %v", err)
	}
}
