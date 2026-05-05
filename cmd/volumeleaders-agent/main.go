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
		// Session expiration gets a dedicated exit code (2) and human-readable
		// message before generic CLI error handling runs.
		if auth.IsSessionExpired(err) {
			fmt.Fprintln(os.Stderr, userFacingError(err))
			os.Exit(exitCode(err))
		}
		fmt.Fprintln(os.Stderr, err.Error())
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
	if strings.Contains(message, "unknown flag") {
		return 12
	}
	if strings.Contains(message, "required flag") {
		return 10
	}
	return 1
}
