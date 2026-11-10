#!/usr/bin/env bash
# Nightly Postgres backup. Run from cron on the EC2 box (see deploy/README.md).
# Writes ~/backups/foundry-YYYY-MM-DD.dump and keeps the last 7 days.
set -euo pipefail

APP_DIR="${APP_DIR:-$HOME/foundry}"
BACKUP_DIR="${BACKUP_DIR:-$HOME/backups}"
KEEP_DAYS="${KEEP_DAYS:-7}"

mkdir -p "$BACKUP_DIR"
cd "$APP_DIR"

out="$BACKUP_DIR/foundry-$(date +%F).dump"
# -T: no TTY (cron has none). -Fc: compressed custom format, restore with pg_restore.
# Write to a temp file first so a failed dump never replaces a good one.
docker compose exec -T db sh -c 'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' > "$out.tmp"
mv "$out.tmp" "$out"

find "$BACKUP_DIR" -name 'foundry-*.dump' -mtime +"$KEEP_DAYS" -delete
echo "$(date -Is) backup ok: $out ($(du -h "$out" | cut -f1))"
