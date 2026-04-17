#!/bin/sh
set -e

# iconkit installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Tendo33/iconkit/main/install.sh | sh

REPO="Tendo33/iconkit"
DEFAULT_INSTALL_DIR="/usr/local/bin"
DEFAULT_USER_INSTALL_DIR=".local/bin"
WINDOWS_USER_INSTALL_DIR="bin"
TMPDIR=""
ORIG_PWD=""
INSTALL_DIR_SOURCE=""
INSTALL_DIR_FALLBACK_REASON=""
RESOLVED_INSTALL_DIR=""

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

cleanup() {
    if [ -n "$ORIG_PWD" ]; then
        cd "$ORIG_PWD" 2>/dev/null || :
    fi
    if [ -n "$TMPDIR" ]; then
        rm -rf "$TMPDIR"
    fi
}

require_cmd() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "Error: required command '$1' is not available"
        return 1
    fi
}

choose_user_bin_dir() {
    os="$1"

    case "$os" in
        windows) printf '%s\n' "$HOME/$WINDOWS_USER_INSTALL_DIR" ;;
        linux|darwin) printf '%s\n' "$HOME/$DEFAULT_USER_INSTALL_DIR" ;;
        *) printf '%s\n' "$HOME/bin" ;;
    esac
}

choose_default_install_dir() {
    os="$1"

    if [ "$os" = "windows" ]; then
        RESOLVED_INSTALL_DIR=$(choose_user_bin_dir "$os")
        return 0
    fi

    if { [ -d "$DEFAULT_INSTALL_DIR" ] && [ -w "$DEFAULT_INSTALL_DIR" ]; } || \
       { [ ! -e "$DEFAULT_INSTALL_DIR" ] && [ -d "$(dirname "$DEFAULT_INSTALL_DIR")" ] && [ -w "$(dirname "$DEFAULT_INSTALL_DIR")" ]; }; then
        RESOLVED_INSTALL_DIR="$DEFAULT_INSTALL_DIR"
        return 0
    fi

    INSTALL_DIR_FALLBACK_REASON="Default install dir '$DEFAULT_INSTALL_DIR' is not writable; falling back to user bin directory."
    RESOLVED_INSTALL_DIR=$(choose_user_bin_dir "$os")
}

resolve_install_dir() {
    os="$1"

    if [ -n "${INSTALL_DIR:-}" ]; then
        INSTALL_DIR_SOURCE="explicit"
        RESOLVED_INSTALL_DIR="$INSTALL_DIR"
        return 0
    fi

    INSTALL_DIR_SOURCE="default"
    choose_default_install_dir "$os"
}

ensure_install_dir() {
    dir="$1"

    if [ -d "$dir" ]; then
        return 0
    fi

    mkdir -p "$dir" 2>/dev/null
}

path_contains_dir() {
    dir="$1"
    case ":$PATH:" in
        *:"$dir":*) return 0 ;;
        *) return 1 ;;
    esac
}

print_path_hint() {
    install_dir="$1"

    if ! path_contains_dir "$install_dir"; then
        echo "Add '${install_dir}' to your PATH if the command is not available in new shells."
    fi
}

install_binary() {
    binary="$1"
    install_dir="$2"
    os="$3"

    if ensure_install_dir "$install_dir" && [ -w "$install_dir" ]; then
        mv "$binary" "$install_dir/"
        return 0
    fi

    if [ "$os" = "windows" ]; then
        echo "Error: cannot write to '${install_dir}' on Windows."
        echo "Choose a user-writable INSTALL_DIR or move '$binary' to '${install_dir}' manually."
        return 1
    fi

    if [ "$INSTALL_DIR_SOURCE" = "default" ] && [ "$install_dir" != "$DEFAULT_INSTALL_DIR" ]; then
        echo "Error: failed to create fallback install dir '${install_dir}'."
        echo "Create it manually or rerun with INSTALL_DIR set to a writable directory."
        return 1
    fi

    if command -v sudo >/dev/null 2>&1; then
        echo "Installing to ${install_dir} with sudo..."
        sudo mkdir -p "$install_dir"
        sudo mv "$binary" "$install_dir/"
        return 0
    fi

    echo "Error: '${install_dir}' is not writable and 'sudo' is unavailable."
    echo "Choose a writable INSTALL_DIR or move '$binary' to '${install_dir}' manually."
    return 1
}

print_success_message() {
    os="$1"
    install_dir="$2"
    binary="$3"

    echo ""
    echo "iconkit ${LATEST} installed successfully!"
    echo "Installed binary: ${install_dir}/${binary}"
    print_path_hint "$install_dir"

    shell_name="${SHELL##*/}"
    if [ "$os" = "windows" ] || [ "$shell_name" = "zsh" ]; then
        echo "If the current shell still says 'command not found', run 'rehash' (zsh) or 'hash -r' (bash), or start a new shell."
    fi

    echo "Run 'iconkit --help' to get started."
}

main() {
    OS=$(get_os)
    ARCH=$(get_arch)
    resolve_install_dir "$OS"
    INSTALL_DIR="$RESOLVED_INSTALL_DIR"

    if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
        echo "Error: unsupported platform $(uname -s)/$(uname -m)"
        exit 1
    fi

    echo "Detecting platform: ${OS}/${ARCH}"
    if [ -n "$INSTALL_DIR_FALLBACK_REASON" ]; then
        echo "$INSTALL_DIR_FALLBACK_REASON"
        echo "Using user install dir: ${INSTALL_DIR}"
    fi

    require_cmd curl || exit 1
    require_cmd grep || exit 1
    require_cmd sed || exit 1
    require_cmd mktemp || exit 1
    if [ "$OS" = "windows" ]; then
        require_cmd unzip || exit 1
    else
        require_cmd tar || exit 1
    fi

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
    ORIG_PWD=$(pwd)
    trap cleanup EXIT

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

    install_binary "$BINARY" "$INSTALL_DIR" "$OS" || exit 1

    print_success_message "$OS" "$INSTALL_DIR" "$BINARY"
}

if [ "${ICONKIT_INSTALL_TEST_MODE:-0}" = "1" ]; then
    return 0 2>/dev/null || exit 0
fi

main "$@"
