import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/features/taxpayers/domain/rfc.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  group('normalizeRfc', () {
    for (final (input, expected) in [
      (' goma-800101 ab6 ', 'GOMA800101AB6'),
      ('GOMA800101AB6', 'GOMA800101AB6'),
      ('abc-680524-p76', 'ABC680524P76'),
      ('peña800101ab1', 'PEÑA800101AB1'),
      ('\tgoma 800101ab6\n', 'GOMA800101AB6'),
    ]) {
      test('"$input" → $expected', () {
        expect(normalizeRfc(input), expected);
      });
    }
  });

  group('checkRfc', () {
    for (final (rfc, type, expected) in [
      ('GOMA800101AB6', PersonType.fisica, null),
      ('GOMA800101AB1', PersonType.fisica, null),
      ('ABC680524P76', PersonType.moral, null),
      ('PEÑA800101AB1', PersonType.fisica, null),
      ('M&A800101AB1', PersonType.moral, null),
      ('GOMA000229AB1', PersonType.fisica, null),
      ('GOMA800101AB6', PersonType.moral, RfcProblem.length),
      ('ABC680524P76', PersonType.fisica, RfcProblem.length),
      ('GOM4800101AB6', PersonType.fisica, RfcProblem.format),
      ('GOMA80O101AB6', PersonType.fisica, RfcProblem.format),
      ('GOMA800101AB*', PersonType.fisica, RfcProblem.format),
      ('GOMA800101ABB', PersonType.fisica, RfcProblem.format),
      ('ABCD991332XX1', PersonType.fisica, RfcProblem.date),
      ('GOMA010229AB1', PersonType.fisica, RfcProblem.date),
      ('GOMA800431AB1', PersonType.fisica, RfcProblem.date),
      ('GOMA800001AB1', PersonType.fisica, RfcProblem.date),
      ('XAXX010101000', PersonType.fisica, RfcProblem.generic),
      ('XEXX010101000', PersonType.fisica, RfcProblem.generic),
    ]) {
      test('$rfc as ${type.name} → ${expected?.name ?? 'accepted'}', () {
        expect(checkRfc(rfc, type), expected);
      });
    }
  });

  group('rfcLength', () {
    test('13 for a persona física, 12 for a persona moral', () {
      expect(rfcLength(PersonType.fisica), 13);
      expect(rfcLength(PersonType.moral), 12);
    });
  });
}
