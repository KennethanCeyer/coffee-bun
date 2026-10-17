package ui

import (
	"fmt"
	"strings"

	"github.com/KennethanCeyer/coffee-bun/internal/model"
)

// PrintStatusSummary renders a clean terminal table of all monitored packages
func PrintStatusSummary(pkgs []model.Package) string {
	var b strings.Builder

	b.WriteString("\n" + RenderASCIIHeader(0) + "\n\n")

	var categories []model.Category
	seenCats := make(map[model.Category]bool)
	for _, p := range pkgs {
		if !seenCats[p.Category] {
			seenCats[p.Category] = true
			categories = append(categories, p.Category)
		}
	}

	for _, cat := range categories {
		b.WriteString(CategoryHeaderStyle.Render(fmt.Sprintf("  [%s]", string(cat))) + "\n")
		b.WriteString(MutedStyle.Render("  "+strings.Repeat("-", 68)) + "\n")

		var catPkgs []model.Package
		for _, p := range pkgs {
			if p.Category == cat {
				catPkgs = append(catPkgs, p)
			}
		}

		for _, p := range catPkgs {
			var statusBadge string
			var versionInfo string

			if p.Installed {
				statusBadge = BadgeSuccess.Render("  [OK]      ")
				if p.Version != "" {
					versionInfo = MutedStyle.Render(fmt.Sprintf("(%s)", p.Version))
				} else {
					versionInfo = MutedStyle.Render("(ready)")
				}
			} else {
				statusBadge = BadgeMissing.Render("  [MISSING] ")
				versionInfo = MutedStyle.Render("- not installed")
			}

			namePadded := fmt.Sprintf("%-26s", p.Name)
			b.WriteString(fmt.Sprintf("%s %s %s\n", statusBadge, NormalStyle.Render(namePadded), versionInfo))
		}
		b.WriteString("\n")
	}

	// Statistics
	var total, installed, missing int
	for _, p := range pkgs {
		total++
		if p.Installed {
			installed++
		} else {
			missing++
		}
	}

	summaryLine := fmt.Sprintf("  Summary: %s installed, %s missing (Total: %d)",
		BadgeSuccess.Render(fmt.Sprintf("%d", installed)),
		BadgeMissing.Render(fmt.Sprintf("%d", missing)),
		total,
	)
	b.WriteString(summaryLine + "\n\n")

	return b.String()
}
