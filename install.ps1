# keystone - install the project harness into the current directory (Windows)
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/<org>/keystone/main/install.ps1 | iex
#   iwr -useb https://raw.githubusercontent.com/<org>/keystone/main/install.ps1 | iex; Install-Keystone -Agent claude-code
#
# Or download and inspect first (recommended):
#   iwr https://raw.githubusercontent.com/<org>/keystone/main/install.ps1 -OutFile install.ps1
#   Get-Content install.ps1 | more
#   .\install.ps1 [-Agent <name>]
#
# Supported -Agent: claude-code, codex, pi, cursor, aider, github-copilot-cli,
#                   continue, cline, goose
# Omit -Agent to be prompted, or detect via existing repo files.

[CmdletBinding()]
param(
    [string]$Agent = "",
    [string]$RepoUrl = $(if ($env:KEYSTONE_REPO_URL) { $env:KEYSTONE_REPO_URL } else { "https://github.com/tacoda/keystone" }),
    [string]$Branch  = $(if ($env:KEYSTONE_BRANCH) { $env:KEYSTONE_BRANCH } else { "main" })
)

$ErrorActionPreference = "Stop"

$SupportedAgents = @(
    "claude-code", "codex", "pi", "cursor", "aider",
    "github-copilot-cli", "continue", "cline", "goose"
)

# ----- Output helpers -------------------------------------------------------

function Write-Info  { param([string]$m) Write-Host "▸ $m" -ForegroundColor Blue }
function Write-Warn  { param([string]$m) Write-Host "! $m" -ForegroundColor Yellow }
function Write-Err   { param([string]$m) Write-Host "✗ $m" -ForegroundColor Red }
function Write-OK    { param([string]$m) Write-Host "✓ $m" -ForegroundColor Green }

# ----- Agent detection ------------------------------------------------------

function Detect-Agent {
    if (Test-Path "CLAUDE.md")          { return "claude-code" }
    if (Test-Path ".claude")            { return "claude-code" }
    if ((Test-Path "AGENTS.md") -and (Test-Path ".pi")) { return "pi" }
    if (Test-Path ".github/copilot-instructions.md")    { return "github-copilot-cli" }
    if (Test-Path ".cursor")            { return "cursor" }
    if (Test-Path ".aider.conf.yml")    { return "aider" }
    if (Test-Path "CONVENTIONS.md")     { return "aider" }
    if (Test-Path "AGENTS.md")          { return "codex" }
    if (Test-Path ".continuerules")     { return "continue" }
    if (Test-Path ".goosehints")        { return "goose" }
    return ""
}

if (-not $Agent) {
    $detected = Detect-Agent
    if ($detected) {
        Write-Info "Detected agent: $detected"
        $Agent = $detected
    } else {
        $choice = Read-Host "Which coding agent are you using? [$($SupportedAgents -join '|')|_generic]"
        $Agent = $choice
    }
}

if ($SupportedAgents -notcontains $Agent -and $Agent -ne "_generic") {
    Write-Err "Unknown agent '$Agent'. Supported: $($SupportedAgents -join ', '), _generic"
    exit 1
}

# ----- Pre-flight checks ----------------------------------------------------

if (Test-Path "harness") {
    Write-Warn "harness/ already exists in this directory."
    $answer = Read-Host "Overwrite? [y/N]"
    if ($answer -notmatch '^(y|Y|yes|YES)$') {
        Write-Err "Aborted."
        exit 1
    }
    Write-Info "Will overwrite existing harness/."
}

Write-Info "Checking prerequisites (informational - keystone runs regardless)..."

if (-not (Test-Path ".git")) {
    Write-Warn "Not a git repository. Keystone assumes git-tracked projects."
}

$CiCandidates = @(
    ".github/workflows", ".gitlab-ci.yml", ".circleci/config.yml",
    ".travis.yml", "azure-pipelines.yml", "bitbucket-pipelines.yml", "Jenkinsfile"
)
$ciFound = $null
foreach ($p in $CiCandidates) { if (Test-Path $p) { $ciFound = $p; break } }
if ($ciFound) {
    Write-OK "CI config: $ciFound"
} else {
    Write-Warn "No CI config detected. The release phase assumes a CI pipeline; consider adding one."
}

if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
    Write-Err "git is required and was not found in PATH."
    exit 1
}

# ----- Fetch keystone -------------------------------------------------------

$TarballUrl = "$RepoUrl/archive/refs/heads/$Branch.tar.gz"
$Tmp = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "keystone-$([guid]::NewGuid().ToString('N'))"))

try {
    Write-Info "Fetching keystone from $RepoUrl ($Branch)..."

    $tarFile = Join-Path $Tmp.FullName "keystone.tar.gz"
    try {
        Invoke-WebRequest -Uri $TarballUrl -OutFile $tarFile -UseBasicParsing
    } catch {
        Write-Err "Failed to fetch keystone. Check RepoUrl and your network."
        throw
    }

    # tar.exe ships on Windows 10 1803+ and Windows 11.
    if (-not (Get-Command tar -ErrorAction SilentlyContinue)) {
        Write-Err "tar is required to extract the keystone archive. Windows 10 1803+ ships tar.exe; older systems need Git Bash or 7-Zip."
        exit 1
    }
    & tar -xzf $tarFile -C $Tmp.FullName
    if ($LASTEXITCODE -ne 0) {
        Write-Err "Failed to extract archive."
        exit 1
    }

    $Src = Get-ChildItem -Path $Tmp.FullName -Directory | Where-Object { $_.Name -like "keystone-*" } | Select-Object -First 1
    if (-not $Src -or -not (Test-Path (Join-Path $Src.FullName "harness"))) {
        Write-Err "Unexpected tarball layout - no harness/ found."
        exit 1
    }

    # ----- Install corpus ---------------------------------------------------

    Write-Info "Installing corpus to .\harness\..."
    if (-not (Test-Path "harness")) { New-Item -ItemType Directory -Path "harness" | Out-Null }
    Copy-Item -Path (Join-Path $Src.FullName "harness/*") -Destination "harness" -Recurse -Force

    # ----- Install agent target ---------------------------------------------

    $targetDir = Join-Path $Src.FullName "targets/$Agent"
    if (Test-Path $targetDir) {
        Write-Info "Installing $Agent target..."
        Get-ChildItem -Path $targetDir -Recurse -File | ForEach-Object {
            $rel = $_.FullName.Substring($targetDir.Length + 1)
            $dest = Join-Path (Get-Location) $rel
            $destDir = Split-Path $dest -Parent
            if ($destDir -and -not (Test-Path $destDir)) {
                New-Item -ItemType Directory -Path $destDir -Force | Out-Null
            }
            if (Test-Path $dest) {
                Write-Warn "  exists: $rel (skipped - review and merge manually)"
            } else {
                Copy-Item -Path $_.FullName -Destination $dest -Force
                Write-OK "  wrote: $rel"
            }
        }
    } elseif (Test-Path (Join-Path $Src.FullName "targets/_generic/AGENTS.md")) {
        Write-Info "Installing generic AGENTS.md..."
        if (Test-Path "AGENTS.md") {
            Write-Warn "  exists: AGENTS.md (skipped - review and merge manually)"
        } else {
            Copy-Item -Path (Join-Path $Src.FullName "targets/_generic/AGENTS.md") -Destination "AGENTS.md"
            Write-OK "  wrote: AGENTS.md"
        }
    } else {
        Write-Warn "No target found for $Agent. Corpus installed; configure activation manually."
    }

    Write-OK "keystone installed for $Agent."
    Write-Host ""
    Write-Host "Next steps:"
    Write-Host "  1. Read harness/README.md"
    Write-Host "  2. Run the bootstrap action in your agent to populate state/CODEBASE_STATE.md"
    Write-Host "     and idioms/<your-stack>/ from your project."
    Write-Host "  3. Commit harness/ and any agent-specific files this installer created."
    Write-Host ""
    Write-Host "Prerequisites the harness assumes (soft):"
    Write-Host "  - a way to track work (tracker card, TODO.md, or conversation)"
    Write-Host "  - lint / type-check / test / build commands"
    Write-Host "  - pull-request workflow"
    Write-Host "  - CI pipeline (CD is even better)"
    Write-Host ""
    Write-Host "Missing one degrades the corresponding phase but does not break the harness."
}
finally {
    if (Test-Path $Tmp.FullName) {
        Remove-Item -Path $Tmp.FullName -Recurse -Force -ErrorAction SilentlyContinue
    }
}
