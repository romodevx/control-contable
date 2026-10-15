import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// Push (M9-09) before and after the Firebase project exists (OB9).
///
/// - FCM never makes a token at launch: auto-init is off on both platforms
///   and the app turns it on only when the user opts in.
/// - The Firebase config files are never committed: the repository is
///   public (GitGuardian), CI writes them from secrets per flavor. Until
///   they exist the Gradle plugin is not applied, so every build works.
const _androidManifest = 'android/app/src/main/AndroidManifest.xml';
const _iosInfoPlist = 'ios/Runner/Info.plist';
const _appGradle = 'android/app/build.gradle.kts';
const _gitignore = '.gitignore';

void main() {
  test('Android: FCM auto-init and Analytics collection are off', () {
    final manifest = File(_androidManifest).readAsStringSync();

    expect(
      manifest,
      matches(
        RegExp(
          r'android:name="firebase_messaging_auto_init_enabled"\s*'
          'android:value="false"',
        ),
      ),
    );
    expect(
      manifest,
      matches(
        RegExp(
          r'android:name="firebase_analytics_collection_enabled"\s*'
          'android:value="false"',
        ),
      ),
    );
  });

  test('iOS: FCM auto-init is off', () {
    final plist = File(_iosInfoPlist).readAsStringSync();

    expect(
      plist,
      matches(
        RegExp(r'<key>FirebaseMessagingAutoInitEnabled</key>\s*<false/>'),
      ),
    );
  });

  test('the Firebase config files are git-ignored', () {
    final ignored = File(_gitignore).readAsLinesSync();

    expect(ignored, contains('**/google-services.json'));
    expect(ignored, contains('**/GoogleService-Info.plist'));
  });

  test('no Firebase config file is in the tree', () {
    final found = [
      for (final root in ['android', 'ios', 'lib'])
        for (final entity in Directory(root).listSync(recursive: true))
          if (entity is File &&
              (entity.path.endsWith('google-services.json') ||
                  entity.path.endsWith('GoogleService-Info.plist') ||
                  entity.path.endsWith('firebase_options.dart')))
            entity.path,
    ];

    expect(found, isEmpty);
  });

  test('Gradle applies google-services only when a config file exists', () {
    final gradle = File(_appGradle).readAsStringSync();

    expect(gradle, contains('google-services.json'));
    expect(
      gradle,
      contains('apply(plugin = "com.google.gms.google-services")'),
    );
  });
}
