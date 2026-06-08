#!/bin/bash

VOLUME="caroption_go"
BACKUP_FILE="$1"   # path to backup .tar.gz
TMP_DIR="/tmp/restore_$VOLUME"

if [ -z "$BACKUP_FILE" ]; then
  echo "Usage: ./restore.sh /path/to/caroption_go-YYYY-MM-DD.tar.gz"
  exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
  echo "Backup file not found: $BACKUP_FILE"
  exit 1
fi

echo "Restoring volume: $VOLUME"
echo "Using backup: $BACKUP_FILE"

# Create temp dir
mkdir -p "$TMP_DIR"

# Extract backup into temp
tar -xzf "$BACKUP_FILE" -C "$TMP_DIR"

# Restore into Docker volume
docker run --rm \
  -v $VOLUME:/volume \
  -v $TMP_DIR:/backup \
  alpine \
  sh -c "rm -rf /volume/* && cp -a /backup/. /volume/"

# Cleanup
rm -rf "$TMP_DIR"

echo "Restore completed successfully."

