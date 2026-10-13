#!/usr/bin/env bash
# Runs once, on an empty data directory (docker-entrypoint-initdb.d).
# Passes the role passwords to psql as variables so they never appear in a file.
set -euo pipefail

psql --no-psqlrc -v ON_ERROR_STOP=1 \
  --username "$POSTGRES_USER" \
  --dbname "$POSTGRES_DB" \
  -v db_name="$CC_DB_NAME" \
  -v cc_owner_password="$CC_OWNER_PASSWORD" \
  -v cc_app_password="$CC_APP_PASSWORD" \
  -v cc_webhook_password="$CC_WEBHOOK_PASSWORD" \
  -v cc_worker_password="$CC_WORKER_PASSWORD" \
  -f /bootstrap/roles.sql
