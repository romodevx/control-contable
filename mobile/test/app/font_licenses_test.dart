import 'package:control_contable/app/font_licenses.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_test/flutter_test.dart';

const _oflTitle = 'SIL Open Font License, Version 1.1';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  tearDown(LicenseRegistry.reset);

  test('every bundled face ships its OFL notice in the app licenses', () async {
    registerFontLicenses();

    final entries = await LicenseRegistry.licenses.toList();
    String noticeFor(String family) => entries
        .where((entry) => entry.packages.contains(family))
        .expand((entry) => entry.paragraphs)
        .map((paragraph) => paragraph.text)
        .join('\n');

    expect(bundledFontFamilies, containsAll(['Public Sans', 'Source Serif 4']));
    for (final family in bundledFontFamilies) {
      expect(noticeFor(family), contains(_oflTitle), reason: family);
    }
  });
}
