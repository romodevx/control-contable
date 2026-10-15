import 'dart:async';

import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/reminders/application/push_messaging.dart';
import 'package:control_contable/features/reminders/application/reminder_push_sync.dart';
import 'package:control_contable/features/reminders/domain/reminder_deep_link.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Wires push into the running app, above the router (composition root):
/// registers the device after each sign-in while `module_reminders` is on,
/// re-registers on token rotation, and follows a tapped notification's
/// deep link (the estimate of its period; Home when the engine is off or
/// the link is unknown). Renders [child] unchanged.
class ReminderPushListener extends ConsumerStatefulWidget {
  /// Listens for push events around [child].
  const new({required this.child, super.key});

  /// The app.
  final Widget child;

  @override
  ConsumerState<ReminderPushListener> createState() =>
      _ReminderPushListenerState();
}

class _ReminderPushListenerState extends ConsumerState<ReminderPushListener> {
  final List<StreamSubscription<Object>> _subscriptions = [];
  late final Session _session;
  bool _active = false;

  bool get _shouldBeActive =>
      _session.isSignedIn &&
      ref.read(moduleEnabledProvider(AppModule.reminders));

  @override
  void initState() {
    super.initState();
    final messaging = ref.read(pushMessagingProvider);
    final sync = ref.read(reminderPushSyncProvider);
    _session = ref.read(sessionProvider)..addListener(_onContextChanged);
    ref.listenManual(
      moduleEnabledProvider(AppModule.reminders),
      (_, _) => _onContextChanged(),
    );
    _subscriptions
      ..add(messaging.openedLinks.listen(_open))
      ..add(
        messaging.tokenRefreshes.listen((token) {
          if (_active) unawaited(sync.onTokenRefreshed(token));
        }),
      );
    unawaited(
      messaging.initialLink().then((link) => link == null ? null : _open(link)),
    );
    _onContextChanged();
  }

  /// Registers on each transition into "signed in with reminders on".
  void _onContextChanged() {
    final active = _shouldBeActive;
    if (active && !_active) {
      unawaited(ref.read(reminderPushSyncProvider).afterSignIn());
    }
    _active = active;
  }

  Future<void> _open(String link) async {
    final enabled = await ref.read(flagsProvider.future);
    if (!mounted) return;
    final router = ref.read(appRouterProvider);
    final period = estimatePeriodOfLink(link);
    if (period != null && enabled.contains(AppModule.taxEngine)) {
      unawaited(router.push<void>(TaxEngineRoutes.estimateOf(period)));
    } else {
      router.go(AppRoutes.home);
    }
  }

  @override
  void dispose() {
    _session.removeListener(_onContextChanged);
    for (final subscription in _subscriptions) {
      unawaited(subscription.cancel());
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => widget.child;
}
