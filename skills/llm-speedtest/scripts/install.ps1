# Install the latest llm-speedtest release binary on Windows.
# Usage: powershell -ExecutionPolicy Bypass -File scripts/install.ps1 [-Version v0.1.0] [-InstallDir C:\tools]
param(
  [string]$Version = "latest",
  [string]$InstallDir = "$env:LOCALAPPDATA\Programs\llm-speedtest"
)
$ErrorActionPreference = "Stop"
$repo = "AnnaofArendelle/llm-speedtest"
$arch = if ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -eq "Arm64") { "arm64" } else { "amd64" }
if ($Version -eq "latest") {
  $Version = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
}
$ver = $Version.TrimStart("v")
$asset = "llm-speedtest_${ver}_windows_${arch}.zip"
$base = "https://github.com/$repo/releases/download/$Version"
$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tmp | Out-Null
Write-Host "downloading $asset"
Invoke-WebRequest "$base/$asset" -OutFile (Join-Path $tmp $asset)
Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp "checksums.txt")
$expected = (Get-Content (Join-Path $tmp "checksums.txt") | Where-Object { $_ -match [regex]::Escape($asset) }).Split(" ")[0]
$actual = (Get-FileHash (Join-Path $tmp $asset) -Algorithm SHA256).Hash.ToLower()
if ($expected -ne $actual) { throw "checksum mismatch for $asset" }
Expand-Archive (Join-Path $tmp $asset) -DestinationPath $tmp -Force
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item (Join-Path $tmp "llm-speedtest.exe") (Join-Path $InstallDir "llm-speedtest.exe") -Force
Remove-Item $tmp -Recurse -Force
Write-Host "installed $InstallDir\llm-speedtest.exe"
Write-Host "add $InstallDir to PATH if it is not there yet"
