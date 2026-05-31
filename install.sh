#!/usr/bin/env bash
#
# keystone — install the project harness into the current directory
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/<org>/keystone/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/<org>/keystone/main/install.sh | sh -s -- <agent>
#
# Or download and inspect first (recommended):
#   curl -fsSL https://raw.githubusercontent.com/<org>/keystone/main/install.sh > install.sh
#   less install.sh
#   sh install.sh [<agent>]
#
# Supported <agent>: claude-code, codex, pi, cursor, aider, github-copilot-cli,
#                    continue, cline, goose
# Omit <agent> to be prompted, or detect via existing repo files.

set -euo pipefail

# ----- Configuration --------------------------------------------------------

# Where to fetch keystone from. Override with KEYSTONE_REPO_URL.
KEYSTONE_REPO_URL="${KEYSTONE_REPO_URL:-https://github.com/tacoda/keystone}"
KEYSTONE_BRANCH="${KEYSTONE_BRANCH:-main}"
TARBALL_URL="${KEYSTONE_REPO_URL}/archive/refs/heads/${KEYSTONE_BRANCH}.tar.gz"

SUPPORTED_AGENTS="claude-code codex pi cursor aider github-copilot-cli continue cline goose"

# ----- Output helpers -------------------------------------------------------

info()  { printf '\033[1;34m▸\033[0m %s\n' "$*"; }
warn()  { printf '\033[1;33m!\033[0m %s\n' "$*" >&2; }
err()   { printf '\033[1;31m✗\033[0m %s\n' "$*" >&2; }
ok()    { printf '\033[1;32m✓\033[0m %s\n' "$*"; }

# ----- Argument parsing -----------------------------------------------------

AGENT="${1:-}"

# ----- Agent detection (when not provided) ---------------------------------

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
    info "Detected agent: $detected"
    AGENT="$detected"
  else
    printf 'Which coding agent are you using? ['
    printf '%s' "$SUPPORTED_AGENTS" | tr ' ' '|'
    printf '|_generic] '
    read -r AGENT
  fi
fi

# Validate agent.
case " $SUPPORTED_AGENTS _generic " in
  *" $AGENT "*) ;;
  *)
    err "Unknown agent '$AGENT'. Supported: $SUPPORTED_AGENTS _generic"
    exit 1
    ;;
esac

# ----- Pre-flight checks ----------------------------------------------------

if [ -d harness ]; then
  warn "harness/ already exists in this directory."
  printf "Overwrite? [y/N] "
  read -r answer
  case "$answer" in
    y|Y|yes|YES) info "Will overwrite existing harness/." ;;
    *) err "Aborted."; exit 1 ;;
  esac
fi

# Soft prerequisite check (documents, does not enforce).
info "Checking prerequisites (informational — keystone runs regardless)..."

if [ ! -d .git ] && [ ! -f .git ]; then
  warn "Not a git repository. Keystone assumes git-tracked projects."
fi

ci_found=""
for ci in .github/workflows .gitlab-ci.yml .circleci/config.yml .travis.yml \
          azure-pipelines.yml bitbucket-pipelines.yml Jenkinsfile; do
  if [ -e "$ci" ]; then ci_found="$ci"; break; fi
done
if [ -z "$ci_found" ]; then
  warn "No CI config detected. The release phase assumes a CI pipeline; consider adding one."
else
  ok "CI config: $ci_found"
fi

if ! command -v git >/dev/null 2>&1; then
  err "git is required and was not found in PATH."
  exit 1
fi

# ----- Fetch keystone -------------------------------------------------------

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

info "Fetching keystone from $KEYSTONE_REPO_URL ($KEYSTONE_BRANCH)..."

if ! curl -fsSL "$TARBALL_URL" | tar -xz -C "$TMP"; then
  err "Failed to fetch keystone. Check KEYSTONE_REPO_URL and your network."
  exit 1
fi

# The tarball extracts to a single top-level dir named keystone-<branch>.
SRC=$(find "$TMP" -mindepth 1 -maxdepth 1 -type d | head -n1)
if [ ! -d "$SRC/harness" ]; then
  err "Unexpected tarball layout — no harness/ at $SRC/harness."
  exit 1
fi

# ----- Install corpus -------------------------------------------------------

info "Installing corpus to ./harness/..."
mkdir -p harness
cp -R "$SRC/harness/." harness/

# ----- Install agent target -------------------------------------------------

if [ -d "$SRC/targets/$AGENT" ]; then
  info "Installing $AGENT target..."
  # Use a portable approach: walk files and cp each.
  ( cd "$SRC/targets/$AGENT" && find . -type f -print0 ) | \
    while IFS= read -r -d '' file; do
      dest="${file#./}"
      mkdir -p "$(dirname "$dest")"
      if [ -f "$dest" ]; then
        warn "  exists: $dest (skipped — review and merge manually)"
      else
        cp "$SRC/targets/$AGENT/$file" "$dest"
        ok "  wrote: $dest"
      fi
    done
elif [ "$AGENT" = "_generic" ] || [ -d "$SRC/targets/_generic" ]; then
  info "Installing generic AGENTS.md..."
  if [ -f AGENTS.md ]; then
    warn "  exists: AGENTS.md (skipped — review and merge manually)"
  else
    cp "$SRC/targets/_generic/AGENTS.md" .
    ok "  wrote: AGENTS.md"
  fi
else
  warn "No target found for $AGENT. Corpus installed; configure activation manually."
fi

# ----- Done -----------------------------------------------------------------

ok "keystone installed for $AGENT."
cat <<'EOF'

Next steps:

  1. Read harness/README.md
  2. Run the bootstrap action in your agent to populate state/CODEBASE_STATE.md
     and idioms/<your-stack>/ from your project.
  3. Commit harness/ and any agent-specific files this installer created.

Prerequisites the harness assumes (soft):

  - a way to track work (tracker card, TODO.md, or conversation)
  - lint / type-check / test / build commands
  - pull-request workflow
  - CI pipeline (CD is even better)

Missing one degrades the corresponding phase but does not break the harness.
EOF
