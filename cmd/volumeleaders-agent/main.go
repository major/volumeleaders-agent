package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/major/volumeleaders-agent/internal/auth"
	cli "github.com/major/volumeleaders-agent/internal/cli"
)

// version is set at build time via ldflags (see .goreleaser.yml).
var version = "dev"

func main() {
	rootCmd := cli.NewRootCmd(version)
	cli.SetupCLI(rootCmd)
	_, err := rootCmd.ExecuteC()
	if err != nil {
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
	message := err.Error()
	if strings.Contains(message, "unknown flag") || strings.Contains(message, "unknown shorthand flag") {
		return 12
	}
	if strings.Contains(message, "required flag") {
		return 10
	}
	return 1
}
