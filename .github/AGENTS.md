# .github

GitHub metadata, ownership, and workflows.

## Review guidelines

- Treat unpinned third-party GitHub Actions, excessive workflow permissions, unsafe `pull_request_target` usage, or secret exposure in logs as P1.
- Verify workflows keep least-privilege permissions and do not expose tokens or credentials to untrusted pull request code.
- Check that CI continues to run linting, tests with race detection, vulnerability checks, and build verification for Go changes.
- For release workflow changes, verify signing, provenance, and GoReleaser behavior remain intentional and safe.

## Maintenance notes

- Update these guidelines whenever CI, release, security scanning, permissions, pinned action policy, signing, provenance, or PR execution behavior changes.
