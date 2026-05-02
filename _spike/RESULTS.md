# Structcli 0.17.0 validation spike results

All seven validation tests pass in the isolated spike module with the current Go toolchain. A2 is not clean for a Go 1.21 migration target because structcli 0.17.0 declares `go 1.24.0` and needs a newer pflag than its own module currently requires.

- A1 Changed after Bind: PASS. `cmd.Flags().Changed("start-date")` is false when omitted and true when explicitly passed after `structcli.Bind` plus `structcli.ExecuteC`.
- A2 go get compatibility: FAIL for strict Go 1.21 compatibility, PASS with current Go. `github.com/leodido/structcli@v0.17.0` declares `go 1.24.0`. Initial `go get github.com/leodido/structcli@v0.17.0` selected `pflag v1.0.6`, which does not provide `FlagSet.BoolFunc` and failed to compile. The spike passes after resolving `github.com/spf13/pflag` to v1.0.10.
- A3 flag tag override: PASS. An embedded struct field tagged `flag:"start-date"` creates `--start-date` without a struct path prefix.
- A4 pre-set defaults: PASS. Pre-set field values become pflag defaults when no non-empty `default` tag exists. An empty `default:""` tag behaves like no default tag. A non-empty `default:"25"` tag overrides the pre-set field for `DefValue`.
- A5 PersistentPreRunE: PASS. The root `PersistentPreRunE` fires for child commands when the child has bound options and execution goes through `structcli.ExecuteC`.
- A6 SilenceErrors and WithFlagErrors: PASS. Unknown flag errors are not double-printed to stderr with `SilenceErrors: true`, `SilenceUsage: true`, and `structcli.WithFlagErrors()`.
- A7 ExecuteC replacement: PASS. `structcli.ExecuteC(root)` returns the executed subcommand, calls the correct `RunE`, and auto-unmarshals bound option structs before `RunE`.

Variable defaults approach: use pre-set struct fields for runtime defaults where possible. Avoid non-empty `default` tags for values that must vary per command, because the tag wins for displayed `DefValue`.
