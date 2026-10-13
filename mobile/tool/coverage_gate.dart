// Coverage gate for `flutter test --coverage` (P0-25): fails when line
// coverage of the hand-written code under lib/ is below the minimum. Generated
// code (gen-l10n output, build_runner / swagger_parser parts, the generated API
// client) is left out.
//
// Usage: dart run tool/coverage_gate.dart [lcov=coverage/lcov.info] [min=80]
import 'dart:convert';
import 'dart:io';

/// Line-coverage floor, in percent (testing.md: >= 80% on business logic).
const defaultMinPercent = 80.0;

/// Generated sources: they are not ours to test.
final generatedSources = [
  RegExp(r'^lib/l10n/app_localizations.*\.dart$'),
  RegExp(r'\.(g|freezed|gr|mocks)\.dart$'),
  RegExp('^lib/api/generated/'),
];

/// Flavor entrypoints only call bootstrap; no unit test loads them.
final entrypoints = [RegExp(r'^lib/main_\w+\.dart$')];

/// Line coverage of the files that survived the exclusions.
class CoverageReport {
  const new({
    required this.files,
    required this.linesFound,
    required this.linesHit,
  });

  /// lib/-relative paths measured, sorted.
  final List<String> files;
  final int linesFound;
  final int linesHit;

  /// 0 when nothing was measured: an empty profile must never pass the gate.
  double get percent => linesFound == 0 ? 0 : linesHit * 100 / linesFound;
}

/// Parses an lcov tracefile and measures it, skipping files matching [exclude].
/// Counts DA records (not LF/LH) so a file listed twice is merged, not double counted.
CoverageReport measureCoverage(String lcov, {required List<RegExp> exclude}) {
  final hitsByFile = <String, Map<int, bool>>{};
  Map<int, bool>? current;
  for (final line in const LineSplitter().convert(lcov)) {
    if (line.startsWith('SF:')) {
      final path = _libRelative(line.substring(3));
      current = _isExcluded(path, exclude)
          ? null
          : hitsByFile.putIfAbsent(path, () => {});
    } else if (line.startsWith('DA:') && current != null) {
      final fields = line.substring(3).split(',');
      final lineNumber = int.parse(fields[0]);
      final hit = int.parse(fields[1]) > 0;
      current[lineNumber] = (current[lineNumber] ?? false) || hit;
    } else if (line == 'end_of_record') {
      current = null;
    }
  }

  final lines = hitsByFile.values.expand((hits) => hits.values);
  return CoverageReport(
    files: hitsByFile.keys.toList()..sort(),
    linesFound: lines.length,
    linesHit: lines.where((hit) => hit).length,
  );
}

/// The [sources] no test loaded (absent from [report]), minus [exclude].
List<String> untestedFiles(
  CoverageReport report, {
  required List<String> sources,
  required List<RegExp> exclude,
}) {
  final measured = report.files.toSet();
  return [
    for (final source in sources.map(_libRelative))
      if (!measured.contains(source) && !_isExcluded(source, exclude)) source,
  ]..sort();
}

String _libRelative(String path) {
  final normalized = path.replaceAll(r'\', '/');
  final index = normalized.lastIndexOf('/lib/');
  if (index >= 0) return normalized.substring(index + 1);
  return normalized;
}

/// Header markers of generated Dart (build_runner, swagger_parser, gen-l10n
/// variants) and of files a generator opted out of coverage.
final _generatedHeader = RegExp(
  '^// (GENERATED CODE - DO NOT MODIFY BY HAND|coverage:ignore-file)',
  multiLine: true,
);

/// Whether [content] is a generated source, whatever its path.
bool isGeneratedSource(String content) => _generatedHeader.hasMatch(content);

bool _isExcluded(String path, List<RegExp> exclude) =>
    exclude.any((pattern) => pattern.hasMatch(path));

void main(List<String> args) {
  final lcovPath = args.isNotEmpty ? args[0] : 'coverage/lcov.info';
  final minPercent = args.length > 1
      ? double.parse(args[1])
      : defaultMinPercent;

  final lcovFile = File(lcovPath);
  if (!lcovFile.existsSync()) {
    stderr.writeln(
      'coverage gate: $lcovPath not found (run flutter test --coverage)',
    );
    exit(2);
  }

  final dartFiles = Directory('lib')
      .listSync(recursive: true)
      .whereType<File>()
      .where((file) => file.path.endsWith('.dart'))
      .toList();
  final sources = [for (final file in dartFiles) file.path];
  final generatedByHeader = [
    for (final file in dartFiles)
      if (isGeneratedSource(file.readAsStringSync()))
        RegExp('^${RegExp.escape(_libRelative(file.path))}\$'),
  ];
  final exclude = [...generatedSources, ...entrypoints, ...generatedByHeader];
  final report = measureCoverage(lcovFile.readAsStringSync(), exclude: exclude);
  for (final path in untestedFiles(
    report,
    sources: sources,
    exclude: exclude,
  )) {
    stdout.writeln('coverage gate: warning: no test loads $path');
  }

  final percent = report.percent.toStringAsFixed(2);
  final summary =
      '$percent% (${report.linesHit}/${report.linesFound} lines, '
      '${report.files.length} files; min $minPercent%)';
  if (report.percent < minPercent) {
    stderr.writeln('coverage gate: FAIL $summary');
    exit(1);
  }
  stdout.writeln('coverage gate: OK $summary');
}
