# Changelog

All notable changes to iconkit are documented here.

---

## [0.1.0] — 2026-04-26

### Added
- **Anti-aliased round corners** — `--radius` and `--radius-percent` now use 4×4 supersampling on edge pixels for smooth arcs instead of hard-edged cutouts
- **`--radius-percent`** — specify corner radius as a percentage of the minimum dimension (0–50), works consistently across all output sizes
- **`--resize-mode`** — `stretch` (default), `fit` (letterbox with background color), `cover` (crop center); fixes distortion for non-square inputs
- **`--output-name` template** — customise output filenames with `{name}`, `{size}`, `{width}`, `{height}`, `{ext}` placeholders
- **`--html`** — generate `icons.html` with `<link>` tags for all output icons
- **`--manifest`** — generate `manifest.json` (Web App Manifest) for all output icons
- **`--maskable`** — automatically apply 18 % padding for Android adaptive/maskable icons
- **`--dry-run`** — preview what would be generated without writing any files
- **`--quiet`** — suppress per-file log lines, only print final summary
- **`--verbose`** — print source→target dimensions and per-file processing time
- **`--continue-on-error`** — batch mode continues after per-file failures and reports a summary at the end
- **`-j / --concurrency`** — parallel worker pool for batch processing (default: `NumCPU`)
- **Batch progress display** — shows `Processing [N/M]...` for batches larger than 5 files
- **`--format webp`** — output images as WebP (requires CGO-enabled build; stubs present for CGO-free binaries)
- **SVG input** — `.svg` files can now be used as input (rasterized to 512 × 512 then resized); powered by `oksvg` + `rasterx`
- **WebP input** — `.webp` files can now be used as input; powered by `golang.org/x/image/webp`
- **New presets**: `macos`, `windows`, `electron`, `tauri`
- **`iconkit.schema.json`** — JSON Schema for `iconkit.json` with VS Code / JetBrains auto-complete support
- **Linux packages** — `.deb`, `.rpm`, `.apk` via GoReleaser `nfpms`
- **Windows Scoop manifest** — auto-published to `Tendo33/scoop-bucket` on release
- **macOS ad-hoc code signing** — reduces Gatekeeper friction for unsigned binaries

### Fixed
- Indentation inconsistency in `background.go` (mixed tab/space inside pixel loop)
- `-s` ignored warning now goes to stderr instead of stdout

---

## [0.0.3] — 2026-04-18

### Added
- Cross-platform `install.sh` with Windows Git Bash support
- `~/bin` fallback on Windows, `~/.local/bin` fallback on macOS/Linux
- `hash -r` / `rehash` hint printed after install when PATH may need updating

### Fixed
- Installer now uses POSIX-compatible shell constructs only (`#!/bin/sh`)

---

## [0.0.2] — 2026-04-17

### Added
- Advanced icon output options: `--pad`, `--bg`, `--ico`
- Batch processing mode (directory input)
- JSON config file support (`iconkit.json`)
- Processing-only mode: `-r`, `--pad`, or `--bg` without `-s`/`-p` keeps original dimensions
- Collision-safe batch output naming (`{name}-{ext}.png`)

---

## [0.0.1] — 2026-04-14

### Added
- Initial release
- Resize to multiple sizes from a single PNG/JPG input
- Round corners (`--radius`)
- Presets: `web`, `ios`, `android`, `chrome-ext`, `firefox-ext`, `pwa`
- Custom output directory (`--out`)
- Force overwrite (`--force`)
- GoReleaser pipeline with Homebrew cask publication
