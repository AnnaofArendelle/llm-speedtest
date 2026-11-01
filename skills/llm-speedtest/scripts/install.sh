#!/usr/bin/env bash
# Install the latest llm-speedtest release binary for this machine.
# Usage: scripts/install.sh [version] [install-dir]
set -euo pipefail

REPO="AnnaofArendelle/llm-speedtest"
VERSION="${1:-latest}"
INSTALL_DIR="${2:-${LLM_SPEEDTEST_INSTALL_DIR:-$HOME/.local/bin}}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;;
  *) echo "unsupported OS: $os (use install.ps1 on Windows)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)"
  [ -n "$VERSION" ] || { echo "could not determine the latest release" >&2; exit 1; }
fi
ver="${VERSION#v}"
asset="llm-speedtest_${ver}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$VERSION"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "downloading $asset"
curl -fsSL "$base/$asset" -o "$tmp/$asset"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"
if command -v sha256sum >/dev/null 2>&1; then
  (cd "$tmp" && grep " $asset\$" checksums.txt | sha256sum -c -)
else # macOS and other BSDs ship shasum instead of sha256sum
  (cd "$tmp" && grep " $asset\$" checksums.txt | shasum -a 256 -c -)
fi
tar -xzf "$tmp/$asset" -C "$tmp"
mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/llm-speedtest" "$INSTALL_DIR/llm-speedtest"
echo "installed $INSTALL_DIR/llm-speedtest ($("$INSTALL_DIR/llm-speedtest" version))"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "note: add $INSTALL_DIR to your PATH" ;;
esac
