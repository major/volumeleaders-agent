package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLiveRunnerRejectsSensitiveOutputBeforeVerbose(t *testing.T) {
	t.Parallel()

	runner := liveRunner{binary: writeFixtureBinary(t, `#!/bin/sh
printf '%s\n' 'x-xsrf-token: secret-value' >&2
exit 1
`), timeout: time.Second, verbose: true}

	_, err := runner.run(context.Background(), "trade", "list")
	if err == nil {
		t.Fatal("expected sensitive output error")
	}
	if strings.Contains(err.Error(), "secret-value") {
		t.Fatalf("error leaked sensitive value: %v", err)
	}
	if !strings.Contains(err.Error(), "leaked sensitive auth material") {
		t.Fatalf("expected sensitive auth material error, got %v", err)
	}
}

func TestCaptureCreatedKeyVerifiesLookupKey(t *testing.T) {
	t.Parallel()

	runner := liveRunner{binary: writeFixtureBinary(t, `#!/bin/sh
case "$*" in
  "alert configs") printf '%s\n' '[{"AlertConfigKey":42,"Name":"smoke-alert"}]' ;;
  *) printf '%s\n' '{}' ;;
esac
`), timeout: time.Second}
	state := &fixtureState{}
	testCase := smokeCase{Command: "alert create", FixtureName: "smoke-alert"}

	if err := captureCreatedKey(context.Background(), runner, testCase, []byte(`{"key":42}`), state); err != nil {
		t.Fatalf("captureCreatedKey returned error: %v", err)
	}
	if state.AlertKey != 42 {
		t.Fatalf("expected alert key 42, got %d", state.AlertKey)
	}
	if len(state.Cleanup) != 1 || strings.Join(state.Cleanup[0].Args, " ") != "alert delete --key 42" {
		t.Fatalf("unexpected cleanup entry: %#v", state.Cleanup)
	}
}

func TestCaptureCreatedKeyRejectsMismatchedResponseKey(t *testing.T) {
	t.Parallel()

	runner := liveRunner{binary: writeFixtureBinary(t, `#!/bin/sh
case "$*" in
  "watchlist configs") printf '%s\n' '[{"SearchTemplateKey":7,"Name":"smoke-watchlist"}]' ;;
  *) printf '%s\n' '{}' ;;
esac
`), timeout: time.Second}
	state := &fixtureState{}
	testCase := smokeCase{Command: "watchlist create", FixtureName: "smoke-watchlist"}

	err := captureCreatedKey(context.Background(), runner, testCase, []byte(`{"key":8}`), state)
	if err == nil {
		t.Fatal("expected mismatched key error")
	}
	if !strings.Contains(err.Error(), "did not match smoke-owned key") {
		t.Fatalf("expected ownership mismatch, got %v", err)
	}
	if state.WatchlistKey != 0 {
		t.Fatalf("unexpected captured watchlist key: %d", state.WatchlistKey)
	}
	if len(state.Cleanup) != 1 || state.Cleanup[0].FixtureName != "smoke-watchlist" || len(state.Cleanup[0].Args) != 0 {
		t.Fatalf("expected name-only cleanup entry, got %#v", state.Cleanup)
	}
}

func TestResolveCleanupCaseUsesExactFixtureName(t *testing.T) {
	t.Parallel()

	runner := liveRunner{binary: writeFixtureBinary(t, `#!/bin/sh
case "$*" in
  "watchlist configs") printf '%s\n' '[{"SearchTemplateKey":7,"Name":"smoke-watchlist"}]' ;;
  *) printf '%s\n' '{}' ;;
esac
`), timeout: time.Second}

	resolved, err := resolveCleanupCase(context.Background(), runner, smokeCase{Command: "watchlist delete", FixtureName: "smoke-watchlist"})
	if err != nil {
		t.Fatalf("resolveCleanupCase returned error: %v", err)
	}
	if got := strings.Join(resolved.Args, " "); got != "watchlist delete --key 7" {
		t.Fatalf("unexpected cleanup args: %s", got)
	}
}

func TestFindUniqueConfigKeyRejectsUnsafeMatches(t *testing.T) {
	t.Parallel()

	configs := []alertConfig{
		{Key: 0, Name: "zero"},
		{Key: 1, Name: "duplicate"},
		{Key: 2, Name: "duplicate"},
		{Key: 3, Name: "owned"},
	}
	fields := func(config alertConfig) (int, string) {
		return config.Key, config.Name
	}

	if key, err := findUniqueConfigKey("owned", configs, fields); err != nil || key != 3 {
		t.Fatalf("expected owned key 3, got key=%d err=%v", key, err)
	}
	for _, name := range []string{"missing", "zero", "duplicate"} {
		if _, err := findUniqueConfigKey(name, configs, fields); err == nil {
			t.Fatalf("expected error for unsafe match %q", name)
		}
	}
}

func TestShouldSkipMutationHonorsReadOnlyMode(t *testing.T) {
	t.Parallel()

	readOnlyOpts := &options{Mode: modeReadOnly}
	allOpts := &options{Mode: modeAll}
	if !shouldSkipMutation(readOnlyOpts, smokeCase{Mutation: true}) {
		t.Fatal("expected read-only mode to skip mutations")
	}
	if shouldSkipMutation(readOnlyOpts, smokeCase{Mutation: false}) {
		t.Fatal("did not expect read-only mode to skip read-only commands")
	}
	if shouldSkipMutation(allOpts, smokeCase{Mutation: true}) {
		t.Fatal("did not expect all mode to skip mutations")
	}
}

func writeFixtureBinary(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "volumeleaders-agent-fixture")
	if err := os.WriteFile(path, []byte(contents), 0o700); err != nil {
		t.Fatalf("write fixture binary: %v", err)
	}
	return path
}
