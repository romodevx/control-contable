import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// iOS shows its own UI (text selection menu, permission prompts, share
/// sheet) in the languages the bundle declares. Without `es` here those stay
/// in English even though the Flutter UI is es-MX.
const _infoPlistPath = 'ios/Runner/Info.plist';
final _declaresSpanish = RegExp(
  r'<key>CFBundleLocalizations</key>\s*<array>\s*<string>es</string>',
);

void main() {
  test('the iOS bundle declares Spanish as a supported localization', () {
    final plist = File(_infoPlistPath).readAsStringSync();

    expect(plist, matches(_declaresSpanish));
  });
}
