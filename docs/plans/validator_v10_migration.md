Validator v10 migration plan

Goal
- Move from github.com/go-playground/validator v9 to v10 (module path v10) to use maintained APIs and better error types.

Steps
1. Update go.mod: replace require github.com/go-playground/validator v9... with github.com/go-playground/validator/v10 latest.
2. Run `go mod tidy` and run unit tests.
3. Replace imports in code from `github.com/go-playground/validator` to `github.com/go-playground/validator/v10` where direct package use exists.
4. Adjust any usage of validator types that changed between v9 and v10 (error types, tag options).
5. Run `golangci-lint` and fix lint failures.
6. Add a short entry to `docs/CONTRIBUTING_TESTING.md` describing the validator version and how to run validator-specific tests.

Rollback
- If migration causes widespread failures, revert go.mod and imports and open a follow-up ticket to migrate incrementally per package.

