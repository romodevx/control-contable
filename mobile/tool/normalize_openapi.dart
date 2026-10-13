// Usage: dart run tool/normalize_openapi.dart <openapi.yaml> <out.json>
//
// Writes the swagger_parser input for `make mobile-client` (ADR-0003).
import 'dart:convert';
import 'dart:io';

import 'package:yaml/yaml.dart';

import 'src/openapi_normalizer.dart';

const _usageExitCode = 64;

void main(List<String> args) {
  if (args.length != 2) {
    stderr.writeln('usage: normalize_openapi.dart <openapi.yaml> <out.json>');
    exit(_usageExitCode);
  }
  final spec = loadYaml(File(args[0]).readAsStringSync());
  final normalized = normalizeOpenApi(spec as Map<Object?, Object?>);
  File(args[1])
    ..createSync(recursive: true)
    ..writeAsStringSync(const JsonEncoder.withIndent('  ').convert(normalized));
}
