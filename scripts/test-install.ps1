$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$installer = Get-Content (Join-Path $root 'install.ps1') -Raw

if ($installer -notmatch 'FileShare\]::None') {
  throw 'installer must check that an existing target is not locked'
}
if ($installer -notmatch 'Close running balde processes') {
  throw 'installer must provide actionable guidance for a locked target'
}

Write-Host 'PowerShell installer checks passed'
