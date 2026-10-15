import 'dart:io';

import 'package:flutter_test/flutter_test.dart';

/// iOS flavors (OB6): `flutter run|build ios --flavor <f>` needs a shared
/// Xcode scheme named `<f>` whose actions use the `Debug-<f>`, `Profile-<f>`
/// and `Release-<f>` build configurations. Without them the Flutter tool
/// refuses `--flavor` ("The Xcode project does not define custom schemes"),
/// so `make mobile-run` and `make mobile-build-ipa` cannot run on iOS.
///
/// Each flavor mirrors its Android product flavor (build.gradle.kts): its
/// own bundle id, so dev, staging and prod install side by side, and its
/// own home-screen name.
const _pbxprojPath = 'ios/Runner.xcodeproj/project.pbxproj';
const _schemesDir = 'ios/Runner.xcodeproj/xcshareddata/xcschemes';
const _infoPlistPath = 'ios/Runner/Info.plist';
const _baseBundleId = 'mx.romodevx.controlContable';
const _testsBundleId = '$_baseBundleId.RunnerTests';

class _Flavor {
  const new(this.name, this.bundleId, this.displayName);

  final String name;
  final String bundleId;
  final String displayName;
}

const _flavors = [
  _Flavor('dev', '$_baseBundleId.dev', 'Control Contable Dev'),
  _Flavor('staging', '$_baseBundleId.staging', 'Control Contable Staging'),
  _Flavor('prod', _baseBundleId, 'Control Contable'),
];

/// Build type → the Flutter xcconfig its configurations must be based on.
const _baseXcconfig = {
  'Debug': 'Debug.xcconfig',
  'Profile': 'Release.xcconfig',
  'Release': 'Release.xcconfig',
};

/// One `XCBuildConfiguration` object of the project file.
class _BuildConfiguration {
  const new(this.name, this.body);

  final String name;
  final String body;

  String? setting(String key) =>
      RegExp('\\b$key = "?([^";]*)"?;').firstMatch(body)?.group(1);

  bool get isRunnerTarget =>
      body.contains('INFOPLIST_FILE = Runner/Info.plist;');

  bool get isTestsTarget =>
      setting('PRODUCT_BUNDLE_IDENTIFIER') == _testsBundleId;
}

List<_BuildConfiguration> _buildConfigurations(String pbxproj) {
  final block = RegExp(
    r'/\* [^*]+ \*/ = \{\s*isa = XCBuildConfiguration;(.*?)\n\t\t\};',
    dotAll: true,
  );
  return [
    for (final match in block.allMatches(pbxproj))
      _BuildConfiguration(
        RegExp(r'\bname = "?([^";]+)"?;')
            .firstMatch(match.group(1)!)!
            .group(1)!,
        match.group(1)!,
      ),
  ];
}

/// The `buildConfiguration` of `<action>` in a scheme file.
String? _schemeConfiguration(String scheme, String action) => RegExp(
  '<$action\\b[^>]*?buildConfiguration = "([^"]+)"',
  dotAll: true,
).firstMatch(scheme)?.group(1);

void main() {
  final pbxproj = File(_pbxprojPath).readAsStringSync();
  final configurations = _buildConfigurations(pbxproj);

  for (final flavor in _flavors) {
    group('flavor ${flavor.name}', () {
      final schemeFile = File('$_schemesDir/${flavor.name}.xcscheme');

      test('has a shared scheme', () {
        expect(schemeFile.existsSync(), isTrue, reason: schemeFile.path);
      });

      test('its scheme runs Debug, profiles Profile, archives Release', () {
        final scheme = schemeFile.readAsStringSync();
        final expected = {
          'TestAction': 'Debug-${flavor.name}',
          'LaunchAction': 'Debug-${flavor.name}',
          'ProfileAction': 'Profile-${flavor.name}',
          'AnalyzeAction': 'Debug-${flavor.name}',
          'ArchiveAction': 'Release-${flavor.name}',
        };
        for (final MapEntry(key: action, value: config) in expected.entries) {
          expect(_schemeConfiguration(scheme, action), config, reason: action);
        }
        expect(scheme, contains('BlueprintName = "Runner"'));
      });

      for (final MapEntry(key: buildType, value: xcconfig)
          in _baseXcconfig.entries) {
        final name = '$buildType-${flavor.name}';

        test('$name exists for the project and both targets', () {
          final named = configurations.where((c) => c.name == name).toList();

          expect(named, hasLength(3), reason: name);
          expect(named.where((c) => c.isRunnerTarget), hasLength(1));
          expect(named.where((c) => c.isTestsTarget), hasLength(1));
        });

        test('$name of Runner: bundle id, name, Flutter base', () {
          final runner = configurations.singleWhere(
            (c) => c.name == name && c.isRunnerTarget,
          );

          expect(runner.setting('PRODUCT_BUNDLE_IDENTIFIER'), flavor.bundleId);
          expect(runner.setting('APP_DISPLAY_NAME'), flavor.displayName);
          expect(
            runner.body,
            contains('baseConfigurationReference = '),
            reason:
                'Generated.xcconfig (FLUTTER_ROOT, build name) comes in '
                'through the Flutter xcconfig',
          );
          expect(runner.body, contains('/* $xcconfig */;'));
        });
      }
    });
  }

  test('the home-screen name comes from the flavor', () {
    final plist = File(_infoPlistPath).readAsStringSync();

    expect(
      plist,
      matches(
        RegExp(
          r'<key>CFBundleDisplayName</key>\s*'
          r'<string>\$\(APP_DISPLAY_NAME\)</string>',
        ),
      ),
    );
  });

  test('the flavor-less Runner configurations keep a display name', () {
    final runnerDefaults = configurations.where(
      (c) => c.isRunnerTarget && _baseXcconfig.containsKey(c.name),
    );

    expect(runnerDefaults, hasLength(3));
    for (final config in runnerDefaults) {
      expect(
        config.setting('APP_DISPLAY_NAME'),
        'Control Contable',
        reason: config.name,
      );
    }
  });
}
