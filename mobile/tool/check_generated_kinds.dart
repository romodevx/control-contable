// Usage: dart run tool/check_generated_kinds.dart <openapi.json> <generated dir>
//
// Fails (exit 1) when a generated Dart type's kind differs from its component
// schema, or two schemas map to one Dart name (ADR-0003 condition 3).
import 'dart:convert';
import 'dart:io';

import 'src/generated_kinds.dart';

const _usageExitCode = 64;
const _findingsExitCode = 1;

void main(List<String> args) {
  if (args.length != 2) {
    stderr.writeln(
      'usage: check_generated_kinds.dart <openapi.json> <generated dir>',
    );
    exit(_usageExitCode);
  }
  final spec = jsonDecode(File(args[0]).readAsStringSync());
  final root = Directory(args[1]);
  final sources = {
    for (final file in root.listSync(recursive: true).whereType<File>())
      if (file.path.endsWith('.dart'))
        file.path.substring(root.path.length + 1): file.readAsStringSync(),
  };
  final findings = checkGeneratedKinds(
    spec: spec as Map<String, Object?>,
    sources: sources,
  );
  if (findings.isEmpty) {
    stdout.writeln('generated kinds: ${sources.length} files, all match');
    return;
  }
  stderr
    ..writeln('generated kinds: ${findings.length} problem(s) (ADR-0003)')
    ..writeAll(findings.map((f) => '  - $f'), '\n')
    ..writeln();
  exit(_findingsExitCode);
}
