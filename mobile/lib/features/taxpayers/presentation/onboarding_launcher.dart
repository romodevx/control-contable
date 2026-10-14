import 'dart:async';

import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/taxpayers_routes.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Opens the onboarding by itself, once, when the signed-in user has no
/// taxpayer at all (US-T1). Leaving the onboarding returns here and it does
/// not reopen. A user whose taxpayers are all archived is left on home,
/// where the switcher can unarchive one or register another.
///
/// Renders nothing. Place it on home behind
/// `FlagGate(module: AppModule.taxpayers, …)`.
class OnboardingLauncher extends ConsumerStatefulWidget {
  /// Creates the launcher.
  const new({super.key});

  @override
  ConsumerState<OnboardingLauncher> createState() => _OnboardingLauncherState();
}

class _OnboardingLauncherState extends ConsumerState<OnboardingLauncher> {
  bool _launched = false;

  @override
  void initState() {
    super.initState();
    ref.listenManual(taxpayersProvider, _launchWhenNone, fireImmediately: true);
  }

  void _launchWhenNone(
    AsyncValue<List<Taxpayer>>? previous,
    AsyncValue<List<Taxpayer>> next,
  ) {
    if (_launched || next.value?.isEmpty != true) return;
    _launched = true;
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) unawaited(context.push<void>(TaxpayersRoutes.onboarding));
    });
  }

  @override
  Widget build(BuildContext context) => const SizedBox.shrink();
}
