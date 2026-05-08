$candidates = @(
  (Join-Path (Get-Location) "ngrok.exe"),
  (Join-Path (Get-Location) "tools/ngrok/ngrok.exe")
)

$ngrok = $candidates | Where-Object { Test-Path $_ } | Select-Object -First 1
if (!$ngrok) {
  Write-Host "ngrok not found. Checked:"
  $candidates | ForEach-Object { Write-Host " - $_" }
  exit 1
}

& $ngrok http 8080
