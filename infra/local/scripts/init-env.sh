#!/usr/bin/env bash
# Writes the LOCAL stack's .env from .env.example, replacing every *_PASSWORD placeholder
# with a fresh random value (openssl, 24 bytes as hex) and every CC_AUTH_* key with 32
# random bytes as hex, mode 0600. Refuses to overwrite.
# Usage: init-env.sh <example> <target>
set -euo pipefail

example="${1:?usage: init-env.sh <example> <target>}"
target="${2:?usage: init-env.sh <example> <target>}"

if [ -e "$target" ]; then
  echo "init-env: $target already exists; not overwriting" >&2
  exit 1
fi

umask 077
tmp="$(mktemp "${target}.XXXXXX")"
trap 'rm -f "$tmp"' EXIT

while IFS= read -r line || [ -n "$line" ]; do
  case "$line" in
    [A-Z]*_PASSWORD=*) printf '%s=%s\n' "${line%%=*}" "$(openssl rand -hex 24)" ;;
    CC_AUTH_*=*) printf '%s=%s\n' "${line%%=*}" "$(openssl rand -hex 32)" ;;
    *) printf '%s\n' "$line" ;;
  esac
done <"$example" >"$tmp"

mv "$tmp" "$target"
trap - EXIT
