package main

import (
	"context"
	"fmt"
	"os"

	"github.com/major/volumeleaders-agent/internal/auth"
	"github.com/major/volumeleaders-agent/internal/commands"
)

// version is set at build time via ldflags (see .goreleaser.yml).
var version = "dev"

func main() {
	if err := commands.NewApp(version).Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, userFacingError(err))
		os.Exit(exitCode(err))
	}
}

func userFacingError(err error) string {
	if auth.IsSessionExpired(err) {
		return auth.SessionExpiredMessage
	}
	return err.Error()
}

func exitCode(err error) int {
	if auth.IsSessionExpired(err) {
		return 2
	}
	return 1
}
