#!/usr/bin/env bash
# Tests for check-release-artifact.sh: each fixture is an APK/AAB-shaped zip that the gate
# must accept or reject. Needs zip, an ELF-producing C compiler and strip:
#   Linux CI: the defaults (cc, strip, readelf).
#   macOS:    CC="<ndk>/clang --target=aarch64-linux-android24" STRIP=<ndk>/llvm-strip
#             READELF=<ndk>/llvm-readelf
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
gate="$here/check-release-artifact.sh"
CC="${CC:-cc}"
STRIP="${STRIP:-strip}"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

# Two shared objects: one with a symbol table (what an unstripped libapp.so looks like),
# one stripped (what the obfuscated release must ship).
printf 'int dart_entry(void) { return 42; }\n' >"$work/app.c"
# shellcheck disable=SC2086 # CC may carry flags (--target=...)
$CC -shared -fPIC -g -o "$work/libapp-unstripped.so" "$work/app.c"
cp "$work/libapp-unstripped.so" "$work/libapp-stripped.so"
"$STRIP" --strip-all "$work/libapp-stripped.so"

# make_zip <name> <entry=source>...: builds $work/<name> from the given entries.
make_zip() {
  local name="$1"
  shift
  local root="$work/root-$name"
  mkdir -p "$root"
  local spec entry source
  for spec in "$@"; do
    entry="${spec%%=*}"
    source="${spec#*=}"
    mkdir -p "$root/$(dirname "$entry")"
    cp "$source" "$root/$entry"
  done
  (cd "$root" && zip -qr "$work/$name" .)
}

printf 'console.log(1);\n//# sourceMappingURL=main.js.map\n' >"$work/with-url.js"
printf '{"version":3}\n' >"$work/map.json"
printf 'plain asset\n' >"$work/asset.txt"

failures=0
# expect <pass|fail> <description> <artifact>
expect() {
  local want="$1" description="$2" artifact="$3" got
  if "$gate" "$work/$artifact" >"$work/out.log" 2>&1; then got=pass; else got=fail; fi
  if [[ "$got" == "$want" ]]; then
    echo "ok   - $description"
  else
    echo "FAIL - $description (want $want, got $got)"
    sed 's/^/       /' "$work/out.log"
    failures=$((failures + 1))
  fi
}

lib="lib/arm64-v8a/libapp.so"
make_zip clean.apk "$lib=$work/libapp-stripped.so" "assets/a.txt=$work/asset.txt"
make_zip symtab.apk "$lib=$work/libapp-unstripped.so"
make_zip map.apk "$lib=$work/libapp-stripped.so" "assets/main.js.map=$work/map.json"
make_zip symbols.apk "$lib=$work/libapp-stripped.so" "assets/app.android-arm64.symbols=$work/asset.txt"
make_zip url.apk "$lib=$work/libapp-stripped.so" "assets/main.js=$work/with-url.js"
make_zip nolib.apk "assets/a.txt=$work/asset.txt"
make_zip clean.aab "base/$lib=$work/libapp-stripped.so" \
  "BUNDLE-METADATA/com.android.tools.build.debugsymbols/arm64-v8a/libapp.so.sym=$work/libapp-unstripped.so" \
  "BUNDLE-METADATA/com.android.tools.build.obfuscation/proguard.map=$work/asset.txt"
make_zip symtab.aab "base/$lib=$work/libapp-unstripped.so"
make_zip unknown-metadata.aab "base/$lib=$work/libapp-stripped.so" \
  "BUNDLE-METADATA/com.example/app.android-arm64.symbols=$work/asset.txt"

expect pass "a stripped APK with plain assets passes" clean.apk
expect fail "an APK whose libapp.so keeps .symtab/.debug_* fails" symtab.apk
expect fail "an APK carrying a .map file fails" map.apk
expect fail "an APK carrying split-debug-info .symbols fails" symbols.apk
expect fail "an APK with a sourceMappingURL comment fails" url.apk
expect fail "an APK without libapp.so fails (nothing was inspected)" nolib.apk
expect pass "an AAB with stripped libs and Play-only BUNDLE-METADATA passes" clean.aab
expect fail "an AAB whose delivered libapp.so keeps symbols fails" symtab.aab
expect fail "an AAB with unknown BUNDLE-METADATA content fails" unknown-metadata.aab

if ((failures > 0)); then
  echo "$failures test(s) failed" >&2
  exit 1
fi
echo "all release-artifact gate tests passed"
