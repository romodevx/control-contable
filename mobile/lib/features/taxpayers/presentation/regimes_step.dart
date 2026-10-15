import 'dart:async';

import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/application/regime_catalog.dart';
import 'package:control_contable/features/taxpayers/application/regime_wizard.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_list.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_selection_field.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/semantics.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Step 3: the regimes of the chosen person type (`GET /v1/regimes`),
/// grouped by what the app can do with them, each group explained in
/// plain words, with the mandatory legend. Multi-select for a persona
/// física, single-select for a persona moral or when adding one regime
/// (regimes already held are not offered again). The month the regimes
/// start in is asked first, above the list, so it is never missed: this
/// month by default (the server's default), back to January 2026 for a
/// user who starts using the app mid-year (M5-29 QA D2, #152 QA O1).
class RegimesStep extends ConsumerWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The wizard (onboarding, or adding a regime).
  final RegimeWizard flow;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context);
    final personType = flow.personType ?? PersonType.fisica;
    final catalog = ref.watch(regimeCatalogProvider(personType));
    return switch (catalog) {
      AsyncData(:final value) when value.isEmpty => _Status(
        flow: flow,
        child: LoadStatus(message: l10n.onboardingRegimesEmpty),
      ),
      AsyncData(:final value) when _offered(value).isEmpty => _Status(
        flow: flow,
        child: LoadStatus(message: l10n.addRegimeNoneLeft),
      ),
      AsyncData(:final value) => _RegimesForm(
        flow: flow,
        catalog: _offered(value),
        held: [
          for (final regime in value)
            if (flow.heldCodes.contains(regime.code)) regime,
        ],
      ),
      AsyncError() => _Status(
        flow: flow,
        child: LoadFailed(
          message: l10n.onboardingRegimesLoadFailed,
          onRetry: () => ref.invalidate(regimeCatalogProvider(personType)),
        ),
      ),
      _ => _Status(
        flow: flow,
        child: LoadStatus(message: l10n.onboardingRegimesLoading, busy: true),
      ),
    };
  }

  List<Regime> _offered(List<Regime> catalog) => [
    for (final regime in catalog)
      if (!flow.heldCodes.contains(regime.code)) regime,
  ];
}

/// The step's header over a load status.
class _Status extends StatelessWidget {
  const new({required this.flow, required this.child});

  final RegimeWizard flow;
  final Widget child;

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.all(AppSpacing.gutter),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        StepHeader(
          flow: flow,
          title: AppLocalizations.of(context).onboardingRegimesTitle,
        ),
        const SizedBox(height: AppSpacing.xl),
        child,
      ],
    ),
  );
}

class _RegimesForm extends StatefulWidget {
  const new({required this.flow, required this.catalog, required this.held});

  final RegimeWizard flow;
  final List<Regime> catalog;
  final List<Regime> held;

  @override
  State<_RegimesForm> createState() => _RegimesFormState();
}

class _RegimesFormState extends State<_RegimesForm> {
  final AppFormController _form = AppFormController();
  RegimeSelectionField? _selectionField;

  RegimeWizard get _flow => widget.flow;
  RegimeSelectionField get _selection => _selectionField!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_selectionField != null) return;
    final l10n = AppLocalizations.of(context);
    _selectionField = _form.add(
      RegimeSelectionField(
        label: l10n.onboardingRegimesLabel,
        single: _flow.singleRegime,
        held: widget.held,
        missingError: CustomFieldError(l10n.onboardingRegimesRequired),
        conflictError: CustomFieldError(l10n.onboardingRegimesConflict),
        initial: _flow.regimes,
      ),
    )..addListener(_remember);
    _toldConflicts = _selection.conflicts.keys.toSet();
  }

  /// Conflicts already announced, by the later regime's code.
  Set<String> _toldConflicts = const {};

  /// Keeps the pick in the flow as it changes (going back keeps it) and
  /// announces a new incompatible pair: its reason appears in a row the
  /// user did not move to, so a live region alone would stay silent.
  void _remember() {
    _flow.rememberRegimes(_selection.selected);
    _announceNewConflicts();
  }

  void _announceNewConflicts() {
    final conflicts = _selection.conflicts;
    final l10n = AppLocalizations.of(context);
    for (final MapEntry(key: code, value: conflict) in conflicts.entries) {
      if (_toldConflicts.contains(code)) continue;
      unawaited(
        SemanticsService.sendAnnouncement(
          View.of(context),
          l10n.onboardingRegimeConflict(conflict.code, conflict.reason),
          Directionality.of(context),
        ),
      );
    }
    _toldConflicts = conflicts.keys.toSet();
  }

  @override
  void dispose() {
    _selectionField?.removeListener(_remember);
    _form.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return ListenableBuilder(
      listenable: Listenable.merge([_selection, _flow]),
      builder: (context, _) {
        // "Crear" only once the choice is complete and asks for nothing more.
        final selected = _selection.selected;
        final createsNow =
            selected.isNotEmpty &&
            selected.every((regime) => regime.optionDefs.isEmpty);
        final notice = _flow.notice;
        return AppForm(
          controller: _form,
          onSave: () => _flow.submitRegimes(_selection.selected),
          primaryLabel: createsNow
              ? finishLabel(_flow.purpose, l10n)
              : l10n.onboardingNext,
          children: [
            StepHeader(
              flow: _flow,
              title: l10n.onboardingRegimesTitle,
              intro: regimesIntro(_flow, l10n),
            ),
            if (notice != null)
              FormNotice(message: describeNotice(notice, l10n)),
            // The month goes first: below a long catalog it was missed and
            // the current-month default was sent unseen (#152 QA O1).
            _StartPeriodField(flow: _flow),
            const EstimateLegend(),
            RegimeList(catalog: widget.catalog, selection: _selection),
          ],
        );
      },
    );
  }
}

/// The month the chosen regimes start in, picked from a list (never out of
/// range), with an always-visible hint: records and estimates exist only
/// from that month, so a user who already paid taxes earlier picks the
/// month they started. One month for every regime of the step. Re-adding
/// an ended regime asks nothing: it reopens from its original start, said
/// in the field's place and room.
class _StartPeriodField extends StatelessWidget {
  const new({required this.flow});

  final RegimeWizard flow;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final months = MaterialLocalizations.of(context);
    final picker = _picker(context);
    final reopenedStart = flow.reopenedStartPeriod;
    if (reopenedStart == null) return picker;
    // The field sits above the regime list: the note takes its room (the
    // hidden picker keeps it) so the list never jumps under the finger
    // that just picked the ended regime.
    return Stack(
      children: [
        ExcludeFocus(
          child: ExcludeSemantics(
            child: IgnorePointer(child: Opacity(opacity: 0, child: picker)),
          ),
        ),
        Text(
          l10n.addRegimeReopensFrom(
            months.formatMonthYear(periodMonth(reopenedStart)),
          ),
        ),
      ],
    );
  }

  Widget _picker(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final months = MaterialLocalizations.of(context);
    final label = flow.singleRegime
        ? l10n.onboardingStartPeriodLabelSingle
        : l10n.onboardingStartPeriodLabelMany;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // The picker below carries this label for screen readers.
        ExcludeSemantics(
          child: Text(label, style: Theme.of(context).textTheme.titleMedium),
        ),
        const SizedBox(height: AppSpacing.xxs),
        Semantics(
          label: label,
          child: DropdownButtonFormField<String>(
            initialValue: flow.startPeriod,
            isExpanded: true,
            decoration: InputDecoration(
              helperText: l10n.onboardingStartPeriodHint,
              helperMaxLines: _hintLines,
            ),
            items: [
              for (final period in flow.startPeriodChoices)
                DropdownMenuItem(
                  value: period,
                  child: Text(months.formatMonthYear(periodMonth(period))),
                ),
            ],
            onChanged: (period) => flow.chooseStartPeriod(period!),
          ),
        ),
      ],
    );
  }
}

/// Lines the start-month hint may take before it wraps no further.
const int _hintLines = 4;
