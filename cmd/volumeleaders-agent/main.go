package main

import (
	"context"
	"fmt"
	"os"

	"github.com/major/volumeleaders-agent/internal/commands"
)

// version is set at build time via ldflags (see .goreleaser.yml).
var version = "dev"

func main() {
	if err := commands.NewApp(version).Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
