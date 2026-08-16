#!/usr/bin/env bash
# ==============================================================================
# ACIS - Automated PostgreSQL Database Backup Script
# Usage: ./scripts/backup_db.sh
# ==============================================================================
set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/acis_backup_${TIMESTAMP}.sql.gz"
RETENTION_DAYS="${RETENTION_DAYS:-14}"

mkdir -p "${BACKUP_DIR}"

echo "🔄 Starting ACIS Database Backup at $(date)..."

# If DATABASE_URL is set, use it; otherwise fallback to individual env variables
if [ -n "${DATABASE_URL:-}" ]; then
  pg_dump "${DATABASE_URL}" | gzip > "${BACKUP_FILE}"
else
  DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="${DB_PORT:-5432}"
  DB_USER="${DB_USER:-acis_user}"
  DB_NAME="${DB_NAME:-acis_db}"
  export PGPASSWORD="${DB_PASSWORD:-acis_secret_password}"

  pg_dump -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" "${DB_NAME}" | gzip > "${BACKUP_FILE}"
fi

echo "✅ Backup created successfully: ${BACKUP_FILE} ($(du -h "${BACKUP_FILE}" | cut -f1))"

# Clean up backups older than RETENTION_DAYS
echo "🧹 Cleaning up backups older than ${RETENTION_DAYS} days..."
find "${BACKUP_DIR}" -name "acis_backup_*.sql.gz" -mtime +"${RETENTION_DAYS}" -delete || true

echo "✨ Backup job completed."
