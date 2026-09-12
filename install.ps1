# balde installer/upgrader — https://github.com/egermano/balde
#
# Install or upgrade with a single command (PowerShell):
#
#   irm https://raw.githubusercontent.com/egermano/balde/main/install.ps1 | iex
#
# Environment overrides:
#   BALDE_VERSION      pin a specific release tag (e.g. v0.1.0-alpha.3)
#   BALDE_INSTALL_DIR  install destination (default: %USERPROFILE%\bin, or the
#                      directory of an existing balde installation)
#   BALDE_FORCE=1      reinstall even when the latest version is already installed
#
# The script leaves nothing behind: all downloads happen in a temporary
# directory that is always removed, and balde.exe is moved into place only
# after its sha256 checksum has been verified.

#Requires -Version 5.1

$ErrorActionPreference = 'Stop'
$Repo = 'egermano/balde'
$Github = "https://github.com/$Repo"

[Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

# --- detect architecture ----------------------------------------------------

$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
  'AMD64' { 'x86_64' }
  'ARM64' { 'arm64' }
  default { throw "Unsupported architecture '$($env:PROCESSOR_ARCHITECTURE)' - supported: AMD64, ARM64" }
}
$asset = "balde_Windows_$arch.zip"

# --- resolve release tag ----------------------------------------------------

# Latest stable release tag, resolved from the releases/latest redirect (no
# GitHub API involved). If the project only has prereleases, that URL returns
# 404 and we fall back to the API (most recent release, prerelease included).
function Get-LatestTag {
  try {
    $req = [Net.WebRequest]::Create("$Github/releases/latest")
    $req.AllowAutoRedirect = $false
    $req.Method = 'HEAD'
    $resp = $req.GetResponse()
    try { return $resp.Headers['Location'] } finally { $resp.Close() }
  } catch {
    return $null
  }
}

if ($env:BALDE_VERSION) {
  $tag = $env:BALDE_VERSION
} else {
  Write-Host '==> resolving latest release'
  $tag = Get-LatestTag
  if ($tag) { $tag = $tag -replace '^.*/tag/', '' }
  if (-not $tag) {
    try {
      $rel = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases?per_page=1" -UseBasicParsing
      $tag = $rel[0].tag_name
    } catch {
      $tag = $null
    }
  }
  if (-not $tag) { throw 'Could not resolve the latest balde release' }
}
if ($tag -notlike 'v*') { $tag = "v$tag" }

# --- pick install destination -----------------------------------------------

$cmd = Get-Command balde -ErrorAction SilentlyContinue
if ($env:BALDE_INSTALL_DIR) {
  $installDir = $env:BALDE_INSTALL_DIR
} elseif ($cmd -and $cmd.Source) {
  $installDir = Split-Path $cmd.Source
} else {
  $installDir = Join-Path $env:USERPROFILE 'bin'
}
$target = Join-Path $installDir 'balde.exe'

$oldVersion = $null
if (Test-Path $target) {
  try {
    $out = & $target --version 2>$null
    if ($LASTEXITCODE -eq 0 -and $out) { $oldVersion = ($out -split '\s+')[2] }
  } catch {
    $oldVersion = $null
  }
}

# --- no-op when already up to date ------------------------------------------

if ($env:BALDE_FORCE -ne '1' -and $oldVersion) {
  if ($oldVersion.TrimStart('v') -eq $tag.TrimStart('v')) {
    Write-Host "==> balde $tag is already installed at $target - up to date, nothing to do"
    return
  }
}

# --- download and verify ----------------------------------------------------

$tmpDir = Join-Path ([IO.Path]::GetTempPath()) ("balde-install-" + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
  $baseUrl = "$Github/releases/download/$tag"
  $archive = Join-Path $tmpDir $asset
  Write-Host "==> downloading balde $tag (Windows/$arch)"
  try {
    Invoke-WebRequest -Uri "$baseUrl/$asset" -OutFile $archive -UseBasicParsing
    Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile (Join-Path $tmpDir 'checksums.txt') -UseBasicParsing
  } catch {
    throw "Download failed: $($_.Exception.Message)"
  }

  $lines = @(Get-Content (Join-Path $tmpDir 'checksums.txt') | Where-Object { $_ -like "*  $asset" })
  if ($lines.Count -eq 0) { throw "Checksum for $asset not found in checksums.txt" }
  $expected = ($lines[0] -split '\s+')[0]
  $actual = (Get-FileHash -Algorithm SHA256 $archive).Hash.ToLower()
  if ($actual -ne $expected.ToLower()) {
    throw "Checksum mismatch for ${asset}:`n  expected: $expected`n  actual:   $actual"
  }

  Expand-Archive -Path $archive -DestinationPath $tmpDir -Force
  $bin = Join-Path $tmpDir 'balde.exe'
  if (-not (Test-Path $bin)) { throw 'Archive did not contain a balde.exe binary' }

  # --- install ---------------------------------------------------------------

  if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
  }

  Write-Host "==> installing balde $tag -> $target"
  $tmpTarget = Join-Path $installDir ('.balde.tmp.' + [Guid]::NewGuid().ToString('N') + '.exe')
  Copy-Item $bin $tmpTarget -Force
  Move-Item -Force -Path $tmpTarget -Destination $target

  $was = if ($oldVersion) { " (was $oldVersion)" } else { '' }
  Write-Host "==> balde $tag installed at $target$was"

  if (($env:PATH -split ';') -notcontains $installDir) {
    Write-Host ''
    Write-Host "NOTE: $installDir is not on your PATH. Add it for your user:"
    Write-Host "      [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$installDir', 'User')"
  }
} finally {
  Remove-Item $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
