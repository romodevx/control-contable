#!/usr/bin/env bash
# Test for init-env.sh (P0-30, PR #5 follow-up 5): the first `make db-up` writes
# infra/local/.env with fresh random passwords instead of the well-known placeholders.
# Usage: infra/local/scripts/init-env_test.sh
# The checks are single-quoted on purpose: check() evals them after the files exist.
# shellcheck disable=SC2016
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
init="$script_dir/init-env.sh"
example="$script_dir/../.env.example"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

failures=0
check() {
  if ! eval "$1"; then
    echo "FAIL: $2"
    failures=$((failures + 1))
  fi
}

"$init" "$example" "$work/a.env"
"$init" "$example" "$work/b.env"

check '! grep -q "change-me" "$work/a.env"' "placeholders replaced"
check '[ "$(grep -c "_PASSWORD=" "$work/a.env")" = "$(grep -c "_PASSWORD=" "$example")" ]' "every password line kept"
check '! grep -qE "_PASSWORD=$" "$work/a.env"' "no empty password"
check 'grep -qE "^CC_APP_PASSWORD=[0-9a-f]{48}$" "$work/a.env"' "48 hex chars from openssl"
check '[ "$(grep _PASSWORD= "$work/a.env")" != "$(grep _PASSWORD= "$work/b.env")" ]' "two runs differ"
check '[ "$(grep -c "^CC_APP_PASSWORD=" "$work/a.env")" = 1 ]' "one line per variable"
check 'diff <(grep -v "_PASSWORD=" "$example") <(grep -v "_PASSWORD=" "$work/a.env") >/dev/null' "other lines unchanged"
check '[ "$(ls -l "$work/a.env" | cut -c1-10)" = "-rw-------" ]' "mode 0600"
check '! "$init" "$example" "$work/a.env" 2>/dev/null' "never overwrites an existing file"

if [ "$failures" -gt 0 ]; then
  echo "init-env_test: $failures failure(s)"
  exit 1
fi
echo "init-env_test: ok"
