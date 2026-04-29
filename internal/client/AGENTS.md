# internal/client

HTTP client code for VolumeLeaders JSON and DataTables requests.

## Review guidelines

- Treat missing response body closure, missing context propagation, or unsafe retry behavior as P1 when it can leak resources, hang commands, or ignore cancellation.
- Verify HTTP status errors include the operation or endpoint context needed to diagnose failures.
- Verify request construction preserves authentication headers and XSRF behavior without logging secret values.
- Check DataTables request encoding carefully because small parameter changes can silently alter backend filtering, sorting, or pagination.
- Do not request style-only changes that `gofmt`, `go vet`, or `golangci-lint` already enforce.

## Maintenance notes

- Update these guidelines whenever request encoding, authentication headers, context propagation, status handling, retry behavior, or DataTables protocol assumptions change.
