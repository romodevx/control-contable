import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/reminders/presentation/reminders_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_reminders`.
abstract final class RemindersRoutes {
  /// The reminder settings and upcoming due dates (M9-10).
  static const settings = '/recordatorios';
}

/// The Reminders slice: unreachable while `module_reminders` is off (the
/// server's `404` is the security boundary).
final FeatureModule remindersFeature = FeatureModule(
  module: AppModule.reminders,
  routes: [
    GoRoute(
      path: RemindersRoutes.settings,
      builder: (context, state) => const RemindersScreen(),
    ),
  ],
);
