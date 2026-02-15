#!/bin/bash
set -euo pipefail

# ical installer — downloads the latest release from GitHub
# Usage: curl -fsSL https://ical.sidv.dev/install | bash

REPO="BRO3886/ical"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="ical"

info() { printf "\033[36m%s\033[0m\n" "$*"; }
error() { printf "\033[31mError: %s\033[0m\n" "$*" >&2; exit 1; }

# --- Pre-flight checks ---

if [ "$(uname -s)" != "Darwin" ]; then
    error "ical only supports macOS"
fi

ARCH="$(uname -m)"
case "$ARCH" in
    arm64|aarch64) ARCH="arm64" ;;
    x86_64)        ARCH="amd64" ;;
    *)             error "Unsupported architecture: $ARCH" ;;
esac

if ! command -v curl >/dev/null 2>&1; then
    error "curl is required but not found"
fi

# --- Resolve latest version ---

info "Fetching latest release..."
LATEST=$(curl -sSL -H "Accept: application/vnd.github+json" \
    "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    error "Could not determine latest release"
fi

info "Latest version: $LATEST"

# --- Download and extract ---

ASSET_NAME="ical-darwin-${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET_NAME}"

TMPDIR_PATH=$(mktemp -d)
trap 'rm -rf "$TMPDIR_PATH"' EXIT

info "Downloading ${ASSET_NAME}..."
HTTP_CODE=$(curl -sSL -w "%{http_code}" -o "${TMPDIR_PATH}/${ASSET_NAME}" "$DOWNLOAD_URL")

if [ "$HTTP_CODE" != "200" ]; then
    error "Download failed (HTTP $HTTP_CODE). Asset '${ASSET_NAME}' may not exist for ${LATEST}."
fi

tar -xzf "${TMPDIR_PATH}/${ASSET_NAME}" -C "${TMPDIR_PATH}"

# --- Install ---

if [ -w "$INSTALL_DIR" ]; then
    mv "${TMPDIR_PATH}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
    info "Requires sudo to install to ${INSTALL_DIR}"
    sudo mv "${TMPDIR_PATH}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
fi

info "Installed ical ${LATEST} to ${INSTALL_DIR}/${BINARY_NAME}"

# --- Verify ---

if command -v ical >/dev/null 2>&1; then
    info "Run 'ical --help' to get started"
else
    info "Note: ${INSTALL_DIR} may not be in your PATH"
fi
