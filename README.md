# iconkit

<p align="center">
  <strong>A developer-friendly CLI for generating polished app and web icons.</strong><br />
  Resize images, round corners, add padding, fill backgrounds, and generate <code>favicon.ico</code> in one command.
</p>

<p align="center">
  <a href="./README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="https://github.com/Tendo33/iconkit/actions/workflows/ci.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/Tendo33/iconkit/ci.yml?branch=main&label=CI&logo=githubactions" alt="CI" />
  </a>
  <a href="https://github.com/Tendo33/iconkit/releases">
    <img src="https://img.shields.io/github/v/release/Tendo33/iconkit?display_name=tag&logo=github" alt="Release" />
  </a>
  <img src="https://img.shields.io/badge/Go-1.26.2-00ADD8?logo=go" alt="Go Version" />
  <a href="./LICENSE">
    <img src="https://img.shields.io/github/license/Tendo33/iconkit" alt="License" />
  </a>
</p>

## Highlights

- Generate multiple icon sizes from a single PNG, JPG, WebP, or SVG input
- Anti-aliased round corners with proportional scaling across output sizes
- Add padding for safe zones and maskable icon layouts
- Fill transparent areas with a solid background color
- Export a multi-size `favicon.ico` alongside PNG outputs
- Batch process every image in a directory with parallel workers
- Generate `manifest.json` and HTML `<link>` tags automatically
- Non-square input support: letterbox (`fit`) and center-crop (`cover`) modes

## Table of Contents

- [Install](#install)
- [Quick Start](#quick-start)
- [Usage](#usage)
- [Options](#options)
- [Presets](#presets)
- [favicon.ico](#faviconico)
- [Padding and Background](#padding-and-background)
- [Maskable Icons](#maskable-icons)
- [JSON Config](#json-config)
- [JSON Schema](#json-schema)
- [Batch Processing](#batch-processing)
- [Output](#output)
- [Development](#development)
- [Release](#release)
- [License](#license)

## Install

### One-liner (macOS / Linux / Windows Git Bash)

```bash
curl -fsSL https://raw.githubusercontent.com/Tendo33/iconkit/main/install.sh | sh
```

The installer chooses a sensible default per platform:

- Windows Git Bash installs to `~/bin`
- macOS / Linux try `/usr/local/bin` first and fall back to `~/.local/bin` when the system directory is not writable

If the chosen user bin directory is not already on your `PATH`, add it before opening a new shell. If `iconkit` is still not found in the current shell after installation, run `hash -r` or restart the shell.

### Homebrew (macOS)

```bash
brew install --cask Tendo33/homebrew-tap/iconkit
```

`iconkit` is currently distributed as an unsigned macOS binary. If macOS blocks it after installation, run:

```bash
xattr -dr com.apple.quarantine "$(brew --prefix)/Caskroom/iconkit/latest/iconkit"
```

### Go

```bash
go install github.com/Tendo33/iconkit@latest
```

For Linux and Windows Git Bash, use the one-liner installer, `go install`, or download a binary from Releases.

### Binary download

Download the latest binary from [GitHub Releases](https://github.com/Tendo33/iconkit/releases).

## Quick Start

```bash
# Default output: 16, 32, 64, 128
iconkit icon.png

# Web preset + favicon.ico
iconkit icon.png -p web --ico

# Rounded corners, keeping the original size
iconkit icon.png -r 20

# Rounded corners as a percentage (works across all output sizes)
iconkit icon.png --radius-percent 25 -p ios

# Rounded corners with a custom output directory
iconkit icon.png -r 20 -o ./dist

# Add padding and a white background
iconkit icon.png --pad 0.1 --bg "#ffffff"

# Letterbox a non-square image to 512x512 with a dark background
iconkit icon.png --resize-mode fit --bg "#1a1a2e" -s 512

# Generate icons + HTML link tags + Web App Manifest
iconkit icon.png -p pwa --html --manifest

# Android maskable icon (18% safe-zone padding + manifest)
iconkit icon.png -p android --maskable --manifest

# Preview without writing files
iconkit icon.png -p ios --dry-run

# Parallel batch processing
iconkit ./assets/ -p web -j 4
```

## Usage

```bash
iconkit [input] [options]
```

`<input>` can be a single `.png`, `.jpg`, `.jpeg`, `.webp`, `.svg`, or a directory containing images.

### Examples

```bash
# Default: generates 16, 32, 64, 128 px icons
iconkit icon.png

# Radius only: keeps the original size and outputs a single PNG
iconkit icon.png -r 20

# Radius as percentage: consistent look across all sizes
iconkit icon.png --radius-percent 25 -p ios

# Padding only: keeps the original size and outputs a single PNG
iconkit icon.png --pad 0.1

# Background fill only: keeps the original size and outputs a single PNG
iconkit icon.png --bg "#ffffff"

# Custom sizes with rounded corners
iconkit icon.png -r 20 -s 16,32,64,128

# Web preset (favicon sizes)
iconkit icon.png -p web

# Chrome extension icons
iconkit icon.png -p chrome-ext

# Firefox add-on icons
iconkit icon.png -p firefox-ext

# iOS AppIcon sizes
iconkit icon.png -p ios

# Android mipmap icons
iconkit icon.png -p android

# macOS App Icon sizes
iconkit icon.png -p macos

# Windows shell icons
iconkit icon.png -p windows

# Electron cross-platform icons
iconkit icon.png -p electron

# Tauri v2 icons (use --output-name for correct filenames)
iconkit icon.png -p tauri --output-name "{width}x{height}"

# PWA icons
iconkit icon.png -p pwa

# Generate favicon.ico alongside PNGs
iconkit icon.png -p web --ico

# Add 10% padding with white background
iconkit icon.png --pad 0.1 --bg "#ffffff"

# Fill transparent areas with a custom color
iconkit icon.png --bg "#1a1a2e" -p chrome-ext

# Non-square input: letterbox to 512x512 with dark background
iconkit banner.png --resize-mode fit --bg "#1a1a2e" -s 512

# Non-square input: center-crop to 512x512
iconkit banner.png --resize-mode cover -s 512

# Batch process all images in a directory
iconkit ./assets/ -p web

# Batch process with 4 parallel workers
iconkit ./assets/ -p web -j 4

# Custom output directory, force overwrite
iconkit icon.png -s 16,32 -o ./dist -f

# Use a JSON config file
iconkit icon.png -c iconkit.json

# Generate Web App Manifest and HTML link tags
iconkit icon.png -p pwa --manifest --html

# Android maskable icon with manifest
iconkit icon.png -p android --maskable --manifest

# Preview what would be generated (no files written)
iconkit icon.png -p ios --dry-run

# Quiet mode — only print the final summary
iconkit icon.png -p web --quiet
```

## Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--sizes` | `-s` | Output sizes, comma-separated | auto (`16,32,64,128`; with `-r` / `--pad` / `--bg` only, keep original dimensions) |
| `--radius` | `-r` | Corner radius in pixels | `0` |
| `--radius-percent` | | Corner radius as % of min dimension (0–50); mutually exclusive with `--radius` | `0` |
| `--preset` | `-p` | Size preset from the table below | none |
| `--resize-mode` | | How to resize non-square inputs: `stretch`, `fit` (letterbox), `cover` (crop center) | `stretch` |
| `--out` | `-o` | Output directory | `./icons` |
| `--force` | `-f` | Overwrite existing files | `false` |
| `--config` | `-c` | Path to a JSON config file | auto-detect `iconkit.json` |
| `--pad` | | Padding ratio from `0.0` to `0.5` | `0` |
| `--bg` | | Background color in hex, such as `#ffffff` | transparent |
| `--ico` | | Also generate `favicon.ico` for sizes `<= 256` | `false` |
| `--maskable` | | Apply 18 % padding for Android maskable/adaptive icons | `false` |
| `--output-name` | | Output filename template: `{name}`, `{size}`, `{width}`, `{height}`, `{ext}` | auto |
| `--html` | | Generate `icons.html` with `<link>` tags for all outputs | `false` |
| `--manifest` | | Generate `manifest.json` (Web App Manifest) | `false` |
| `--dry-run` | | Preview output without writing any files | `false` |
| `--quiet` | | Suppress per-file output; only print final summary | `false` |
| `--verbose` | | Print source/target dimensions and per-file timing | `false` |
| `--continue-on-error` | | In batch mode, continue after per-file failures | `false` |
| `--concurrency` | `-j` | Number of parallel workers for batch processing | `NumCPU` |
| `--format` | | Output image format: `png`, `webp` | `png` |
| `--webp-quality` | | WebP output quality (0–100) | `90` |
| `--version` | `-v` | Print version | none |

When `-p` is specified, `-s` is ignored.
When `-r`, `--pad`, or `--bg` is used without `-s` or `-p`, iconkit writes a single PNG that keeps the source dimensions.
When `--ico` is also enabled, iconkit keeps the existing multi-size favicon flow.

## Presets

| Name | Sizes | Use case |
|------|-------|----------|
| `web` | 16, 32, 48, 64, 128, 256 | Favicons and PWA icons |
| `chrome-ext` | 16, 32, 48, 128 | Chrome Extension (Manifest V3) |
| `firefox-ext` | 32, 48, 64, 96, 128 | Firefox Add-on |
| `pwa` | 192, 512 | Progressive Web App |
| `ios` | 20, 29, 40, 58, 60, 76, 80, 87, 120, 152, 167, 180, 1024 | iOS AppIcon |
| `android` | 36, 48, 72, 96, 144, 192, 512 | Android mipmap and Play Store |
| `macos` | 16, 32, 64, 128, 256, 512, 1024 | macOS App Icon (`.icns` sources) |
| `windows` | 16, 24, 32, 48, 64, 128, 256 | Windows Shell + Microsoft Store |
| `electron` | 16, 32, 48, 64, 128, 256, 512, 1024 | Electron cross-platform apps |
| `tauri` | 32, 128, 256 | Tauri v2 apps |

For Tauri, use `--output-name "{width}x{height}"` to match the expected filenames:

```bash
iconkit icon.png -p tauri --output-name "{width}x{height}"
```

## favicon.ico

Use `--ico` to generate a multi-size `.ico` file alongside PNG outputs:

```bash
iconkit icon.png -p web --ico
```

Only sizes `<= 256` are embedded in the generated `.ico` file.

## Padding and Background

Use `--pad` to add breathing room around the original icon:

```bash
iconkit icon.png --pad 0.1 -p ios
```

Use `--bg` to fill transparent areas with a solid color:

```bash
iconkit icon.png --bg "#ffffff" -p android
```

You can combine both:

```bash
iconkit icon.png --pad 0.1 --bg "#1a1a2e" -r 20 -p web --ico
```

## Maskable Icons

Android adaptive icons require the icon content to sit within a 72 % safe zone (18 % padding on each side). Use `--maskable` to apply this automatically:

```bash
iconkit icon.png -p android --maskable --manifest
```

`--maskable` sets padding to 18 % (or keeps your existing `--pad` if it is already larger) and marks icons as `"purpose": "maskable"` in the generated `manifest.json`.

## JSON Config

Create an `iconkit.json` in your project root:

```json
{
  "$schema": "https://raw.githubusercontent.com/Tendo33/iconkit/main/iconkit.schema.json",
  "input": "icon.png",
  "sizes": [16, 32, 64, 128],
  "radius": 20,
  "preset": "web",
  "out": "./dist",
  "force": true,
  "pad": 0.1,
  "bg": "#112233",
  "ico": true
}
```

The JSON config supports all CLI flags. CLI flags always override values from the config file.

If `input` is set in `iconkit.json`, you can run `iconkit` without passing a positional input path.

## JSON Schema

Add `"$schema"` to your `iconkit.json` to get auto-complete and validation in VS Code, JetBrains, and other editors that support JSON Schema:

```json
{
  "$schema": "https://raw.githubusercontent.com/Tendo33/iconkit/main/iconkit.schema.json"
}
```

## Batch Processing

Pass a directory to process all `.png`, `.jpg`, `.jpeg`, `.webp`, and `.svg` files:

```bash
iconkit ./assets/ -s 32,64
```

Output filenames preserve the original base name:

```text
logo-32.png
logo-64.png
badge-32.png
badge-64.png
```

Use `-j` to set the number of parallel workers (default: number of logical CPUs):

```bash
iconkit ./assets/ -p web -j 4
```

Use `--continue-on-error` to process all files even when some fail:

```bash
iconkit ./assets/ -p web --continue-on-error
```

If you run a "processing-only" command in batch mode without `-s` or `-p`, each source image keeps its original dimensions and writes one PNG:

```bash
iconkit ./assets/ -r 20
```

Typical output:

```text
logo.png
badge.png
```

If two source files share the same base name but have different extensions, iconkit appends the source extension to avoid collisions:

```text
logo-png.png
logo-jpg.png
```

When `--ico` is used in batch mode, each image gets its own `.ico` file named after the source file.

## Output

iconkit chooses the output shape using these rules:

1. If `-p` is set, the preset sizes are used.
2. Else if `-s` is set, the explicit sizes are used.
3. Else if one of `-r`, `--pad`, or `--bg` is set, iconkit writes a single PNG that keeps the source dimensions.
4. Else iconkit falls back to the default sizes `16,32,64,128`.
5. If `--ico` is enabled, iconkit keeps the existing multi-size favicon flow.

Single-file input with default sizes uses this structure:

```text
./icons/
|- icon-16.png
|- icon-32.png
|- icon-64.png
`- icon-128.png
```

Single-file input in processing-only mode uses the original base name:

```bash
iconkit icon.jpg -r 20
```

```text
./icons/
`- icon.png
```

Batch input with multi-size output uses `{name}-{size}.png`.

Batch input in processing-only mode uses `{name}.png`, unless a same-name conflict is detected, in which case iconkit uses `{name}-{source-ext}.png`.

With `--ico`, iconkit also writes `favicon.ico` for single-file input, or `{name}.ico` for batch input.

## Development

```bash
go test ./... -v
go build -o iconkit .
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for a full contributor guide.

## Release

```bash
git tag v2.1.0
git push origin v2.1.0
```

Releases are automated with GoReleaser and GitHub Actions. Each release publishes:
- Linux `.deb`, `.rpm`, `.apk` packages
- macOS Homebrew cask (`Tendo33/homebrew-tap`)
- Windows Scoop manifest (`Tendo33/scoop-bucket`)
- GitHub release with checksums

Before pushing a release tag, add a `GORELEASER_GITHUB_TOKEN` repository secret.
Use a GitHub PAT that can write to `Tendo33/iconkit`, `Tendo33/homebrew-tap`, and `Tendo33/scoop-bucket`.

## License

MIT
