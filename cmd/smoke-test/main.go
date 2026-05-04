package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	defaultBinary    = "./volumeleaders-agent"
	defaultSmokeDate = "2026-04-28"
	modeAll          = "all"
	modeReadOnly     = "read-only"
)

var sensitiveTokens = []string{
	"asp.net_sessionid",
	"sessionid",
	"xsrf-token",
	"x-xsrf-token",
	"cookie:",
	"set-cookie:",
}

// options contains command-line settings for the local live smoke harness.
type options struct {
	Binary  string
	Date    string
	Mode    string
	Command string
	Timeout time.Duration
	Verbose bool
}

// schemaEntry is the small subset of structcli JSON Schema output needed to
// discover runnable command paths.
type schemaEntry struct {
	Title       string   `json:"title"`
	Subcommands []string `json:"x-structcli-subcommands"`
}

// outputContract is the subset of volumeleaders-agent outputschema output used
// to validate coverage and default stdout format.
type outputContract struct {
	Command       string   `json:"command"`
	Formats       []string `json:"formats"`
	DefaultFormat string   `json:"default_format"`
}

// smokeCase describes a single command invocation that should return successful
// stdout from the live CLI.
type smokeCase struct {
	Command     string
	Args        []string
	Mutation    bool
	Dynamic     bool
	FixtureName string
}

// commandResult captures a completed subprocess invocation.
type commandResult struct {
	Stdout   []byte
	Stderr   []byte
	Duration time.Duration
}

// reportRow stores one command result for the final pass/fail table.
type reportRow struct {
	Command  string
	Status   string
	Duration time.Duration
	Message  string
}

// fixtureState tracks keys created during mutating smoke tests so every live
// mutation can target smoke-owned data and be cleaned up before exit.
type fixtureState struct {
	AlertKey     int
	WatchlistKey int
	Cleanup      []smokeCase
}

// alertConfig contains the live fields needed to find a smoke-owned alert after
// creation. The create endpoint can return key zero, so the configs endpoint is
// the source of truth for the key used by follow-up edits and cleanup.
type alertConfig struct {
	Key  int    `json:"AlertConfigKey"`
	Name string `json:"Name"`
}

// watchlistConfig contains the live fields needed to find a smoke-owned
// watchlist after creation when the create response omits the real key.
type watchlistConfig struct {
	Key  int    `json:"SearchTemplateKey"`
	Name string `json:"Name"`
}

func main() {
	ctx := context.Background()
	opts := parseOptions(os.Args[1:])
	if err := run(ctx, &opts); err != nil {
		fmt.Fprintf(os.Stderr, "smoke failed: %v\n", err)
		os.Exit(1)
	}
}

// parseOptions turns command-line flags into smoke harness options.
func parseOptions(args []string) options {
	flags := flag.NewFlagSet("smoke-test", flag.ExitOnError)
	opts := options{}
	flags.StringVar(&opts.Binary, "binary", defaultBinary, "volumeleaders-agent binary to execute")
	flags.StringVar(&opts.Date, "date", defaultSmokeDate, "YYYY-MM-DD date used by commands that require one")
	flags.StringVar(&opts.Mode, "mode", modeAll, "commands to run: all or read-only")
	flags.StringVar(&opts.Command, "command", "", "run one command path, for example 'trade list'")
	flags.DurationVar(&opts.Timeout, "timeout", 45*time.Second, "timeout for each live command")
	flags.BoolVar(&opts.Verbose, "verbose", false, "print command stdout/stderr details")
	_ = flags.Parse(args)
	return opts
}

// run executes the live smoke suite and prints a compact summary table.
func run(ctx context.Context, opts *options) error {
	if err := validateOptions(opts); err != nil {
		return err
	}

	runner := liveRunner{binary: opts.Binary, timeout: opts.Timeout, verbose: opts.Verbose}
	commands, err := discoverCommands(ctx, runner)
	if err != nil {
		return err
	}
	contracts, err := discoverContracts(ctx, runner)
	if err != nil {
		return err
	}
	if err := validateCoverage(commands, contracts, buildStaticCases(opts.Date)); err != nil {
		return err
	}

	rows, err := runSmokeCases(ctx, runner, opts, commands)
	printReport(rows)
	return err
}

// validateOptions rejects unsupported modes before any live commands run.
func validateOptions(opts *options) error {
	switch opts.Mode {
	case modeAll, modeReadOnly:
		return nil
	default:
		return fmt.Errorf("unsupported --mode %q; use %q or %q", opts.Mode, modeAll, modeReadOnly)
	}
}

// liveRunner shells out to the built volumeleaders-agent binary.
type liveRunner struct {
	binary  string
	timeout time.Duration
	verbose bool
}

// run executes one CLI command and returns captured stdout and stderr.
func (runner liveRunner) run(ctx context.Context, args ...string) (commandResult, error) {
	commandCtx, cancel := context.WithTimeout(ctx, runner.timeout)
	defer cancel()

	started := time.Now()
	// #nosec G204 - this local smoke harness intentionally executes the configured
	// volumeleaders-agent binary with explicit fixture arguments.
	cmd := exec.CommandContext(commandCtx, runner.binary, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result := commandResult{Stdout: stdout.Bytes(), Stderr: stderr.Bytes(), Duration: time.Since(started)}
	if commandCtx.Err() != nil {
		return result, fmt.Errorf("%s timed out after %s", strings.Join(args, " "), runner.timeout)
	}
	if leaksSensitiveData(result.Stdout) || leaksSensitiveData(result.Stderr) {
		return result, fmt.Errorf("%s leaked sensitive auth material", strings.Join(args, " "))
	}
	if runner.verbose {
		printVerbose(args, result)
	}
	if err != nil {
		return result, fmt.Errorf("%s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(result.Stderr)))
	}
	return result, nil
}

// printVerbose writes raw command output when requested for debugging failures.
func printVerbose(args []string, result commandResult) {
	fmt.Fprintf(os.Stderr, "\n$ volumeleaders-agent %s\n", strings.Join(args, " "))
	fmt.Fprintf(os.Stderr, "duration: %s\n", result.Duration.Round(time.Millisecond))
	if len(result.Stdout) > 0 {
		fmt.Fprintf(os.Stderr, "stdout: %s\n", strings.TrimSpace(string(result.Stdout)))
	}
	if len(result.Stderr) > 0 {
		fmt.Fprintf(os.Stderr, "stderr: %s\n", strings.TrimSpace(string(result.Stderr)))
	}
}

// discoverCommands asks the real CLI for its command tree and returns runnable
// domain commands, excluding parent command groups.
func discoverCommands(ctx context.Context, runner liveRunner) (map[string]struct{}, error) {
	result, err := runner.run(ctx, "--jsonschema=tree")
	if err != nil {
		return nil, fmt.Errorf("discover command tree: %w", err)
	}

	var entries []schemaEntry
	if err := json.Unmarshal(result.Stdout, &entries); err != nil {
		return nil, fmt.Errorf("decode command tree: %w", err)
	}

	commands := make(map[string]struct{})
	for _, entry := range entries {
		command, ok := strings.CutPrefix(entry.Title, "volumeleaders-agent ")
		if !ok || command == "" || len(entry.Subcommands) > 0 || isBuiltinCommand(command) {
			continue
		}
		commands[command] = struct{}{}
	}
	return commands, nil
}

// isBuiltinCommand excludes Cobra/structcli helper commands that are not live
// VolumeLeaders operations and do not follow the default JSON stdout contract.
func isBuiltinCommand(command string) bool {
	return command == "help" || strings.HasPrefix(command, "completion ")
}

// discoverContracts loads machine-readable stdout contracts from the real CLI.
func discoverContracts(ctx context.Context, runner liveRunner) (map[string]outputContract, error) {
	result, err := runner.run(ctx, "outputschema")
	if err != nil {
		return nil, fmt.Errorf("discover output contracts: %w", err)
	}

	var contracts []outputContract
	if err := json.Unmarshal(result.Stdout, &contracts); err != nil {
		return nil, fmt.Errorf("decode output contracts: %w", err)
	}

	byCommand := make(map[string]outputContract, len(contracts))
	for _, contract := range contracts {
		byCommand[contract.Command] = contract
	}
	return byCommand, nil
}

// validateCoverage ensures new CLI commands fail smoke setup until they have an
// explicit fixture and stdout contract.
func validateCoverage(commands map[string]struct{}, contracts map[string]outputContract, cases map[string]smokeCase) error {
	var missingFixtures []string
	var missingContracts []string
	for command := range commands {
		if _, ok := cases[command]; !ok {
			missingFixtures = append(missingFixtures, command)
		}
		if command != "outputschema" {
			if _, ok := contracts[command]; !ok {
				missingContracts = append(missingContracts, command)
			}
		}
	}
	if len(missingFixtures) > 0 || len(missingContracts) > 0 {
		sort.Strings(missingFixtures)
		sort.Strings(missingContracts)
		return fmt.Errorf("smoke coverage incomplete: missing fixtures=%v missing output contracts=%v", missingFixtures, missingContracts)
	}
	return nil
}

// buildStaticCases returns safe fixture arguments for commands that do not need
// keys created during this smoke run.
func buildStaticCases(date string) map[string]smokeCase {
	return map[string]smokeCase{
		"alert configs":           readOnly("alert configs"),
		"alert create":            mutation("alert create"),
		"alert delete":            mutation("alert delete"),
		"alert edit":              mutation("alert edit"),
		"market earnings":         readOnly("market earnings", "--days", "5"),
		"market exhaustion":       readOnly("market exhaustion"),
		"market snapshots":        readOnly("market snapshots"),
		"outputschema":            readOnly("outputschema"),
		"trade alerts":            readOnly("trade alerts", "--date", date, "--length", "10"),
		"trade cluster-alerts":    readOnly("trade cluster-alerts", "--date", date, "--length", "10"),
		"trade cluster-bombs":     readOnly("trade cluster-bombs", "AAPL", "--days", "1"),
		"trade clusters":          readOnly("trade clusters", "AAPL", "--days", "1"),
		"trade level-touches":     readOnly("trade level-touches", "AAPL", "--days", "14", "--length", "10"),
		"trade levels":            readOnly("trade levels", "AAPL", "--trade-level-count", "5"),
		"trade list":              readOnly("trade list", "AAPL", "--days", "1"),
		"trade preset-tickers":    readOnly("trade preset-tickers", "--preset", "All Trades"),
		"trade presets":           readOnly("trade presets"),
		"trade sentiment":         readOnly("trade sentiment", "--days", "5"),
		"volume ah-institutional": readOnly("volume ah-institutional", "--date", date, "--length", "10"),
		"volume institutional":    readOnly("volume institutional", "--date", date, "--length", "10"),
		"volume total":            readOnly("volume total", "--date", date, "--length", "10"),
		"watchlist add-ticker":    mutation("watchlist add-ticker"),
		"watchlist configs":       readOnly("watchlist configs"),
		"watchlist create":        mutation("watchlist create"),
		"watchlist delete":        mutation("watchlist delete"),
		"watchlist edit":          mutation("watchlist edit"),
		"watchlist tickers":       dynamicReadOnly("watchlist tickers"),
	}
}

// readOnly constructs a non-mutating smoke case.
func readOnly(command string, args ...string) smokeCase {
	return smokeCase{Command: command, Args: commandArgs(command, args...), Mutation: false}
}

// dynamicReadOnly marks a read-only command whose args are filled after fixture
// creation, such as querying the smoke-owned watchlist key.
func dynamicReadOnly(command string) smokeCase {
	return smokeCase{Command: command, Mutation: false, Dynamic: true}
}

// mutation constructs a command case whose args are built during fixture setup.
func mutation(command string) smokeCase {
	return smokeCase{Command: command, Mutation: true, Dynamic: true}
}

// commandArgs joins a command path and its flags into subprocess args.
func commandArgs(command string, args ...string) []string {
	parts := strings.Fields(command)
	return append(parts, args...)
}

// runSmokeCases executes requested cases and always attempts cleanup for created
// live fixtures before returning.
func runSmokeCases(ctx context.Context, runner liveRunner, opts *options, commands map[string]struct{}) ([]reportRow, error) {
	cases := buildStaticCases(opts.Date)
	selected, err := selectCases(opts, commands, cases)
	if err != nil {
		return nil, err
	}

	state := &fixtureState{}
	var rows []reportRow
	var failures []error
	for _, smokeCase := range selected {
		if shouldSkipMutation(opts, smokeCase) {
			rows = append(rows, reportRow{Command: smokeCase.Command, Status: "skip", Message: "mutating command skipped in read-only mode"})
			continue
		}

		resolved, err := resolveCase(smokeCase, state)
		if err != nil {
			rows = append(rows, reportRow{Command: smokeCase.Command, Status: "fail", Message: err.Error()})
			failures = append(failures, err)
			continue
		}

		row, err := runCase(ctx, runner, resolved, state)
		rows = append(rows, row)
		if err != nil {
			failures = append(failures, err)
		}
	}

	cleanupRows, cleanupErr := cleanupFixtures(ctx, runner, state)
	rows = append(rows, cleanupRows...)
	if cleanupErr != nil {
		failures = append(failures, cleanupErr)
	}
	return rows, errors.Join(failures...)
}

// selectCases returns cases in stable order, optionally narrowed to one command.
func selectCases(opts *options, commands map[string]struct{}, cases map[string]smokeCase) ([]smokeCase, error) {
	if opts.Command != "" {
		if _, ok := commands[opts.Command]; !ok {
			return nil, fmt.Errorf("unknown smoke command %q", opts.Command)
		}
		return []smokeCase{cases[opts.Command]}, nil
	}

	ordered := orderedCommands(commands)
	selected := make([]smokeCase, 0, len(ordered))
	for _, name := range ordered {
		selected = append(selected, cases[name])
	}
	return selected, nil
}

// orderedCommands returns a stable smoke sequence that creates live fixtures
// before any command needs their keys and deletes them at the end.
func orderedCommands(commands map[string]struct{}) []string {
	preferred := []string{
		"outputschema",
		"trade presets",
		"trade preset-tickers",
		"market snapshots",
		"market exhaustion",
		"market earnings",
		"trade list",
		"trade sentiment",
		"trade clusters",
		"trade cluster-bombs",
		"trade alerts",
		"trade cluster-alerts",
		"trade levels",
		"trade level-touches",
		"volume institutional",
		"volume ah-institutional",
		"volume total",
		"alert configs",
		"alert create",
		"alert edit",
		"alert delete",
		"watchlist configs",
		"watchlist create",
		"watchlist tickers",
		"watchlist add-ticker",
		"watchlist edit",
		"watchlist delete",
	}

	seen := make(map[string]struct{}, len(preferred))
	ordered := make([]string, 0, len(commands))
	for _, command := range preferred {
		if _, ok := commands[command]; ok {
			ordered = append(ordered, command)
			seen[command] = struct{}{}
		}
	}

	var unknown []string
	for command := range commands {
		if _, ok := seen[command]; !ok {
			unknown = append(unknown, command)
		}
	}
	sort.Strings(unknown)
	return append(ordered, unknown...)
}

// shouldSkipMutation reports whether a mutating command is disabled by mode.
func shouldSkipMutation(opts *options, smokeCase smokeCase) bool {
	return opts.Mode == modeReadOnly && smokeCase.Mutation
}

// resolveCase fills dynamic fixture arguments once the smoke-owned keys exist.
func resolveCase(smokeCase smokeCase, state *fixtureState) (smokeCase, error) {
	if !smokeCase.Dynamic {
		return smokeCase, nil
	}

	switch smokeCase.Command {
	case "alert create":
		smokeCase.FixtureName = smokeName("alert")
		smokeCase.Args = commandArgs(smokeCase.Command, "--name", smokeCase.FixtureName, "--tickers", "AAPL", "--trade-rank-lte", "10")
	case "alert edit":
		if state.AlertKey == 0 {
			return smokeCase, fmt.Errorf("alert edit requires smoke-created alert key")
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--key", fmt.Sprint(state.AlertKey), "--name", smokeName("alert-edited"))
	case "alert delete":
		if state.AlertKey == 0 {
			return smokeCase, fmt.Errorf("alert delete requires smoke-created alert key")
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--key", fmt.Sprint(state.AlertKey))
	case "watchlist add-ticker":
		if state.WatchlistKey == 0 {
			return smokeCase, fmt.Errorf("watchlist add-ticker requires smoke-created watchlist key")
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--watchlist-key", fmt.Sprint(state.WatchlistKey), "--ticker", "MSFT")
	case "watchlist create":
		smokeCase.FixtureName = smokeName("watchlist")
		smokeCase.Args = commandArgs(smokeCase.Command, "--name", smokeCase.FixtureName, "--tickers", "AAPL")
	case "watchlist delete":
		if state.WatchlistKey == 0 {
			return smokeCase, fmt.Errorf("watchlist delete requires smoke-created watchlist key")
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--key", fmt.Sprint(state.WatchlistKey))
	case "watchlist edit":
		if state.WatchlistKey == 0 {
			return smokeCase, fmt.Errorf("watchlist edit requires smoke-created watchlist key")
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--key", fmt.Sprint(state.WatchlistKey), "--name", smokeName("watchlist-edited"))
	case "watchlist tickers":
		if state.WatchlistKey == 0 {
			smokeCase.Args = commandArgs(smokeCase.Command)
			return smokeCase, nil
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--watchlist-key", fmt.Sprint(state.WatchlistKey))
	default:
		return smokeCase, fmt.Errorf("no dynamic fixture resolver for %q", smokeCase.Command)
	}
	return smokeCase, nil
}

// smokeName creates short, recognizable names for records that cleanup may need
// to find in VolumeLeaders logs or UI during troubleshooting.
func smokeName(kind string) string {
	return fmt.Sprintf("smoke-%s-%s", kind, time.Now().UTC().Format("20060102-150405"))
}

// runCase executes one resolved smoke case and captures keys from create calls.
func runCase(ctx context.Context, runner liveRunner, smokeCase smokeCase, state *fixtureState) (reportRow, error) {
	result, err := runner.run(ctx, smokeCase.Args...)
	row := reportRow{Command: smokeCase.Command, Duration: result.Duration}
	if err != nil {
		row.Status = "fail"
		row.Message = err.Error()
		return row, err
	}
	if err := validateJSONStdout(smokeCase.Command, result.Stdout); err != nil {
		row.Status = "fail"
		row.Message = err.Error()
		return row, err
	}
	if err := captureCreatedKey(ctx, runner, smokeCase, result.Stdout, state); err != nil {
		row.Status = "fail"
		row.Message = err.Error()
		return row, err
	}
	clearDeletedKey(smokeCase, state)
	row.Status = "pass"
	return row, nil
}

// validateJSONStdout checks the default smoke contract: compact JSON on stdout.
func validateJSONStdout(command string, stdout []byte) error {
	trimmed := bytes.TrimSpace(stdout)
	if len(trimmed) == 0 {
		return fmt.Errorf("%s produced empty stdout", command)
	}
	var decoded any
	decoder := json.NewDecoder(bytes.NewReader(trimmed))
	decoder.UseNumber()
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("%s produced invalid JSON: %w", command, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s produced multiple JSON values", command)
	}
	return nil
}

// captureCreatedKey stores fixture keys from create commands. It verifies the
// response against an exact-name lookup so later mutations target only the record
// this smoke run created. Some VolumeLeaders mutation responses report key zero,
// so the exact-name lookup is also the fallback source of truth for those cases.
func captureCreatedKey(ctx context.Context, runner liveRunner, testCase smokeCase, stdout []byte, state *fixtureState) error {
	if testCase.Command != "alert create" && testCase.Command != "watchlist create" {
		return nil
	}
	if testCase.FixtureName == "" {
		return fmt.Errorf("%s had no fixture name for ownership verification", testCase.Command)
	}

	responseKey, responseErr := decodeMutationKey(stdout)
	lookupKey, lookupErr := lookupCreatedKey(ctx, runner, testCase.Command, testCase.FixtureName)
	if lookupErr != nil {
		queueNameCleanup(testCase, state)
		return fmt.Errorf("%s could not verify smoke-owned key by name: %w", testCase.Command, lookupErr)
	}
	if responseErr == nil && responseKey != lookupKey {
		queueNameCleanup(testCase, state)
		return fmt.Errorf("%s response key %d did not match smoke-owned key %d", testCase.Command, responseKey, lookupKey)
	}
	key := lookupKey
	switch testCase.Command {
	case "alert create":
		state.AlertKey = key
		state.Cleanup = append(state.Cleanup, smokeCase{Command: "alert delete", Args: commandArgs("alert delete", "--key", fmt.Sprint(key)), Mutation: true})
	case "watchlist create":
		state.WatchlistKey = key
		state.Cleanup = append(state.Cleanup, smokeCase{Command: "watchlist delete", Args: commandArgs("watchlist delete", "--key", fmt.Sprint(key)), Mutation: true})
	}
	return nil
}

// queueNameCleanup preserves enough ownership data to attempt cleanup even when
// key discovery fails after a successful create response.
func queueNameCleanup(testCase smokeCase, state *fixtureState) {
	switch testCase.Command {
	case "alert create":
		state.Cleanup = append(state.Cleanup, smokeCase{Command: "alert delete", Mutation: true, FixtureName: testCase.FixtureName})
	case "watchlist create":
		state.Cleanup = append(state.Cleanup, smokeCase{Command: "watchlist delete", Mutation: true, FixtureName: testCase.FixtureName})
	}
}

// lookupCreatedKey finds the exact smoke-owned record created by this run.
func lookupCreatedKey(ctx context.Context, runner liveRunner, command, fixtureName string) (int, error) {
	switch command {
	case "alert create":
		return lookupAlertKey(ctx, runner, fixtureName)
	case "watchlist create":
		return lookupWatchlistKey(ctx, runner, fixtureName)
	default:
		return 0, fmt.Errorf("no key lookup configured for %q", command)
	}
}

// lookupAlertKey loads alert configs and returns the key for one exact name.
func lookupAlertKey(ctx context.Context, runner liveRunner, fixtureName string) (int, error) {
	result, err := runner.run(ctx, "alert", "configs")
	if err != nil {
		return 0, fmt.Errorf("load alert configs: %w", err)
	}
	if err := validateJSONStdout("alert configs lookup", result.Stdout); err != nil {
		return 0, err
	}

	var configs []alertConfig
	if err := json.Unmarshal(result.Stdout, &configs); err != nil {
		return 0, fmt.Errorf("decode alert configs: %w", err)
	}
	return findUniqueConfigKey(fixtureName, configs, func(config alertConfig) (int, string) {
		return config.Key, config.Name
	})
}

// lookupWatchlistKey loads watchlist configs and returns the key for one exact
// name.
func lookupWatchlistKey(ctx context.Context, runner liveRunner, fixtureName string) (int, error) {
	result, err := runner.run(ctx, "watchlist", "configs")
	if err != nil {
		return 0, fmt.Errorf("load watchlist configs: %w", err)
	}
	if err := validateJSONStdout("watchlist configs lookup", result.Stdout); err != nil {
		return 0, err
	}

	var configs []watchlistConfig
	if err := json.Unmarshal(result.Stdout, &configs); err != nil {
		return 0, fmt.Errorf("decode watchlist configs: %w", err)
	}
	return findUniqueConfigKey(fixtureName, configs, func(config watchlistConfig) (int, string) {
		return config.Key, config.Name
	})
}

// findUniqueConfigKey rejects ambiguous or malformed live config matches so the
// smoke harness never mutates an unintended record.
func findUniqueConfigKey[T any](fixtureName string, configs []T, fields func(T) (int, string)) (int, error) {
	var matches []int
	for _, config := range configs {
		key, name := fields(config)
		if name == fixtureName {
			matches = append(matches, key)
		}
	}
	switch len(matches) {
	case 0:
		return 0, fmt.Errorf("no config named %q", fixtureName)
	case 1:
		if matches[0] == 0 {
			return 0, fmt.Errorf("config named %q had key zero", fixtureName)
		}
		return matches[0], nil
	default:
		return 0, fmt.Errorf("multiple configs named %q", fixtureName)
	}
}

// decodeMutationKey extracts the created record key from mutation JSON output.
func decodeMutationKey(stdout []byte) (int, error) {
	var payload struct {
		Key int `json:"key"`
	}
	if err := json.Unmarshal(stdout, &payload); err != nil {
		return 0, err
	}
	if payload.Key == 0 {
		return 0, fmt.Errorf("response key was zero")
	}
	return payload.Key, nil
}

// clearDeletedKey prevents cleanup from deleting the same fixture twice when the
// delete command itself is part of the selected smoke suite.
func clearDeletedKey(smokeCase smokeCase, state *fixtureState) {
	switch smokeCase.Command {
	case "alert delete":
		state.AlertKey = 0
		state.Cleanup = removeCleanup(state.Cleanup, "alert delete")
	case "watchlist delete":
		state.WatchlistKey = 0
		state.Cleanup = removeCleanup(state.Cleanup, "watchlist delete")
	}
}

// removeCleanup removes queued cleanup entries by command path.
func removeCleanup(cleanup []smokeCase, command string) []smokeCase {
	filtered := cleanup[:0]
	for _, smokeCase := range cleanup {
		if smokeCase.Command != command {
			filtered = append(filtered, smokeCase)
		}
	}
	return filtered
}

// cleanupFixtures deletes remaining smoke-owned records in reverse order.
func cleanupFixtures(ctx context.Context, runner liveRunner, state *fixtureState) ([]reportRow, error) {
	var rows []reportRow
	var failures []error
	for index := len(state.Cleanup) - 1; index >= 0; index-- {
		smokeCase := state.Cleanup[index]
		resolved, err := resolveCleanupCase(ctx, runner, smokeCase)
		if err != nil {
			row := reportRow{Command: smokeCase.Command + " cleanup", Status: "fail", Message: err.Error()}
			rows = append(rows, row)
			failures = append(failures, err)
			continue
		}
		result, err := runner.run(ctx, resolved.Args...)
		row := reportRow{Command: smokeCase.Command + " cleanup", Duration: result.Duration}
		if err != nil {
			row.Status = "fail"
			row.Message = err.Error()
			failures = append(failures, err)
		} else if err := validateJSONStdout(row.Command, result.Stdout); err != nil {
			row.Status = "fail"
			row.Message = err.Error()
			failures = append(failures, err)
		} else {
			row.Status = "pass"
		}
		rows = append(rows, row)
	}
	return rows, errors.Join(failures...)
}

// resolveCleanupCase converts name-only cleanup entries into delete commands by
// exact smoke fixture name. This gives failed create-key capture a second chance
// to clean up records without ever touching non-smoke data.
func resolveCleanupCase(ctx context.Context, runner liveRunner, smokeCase smokeCase) (smokeCase, error) {
	if len(smokeCase.Args) > 0 {
		return smokeCase, nil
	}
	if smokeCase.FixtureName == "" {
		return smokeCase, fmt.Errorf("%s cleanup missing args and fixture name", smokeCase.Command)
	}

	switch smokeCase.Command {
	case "alert delete":
		key, err := lookupAlertKey(ctx, runner, smokeCase.FixtureName)
		if err != nil {
			return smokeCase, fmt.Errorf("lookup alert cleanup key: %w", err)
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--key", fmt.Sprint(key))
	case "watchlist delete":
		key, err := lookupWatchlistKey(ctx, runner, smokeCase.FixtureName)
		if err != nil {
			return smokeCase, fmt.Errorf("lookup watchlist cleanup key: %w", err)
		}
		smokeCase.Args = commandArgs(smokeCase.Command, "--key", fmt.Sprint(key))
	default:
		return smokeCase, fmt.Errorf("no name cleanup resolver for %q", smokeCase.Command)
	}
	return smokeCase, nil
}

// leaksSensitiveData performs a conservative check that auth material did not
// appear in stdout or stderr.
func leaksSensitiveData(output []byte) bool {
	lowered := strings.ToLower(string(output))
	for _, token := range sensitiveTokens {
		if strings.Contains(lowered, token) {
			return true
		}
	}
	return false
}

// printReport writes a stable summary table for humans and automation logs.
func printReport(rows []reportRow) {
	fmt.Println("Command                         Status  Duration  Message")
	fmt.Println("------------------------------  ------  --------  -------")
	for _, row := range rows {
		message := row.Message
		if message == "" {
			message = "-"
		}
		fmt.Printf("%-30s  %-6s  %-8s  %s\n", row.Command, row.Status, row.Duration.Round(time.Millisecond), message)
	}
}
