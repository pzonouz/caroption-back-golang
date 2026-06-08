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



