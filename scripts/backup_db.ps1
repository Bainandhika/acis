# ==============================================================================
# ACIS - Automated PostgreSQL Database Backup Script (PowerShell for Windows)
# Usage: .\scripts\backup_db.ps1
# ==============================================================================
param (
    [string]$BackupDir = "./backups",
    [int]$RetentionDays = 14
)

$ErrorActionPreference = "Stop"
$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"

if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir | Out-Null
}

$BackupFile = Join-Path $BackupDir "acis_backup_$Timestamp.sql"
Write-Host "🔄 Starting ACIS Database Backup..." -ForegroundColor Cyan

if ($env:DATABASE_URL) {
    & pg_dump "$env:DATABASE_URL" -f $BackupFile
} else {
    $dbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
    $dbPort = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
    $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "acis_user" }
    $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "acis_db" }
    if ($env:DB_PASSWORD) { $env:PGPASSWORD = $env:DB_PASSWORD }

    & pg_dump -h $dbHost -p $dbPort -U $dbUser -d $dbName -f $BackupFile
}

Write-Host "✅ Backup created successfully: $BackupFile" -ForegroundColor Green

# Clean up older backups
$CutoffDate = (Get-Date).AddDays(-$RetentionDays)
Get-ChildItem -Path $BackupDir -Filter "acis_backup_*.sql" | Where-Object { $_.LastWriteTime -lt $CutoffDate } | Remove-Item -Force

Write-Host "✨ Backup job finished." -ForegroundColor Cyan
