import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/auth/auth_routes.dart';
import 'package:control_contable/features/expenses/expenses_routes.dart';
import 'package:control_contable/features/history/application/history_view_marker.dart';
import 'package:control_contable/features/history/history_routes.dart';
import 'package:control_contable/features/incomes/incomes_routes.dart';
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
];

/// How the slices talk without importing each other, wired at the app
/// root: an estimate shown tells History, which marks the month viewed
/// (M8, only while `module_history` is on).
final List<Override> appFeatureWiring = [
  estimateRenderedListenerProvider.overrideWith(
    (ref) => HistoryViewMarker(ref).onEstimateRendered,
  ),
];
