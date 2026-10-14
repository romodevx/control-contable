import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// M1-12 rule (mobile/README.md, "Session tokens"): session tokens live only
/// in `flutter_secure_storage` (Keystore / Keychain), never in the
/// `shared_preferences` plugin, which writes plain XML on Android. The
/// backup exclusions (android_backup_test.dart) keep the encrypted store on
/// the device; they do not make a plaintext store safe.
final _sharedPreferencesImport = RegExp(
  r'''import\s+['"]package:shared_preferences/''',
);
final _mentionsToken = RegExp('token', caseSensitive: false);
final _secureStorageImport = RegExp(
  r'''import\s+['"]package:flutter_secure_storage/''',
);

/// The one adapter allowed to touch the platform's secure storage.
const _secureStoreAdapter = 'lib/core/session/refresh_token_store.dart';

void main() {
  test('no source that uses shared_preferences handles a token', () {
    final offenders = [
      for (final file in Directory('lib').listSync(recursive: true))
        if (file is File && file.path.endsWith('.dart'))
          if (_usesPrefsForTokens(file.readAsStringSync())) file.path,
    ];

    expect(offenders, isEmpty);
  });

  test('one adapter owns flutter_secure_storage (one key, one policy)', () {
    final users = [
      for (final file in Directory('lib').listSync(recursive: true))
        if (file is File && file.path.endsWith('.dart'))
          if (_secureStorageImport.hasMatch(file.readAsStringSync()))
            file.uri.path,
    ];

    expect(users, [_secureStoreAdapter]);
  });
}

bool _usesPrefsForTokens(String source) =>
    _sharedPreferencesImport.hasMatch(source) &&
    _mentionsToken.hasMatch(source);
