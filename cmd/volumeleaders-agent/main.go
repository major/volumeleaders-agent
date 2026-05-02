package main

import (
	"fmt"
	"os"

	"github.com/leodido/structcli"

	"github.com/major/volumeleaders-agent/internal/auth"
	cli "github.com/major/volumeleaders-agent/internal/cli"
)

// version is set at build time via ldflags (see .goreleaser.yml).
var version = "dev"

func main() {
	rootCmd := cli.NewRootCmd(version)
	cli.SetupCLI(rootCmd)
	c, err := structcli.ExecuteC(rootCmd)
	if err != nil {
		// Session expiration gets a dedicated exit code (2) and human-readable
		// message before structcli's structured error handler runs.
		if auth.IsSessionExpired(err) {
			fmt.Fprintln(os.Stderr, userFacingError(err))
			os.Exit(exitCode(err))
		}
		os.Exit(structcli.HandleError(c, err, os.Stderr))
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
