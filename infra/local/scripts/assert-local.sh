#!/usr/bin/env bash
# Refuses to continue unless the target database is the LOCAL docker compose stack.
# Usage: assert-local.sh <env-file>
set -euo pipefail

env_file="${1:?usage: assert-local.sh <env-file>}"

fail() {
  echo "REFUSED (not local): $*" >&2
  exit 1
}

[ -f "$env_file" ] || fail "missing $env_file"

# The Docker daemon must be on this machine (a unix socket), not a remote engine.
docker_host="${DOCKER_HOST:-}"
if [ -z "$docker_host" ]; then
  docker_host="$(docker context inspect --format '{{.Endpoints.docker.Host}}' 2>/dev/null || true)"
fi
case "$docker_host" in
  unix://*|npipe://*) ;;
  *) fail "Docker endpoint '$docker_host' is not a local socket" ;;
esac

# The database host the stack advertises must be loopback.
db_host="$(sed -n 's/^CC_DB_HOST=//p' "$env_file" | tail -1)"
case "$db_host" in
  127.0.0.1|localhost|::1) ;;
  *) fail "CC_DB_HOST='$db_host' in $env_file is not loopback" ;;
esac

# url_host prints the host of a postgres:// or postgresql:// URL, or nothing if it is
# not one. The authority ends at the first / ? or #; the userinfo ends at its LAST @
# (a raw @ may sit in a password); an IPv6 host keeps its brackets; the port goes.
# Matching a glob on the whole URL is not enough: "@localhost/" could sit in the query.
url_host() {
  local rest authority host
  case "$1" in
    postgres://*) rest="${1#postgres://}" ;;
    postgresql://*) rest="${1#postgresql://}" ;;
    *) return 0 ;;
  esac
  authority="${rest%%[/?#]*}"
  host="${authority##*@}"
  case "$host" in
    \[*\]*) host="${host%%]*}]" ;;
    *) host="${host%%:*}" ;;
  esac
  printf '%s' "$host"
}

# An exported DATABASE_URL to a remote host is a sign the shell targets a shared env.
if [ -n "${DATABASE_URL:-}" ]; then
  case "$(url_host "$DATABASE_URL")" in
    127.0.0.1|localhost|\[::1\]) ;;
    *) fail "DATABASE_URL points to a non-loopback host; unset it first" ;;
  esac
fi

echo "local target confirmed: docker=$docker_host db_host=$db_host"
