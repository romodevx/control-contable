import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/application/add_regime_flow.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_settings.dart';
import 'package:control_contable/features/taxpayers/presentation/read_only_notice.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_options_step.dart';
import 'package:control_contable/features/taxpayers/presentation/regimes_step.dart';
import 'package:control_contable/features/taxpayers/presentation/settings/settings_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

/// Adds a regime to the taxpayer [taxpayerId] (wizard steps 3–4 of the
/// onboarding, one regime). Back goes one step at a time; once added it
/// returns to the settings and says so (a CPA-review warning stays until
/// dismissed).
class AddRegimeScreen extends ConsumerWidget {
  /// The screen for [taxpayerId].
  const new({required this.taxpayerId, super.key});

  /// The taxpayer's id (from the location).
  final String taxpayerId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final taxpayer = ref
        .watch(taxpayersProvider)
        .value
        ?.where((t) => t.id == taxpayerId)
        .firstOrNull;
    if (taxpayer == null) {
      return Scaffold(
        appBar: AppBar(title: Text(l10n.addRegimeTitle)),
        body: Padding(
          padding: const EdgeInsets.all(AppSpacing.gutter),
          child: LoadStatus(message: l10n.settingsNotFound),
        ),
      );
    }
    return _AddRegime(taxpayer: taxpayer);
  }
}

class _AddRegime extends ConsumerStatefulWidget {
  const new({required this.taxpayer});

  final Taxpayer taxpayer;

  @override
  ConsumerState<_AddRegime> createState() => _AddRegimeState();
}

class _AddRegimeState extends ConsumerState<_AddRegime> {
  late final AddRegimeFlow _flow = AddRegimeFlow(
    settings: ref.read(taxpayerSettingsProvider),
    taxpayerId: widget.taxpayer.id,
    personType: widget.taxpayer.personType,
    heldCodes: {
      for (final regime in widget.taxpayer.regimes)
        if (regime.endPeriod == null) regime.regimeCode,
    },
    knownReadOnly: ref.read(taxpayerSettingsProvider).knownReadOnly,
  )..addListener(_handleStepChanged);

  bool _isLeaving = false;

  void _handleStepChanged() {
    final added = _flow.added;
    if (_flow.step != AddRegimeStep.done || added == null) return;
    final l10n = AppLocalizations.of(context);
    final reasons = [
      for (final warning in added.warnings ?? const <Warning>[])
        if (warning.code == cpaReviewWarning) warning.message,
    ];
    showSettingsToast(
      ScaffoldMessenger.of(context),
      reasons.isEmpty
          ? l10n.addRegimeAdded(added.regimeCode)
          : l10n.addRegimeCpaReview(added.regimeCode, reasons.join(' ')),
      action: reasons.isEmpty
          ? null
          : SnackBarAction(label: l10n.paywallClose, onPressed: () {}),
    );
    _leave();
  }

  void _back() {
    if (!_flow.back()) _leave();
  }

  void _leave() {
    if (_isLeaving || !mounted) return;
    _isLeaving = true;
    context.pop();
  }

  @override
  void dispose() {
    _flow.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, _) {
        if (!didPop) _back();
      },
      child: Scaffold(
        appBar: AppBar(
          title: Text(l10n.addRegimeTitle),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            tooltip: MaterialLocalizations.of(context).backButtonTooltip,
            onPressed: _back,
          ),
        ),
        body: ListenableBuilder(
          listenable: _flow,
          builder: (context, _) => KeyedSubtree(
            key: ValueKey(_flow.step),
            child: switch (_flow.step) {
              AddRegimeStep.regimes => RegimesStep(flow: _flow),
              AddRegimeStep.options => RegimeOptionsStep(flow: _flow),
              AddRegimeStep.readOnly => ReadOnlyNotice(
                onLeave: _leave,
                body: l10n.paywallWriteBody,
                actionLabel: l10n.paywallClose,
              ),
              AddRegimeStep.done => const SizedBox.shrink(),
            },
          ),
        ),
      ),
    );
  }
}
