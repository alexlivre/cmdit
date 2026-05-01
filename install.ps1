# cmdit installer — one-liner for Windows (PowerShell)
# Usage: irm https://raw.githubusercontent.com/alexlivre/cmdit/main/install.ps1 | iex

param(
    [string]$Version = "latest"
)

$repo = "alexlivre/cmdit"
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$binary = "cmdit-windows-$arch.exe"

if ($Version -eq "latest") {
    $url = "https://github.com/$repo/releases/latest/download/$binary"
} else {
    $url = "https://github.com/$repo/releases/download/$Version/$binary"
}

Write-Host "→ Installing cmdit for Windows/$arch..." -ForegroundColor Cyan
Write-Host "  Downloading: $url"

$tmp = "$env:TEMP\cmdit.exe"
Invoke-WebRequest -Uri $url -OutFile $tmp

# Install to a directory in PATH
$installDir = "$env:LOCALAPPDATA\cmdit"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Move-Item -Force $tmp "$installDir\cmdit.exe"

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path += ";$installDir"
}

Write-Host "✅ cmdit installed to $installDir\cmdit.exe" -ForegroundColor Green
Write-Host "   Run: cmdit [file]"
