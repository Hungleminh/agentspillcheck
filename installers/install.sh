#!/bin/sh
# AgentSpillCheck installer for macOS/Linux (PLAN 16.3).
# Chooses Docker when it is running, otherwise installs the native
# binary into ~/.local/bin and verifies its SHA-256 checksum.
set -eu

REPO="Hungleminh/agentspillcheck"
MODE="auto"
for arg in "$@"; do
  case "$arg" in
    --docker) MODE="docker" ;;
    --native) MODE="native" ;;
  esac
done

say() { printf '%s\n' "$1"; }
say "AgentSpillCheck installer / Trinh cai dat AgentSpillCheck"

if [ "$MODE" = "auto" ]; then
  if command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then
    MODE="docker"
  else
    MODE="native"
  fi
fi

if [ "$MODE" = "docker" ]; then
  say "Docker detected: using docker compose / Phat hien Docker: dung docker compose"
  mkdir -p "$HOME/.spillcheck"
  curl -fsSL "https://raw.githubusercontent.com/$REPO/main/deploy/docker-compose.yml" -o "$HOME/.spillcheck/docker-compose.yml"
  (cd "$HOME/.spillcheck" && docker compose up -d)
  say "Dashboard: http://127.0.0.1:7700/"
  exit 0
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac
VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -1)"
if [ -z "$VERSION" ]; then
  say "Cannot find the latest release / Khong tim thay ban phat hanh moi nhat" >&2
  exit 1
fi
ASSET="agentspillcheck_${VERSION#v}_${OS}_${ARCH}.tar.gz"
BASE="https://github.com/$REPO/releases/download/$VERSION"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
curl -fsSL "$BASE/$ASSET" -o "$TMP/$ASSET"
curl -fsSL "$BASE/checksums.txt" -o "$TMP/checksums.txt"
(cd "$TMP" && grep "$ASSET" checksums.txt | shasum -a 256 -c -)
mkdir -p "$HOME/.local/bin"
tar -xzf "$TMP/$ASSET" -C "$TMP"
install -m 0755 "$TMP/spillcheck" "$HOME/.local/bin/spillcheck"
"$HOME/.local/bin/spillcheck" doctor || true
say "Installed / Da cai: $HOME/.local/bin/spillcheck"
say "Run / Chay: spillcheck demo"
