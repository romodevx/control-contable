import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// P0-31 F-03 (ahead of M1-12): the refresh token must never leave the
/// device in a backup. Auto Backup (cloud) and device-to-device transfer are
/// separate channels on Android 12+, and `allowBackup="false"` alone does
/// not stop the second one, so each channel gets its own rules. Encrypted
/// shared preferences restored without their Keystore key are also
/// unreadable, which breaks the restored app.
const _mainManifestPath = 'android/app/src/main/AndroidManifest.xml';
const _xmlDir = 'android/app/src/main/res/xml';

/// `<exclude domain="sharedpref" path="."/>`, whitespace-tolerant.
final _excludesSharedPrefs = RegExp(
  r'<exclude\s+domain="sharedpref"\s+path="\."\s*/>',
);

String _read(String path) => File(path).readAsStringSync();

/// The body of `<tag>…</tag>` in [xml], or `null` when it is missing.
String? _section(String xml, String tag) =>
    RegExp('<$tag[^>]*>(.*?)</$tag>', dotAll: true).firstMatch(xml)?.group(1);

void main() {
  final manifest = _read(_mainManifestPath);

  test('the app opts out of Auto Backup', () {
    expect(manifest, contains('android:allowBackup="false"'));
  });

  test('Android 12+ rules exclude the token store from both channels', () {
    expect(
      manifest,
      contains('android:dataExtractionRules="@xml/data_extraction_rules"'),
    );
    final rules = _read('$_xmlDir/data_extraction_rules.xml');

    for (final channel in ['cloud-backup', 'device-transfer']) {
      final body = _section(rules, channel);
      expect(body, isNotNull, reason: '<$channel> must be declared');
      expect(body, matches(_excludesSharedPrefs), reason: channel);
    }
  });

  test('Android 11 and lower rules exclude the token store', () {
    expect(manifest, contains('android:fullBackupContent="@xml/backup_rules"'));
    final rules = _read('$_xmlDir/backup_rules.xml');

    expect(
      _section(rules, 'full-backup-content'),
      matches(_excludesSharedPrefs),
    );
  });
}
