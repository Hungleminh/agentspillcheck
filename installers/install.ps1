# AgentSpillCheck native installer for Windows 11 (PLAN 16.2).
# Downloads the matching release binary, verifies SHA-256, installs
# under %LOCALAPPDATA%\AgentSpillCheck\bin and adds it to the user PATH.
$ErrorActionPreference = "Stop"
$repo = "Hungleminh/agentspillcheck"
$installDir = Join-Path $env:LOCALAPPDATA "AgentSpillCheck\bin"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "amd64" }
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { $arch = "arm64" }
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$asset = "agentspillcheck_$($version.TrimStart('v'))_windows_$arch.zip"
$base = "https://github.com/$repo/releases/download/$version"
$tmp = Join-Path $env:TEMP "agentspillcheck-install"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
Invoke-WebRequest "$base/$asset" -OutFile (Join-Path $tmp $asset)
Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp "checksums.txt")
$expected = (Select-String -Path (Join-Path $tmp "checksums.txt") -Pattern $asset).Line.Split(" ")[0]
$actual = (Get-FileHash (Join-Path $tmp $asset) -Algorithm SHA256).Hash.ToLower()
if ($actual -ne $expected.ToLower()) { throw "Checksum mismatch / Sai checksum" }
Expand-Archive (Join-Path $tmp $asset) -DestinationPath $tmp -Force
Copy-Item (Join-Path $tmp "spillcheck.exe") (Join-Path $installDir "spillcheck.exe") -Force
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
}
Write-Host "Installed / Da cai: $installDir\spillcheck.exe"
