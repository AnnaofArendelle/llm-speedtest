#!/usr/bin/env sh
# Installs the latest llm-speedtest release binary from GitHub.
#   curl -fsSL https://raw.githubusercontent.com/AnnaofArendelle/llm-speedtest/main/scripts/install.sh | sh
# Environment:
#   LLM_SPEEDTEST_VERSION  tag to install (default: latest)
#   LLM_SPEEDTEST_INSTALL  install directory (default: /usr/local/bin if writable, else ~/.local/bin)
set -eu

REPO="AnnaofArendelle/llm-speedtest"
VERSION="${LLM_SPEEDTEST_VERSION:-latest}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;;
  *) echo "unsupported OS: $os (download the Windows zip from the releases page)" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)
  [ -n "$VERSION" ] || { echo "could not determine the latest version" >&2; exit 1; }
fi
ver=${VERSION#v}
asset="llm-speedtest_${ver}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/$VERSION"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
echo "downloading $asset ..."
curl -fsSL "$base/$asset" -o "$tmp/$asset"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"
(cd "$tmp" && grep " $asset\$" checksums.txt | sha256sum -c - >/dev/null) || { echo "checksum verification failed" >&2; exit 1; }
tar -xzf "$tmp/$asset" -C "$tmp"

dest="${LLM_SPEEDTEST_INSTALL:-}"
if [ -z "$dest" ]; then
  if [ -w /usr/local/bin ]; then dest=/usr/local/bin; else dest="$HOME/.local/bin"; fi
fi
mkdir -p "$dest"
install -m 0755 "$tmp/llm-speedtest" "$dest/llm-speedtest"
echo "installed llm-speedtest $VERSION to $dest/llm-speedtest"
case ":$PATH:" in
  *":$dest:"*) ;;
  *) echo "note: add $dest to your PATH" ;;
esac
"$dest/llm-speedtest" version
