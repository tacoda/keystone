# Keystone installer (Windows) — downloads the `keystone.exe` binary into
# $env:LOCALAPPDATA\Programs\keystone (or -Prefix) and adds the install
# directory to the user PATH if it is not already there.
#
# Usage:
#   iwr -useb https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 | iex
#   iwr https://raw.githubusercontent.com/tacoda/keystone/main/install.ps1 -OutFile install.ps1; .\install.ps1 -Version v0.7.0
#
# Parameters:
#   -Version <tag>  pin a release tag (default: latest)
#   -Prefix <path>  override install dir
#
# This installer does NOT run `keystone init`. Once installed, open a new
# terminal so the updated PATH is picked up, then run `keystone init` in
# any project to scaffold the harness.

[CmdletBinding()]
param(
    [string]$Version = $(if ($env:KEYSTONE_VERSION) { $env:KEYSTONE_VERSION } else { "latest" }),
    [string]$Prefix = $(if ($env:KEYSTONE_PREFIX) { $env:KEYSTONE_PREFIX } else { Join-Path $env:LOCALAPPDATA "Programs\keystone" })
)

$ErrorActionPreference = "Stop"

$Repo = "tacoda/keystone"

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

    # ----- Ensure PATH ------------------------------------------------------

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -like "*$Prefix*") {
        Write-Info "$Prefix already on user PATH"
    } else {
        $newPath = if ([string]::IsNullOrEmpty($userPath)) { $Prefix } else { "$userPath;$Prefix" }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-OK "added $Prefix to user PATH"
        Write-Warn "open a new terminal to pick up the change"
    }
}
finally {
    if (Test-Path $Tmp.FullName) { Remove-Item -Path $Tmp.FullName -Recurse -Force -ErrorAction SilentlyContinue }
}

# ----- Next steps -----------------------------------------------------------

Write-Host ""
Write-Host "Run keystone init in any project to scaffold the harness:"
Write-Host ""
Write-Host "  > cd C:\path\to\your\project"
Write-Host "  > keystone init"
Write-Host ""
Write-Host "See 'keystone help' for options."
