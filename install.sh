#!/usr/bin/env bash
#
# Keystone installer — downloads the `keystone` binary into ~/.local/bin
# (or $KEYSTONE_PREFIX) and ensures that directory is on your PATH.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh | sh
#
# Environment overrides:
#   KEYSTONE_VERSION    pin a release tag (default: latest)
#   KEYSTONE_PREFIX     install dir (default: $HOME/.local/bin)
#
# This installer does NOT run `keystone init`. Once installed, open a new
# shell (or `source` your shell rc) and run `keystone init` in any project
# to scaffold the harness.

set -euo pipefail

REPO="tacoda/keystone"
PREFIX="${KEYSTONE_PREFIX:-$HOME/.local/bin}"
VERSION="${KEYSTONE_VERSION:-latest}"

info()  { printf '\033[1;34m▸\033[0m %s\n' "$*"; }
warn()  { printf '\033[1;33m!\033[0m %s\n' "$*" >&2; }
err()   { printf '\033[1;31m✗\033[0m %s\n' "$*" >&2; }
ok()    { printf '\033[1;32m✓\033[0m %s\n' "$*"; }

# ----- Detect OS / arch -----------------------------------------------------

uname_s=$(uname -s)
uname_m=$(uname -m)

case "$uname_s" in
  Darwin) os="darwin" ;;
  Linux)  os="linux" ;;
  *)      err "unsupported OS: $uname_s (Windows users: use install.ps1)"; exit 1 ;;
esac

case "$uname_m" in
  x86_64|amd64) arch="x86_64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) err "unsupported arch: $uname_m"; exit 1 ;;
esac

# ----- Resolve version ------------------------------------------------------

if [ "$VERSION" = "latest" ]; then
  info "resolving latest release..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | grep -m1 '"tag_name":' \
    | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
  if [ -z "$VERSION" ]; then
    err "could not resolve latest release tag"
    exit 1
  fi
fi
# Strip leading "v" if present for the archive path.
VERSION_NO_V="${VERSION#v}"

# ----- Download and extract -------------------------------------------------

ARCHIVE="keystone_${VERSION_NO_V}_${os}_${arch}.tar.gz"
URL="https://github.com/$REPO/releases/download/${VERSION}/${ARCHIVE}"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

info "downloading $ARCHIVE ..."
if ! curl -fsSL -o "$TMP/$ARCHIVE" "$URL"; then
  err "failed to download $URL"
  exit 1
fi

info "extracting ..."
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

if [ ! -f "$TMP/keystone" ]; then
  err "extracted archive missing keystone binary"
  exit 1
fi

# ----- Install --------------------------------------------------------------

mkdir -p "$PREFIX"
mv "$TMP/keystone" "$PREFIX/keystone"
chmod +x "$PREFIX/keystone"
ok "installed $PREFIX/keystone ($VERSION)"

# ----- Ensure PATH ----------------------------------------------------------

ensure_on_path() {
  ep_prefix="$1"

  # Already on the current PATH? Nothing to do.
  case ":$PATH:" in
    *":$ep_prefix:"*) info "$ep_prefix already on PATH"; return 0 ;;
  esac

  # Detect the user's login shell and choose its rc file.
  ep_shell=$(basename "${SHELL:-/bin/sh}")
  ep_rc=""
  ep_line=""

  case "$ep_shell" in
    zsh)
      ep_rc="${ZDOTDIR:-$HOME}/.zshrc"
      ep_line="export PATH=\"$ep_prefix:\$PATH\""
      ;;
    bash)
      if [ -f "$HOME/.bashrc" ]; then
        ep_rc="$HOME/.bashrc"
      elif [ "$os" = "darwin" ]; then
        ep_rc="$HOME/.bash_profile"
      else
        ep_rc="$HOME/.profile"
      fi
      ep_line="export PATH=\"$ep_prefix:\$PATH\""
      ;;
    fish)
      ep_rc="$HOME/.config/fish/config.fish"
      ep_line="fish_add_path \"$ep_prefix\""
      ;;
    *)
      warn "unrecognized shell ($ep_shell); add this to your shell rc by hand:"
      printf '    export PATH="%s:$PATH"\n' "$ep_prefix" >&2
      return 0
      ;;
  esac

  # Idempotent: skip if the prefix is already referenced anywhere in the rc.
  if [ -f "$ep_rc" ] && grep -Fq "$ep_prefix" "$ep_rc" 2>/dev/null; then
    warn "$ep_prefix already referenced in $ep_rc but not on the current PATH"
    warn "open a new shell, or run: source $ep_rc"
    return 0
  fi

  mkdir -p "$(dirname "$ep_rc")"
  {
    printf '\n# Added by keystone installer\n'
    printf '%s\n' "$ep_line"
  } >> "$ep_rc"
  ok "added $ep_prefix to PATH via $ep_rc"
  warn "open a new shell, or run: source $ep_rc"
}

ensure_on_path "$PREFIX"

# ----- Next steps -----------------------------------------------------------

cat <<EOF

Run keystone init in any project to scaffold the harness:

  \$ cd /path/to/your/project
  \$ keystone init

See 'keystone help' for options.
EOF
