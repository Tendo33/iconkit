# iconkit

A developer-friendly CLI tool for icon processing. Resize and round corners in one command.

## Install

### One-liner (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/Tendo33/iconkit/main/install.sh | sh
```

### Homebrew (macOS / Linux)

```bash
brew install Tendo33/tap/iconkit
```

### Go

```bash
go install github.com/Tendo33/iconkit@latest
```

### Binary download

Download the latest binary from [GitHub Releases](https://github.com/Tendo33/iconkit/releases), extract, and add to your `PATH`.

## Usage

```bash
iconkit <input> [options]
```

### Quick start

```bash
# Default: generates 16, 32, 64, 128 px icons
iconkit icon.png

# Custom sizes
iconkit icon.png -s 16,32,64,128

# With rounded corners
iconkit icon.png -r 20 -s 16,32,64,128

# Web preset (16, 32, 48, 64, 128, 256)
iconkit icon.png -p web

# iOS AppIcon (all required sizes including 1024 for App Store)
iconkit icon.png -p ios

# Android mipmap (mdpi → xxxhdpi + Play Store 512)
iconkit icon.png -p android

# Custom output directory
iconkit icon.png -s 16,32 -o ./dist

# Batch process all images in a directory
iconkit ./assets/ -p web

# Use a JSON config file
iconkit icon.png -c iconkit.json

# Force overwrite existing files
iconkit icon.png -f
```

## Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--sizes` | `-s` | Output sizes, comma-separated | `16,32,64,128` |
| `--radius` | `-r` | Corner radius in pixels | `0` (no rounding) |
| `--preset` | `-p` | Size preset (`web`, `ios`, `android`) | — |
| `--out` | `-o` | Output directory | `./icons` |
| `--force` | `-f` | Overwrite existing files | `false` |
| `--config` | `-c` | Path to config file | auto-detect `iconkit.json` |
| `--version` | `-v` | Print version | — |

## Presets

| Name | Sizes | Use case |
|------|-------|----------|
| `web` | 16, 32, 48, 64, 128, 256 | Favicons & PWA |
| `ios` | 20, 29, 40, 58, 60, 76, 80, 87, 120, 152, 167, 180, 1024 | iOS AppIcon |
| `android` | 36, 48, 72, 96, 144, 192, 512 | Android mipmap |

When a preset is specified with `-p`, the `-s` flag is ignored.

## JSON Config

Create an `iconkit.json` in your project root to set defaults:

```json
{
  "input": "icon.png",
  "sizes": [16, 32, 64, 128],
  "radius": 20,
  "preset": "web",
  "out": "./dist",
  "force": true
}
```

CLI flags always override config file values.

## Batch Processing

Pass a directory as input to process all PNG/JPG files:

```bash
iconkit ./assets/ -s 32,64
```

Output files are named `{original-name}-{size}.png`.

## Output

Single file input produces `icon-{size}.png`:

```
./icons/
├── icon-16.png
├── icon-32.png
├── icon-64.png
└── icon-128.png
```

Directory input preserves original filenames:

```
./icons/
├── logo-16.png
├── logo-32.png
├── avatar-16.png
└── avatar-32.png
```

## Development

```bash
# Run tests
go test ./... -v

# Build
go build -o iconkit .
```

## Release

Releases are automated via [GoReleaser](https://goreleaser.com/) + GitHub Actions.

```bash
git tag v2.0.0
git push origin v2.0.0
```

## License

MIT
