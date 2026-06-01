#!/usr/bin/env bash
#
# keystone bootstrap — installs the `keystone` binary into ~/.local/bin
# and (optionally) runs `keystone init` in the current directory.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/tacoda/keystone/main/install.sh | sh -s -- <agent>
#
# Environment overrides:
#   KEYSTONE_VERSION    pin a release tag (default: latest)
#   KEYSTONE_PREFIX     install dir (default: $HOME/.local/bin)
#   KEYSTONE_NO_INIT=1  skip the post-install `keystone init` step
#
# After install, `keystone init` is invoked unless KEYSTONE_NO_INIT=1.

set -euo pipefail

REPO="tacoda/keystone"
PREFIX="${KEYSTONE_PREFIX:-$HOME/.local/bin}"
VERSION="${KEYSTONE_VERSION:-latest}"

SUPPORTED_AGENTS="claude-code codex pi cursor aider github-copilot-cli continue cline goose _generic"

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

case ":$PATH:" in
  *":$PREFIX:"*) ;;
  *) warn "$PREFIX is not on your PATH; add it to your shell rc:"
     printf '    export PATH="%s:$PATH"\n' "$PREFIX" >&2
     ;;
esac

# ----- Optional: run init in current directory ------------------------------

if [ "${KEYSTONE_NO_INIT:-}" = "1" ]; then
  info "KEYSTONE_NO_INIT=1 set — skipping init"
  exit 0
fi

AGENT="${1:-}"

# Best-effort detection (mirrors `keystone init`'s built-in detection but lets
# us prompt before invoking the binary).
detect_agent() {
  if   [ -f CLAUDE.md ]              || [ -d .claude ];   then echo "claude-code"
  elif [ -f AGENTS.md ]              && [ -d .pi ];       then echo "pi"
  elif [ -f .github/copilot-instructions.md ];            then echo "github-copilot-cli"
  elif [ -d .cursor ];                                    then echo "cursor"
  elif [ -f .aider.conf.yml ]        || [ -f CONVENTIONS.md ]; then echo "aider"
  elif [ -f AGENTS.md ];                                  then echo "codex"
  elif [ -f .continuerules ];                             then echo "continue"
  elif [ -f .goosehints ];                                then echo "goose"
  else                                                         echo ""
  fi
}

if [ -z "$AGENT" ]; then
  detected=$(detect_agent)
  if [ -n "$detected" ]; then
    info "detected agent: $detected"
    AGENT="$detected"
  else
    printf 'which coding agent are you using? ['
    printf '%s' "$SUPPORTED_AGENTS" | tr ' ' '|'
    printf '] '
    read -r AGENT </dev/tty
  fi
fi

case " $SUPPORTED_AGENTS " in
  *" $AGENT "*) ;;
  *) err "unknown agent '$AGENT'. supported: $SUPPORTED_AGENTS"; exit 1 ;;
esac

FORCE_FLAG=""
if [ -d harness ]; then
  warn "harness/ already exists in this directory."
  printf "overwrite? [y/N] "
  read -r answer </dev/tty
  case "$answer" in
    y|Y|yes|YES) FORCE_FLAG="--force" ;;
    *) err "aborted."; exit 1 ;;
  esac
fi

info "running: keystone init --agent $AGENT $FORCE_FLAG"
"$PREFIX/keystone" init --agent "$AGENT" $FORCE_FLAG
