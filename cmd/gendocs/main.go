package main

import (
	"fmt"
	"log/slog"
	"os"

	cli "github.com/major/volumeleaders-agent/internal/cli"
	"github.com/major/volumeleaders-agent/internal/cli/docs"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <output-dir>\n", os.Args[0])
		os.Exit(1)
	}

	outputDir := os.Args[1]

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		slog.Error("failed to create output directory", "error", err)
		os.Exit(1)
	}

	rootCmd := cli.NewRootCmd("dev")

	if err := docs.GenerateDocs(rootCmd, outputDir); err != nil {
		slog.Error("failed to generate docs", "error", err)
		os.Exit(1)
	}
}
