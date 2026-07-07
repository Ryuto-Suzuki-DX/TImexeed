#!/usr/bin/env bash

set -Eeuo pipefail

# ==================================================
# Timexeed PostgreSQL backup
# ==================================================

CONTAINER_NAME="timexeed-db"
DB_USER="timexeed"
DB_NAME="timexeed"

AWS_REGION="ap-northeast-3"
S3_BUCKET="timexeed-db-backup-891340029892-ap-northeast-3-an"
S3_PREFIX="database-backups"

BACKUP_DIR="/opt/timexeed/backups"
TIMESTAMP="$(TZ=Asia/Tokyo date '+%Y%m%d_%H%M%S')"
BACKUP_FILE="timexeed_${TIMESTAMP}.sql.gz"
LOCAL_PATH="${BACKUP_DIR}/${BACKUP_FILE}"
S3_PATH="s3://${S3_BUCKET}/${S3_PREFIX}/${BACKUP_FILE}"

cleanup() {
  rm -f "${LOCAL_PATH}"
}

trap cleanup EXIT

mkdir -p "${BACKUP_DIR}"

echo "[INFO] PostgreSQL backup started: ${TIMESTAMP}"

if ! docker inspect "${CONTAINER_NAME}" >/dev/null 2>&1; then
  echo "[ERROR] PostgreSQL container not found: ${CONTAINER_NAME}" >&2
  exit 1
fi

if [ "$(docker inspect -f '{{.State.Running}}' "${CONTAINER_NAME}")" != "true" ]; then
  echo "[ERROR] PostgreSQL container is not running: ${CONTAINER_NAME}" >&2
  exit 1
fi

docker exec "${CONTAINER_NAME}" \
  pg_dump \
  --username="${DB_USER}" \
  --dbname="${DB_NAME}" \
  --no-owner \
  --no-privileges \
  | gzip -9 > "${LOCAL_PATH}"

if [ ! -s "${LOCAL_PATH}" ]; then
  echo "[ERROR] Backup file is empty." >&2
  exit 1
fi

gzip -t "${LOCAL_PATH}"

aws s3 cp \
  "${LOCAL_PATH}" \
  "${S3_PATH}" \
  --region "${AWS_REGION}" \
  --sse AES256 \
  --only-show-errors

echo "[SUCCESS] Backup uploaded: ${S3_PATH}"
