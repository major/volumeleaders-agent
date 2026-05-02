package main

import (
	"fmt"
	"os"

	"github.com/major/volumeleaders-agent/internal/auth"
	cli "github.com/major/volumeleaders-agent/internal/cli"
)

// version is set at build time via ldflags (see .goreleaser.yml).
var version = "dev"

func main() {
	rootCmd := cli.NewRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
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
