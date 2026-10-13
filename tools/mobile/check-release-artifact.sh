#!/usr/bin/env bash
# Release-artifact gate (P0-25, frontend-build.md): fails when an APK/AAB carries symbols,
# debug info, source maps or sourceMappingURL comments in what reaches a device.
#
#   check-release-artifact.sh <app.apk|app.aab>...
#
# Every entry is checked except an AAB's BUNDLE-METADATA/: Play keeps that directory for
# crash symbolication and never delivers it to devices. `flutter build appbundle` refuses
# to finish unless AGP moved libapp.so/libflutter.so symbol tables there (`.sym`, i.e.
# debugSymbolLevel SYMBOL_TABLE); since M1-12 also libdartjni.so's (package:jni, pulled in
# by flutter_secure_storage 11 on Android). Only those exact files, R8's proguard.map and AGP's fixed
# build metadata are allowed, and only inside an .aab; full DWARF (`.dbg`) or anything else
# fails. An ELF the reader cannot parse fails too: the gate fails closed.
#
# Needs unzip, grep and an ELF reader (READELF, default `readelf`; llvm-readelf works).
set -euo pipefail

READELF="${READELF:-readelf}"

# Names that only debug builds or symbol uploads should ever contain.
readonly FORBIDDEN_NAMES='\.(map|symbols|sym|dbg|debug|dwo|dSYM)(/|$)|(^|/)mapping\.txt$|(^|/)debug-info/'
# ELF sections that leak symbol names or source-level debug info.
readonly FORBIDDEN_SECTIONS='^\.(symtab|strtab|debug_[a-z_]+|zdebug_[a-z_]+|gnu_debugdata|gnu_debuglink)$'
# Play-only metadata a Flutter AAB is expected to carry (see header).
readonly BUNDLE_METADATA_ALLOWED='^BUNDLE-METADATA/(com\.android\.tools\.build\.debugsymbols/(arm64-v8a|armeabi-v7a|x86_64)/lib(app|flutter|dartjni)\.so\.sym|com\.android\.tools\.build\.obfuscation/proguard\.map|com\.android\.tools\.build\.gradle/app-metadata\.properties|com\.android\.tools\.build\.libraries/dependencies\.pb|com\.android\.tools\.build\.profiles/baseline\.profm?|com\.android\.tools/r8\.json)$'

problems=0
report() {
  echo "  FAIL: $*" >&2
  problems=$((problems + 1))
}

# elf_sections <file>: one section name per line; fails when the reader cannot parse the file.
elf_sections() {
  local table
  table="$("$READELF" -S -W "$1")" || return 1
  sed -n 's/^ *\[ *[0-9]*\] \([^ ]*\).*/\1/p' <<<"$table"
}

check_artifact() {
  local artifact="$1"
  echo "== $artifact"
  [[ -f "$artifact" ]] || { report "$artifact does not exist"; return; }

  local entries delivered
  local -a play_only=()
  entries="$(unzip -Z1 "$artifact")"
  if [[ "$artifact" == *.aab ]]; then
    play_only=(-x 'BUNDLE-METADATA/*')
    local metadata
    metadata="$(grep '^BUNDLE-METADATA/' <<<"$entries" | grep -v '/$' || true)"
    while IFS= read -r entry; do
      [[ -z "$entry" ]] && continue
      if ! grep -Eq "$BUNDLE_METADATA_ALLOWED" <<<"$entry"; then
        report "unexpected Play metadata: $entry"
      fi
    done <<<"$metadata"
    delivered="$(grep -v '^BUNDLE-METADATA/' <<<"$entries" || true)"
  else
    delivered="$entries"
  fi

  local named
  named="$(grep -E "$FORBIDDEN_NAMES" <<<"$delivered" || true)"
  [[ -z "$named" ]] || while IFS= read -r entry; do report "symbol/map file shipped: $entry"; done <<<"$named"

  local unpacked
  unpacked="$(mktemp -d)"
  # The artifact is ours, but unzip still writes only below the fresh directory.
  unzip -q -o "$artifact" ${play_only[@]+"${play_only[@]}"} -d "$unpacked"

  local urls
  urls="$(grep -rlaF 'sourceMappingURL' "$unpacked" || true)"
  [[ -z "$urls" ]] || while IFS= read -r file; do report "sourceMappingURL in ${file#"$unpacked"/}"; done <<<"$urls"

  local libs=0 so section sections
  while IFS= read -r so; do
    [[ -z "$so" ]] && continue
    if [[ "$(basename "$so")" == libapp.so ]]; then
      libs=$((libs + 1))
    fi
    if ! sections="$(elf_sections "$so" 2>/dev/null)"; then
      report "${so#"$unpacked"/} is not a readable ELF (or $READELF is unusable): not inspected"
      continue
    fi
    while IFS= read -r section; do
      [[ -z "$section" ]] && continue
      if grep -Eq "$FORBIDDEN_SECTIONS" <<<"$section"; then
        report "${so#"$unpacked"/} has section $section"
      fi
    done <<<"$sections"
  done < <(find "$unpacked" -type f -name '*.so')
  ((libs > 0)) || report "no libapp.so inside: not a Flutter release build, nothing inspected"
  echo "  inspected $libs libapp.so and $(grep -c . <<<"$delivered") entries"

  rm -rf "$unpacked"
}

command -v "$READELF" >/dev/null || { echo "ELF reader '$READELF' not found (set READELF)" >&2; exit 2; }
if (($# == 0)); then
  echo "usage: $0 <app.apk|app.aab>..." >&2
  exit 2
fi
for artifact in "$@"; do
  check_artifact "$artifact"
done
if ((problems > 0)); then
  echo "release-artifact gate: $problems problem(s)" >&2
  exit 1
fi
echo "release-artifact gate: OK"
