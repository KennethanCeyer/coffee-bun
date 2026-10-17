package ui

import (
	"strings"
	"testing"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

func TestPrintStatusSummary(t *testing.T) {
	pkgs := []model.Package{
		{
			ID:          "go",
			Name:        "Go",
			Category:    model.CategoryLanguages,
			Installed:   true,
			Version:     "1.26.0",
		},
		{
			ID:          "python",
			Name:        "Python 3",
			Category:    model.CategoryLanguages,
			Installed:   true,
		},
		{
			ID:          "aws",
			Name:        "AWS CLI",
			Category:    model.CategoryCloud,
			Installed:   false,
		},
		{
			ID:          "acme-tool",
			Name:        "Acme Tool",
			Category:    model.Category("Acme Internal"),
			Installed:   true,
			Version:     "v1.0",
		},
	}

	output := PrintStatusSummary(pkgs)
	if !strings.Contains(output, "Go") {
		t.Errorf("output missing 'Go'")
	}
	if !strings.Contains(output, "AWS CLI") {
		t.Errorf("output missing 'AWS CLI'")
	}
	if !strings.Contains(output, "Acme Internal") {
		t.Errorf("output missing custom category 'Acme Internal'")
	}
	if !strings.Contains(output, "3 installed, 1 missing") {
		t.Errorf("expected '3 installed, 1 missing' in summary, got:\n%s", output)
	}

	// Test all up to date
	allUpToDate := []model.Package{
		{ID: "go", Name: "Go", Category: model.CategoryLanguages, Installed: true, Version: "1.26.0"},
	}
	out2 := PrintStatusSummary(allUpToDate)
	if !strings.Contains(out2, "installed and up to date") {
		t.Errorf("expected up to date message in output, got:\n%s", out2)
	}
}

func TestStylesHelpers(t *testing.T) {
	b := Bold("test")
	if !strings.Contains(b, "test") {
		t.Errorf("Bold helper failed")
	}

	btn := renderButton("TestButton", true, true)
	if !strings.Contains(btn, "TestButton") {
		t.Errorf("renderButton active failed")
	}

	btnInactive := renderButton("InactiveButton", false, false)
	if !strings.Contains(btnInactive, "InactiveButton") {
		t.Errorf("renderButton inactive failed")
	}
}
