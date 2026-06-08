#!/bin/bash

VOLUME="caroption_go"
UPLOADS_DIR="./uploads"
TMP_DIR="/tmp"
DATE=$(date +%F)

UPLOADS_ARCHIVE="$TMP_DIR/uploads-$DATE.tar.gz"
VOLUME_ARCHIVE="$TMP_DIR/volume-$DATE.tar.gz"
FINAL_ARCHIVE="$TMP_DIR/backup-$DATE.tar.gz"

echo "Backing up uploads folder..."
tar -czf "$UPLOADS_ARCHIVE" -C "$UPLOADS_DIR" .

echo "Backing up docker volume..."
docker run --rm \
  -v $VOLUME:/volume \
  -v $TMP_DIR:/backup \
  alpine \
  tar -czf "/backup/volume-$DATE.tar.gz" -C /volume .

echo "Combining both backups..."
tar -czf "$FINAL_ARCHIVE" \
  -C "$TMP_DIR" "uploads-$DATE.tar.gz" \
  -C "$TMP_DIR" "volume-$DATE.tar.gz"

echo "DONE"
echo "Final backup: $FINAL_ARCHIVE"

