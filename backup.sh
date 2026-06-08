#!/bin/bash

VOLUME="caroption_go"
BACKUP_DIR="/Caroption/Database/Postgres-Docker"
TMP_DIR="/tmp"
KEEP=7

DATE=$(date +%F)
ARCHIVE="$TMP_DIR/${VOLUME}-${DATE}.tar.gz"

# Create backup
docker run --rm \
  -v $VOLUME:/volume \
  -v $TMP_DIR:/backup \
  alpine \
  tar -czf /backup/${VOLUME}-${DATE}.tar.gz -C /volume .

# Upload to MEGA
mega-put "$ARCHIVE" "$BACKUP_DIR/" -c

# Delete local temp
rm -f "$ARCHIVE"

# Rotate old backups
FILES=$(mega-ls -l "$BACKUP_DIR" | grep "$VOLUME" | sort | awk '{print $2}')
COUNT=$(echo "$FILES" | wc -l)

if [ "$COUNT" -gt "$KEEP" ]; then
  DEL_COUNT=$((COUNT-KEEP))
  OLD_FILES=$(echo "$FILES" | head -n $DEL_COUNT)
  for f in $OLD_FILES; do
    mega-rm "$BACKUP_DIR/$f"
  done
fi


