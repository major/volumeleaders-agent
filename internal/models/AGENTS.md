# internal/models

Response models for VolumeLeaders API data.

## Review guidelines

- Verify JSON tags match VolumeLeaders response fields exactly, including capitalization and unusual API names.
- Treat model changes that silently drop data needed by commands, summaries, or CSV/TSV output as P1.
- Check whether new nullable or optional fields need pointer types, zero-value handling, or explicit formatting behavior.
- Verify model changes are reflected in command field selection, summaries, tests, `--jsonschema=tree`, and `outputschema` output where applicable. Update relevant command Long descriptions when user-visible behavior or semantic guidance changes.
- Do not request style-only changes that `gofmt`, `go vet`, or `golangci-lint` already enforce.

## Maintenance notes

- Update these guidelines whenever VolumeLeaders response shapes, JSON tags, nullable-field conventions, compact summaries, CSV/TSV columns, or field selection behavior change.
