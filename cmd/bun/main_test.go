package main

import (
	"os"
	"testing"
)

func TestMainHelp(t *testing.T) {
	// Temporarily set os.Args to trigger help and avoid Exit(1)
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"bun", "--version"}
	main()
}
