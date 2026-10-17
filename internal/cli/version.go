package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/KennethanCeyer/coffee-bun/internal/ui"
)

const Version = "1.0.0"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("\n" + ui.RenderASCIIHeader(0) + "\n")
			fmt.Printf("  Coffee Bun v%s (%s/%s)\n\n", Version, runtime.GOOS, runtime.GOARCH)
		},
	}
}
