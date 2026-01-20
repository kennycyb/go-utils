# AGENTS.md

## Project Overview

This is a lightweight Go utilities library providing string manipulation functions and asynchronous utilities. The project consists of two main packages:

- `strutil`: String utility functions like `IsEmpty` and `ToSnakeCase`
- `future`: Asynchronous utilities for working with goroutines and futures

The project uses Go 1.21 and has no external dependencies, relying only on the standard library.

## Setup Commands

- Install Go 1.21 or later if not already installed
- Clone the repository: `git clone https://github.com/kennycyb/go-utils.git`
- Navigate to the project directory: `cd go-utils`
- No additional setup required - the project uses Go modules with no external dependencies

## Development Workflow

- Format code: `go fmt ./...`
- Organize imports: `goimports -w .` (if goimports is installed)
- Run tests: `go test ./...`
- Run tests with verbose output: `go test -v ./...`
- Run tests with coverage: `go test -cover ./...`
- Lint code: `go vet ./...`
- Build the project: `go build ./...`

## Testing Instructions

- Run all tests: `go test ./...`
- Run tests for a specific package: `go test ./strutil` or `go test ./future`
- Run tests with verbose output: `go test -v ./...`
- Run tests with coverage: `go test -cover ./...`
- Run tests with race detection: `go test -race ./...`

Test files follow Go conventions with `_test.go` suffix and are located alongside the code they test.

## Code Style Guidelines

- Follow idiomatic Go practices as outlined in Effective Go and Go Code Review Comments
- Use `go fmt` to format code before committing
- Use `go vet` to check for common mistakes
- Package names should be lowercase and singular
- Use MixedCaps for exported identifiers, mixedCaps for unexported
- Write self-documenting code with clear, descriptive names
- Document exported functions, types, and packages
- Handle errors immediately after function calls
- Prefer early returns to reduce nesting
- Keep interfaces small and focused, using -er suffix when possible

## Build and Deployment

- Build all packages: `go build ./...`
- Build a specific package: `go build ./strutil`
- The project produces no binaries - it's a library for import by other Go projects
- No deployment steps required for the library itself

## Pull Request Guidelines

- Ensure all tests pass: `go test ./...`
- Run linting: `go vet ./...`
- Format code: `go fmt ./...`
- Check that code is properly formatted: `gofmt -l .` should return no files
- Add tests for new functionality
- Follow conventional commit messages
- Title format: Brief description of changes

## CI/CD

The project uses GitHub Actions for CI with the following checks:

- Tests run on Ubuntu with Go 1.21
- Runs `go test -v ./...`
- Runs `go vet ./...`
- Checks code formatting with `gofmt`

## Additional Notes

- This is a library project - no main package or executable output
- All functionality is tested with unit tests
- The project follows standard Go project layout
- No database or external services required
- Development can be done on any platform supporting Go 1.21+
