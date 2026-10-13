import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// The manifest every build type merges. The INTERNET permission in the
/// debug and profile manifests only covers `flutter run`; a release build
/// without it in main cannot open a socket to the API.
const _mainManifestPath = 'android/app/src/main/AndroidManifest.xml';
const _internetPermission =
    '<uses-permission android:name="android.permission.INTERNET"';

void main() {
  test('the main Android manifest grants INTERNET to release builds', () {
    final manifest = File(_mainManifestPath).readAsStringSync();

    expect(manifest, contains(_internetPermission));
  });
}
