package main

import (
	"flag"
	"log"

	"github.com/major/volumeleaders-agent/internal/discovery"
)

func main() {
	outputDir := flag.String("output", discovery.DefaultOutputDir, "directory for generated discovery files")
	version := flag.String("version", "dev", "version to record in generated discovery files")
	flag.Parse()

	if err := discovery.Generate(*outputDir, *version); err != nil {
		log.Fatal(err)
	}
}
