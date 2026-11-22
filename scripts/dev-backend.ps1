# Start the Go backend for local mini-program joint debugging.
#   1) docker compose up -d mysql
#   2) powershell -ExecutionPolicy Bypass -File scripts\dev-backend.ps1
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

if (-not (Test-Path ".env")) {
  throw "missing .env at repo root (DASHSCOPE_API_KEY etc.)"
}
Get-Content ".env" | ForEach-Object {
  $line = $_.Trim()
  if ($line -and -not $line.StartsWith("#") -and $line.Contains("=")) {
    $name, $value = $line.Split("=", 2)
    Set-Item -Path "Env:$($name.Trim())" -Value $value.Trim()
  }
}

if (-not $env:DB_HOST) { $env:DB_HOST = "127.0.0.1" }
if (-not $env:DB_PORT) { $env:DB_PORT = "3306" }
if (-not $env:DB_USER) { $env:DB_USER = "travelgd" }
if (-not $env:DB_PASSWORD) { $env:DB_PASSWORD = "travelgd" }
if (-not $env:DB_NAME) { $env:DB_NAME = "travel_gd" }
if (-not $env:JWT_SECRET) { $env:JWT_SECRET = "local-dev-secret-please-change-0123456789" }
if (-not $env:PORT) { $env:PORT = "8081" }
if (-not $env:UPLOAD_DIR) { $env:UPLOAD_DIR = (Join-Path (Get-Location) ".local-uploads") }
New-Item -ItemType Directory -Force -Path $env:UPLOAD_DIR | Out-Null

Write-Host "backend: http://127.0.0.1:$($env:PORT)  (uploads: $($env:UPLOAD_DIR))"
Set-Location server-go
go run .
