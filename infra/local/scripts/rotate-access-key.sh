#!/usr/bin/env bash
# Rotates the LOCAL access-token signing key (M1-07) in infra/local/.env, mode 0600:
#   * the current CC_AUTH_ACCESS_TOKEN_KEY / CC_ACCESS_TOKEN_KEY_ID become the previous key
#     (CC_AUTH_ACCESS_TOKEN_PREVIOUS_KEY / CC_ACCESS_TOKEN_PREVIOUS_KEY_ID): it keeps verifying
#     the tokens it signed until they expire (15 min), and never signs again;
#   * a new random Ed25519 seed (32 bytes, hex) becomes the active key under a new kid.
# A key retired by an earlier rotation is dropped. `make api` passes the four values as AUTH_*.
# Shared environments rotate through the secret manager, never with this script.
# Usage: rotate-access-key.sh <env-file>
set -euo pipefail

env_file="${1:?usage: rotate-access-key.sh <env-file>}"
[ -f "$env_file" ] || { echo "rotate-access-key: $env_file does not exist (run make init-env)" >&2; exit 1; }

value_of() { # last assignment of $1 in the env file, or empty
  awk -F= -v k="$1" '$1 == k { v = substr($0, length(k) + 2) } END { print v }' "$env_file"
}

old_key="$(value_of CC_AUTH_ACCESS_TOKEN_KEY)"
old_id="$(value_of CC_ACCESS_TOKEN_KEY_ID)"
old_id="${old_id:-k1}"
case "$old_key" in
  [0-9a-f]*) ;;
  *) echo "rotate-access-key: CC_AUTH_ACCESS_TOKEN_KEY is not a hex key" >&2; exit 1 ;;
esac
new_key="$(openssl rand -hex 32)"
new_id="k$(date -u +%Y%m%d%H%M%S)"

umask 077
tmp="$(mktemp "${env_file}.XXXXXX")"
trap 'rm -f "$tmp"' EXIT
grep -v -E '^(CC_AUTH_ACCESS_TOKEN_KEY|CC_ACCESS_TOKEN_KEY_ID|CC_AUTH_ACCESS_TOKEN_PREVIOUS_KEY|CC_ACCESS_TOKEN_PREVIOUS_KEY_ID)=' \
  "$env_file" >"$tmp" || true
{
  printf 'CC_AUTH_ACCESS_TOKEN_KEY=%s\n' "$new_key"
  printf 'CC_ACCESS_TOKEN_KEY_ID=%s\n' "$new_id"
  printf 'CC_AUTH_ACCESS_TOKEN_PREVIOUS_KEY=%s\n' "$old_key"
  printf 'CC_ACCESS_TOKEN_PREVIOUS_KEY_ID=%s\n' "$old_id"
} >>"$tmp"
mv "$tmp" "$env_file"
trap - EXIT
echo "rotated the access-token key: active kid $new_id, previous kid $old_id (restart make api)"
