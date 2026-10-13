#!/usr/bin/env bash
# Table test for assert-local.sh (P0-30). Runs without Docker: DOCKER_HOST is set to a
# local socket path, so only the env-file and DATABASE_URL checks are exercised.
# Usage: infra/local/scripts/assert-local_test.sh
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
guard="$script_dir/assert-local.sh"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

env_file="$work/.env"
printf 'CC_DB_HOST=127.0.0.1\n' >"$env_file"

failures=0

# expect <accept|refuse> <DATABASE_URL> <description>
expect() {
  local want="$1" url="$2" desc="$3" got
  if DOCKER_HOST="unix:///var/run/docker.sock" DATABASE_URL="$url" "$guard" "$env_file" >/dev/null 2>&1; then
    got=accept
  else
    got=refuse
  fi
  if [ "$got" != "$want" ]; then
    echo "FAIL: $desc: want $want, got $got (DATABASE_URL=$url)"
    failures=$((failures + 1))
  fi
}

expect accept "" "unset DATABASE_URL"
expect accept "postgres://u:p@127.0.0.1:5434/db" "IPv4 loopback with port"
expect accept "postgres://u:p@localhost/db" "localhost without port"
expect accept "postgres://u:p@[::1]:5434/db" "IPv6 loopback"
expect accept "postgresql://u:p@localhost:5434/db?sslmode=disable" "postgresql scheme with query"
expect accept "postgres://localhost:5434/db" "no userinfo"
expect accept "postgres://u:p%40ss@127.0.0.1:5434/db" "escaped @ in password"
expect accept "postgres://u:p@ss@127.0.0.1:5434/db" "raw @ in password: host is after the last @"

expect refuse "postgres://u:p@prod.example.com:5432/db" "remote host"
# PR #5 follow-up 1: the old glob matched these because "@localhost/" appears later.
expect refuse "postgres://u:p@prod.example.com/x?a=@localhost/" "loopback only in the query string"
expect refuse "postgres://u:p@prod.example.com/x#@127.0.0.1:1" "loopback only in the fragment"
expect refuse "postgres://u:p@prod.example.com/@localhost:5434/db" "loopback only in the path"
expect refuse "postgres://u:p@localhost.evil.example:5434/db" "loopback as a host prefix"
expect refuse "postgres://u:p@127.0.0.1.evil.example/db" "loopback IP as a host prefix"
expect refuse "host=prod.example.com user=u dbname=db" "keyword/value DSN"
expect refuse "not a url" "garbage"

if [ "$failures" -gt 0 ]; then
  echo "assert-local_test: $failures failure(s)"
  exit 1
fi
echo "assert-local_test: ok"
