import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/pump_app.dart';
import 'taxpayers_fakes.dart';

/// The client id every onboarding test creates with.
const String testClientId = '0192f0c4-0000-4000-8000-0000000000aa';

/// A taxpayer the user already manages, so home does not open the
/// onboarding by itself (that happens only with none at all).
final Taxpayer existingTaxpayer = taxpayer(
  id: '0192f0c4-7b1e-7c3a-9d2e-00000000000d',
  displayName: 'Existente',
  createdAt: DateTime.utc(2026, 1, 2),
);

/// What a taxpayers widget test scripts and observes.
class OnboardingHarness {
  final FakeTaxpayersRepository repository = FakeTaxpayersRepository();

  /// The device's active-taxpayer id.
  final InMemoryActiveTaxpayerStore store = InMemoryActiveTaxpayerStore();

  /// The running app's providers.
  late ProviderContainer container;

  /// Boots the app with `module_taxpayers` [enabled] (plus [alsoEnabled])
  /// and opens [location] (the onboarding by default) on top of home. The
  /// server holds [taxpayers] (one existing taxpayer by default).
  Future<void> pump(
    WidgetTester tester, {
    bool enabled = true,
    Set<AppModule> alsoEnabled = const {},
    String? location = TaxpayersRoutes.onboarding,
    Brightness brightness = Brightness.light,
    List<Taxpayer>? taxpayers,
  }) async {
    repository.taxpayers.addAll(taxpayers ?? [existingTaxpayer]);
    await pumpControlContableApp(
      tester,
      brightness: brightness,
      enabledModules: {if (enabled) AppModule.taxpayers, ...alsoEnabled},
      overrides: [
        taxpayersRepositoryProvider.overrideWithValue(repository),
        clientIdProvider.overrideWithValue(() => testClientId),
        clockProvider.overrideWithValue(() => DateTime.utc(2026, 10, 13, 18)),
        activeTaxpayerStoreProvider.overrideWithValue(store),
      ],
    );
    container = ProviderScope.containerOf(
      tester.element(find.byType(MaterialApp)),
    );
    if (location != null) {
      unawaited(container.read(appRouterProvider).push<void>(location));
      await tester.pumpAndSettle();
    }
  }

  /// The location on top (a pushed one included).
  String get location => container.read(appRouterProvider).state.uri.toString();
}

/// Taps the widget showing [text], scrolling it into view first.
Future<void> tapText(WidgetTester tester, String text) async {
  final target = find.text(text).last;
  await tester.ensureVisible(target);
  await tester.pumpAndSettle();
  await tester.tap(target);
  await tester.pumpAndSettle();
}

/// The text input under the visible [label] of an `AppTextField`.
Finder inputLabeled(String label) => find.descendant(
  of: find.ancestor(of: find.text(label), matching: find.byType(Column)).first,
  matching: find.byType(TextField),
);

/// Types [text] into the field labeled [label] and leaves it.
Future<void> typeInto(WidgetTester tester, String label, String text) async {
  final field = inputLabeled(label);
  await tester.ensureVisible(field);
  await tester.enterText(field, text);
  await tester.pumpAndSettle();
}

/// The person type and RFC steps done for a persona física, landing on
/// the regimes.
Future<void> passIdentity(
  WidgetTester tester, {
  String card = 'Persona física',
  String rfc = 'GOMA800101AB6',
}) async {
  await tapText(tester, card);
  await typeInto(tester, 'Nombre para identificarlo', 'Ana');
  await typeInto(tester, 'RFC', rfc);
  await tapText(tester, 'Siguiente');
}
