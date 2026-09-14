#!/usr/bin/env bash
#
# install.sh - Installer and updater for streamer-vm
#
# Usage:
#   curl -fsSL https://github.com/codigolandia/streamer-vm/releases/latest/download/install.sh | bash
#   curl -fsSL https://github.com/codigolandia/streamer-vm/releases/latest/download/install.sh | bash -s -- --rollback
#   ./install.sh [--rollback|-r] [--version <vX.Y.Z>] [--dir <path>]

set -euo pipefail

REPO="codigolandia/streamer-vm"

# Color support
if [ -t 1 ]; then
    COLOR_INFO="\033[0;34m"
    COLOR_WARN="\033[0;33m"
    COLOR_ERRO="\033[0;31m"
    COLOR_OK="\033[0;32m"
    COLOR_RESET="\033[0m"
else
    COLOR_INFO=""
    COLOR_WARN=""
    COLOR_ERRO=""
    COLOR_OK=""
    COLOR_RESET=""
fi

log_info() {
    printf "${COLOR_INFO}[INFO]${COLOR_RESET} %s\n" "$*"
}

log_warn() {
    printf "${COLOR_WARN}[WARN]${COLOR_RESET} %s\n" "$*" >&2
}

log_erro() {
    printf "${COLOR_ERRO}[ERRO]${COLOR_RESET} %s\n" "$*" >&2
}

log_ok() {
    printf "${COLOR_OK}[OK  ]${COLOR_RESET} %s\n" "$*"
}

print_usage() {
    cat <<EOF
streamer-vm installer

Usage:
  install.sh [options]

Options:
  -r, --rollback        Roll back to the previous backup binary
  -v, --version <tag>   Install a specific version (e.g. v0.4.0 or 0.4.0)
  -d, --dir <path>      Install into a custom directory
  -h, --help            Show this help message

Environment variables:
  VERSION               Target version to install (defaults to latest)
  INSTALL_DIR           Target installation directory (defaults to PATH priority)
EOF
}

resolve_target_dir() {
    if [ -n "${INSTALL_DIR:-}" ]; then
        echo "$INSTALL_DIR"
        return
    fi

    local local_bin="$HOME/.local/bin"
    local home_bin="$HOME/bin"

    local IFS=':'
    for dir in $PATH; do
        dir="${dir%/}"
        if [ "$dir" = "$local_bin" ] || [ "$dir" = "$home_bin" ]; then
            echo "$dir"
            return
        fi
    done

    # Fallback to ~/.local/bin if neither is in PATH
    echo "$local_bin"
}

check_path() {
    local target_dir="$1"
    local in_path=false
    local old_ifs="$IFS"
    IFS=':'
    for d in $PATH; do
        if [ "${d%/}" = "$target_dir" ]; then
            in_path=true
            break
        fi
    done
    IFS="$old_ifs"

    if [ "$in_path" = false ]; then
        log_warn "Target directory '$target_dir' is not in your \$PATH."
        log_warn "Add it to your shell configuration (e.g. ~/.bashrc or ~/.zshrc):"
        log_warn "  export PATH=\"\$PATH:$target_dir\""
    fi
}

detect_platform() {
    local os
    os="$(uname -s)"
    if [ "$os" != "Linux" ]; then
        log_erro "streamer-vm only supports Linux systems (QEMU/KVM). Detected: $os"
        exit 1
    fi

    local arch
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            log_erro "Unsupported CPU architecture: $arch. Only amd64 and arm64 are supported."
            exit 1
            ;;
    esac
    OS="linux"
}

resolve_version() {
    if [ -n "${VERSION:-}" ]; then
        TARGET_TAG="${VERSION}"
        if [[ "$TARGET_TAG" != v* ]]; then
            TARGET_TAG="v${TARGET_TAG}"
        fi
        CLEAN_VERSION="${TARGET_TAG#v}"
        return
    fi

    log_info "Resolving latest release from GitHub..."
    local latest_url
    latest_url="$(curl -fsIL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest" 2>/dev/null || true)"

    if [ -z "$latest_url" ] || [ "$latest_url" = "https://github.com/${REPO}/releases/latest" ]; then
        log_erro "Could not resolve latest release URL from GitHub."
        exit 1
    fi

    TARGET_TAG="${latest_url##*/}"
    CLEAN_VERSION="${TARGET_TAG#v}"
    log_info "Latest version detected: $TARGET_TAG"
}

do_rollback() {
    local target_dir="$1"
    local binary="$target_dir/streamer-vm"
    local backup="$target_dir/streamer-vm.backup"

    log_info "Looking for backup in $target_dir..."

    if [ ! -f "$backup" ]; then
        log_erro "No backup found at $backup"
        exit 1
    fi

    cp -p "$backup" "$binary"
    chmod +x "$binary"

    if "$binary" help >/dev/null 2>&1; then
        log_ok "Successfully rolled back to backup binary at $binary"
    else
        log_warn "Rolled back binary failed sanity test. Check $binary manually."
    fi
    exit 0
}

main() {
    local rollback=false

    while [ $# -gt 0 ]; do
        case "$1" in
            -r|--rollback)
                rollback=true
                shift
                ;;
            -v|--version)
                if [ $# -lt 2 ]; then
                    log_erro "Missing argument for --version"
                    exit 1
                fi
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                if [ $# -lt 2 ]; then
                    log_erro "Missing argument for --dir"
                    exit 1
                fi
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                print_usage
                exit 0
                ;;
            *)
                log_erro "Unknown option: $1"
                print_usage
                exit 1
                ;;
        esac
    done

    local target_dir
    target_dir="$(resolve_target_dir)"

    if [ "$rollback" = true ]; then
        do_rollback "$target_dir"
    fi

    detect_platform
    resolve_version

    check_path "$target_dir"

    local tarball_name="streamer-vm_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
    local download_url="https://github.com/${REPO}/releases/download/${TARGET_TAG}/${tarball_name}"
    local checksums_url="https://github.com/${REPO}/releases/download/${TARGET_TAG}/checksums.txt"

TMP_DIR=""
cleanup() {
    if [ -n "${TMP_DIR:-}" ] && [ -d "$TMP_DIR" ]; then
        rm -rf "$TMP_DIR"
    fi
}
trap cleanup EXIT

    TMP_DIR="$(mktemp -d)"

    log_info "Downloading $tarball_name..."
    if ! curl -fsSL -o "$TMP_DIR/$tarball_name" "$download_url"; then
        log_erro "Failed to download release asset from $download_url"
        exit 1
    fi

    log_info "Verifying SHA256 checksum..."
    if curl -fsSL -o "$TMP_DIR/checksums.txt" "$checksums_url" 2>/dev/null; then
        (
            cd "$TMP_DIR"
            if command -v sha256sum >/dev/null 2>&1; then
                sha256sum --check --ignore-missing checksums.txt >/dev/null 2>&1 || {
                    log_erro "Checksum verification failed for $tarball_name!"
                    exit 1
                }
            elif command -v shasum >/dev/null 2>&1; then
                shasum -a 256 --check --ignore-missing checksums.txt >/dev/null 2>&1 || {
                    log_erro "Checksum verification failed for $tarball_name!"
                    exit 1
                }
            fi
        )
        log_ok "Checksum verified successfully."
    else
        log_warn "Checksum file not available; skipping checksum verification."
    fi

    log_info "Extracting streamer-vm..."
    tar -xzf "$TMP_DIR/$tarball_name" -C "$TMP_DIR" streamer-vm

    # Sanity check on downloaded binary
    if ! "$TMP_DIR/streamer-vm" help >/dev/null 2>&1; then
        log_erro "Sanity check failed for extracted binary."
        exit 1
    fi

    mkdir -p "$target_dir"
    local dest_binary="$target_dir/streamer-vm"
    local backup_binary="$target_dir/streamer-vm.backup"

    if [ -f "$dest_binary" ]; then
        log_info "Creating backup of existing binary at $backup_binary..."
        cp -p "$dest_binary" "$backup_binary"
    fi

    log_info "Installing to $dest_binary..."
    install -m 755 "$TMP_DIR/streamer-vm" "$dest_binary"

    if "$dest_binary" help >/dev/null 2>&1; then
        log_ok "Successfully installed streamer-vm $TARGET_TAG to $dest_binary"
    else
        log_erro "Installed binary sanity check failed! Attempting recovery from backup..."
        if [ -f "$backup_binary" ]; then
            cp -p "$backup_binary" "$dest_binary"
            log_ok "Restored previous binary from backup."
        fi
        exit 1
    fi
}

main "$@"
