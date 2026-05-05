---
applyTo: "internal/auth/**/*.go"
---

# Auth review instructions

- Check cookie extraction, browser profile handling, and token lookup paths for credential safety.
- Authentication failures must degrade gracefully and provide useful error messages.
- Never expose browser cookies, XSRF tokens, session values, or API credentials in logs or errors.
