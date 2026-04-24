# Security Policy

## Supported Versions

This project is pre-1.0. Only the latest release receives security updates. Older releases are not patched.

## Reporting a Vulnerability

Please report security vulnerabilities through [GitHub Security Advisories](https://github.com/major/volumeleaders-agent/security/advisories/new). Avoid opening public issues for security-related bugs.

## Security Considerations

### Browser Cookie Access

This tool reads browser cookies from local storage to authenticate with volumeleaders.com. Cookies are passed directly to the API and are never logged, stored in files, or transmitted to any third party. The cookie extraction happens locally using the [kooky](https://github.com/browserutils/kooky) library.

### Binary Signing

Official releases sign the checksums file with [cosign](https://github.com/sigstore/cosign) using keyless OIDC via GitHub Actions. To verify a release:

```bash
cosign verify-blob \
  --bundle volumeleaders-agent_*_checksums.txt.sigstore.json \
  --certificate-identity-regexp "github.com/major/volumeleaders-agent" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com" \
  volumeleaders-agent_*_checksums.txt
```

Then verify the binary checksum matches:

```bash
sha256sum --check --ignore-missing volumeleaders-agent_*_checksums.txt
```

### Protecting Session Tokens

Command output may include session tokens or authenticated API responses. Don't share raw output from this tool in public forums, bug reports, or logs without scrubbing any cookie or token values first.

## License

This project is licensed under the [Apache License 2.0](LICENSE).
