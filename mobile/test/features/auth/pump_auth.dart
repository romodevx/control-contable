import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/links/url_opener.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/auth/application/session_store.dart';
import 'package:control_contable/features/auth/data/auth_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import 'auth_fakes.dart';

/// What an auth widget test scripts and observes.
class AuthHarness {
  final FakeAuthRepository repository = FakeAuthRepository();
  final RecordingSessionStore sessions = RecordingSessionStore();
  final FakeClock clock = FakeClock();
  final List<Uri> opened = [];
  bool canOpen = true;

  /// The running app's providers.
  late ProviderContainer container;

  /// Boots the app with `module_auth` [enabled] and opens [location].
  Future<void> pump(
    WidgetTester tester, {
    String? location,
    bool enabled = true,
    Brightness brightness = Brightness.light,
  }) async {
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: {if (enabled) AppModule.auth},
      overrides: [
        authRepositoryProvider.overrideWithValue(repository),
        sessionStoreProvider.overrideWithValue(sessions),
        clockProvider.overrideWithValue(clock.call),
        urlOpenerProvider.overrideWithValue((uri) async {
          opened.add(uri);
          return canOpen;
        }),
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
    if (location != null) {
      container.read(appRouterProvider).go(location);
      await tester.pumpAndSettle();
    }
  }

  /// Moves the clock and the test's fake time together.
  Future<void> elapse(WidgetTester tester, Duration by) async {
    clock.advance(by);
    await tester.pump(by);
  }
}

/// The text input under the visible [label] of an `AppTextField`.
Finder inputLabeled(String label) => find.descendant(
  of: find.ancestor(of: find.text(label), matching: find.byType(Column)).first,
  matching: find.byType(TextField),
);
