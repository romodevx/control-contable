import 'package:control_contable/api/generated/models/regime_incompatibility.dart';
import 'package:control_contable/features/taxpayers/domain/regime_compatibility.dart';
import 'package:flutter_test/flutter_test.dart';

import 'taxpayers_fakes.dart';

void main() {
  const reason = 'RESICO PF sustituye al 612.';
  final resico = regime(
    '626',
    incompatibleWith: [
      const RegimeIncompatibility(code: '612', reason: reason),
    ],
  );
  final business = regime(
    '612',
    incompatibleWith: [
      const RegimeIncompatibility(code: '626', reason: reason),
    ],
  );
  final rentals = regime('606');

  test('compatible regimes have no conflict', () {
    expect(regimeConflicts([resico, rentals]), isEmpty);
    expect(regimeConflicts([]), isEmpty);
  });

  test('the later of an incompatible pair carries the conflict', () {
    final conflicts = regimeConflicts([business, rentals, resico]);

    expect(conflicts.keys, ['626']);
    expect(conflicts['626']!.code, '612');
    expect(conflicts['626']!.reason, reason);
  });

  test('a pair declared on one side only is still a conflict', () {
    final oneSided = regime('612');

    final conflicts = regimeConflicts([oneSided, resico]);

    expect(conflicts['626']!.code, '612');
  });
}
