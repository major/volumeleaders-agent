package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/major/volumeleaders-agent/internal/discovery"
)

func main() {
	outputDir := flag.String("output", discovery.DefaultOutputDir, "directory for generated discovery files")
	skillPath := flag.String("skill", discovery.DefaultSkillPath, "path for the generated root skill file")
	version := flag.String("version", "dev", "version to record in generated discovery files")
	flag.Parse()

	if err := discovery.Generate(*outputDir, *skillPath, *version); err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
		logger.Error("failed to generate discovery files", "error", err)
		os.Exit(1)
	}
}
