import 'package:flutter_test/flutter_test.dart';

import '../../tool/coverage_gate.dart';

const _lcov = '''
SF:lib/app/app.dart
DA:1,1
DA:2,1
DA:3,0
DA:4,1
LF:4
LH:3
end_of_record
SF:lib/l10n/app_localizations.dart
DA:1,0
DA:2,0
LF:2
LH:0
end_of_record
SF:lib/api/model.g.dart
DA:1,0
LF:1
LH:0
end_of_record
''';

final _generated = [
  RegExp(r'^lib/l10n/app_localizations.*\.dart$'),
  RegExp(r'\.g\.dart$'),
];

void main() {
  group('measureCoverage', () {
    test('counts DA lines of every file when nothing is excluded', () {
      final report = measureCoverage(_lcov, exclude: const []);

      expect(report.linesFound, 7);
      expect(report.linesHit, 3);
      expect(report.percent, closeTo(42.857, 0.001));
    });

    test('leaves generated files out of the total', () {
      final report = measureCoverage(_lcov, exclude: _generated);

      expect(report.linesFound, 4);
      expect(report.linesHit, 3);
      expect(report.percent, 75);
      expect(report.files, ['lib/app/app.dart']);
    });

    test('normalizes absolute and Windows paths to lib/-relative ones', () {
      const lcov = r'''
SF:/home/runner/work/x/mobile/lib/a.g.dart
DA:1,0
end_of_record
SF:C:\src\mobile\lib\b.dart
DA:1,1
end_of_record
''';

      final report = measureCoverage(lcov, exclude: _generated);

      expect(report.files, ['lib/b.dart']);
      expect(report.percent, 100);
    });

    test('a profile with no measurable lines is 0%, never a pass', () {
      final report = measureCoverage('', exclude: const []);

      expect(report.linesFound, 0);
      expect(report.percent, 0);
    });
  });

  group('untestedFiles', () {
    test('lists lib files no test loaded, minus the excluded ones', () {
      final report = measureCoverage(_lcov, exclude: _generated);

      final missing = untestedFiles(
        report,
        sources: const [
          'lib/app/app.dart',
          'lib/app/bootstrap.dart',
          'lib/main_prod.dart',
          'lib/api/model.g.dart',
        ],
        exclude: [..._generated, RegExp(r'^lib/main_\w+\.dart$')],
      );

      expect(missing, ['lib/app/bootstrap.dart']);
    });
  });

  group('isGeneratedSource', () {
    test('recognises the build_runner and swagger_parser headers', () {
      expect(
        isGeneratedSource('// GENERATED CODE - DO NOT MODIFY BY HAND\n'),
        isTrue,
      );
      expect(isGeneratedSource('// coverage:ignore-file\nclass A {}'), isTrue);
    });

    test('hand-written code is measured', () {
      expect(isGeneratedSource('class Generated {}\n'), isFalse);
    });
  });

  group('parseMinPercent', () {
    test('defaults to 80 when no minimum is given', () {
      expect(parseMinPercent(null), defaultMinPercent);
    });

    test('accepts a percentage between 0 and 100', () {
      expect(parseMinPercent('85.5'), 85.5);
    });

    test(
      'rejects values that would turn the gate off or make it unpassable',
      () {
        for (final bad in ['-1', '101', 'NaN', 'abc', '']) {
          expect(
            () => parseMinPercent(bad),
            throwsFormatException,
            reason: bad,
          );
        }
      },
    );
  });
}
