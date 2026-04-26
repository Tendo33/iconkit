# Contributing to iconkit

Thank you for your interest in contributing to iconkit. This guide covers everything you need to get started.

---

## Development Environment

### Prerequisites

- Go 1.22 or later (`go version`)
- Git

### Setup

```bash
git clone https://github.com/Tendo33/iconkit.git
cd iconkit
go mod tidy
go build -o iconkit .
```

To confirm everything works:

```bash
go test ./...
```

---

## Running Tests

```bash
# All packages
go test ./...

# With verbose output
go test ./... -v

# A specific package
go test ./internal/processor/...

# With race detector (recommended before submitting a PR)
go test -race ./...
```

All tests must pass before opening a pull request.

---

## Code Style

- Run `gofmt -w .` before committing. The CI pipeline enforces this.
- Run `go vet ./...` to catch common mistakes.
- Keep functions focused and small. Prefer named return values only when they genuinely clarify intent.
- Do not add comments that just restate what the code does. Comments should explain *why*, not *what*.

---

## Commit Message Convention

iconkit uses [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>(<scope>): <short description>

[optional body]
```

Common types:

| Type | When to use |
|------|-------------|
| `feat` | New feature or flag |
| `fix` | Bug fix |
| `refactor` | Code change that is neither a fix nor a feature |
| `test` | Adding or updating tests |
| `docs` | Documentation changes only |
| `ci` | CI/CD configuration |
| `chore` | Build tooling, dependency updates |

Examples:

```
feat(runner): add --dry-run mode
fix(round): correct corner alpha for small radii
docs: add maskable icon example to README
```

---

## How to Add a New Preset

Presets live in `internal/preset/preset.go`. Adding one takes two steps.

**1. Add an entry to the Registry:**

```go
"mypreset": {
    Sizes:       []int{16, 32, 64, 128},
    Description: "Short description of what this preset is for",
},
```

**2. Update the README tables:**

- Add a row to the Presets table in `README.md`
- Add the same row to `README.zh-CN.md`
- Add the preset name to the `--preset` flag description in `cmd/root.go`

**3. Add a test:**

In `internal/runner/runner_test.go`, extend `TestRun_NewPresets` (or add a dedicated `TestRun_<Name>Preset` function) to verify the preset produces the expected number of icons.

---

## How to Add a New Image Processor

Processors live in `internal/processor/`. Each processor is a pure function that takes an `image.Image` and returns an `image.Image`.

**1. Create the file:**

```go
// internal/processor/myeffect.go
package processor

import "image"

// MyEffect applies ... to the image.
func MyEffect(img image.Image, param int) image.Image {
    // ...
}
```

**2. Wire it into the pipeline:**

Processors are called in `runner.go` inside `processOne()`. Add your call in the appropriate order:

```
Pad → Resize → RoundCorners → FillBackground → [your effect]
```

**3. Add a flag:**

- Add a variable in `cmd/root.go` (e.g. `var myParam int`)
- Register it in `init()` with `rootCmd.Flags().IntVar(...)`
- Pass it through `buildOptions()` into the relevant `runner.Options` field
- Add the field to `runner.Options` in `internal/runner/runner.go`

**4. Write tests:**

Add unit tests in `internal/processor/processor_test.go` (or a dedicated `myeffect_test.go`) and integration tests in `internal/runner/runner_test.go`.

---

## Pull Request Guidelines

1. Fork the repository and create a branch: `git checkout -b feat/my-feature`
2. Make your changes with tests
3. Run `go test -race ./...` and confirm all tests pass
4. Run `gofmt -w .` and `go vet ./...`
5. Open a PR against `main` with a clear description of what was changed and why
6. Reference any related issues with `Fixes #123` in the PR description

---

## CI

The CI pipeline (`.github/workflows/ci.yml`) runs on every push and pull request:

- `go vet ./...`
- `go test -race ./...`
- Build check for Linux, macOS, Windows (amd64 and arm64)

PRs that fail CI will not be merged.

---

## Release Process

Releases are automated with [GoReleaser](https://goreleaser.com/) via GitHub Actions. To cut a release:

```bash
git tag v1.2.3
git push origin v1.2.3
```

This triggers the release workflow which:

1. Builds binaries for all supported platforms
2. Creates a GitHub Release with the changelog
3. Publishes the Homebrew cask to `Tendo33/homebrew-tap`
4. Publishes Linux packages (`.deb`, `.rpm`, `.apk`)
5. Updates the Scoop bucket at `Tendo33/scoop-bucket`

Before pushing a tag, ensure the `GORELEASER_GITHUB_TOKEN` repository secret is set to a PAT with write access to `Tendo33/iconkit` and `Tendo33/homebrew-tap`.
