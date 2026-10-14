import 'dart:async';

import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/flags/flags_provider.dart';
import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/taxpayers/application/active_taxpayer.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:control_contable/features/taxpayers/domain/taxpayer_cap.dart';
import 'package:control_contable/features/taxpayers/presentation/identity_step.dart';
import 'package:control_contable/features/taxpayers/presentation/person_type_step.dart';
import 'package:control_contable/features/taxpayers/presentation/read_only_notice.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_options_step.dart';
import 'package:control_contable/features/taxpayers/presentation/regimes_step.dart';
import 'package:control_contable/features/taxpayers/presentation/review_step.dart';
import 'package:control_contable/features/taxpayers/presentation/stop_notice.dart';
import 'package:control_contable/features/taxpayers/presentation/tax_inputs_step.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// How long a step takes to replace the previous one. Onboarding is seen
/// rarely, so a short, quiet transition is worth it; none with reduced
/// motion.
const Duration _stepTransition = Duration(milliseconds: 220);

/// How far (a fraction of its width) a step slides in from.
const double _stepSlide = 0.04;

/// The onboarding of a new taxpayer (US-T1, US-T1b): one screen per step,
/// a progress indicator, back (app bar or system) one step at a time with
/// every value kept. Leaves once the taxpayer exists (or from step 1).
class OnboardingScreen extends ConsumerStatefulWidget {
  /// Creates the screen.
  const new({super.key});

  @override
  ConsumerState<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends ConsumerState<OnboardingScreen> {
  late final OnboardingFlow _flow = OnboardingFlow(
    repository: ref.read(taxpayersRepositoryProvider),
    newId: ref.read(clientIdProvider),
    clock: ref.read(clockProvider),
    taxEngineEnabled: false,
    knownReadOnly: ref.read(readOnlyAccessProvider),
    // An unloaded list says nothing: the server's `422` still decides.
    capReached: isAtTaxpayerCap(ref.read(taxpayersProvider).value?.length ?? 0),
  )..addListener(_handleStepChanged);

  bool _isLeaving = false;

  void _handleStepChanged() {
    if (_flow.step == OnboardingStep.readOnly) {
      ref.read(readOnlyAccessProvider.notifier).markReadOnly();
    }
    if (_flow.step == OnboardingStep.done) _leave();
  }

  void _back() {
    if (!_flow.back()) _leave();
  }

  void _leave() {
    if (_isLeaving || !mounted) return;
    _isLeaving = true;
    // A created taxpayer becomes the active one (US-T1), also when the user
    // backs out of the review.
    if (_flow.created case final created?) {
      unawaited(ref.read(activeTaxpayerProvider.notifier).adopt(created.id));
    }
    if (context.canPop()) {
      context.pop();
    } else {
      context.go(AppRoutes.home);
    }
  }

  @override
  void dispose() {
    _flow.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    _flow.taxEngineEnabled = ref.useFlag(AppModule.taxEngine);
    final reduceMotion = MediaQuery.disableAnimationsOf(context);
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) _back();
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(AppLocalizations.of(context).onboardingTitle),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            tooltip: MaterialLocalizations.of(context).backButtonTooltip,
            onPressed: _back,
          ),
        ),
        body: ListenableBuilder(
          listenable: _flow,
          builder: (context, _) => AnimatedSwitcher(
            duration: reduceMotion ? Duration.zero : _stepTransition,
            switchInCurve: Curves.easeOutCubic,
            switchOutCurve: Curves.easeOutCubic,
            transitionBuilder: _fadeAndSlide,
            child: KeyedSubtree(
              key: ValueKey((_flow.step, _flow.page)),
              child: _stepView(),
            ),
          ),
        ),
      ),
    );
  }

  Widget _stepView() => switch (_flow.step) {
    OnboardingStep.personType => PersonTypeStep(flow: _flow),
    OnboardingStep.identity => IdentityStep(flow: _flow),
    OnboardingStep.regimes => RegimesStep(flow: _flow),
    OnboardingStep.options => RegimeOptionsStep(flow: _flow),
    OnboardingStep.review => ReviewStep(flow: _flow),
    OnboardingStep.taxInputs => TaxInputsStep(flow: _flow),
    OnboardingStep.readOnly => ReadOnlyNotice(onLeave: _leave),
    OnboardingStep.limitReached => _LimitReachedNotice(onLeave: _leave),
    OnboardingStep.done => const SizedBox.shrink(),
  };
}

/// The cap of taxpayers is reached (OD-2): said before step 1 (QA M2-12
/// D1), with the way to free a slot.
class _LimitReachedNotice extends StatelessWidget {
  const new({required this.onLeave});

  final VoidCallback onLeave;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return StopNotice(
      icon: Icons.group_outlined,
      title: l10n.onboardingLimitTitle,
      body: l10n.onboardingLimitBody(taxpayerCap),
      actionLabel: l10n.onboardingPaywallAction,
      onLeave: onLeave,
    );
  }
}

Widget _fadeAndSlide(Widget child, Animation<double> animation) =>
    FadeTransition(
      opacity: animation,
      child: SlideTransition(
        position: Tween(
          begin: const Offset(_stepSlide, 0),
          end: Offset.zero,
        ).animate(animation),
        child: child,
      ),
    );
