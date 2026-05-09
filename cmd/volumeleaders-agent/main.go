package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/major/volumeleaders-agent/internal/auth"
	cli "github.com/major/volumeleaders-agent/internal/cli"
)

// version is set at build time via ldflags (see .goreleaser.yml).
var version = "dev"

// revision is set by local builds so dev binaries identify their source commit.
var revision = ""

func main() {
	rootCmd := cli.NewRootCmd(displayVersion(version, buildRevision()))
	cli.SetupCLI(rootCmd)
	cli.ConfigureCompletions(rootCmd)
	_, err := rootCmd.ExecuteC()
	if err != nil {
		fmt.Fprintln(os.Stderr, userFacingError(err))
		os.Exit(exitCode(err))
	}
}

func buildRevision() string {
	if revision != "" {
		return revision
	}
	return vcsRevision()
}

func displayVersion(version, revision string) string {
	if version != "dev" || revision == "" {
		return version
	}
	return fmt.Sprintf("dev-%s", shortRevision(revision))
}

func shortRevision(revision string) string {
	const shortLength = 7
	if len(revision) <= shortLength {
		return revision
	}
	return revision[:shortLength]
}

func vcsRevision() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return setting.Value
		}
	}
	return ""
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
