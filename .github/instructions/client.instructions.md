---
applyTo: "internal/client/**/*.go"
---

# Client review instructions

- Check HTTP status handling, request context propagation, and response body closure.
- DataTables encoding should preserve expected column and pagination semantics.
- Errors should explain the endpoint or operation that failed without leaking credentials.
