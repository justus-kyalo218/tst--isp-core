$goBin = "C:\\Program Files\\Go\\bin"
$root = Split-Path -Parent $PSScriptRoot

if (!(Get-Command go -ErrorAction SilentlyContinue)) {
  if (Test-Path (Join-Path $goBin "go.exe")) {
    $env:Path = "$goBin;$env:Path"
  }
}

if (!(Get-Command go -ErrorAction SilentlyContinue)) {
  Write-Host "Go not found. Install Go or ensure it is on PATH."
  exit 1
}

$backendCmd = "Set-Location `"$root\\backend`"; go run ./cmd/api"
$frontendCmd = "Set-Location `"$root\\frontend`"; npm run dev"

Start-Process powershell -ArgumentList "-NoExit", "-Command", $backendCmd
Start-Process powershell -ArgumentList "-NoExit", "-Command", $frontendCmd
