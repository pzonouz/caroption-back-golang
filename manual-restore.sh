#!/bin/bash
set -e

BACKUP_FILE=$1
TMP_DIR="/tmp/restore-$$"
UPLOADS_DIR="./uploads"
VOLUME="caroption_go"

echo "Creating temp workspace..."
mkdir -p $TMP_DIR

echo "Extracting backup..."
tar -xzf "$BACKUP_FILE" -C "$TMP_DIR"

UPLOADS_ARCHIVE=$(ls $TMP_DIR/uploads-*.tar.gz)
VOLUME_ARCHIVE=$(ls $TMP_DIR/volume-*.tar.gz)

if [ -z "$UPLOADS_ARCHIVE" ] || [ -z "$VOLUME_ARCHIVE" ]; then
  echo "Backup structure invalid"
  exit 1
fi

echo "Stopping containers..."
make stop-postgres

echo "Restoring uploads..."
rm -rf $UPLOADS_DIR/*
tar -xzf "$UPLOADS_ARCHIVE" -C "$UPLOADS_DIR"

echo "Restoring docker volume..."
docker run --rm \
  -v $VOLUME:/volume \
  -v $TMP_DIR:/backup \
  alpine \
  sh -c "rm -rf /volume/* && tar -xzf /backup/$(basename $VOLUME_ARCHIVE) -C /volume"

echo "Starting containers..."
make start-postgres

echo "Cleaning temp files..."
rm -rf $TMP_DIR

echo " Restore completed"

