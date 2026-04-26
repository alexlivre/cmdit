# Contributing to cmdit

🎉 Thanks for contributing! Here's how to get started.

## Development Setup

```bash
git clone https://github.com/alexlivre/cmdit.git
cd cmdit
go build -o bin/cmdit.exe ./cmd/cmdit
go test ./...
```

## Branching Strategy

We use a simplified Git Flow:

- `main` — production code, tagged releases only
- `develop` — integration branch for features
- `feature/*` — individual features (e.g. `feature/lsp-client`)
- `fix/*` — bug fixes (e.g. `fix/clipboard-os52`)
- `refactor/*` — refactoring (e.g. `refactor/rope-buffer`)

## Commit Convention

Follow [Conventional Commits](https://www.conventionalcommits.org/):

| Type | Description |
|------|-------------|
| `feat:` | New feature |
| `fix:` | Bug fix |
| `docs:` | Documentation only |
| `refactor:` | Code change that neither fixes a bug nor adds a feature |
| `test:` | Adding or correcting tests |
| `chore:` | Maintenance tasks (deps, CI, configs) |
| `perf:` | Performance improvement |
| `ci:` | CI/CD changes |

Examples:
```bash
git commit -m "feat: add LSP client for Go"
git commit -m "fix: clipboard OSC52 on Windows SSH"
git commit -m "refactor: replace gap buffer with rope"
```

## Pull Request Process

1. Fork the repo and create your branch from `develop`
2. Make your changes and add tests
3. Ensure all tests pass: `go test ./...`
4. Run the linter: `go vet ./...`
5. Open a PR with a clear title and description
6. Link the related issue (e.g. "Closes #12")

## Code Style

- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch issues
- Keep functions small and focused (single responsibility)
- Add comments for non-obvious logic

## Reporting Bugs

Open an issue with:
- Go version (`go version`)
- OS and terminal
- Steps to reproduce
- Expected vs actual behavior
- If possible, a minimal reproduction case

## Questions?

Open a discussion or ping us on the issue tracker.
