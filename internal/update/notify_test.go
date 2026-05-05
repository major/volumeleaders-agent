package update

import (
	"os"
	"testing"
)

func TestShouldSkipNotification(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		commandPath string
		ci          string
		want        bool
	}{
		{name: "dev version", version: "dev", commandPath: "volumeleaders-agent trade list", want: true},
		{name: "empty version", version: "", commandPath: "volumeleaders-agent trade list", want: true},
		{name: "ci", version: "0.8.1", commandPath: "volumeleaders-agent trade list", ci: "true", want: true},
		{name: "update command", version: "0.8.1", commandPath: "volumeleaders-agent update check", want: true},
		{name: "non interactive", version: "0.8.1", commandPath: "volumeleaders-agent trade list", want: true},
		// The non-skip case requires both os.Stdout and os.Stderr to be real TTY character devices.
		// Keep this unit test deterministic by replacing them with regular files and covering the
		// positive terminal behavior through manual CLI smoke checks in an interactive shell.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CI", tt.ci)
			stdout, stderr := replaceStandardFiles(t)
			defer stdout.Close()
			defer stderr.Close()
			if got := shouldSkipNotification(tt.version, tt.commandPath); got != tt.want {
				t.Fatalf("shouldSkipNotification(%q, %q) = %v, want %v", tt.version, tt.commandPath, got, tt.want)
			}
		})
	}
}

func replaceStandardFiles(t *testing.T) (stdout, stderr *os.File) {
	t.Helper()
	originalStdout := os.Stdout
	originalStderr := os.Stderr
	stdout, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatalf("create stdout replacement: %v", err)
	}
	stderr, err = os.CreateTemp(t.TempDir(), "stderr")
	if err != nil {
		_ = stdout.Close()
		t.Fatalf("create stderr replacement: %v", err)
	}
	os.Stdout = stdout
	os.Stderr = stderr
	t.Cleanup(func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})
	return stdout, stderr
}
