# internal/auth

Browser cookie and XSRF token extraction for VolumeLeaders authentication.

## Review guidelines

- Treat any change that can leak browser cookies, XSRF tokens, session values, profile paths, or other credentials as P1.
- Verify authentication failures degrade gracefully when browsers, profiles, cookie stores, or tokens are unavailable.
- Check that errors give enough context to troubleshoot without exposing sensitive values.
- Verify context cancellation is preserved through authentication and token lookup paths.
- Do not request style-only changes that `gofmt`, `go vet`, or `golangci-lint` already enforce.

## Maintenance notes

- Update these guidelines whenever browser support, cookie extraction, token discovery, authentication error handling, or secret-handling assumptions change.
