# keystone bootstrap (Windows) - installs the `keystone` binary into
# $env:LOCALAPPDATA\Programs\keystone and (optionally) runs `keystone init` in
# the current directory.
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 | iex
#   iwr https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 -OutFile install.ps1; .\install.ps1 -Agent claude-code
#
# Parameters:
#   -Agent <name>   skip the agent prompt
#   -Version <tag>  pin a release tag (default: latest)
#   -Prefix <path>  override install dir
#   -NoInit         install the binary but skip `keystone init`

[CmdletBinding()]
param(
    [string]$Agent = "",
    [string]$Version = $(if ($env:KEYSTONE_VERSION) { $env:KEYSTONE_VERSION } else { "latest" }),
    [string]$Prefix = $(if ($env:KEYSTONE_PREFIX) { $env:KEYSTONE_PREFIX } else { Join-Path $env:LOCALAPPDATA "Programs\keystone" }),
    [switch]$NoInit
)

$ErrorActionPreference = "Stop"

$Repo = "tacoda/keystone"
$SupportedAgents = @(
    "claude-code", "codex", "pi", "cursor", "aider",
    "github-copilot-cli", "continue", "cline", "goose", "_generic"
)

function Write-Info { param([string]$m) Write-Host "> $m" -ForegroundColor Blue }
function Write-Warn { param([string]$m) Write-Host "! $m" -ForegroundColor Yellow }
function Write-Err  { param([string]$m) Write-Host "x $m" -ForegroundColor Red }
function Write-OK   { param([string]$m) Write-Host "+ $m" -ForegroundColor Green }

# ----- Resolve version ------------------------------------------------------

if ($Version -eq "latest") {
    Write-Info "resolving latest release..."
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
    $Version = $latest.tag_name
    if (-not $Version) { Write-Err "could not resolve latest release tag"; exit 1 }
}
$VersionNoV = $Version.TrimStart("v")

# ----- Download and extract -------------------------------------------------

$Archive = "keystone_${VersionNoV}_windows_x86_64.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"

$Tmp = New-Item -ItemType Directory -Path ([System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "keystone-$([guid]::NewGuid().ToString('N'))"))
try {
    $ZipPath = Join-Path $Tmp.FullName $Archive
    Write-Info "downloading $Archive ..."
    Invoke-WebRequest -Uri $Url -OutFile $ZipPath -UseBasicParsing

    Write-Info "extracting ..."
    Expand-Archive -Path $ZipPath -DestinationPath $Tmp.FullName -Force

    $Binary = Get-ChildItem -Path $Tmp.FullName -Recurse -Filter "keystone.exe" | Select-Object -First 1
    if (-not $Binary) { Write-Err "extracted archive missing keystone.exe"; exit 1 }

    # ----- Install ----------------------------------------------------------

    if (-not (Test-Path $Prefix)) { New-Item -ItemType Directory -Path $Prefix -Force | Out-Null }
    $InstallPath = Join-Path $Prefix "keystone.exe"
    Copy-Item -Path $Binary.FullName -Destination $InstallPath -Force
    Write-OK "installed $InstallPath ($Version)"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$Prefix*") {
        Write-Warn "$Prefix is not on your PATH. add it with:"
        Write-Host "    [Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$Prefix`", 'User')"
    }
}
finally {
    if (Test-Path $Tmp.FullName) { Remove-Item -Path $Tmp.FullName -Recurse -Force -ErrorAction SilentlyContinue }
}

if ($NoInit) {
    Write-Info "-NoInit set, skipping init"
    return
}

# ----- Agent prompt + init --------------------------------------------------

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
        Write-Info "detected agent: $detected"
        $Agent = $detected
    } else {
        $Agent = Read-Host "which coding agent are you using? [$($SupportedAgents -join '|')]"
    }
}

if ($SupportedAgents -notcontains $Agent) {
    Write-Err "unknown agent '$Agent'. supported: $($SupportedAgents -join ', ')"
    exit 1
}

$ForceArgs = @()
if (Test-Path "harness") {
    Write-Warn "harness/ already exists in this directory."
    $answer = Read-Host "overwrite? [y/N]"
    if ($answer -match '^(y|Y|yes|YES)$') {
        $ForceArgs = @("--force")
    } else {
        Write-Err "aborted."
        exit 1
    }
}

Write-Info "running: keystone init --agent $Agent $($ForceArgs -join ' ')"
& $InstallPath init --agent $Agent @ForceArgs
