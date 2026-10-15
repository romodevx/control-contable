import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/core/session/sign_out_steps.dart';
import 'package:control_contable/features/annual/annual_routes.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/history/application/history_view_marker.dart';
import 'package:control_contable/features/history/history_routes.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
import 'package:control_contable/features/reminders/application/push_device.dart';
import 'package:control_contable/features/reminders/reminders_routes.dart';
import 'package:control_contable/features/tax_engine/application/estimate_rendered.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:flutter_riverpod/misc.dart';

/// Every module slice this build ships, composed at the app root so `core/`
/// never imports a feature. The router lets the user reach only the enabled
/// ones (`featureModulesProvider`).
final List<FeatureModule> appFeatureModules = [
  authFeature,
  taxpayersFeature,
  incomesFeature,
  expensesFeature,
  taxEngineFeature,
  historyFeature,
  remindersFeature,
  annualFeature,
];

/// How the slices talk without importing each other, wired at the app
/// root: an estimate shown tells History, which marks the month viewed
/// (M8, only while `module_history` is on); signing out first removes this
/// device's push registration (M9-09, only while `module_reminders` is on;
/// a no-op when never registered), and a registration that may survive
/// that (a failed delete, or the module OFF) goes with the session on the
/// server (`logout`'s `device_id`, M9 security gate F1).
final List<Override> appFeatureWiring = [
  estimateRenderedListenerProvider.overrideWith(
    (ref) => HistoryViewMarker(ref).onEstimateRendered,
  ),
  signOutStepsProvider.overrideWith(
    (ref) => [
      () async {
        if (!ref.read(moduleEnabledProvider(AppModule.reminders))) return;
        await ref.read(pushDeviceProvider).unregister();
      },
    ],
  ),
  signOutDeviceIdProvider.overrideWith(
    (ref) =>
        () => ref.read(pushDeviceProvider).registeredId(),
  ),
];
