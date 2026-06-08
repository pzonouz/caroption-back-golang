#!/bin/bash

TMP_DIR="/tmp"

DATE=$(date +%F)
ARCHIVE="$TMP_DIR/uploads-${DATE}.tar.gz"
tar -cvf $ARCHIVE ./uploads





