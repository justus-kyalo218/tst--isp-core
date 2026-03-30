$ngrok = Join-Path (Get-Location) "tools/ngrok/ngrok.exe"
if (!(Test-Path $ngrok)) {
  Write-Host "ngrok not found at $ngrok"
  exit 1
}

& $ngrok http 8080
