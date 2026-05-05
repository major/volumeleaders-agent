---
applyTo: ".github/workflows/**"
---

# GitHub Actions review instructions

- Treat unpinned third-party actions, excessive permissions, unsafe `pull_request_target` usage, or secret exposure in logs as P1.
- Workflows should keep least-privilege permissions and must not expose tokens or credentials to untrusted pull request code.
- CI should continue to run linting, tests with race detection, vulnerability checks, and build verification for Go changes.
