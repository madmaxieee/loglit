# Development Guidelines

## Commands

- **Build**: `go build ./...`
- **Run**: `go run main.go`
- **Test (All)**: `go test ./...`
- **Test (Single)**: `go test -v -run TestName ./package/path`
- **Lint**: `go vet ./...`
- **Format**: `go fmt ./...`

## Code Style

- **Imports**: Group standard library imports first, separated from third-party imports.
- **Formatting**: Strictly adhere to `gofmt`.
- **Naming**: Use `PascalCase` for exported symbols and `camelCase` for unexported ones.
- **Error Handling**: Handle errors explicitly; prefer returning errors over panicking. Use `if err != nil { return ..., err }`.
- **Structure**: Follow the standard project layout: `cmd/` for entry points, `internal/` for private app code.
- **Types**: Define structs with explicit types; avoid `interface{}` unless absolutely necessary.
- **Comments**: Document exported functions and types. Explain _why_, not _what_.

## Rules

- Adhere to Go idioms and conventions.
- Ensure all new code is covered by tests where possible.
- Do not commit secrets or sensitive data.
- Do not run any git commands
