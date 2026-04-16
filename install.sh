#!/bin/sh
set -e

# iconkit installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Tendo33/iconkit/main/install.sh | sh

REPO="Tendo33/iconkit"
INSTALL_DIR="/usr/local/bin"

get_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
        *) echo "unsupported" ;;
    esac
}

get_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) echo "unsupported" ;;
    esac
}

OS=$(get_os)
ARCH=$(get_arch)

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
    echo "Error: unsupported platform $(uname -s)/$(uname -m)"
    exit 1
fi

echo "Detecting platform: ${OS}/${ARCH}"

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
    echo "Error: failed to fetch latest release"
    exit 1
fi

VERSION="${LATEST#v}"
echo "Latest version: ${LATEST}"

# Build download URL
if [ "$OS" = "windows" ]; then
    ARCHIVE="iconkit_${VERSION}_${OS}_${ARCH}.zip"
else
    ARCHIVE="iconkit_${VERSION}_${OS}_${ARCH}.tar.gz"
fi
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"

echo "Downloading ${URL}..."

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"

# Extract
cd "$TMPDIR"
if [ "$OS" = "windows" ]; then
    unzip -q "$ARCHIVE"
else
    tar xzf "$ARCHIVE"
fi

# Install
BINARY="iconkit"
if [ "$OS" = "windows" ]; then
    BINARY="iconkit.exe"
fi

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY" "$INSTALL_DIR/"
else
    echo "Installing to ${INSTALL_DIR} (requires sudo)..."
    sudo mv "$BINARY" "$INSTALL_DIR/"
fi

echo ""
echo "iconkit ${LATEST} installed successfully!"
echo "Run 'iconkit --help' to get started."
