import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/estimate_legend.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_flow.dart';
import 'package:control_contable/features/taxpayers/application/regime_catalog.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_list.dart';
import 'package:control_contable/features/taxpayers/presentation/regime_selection_field.dart';
import 'package:control_contable/features/taxpayers/presentation/step_header.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Step 3: the regimes of the chosen person type (`GET /v1/regimes`),
/// grouped by what the app can do with them, each group explained in
/// plain words, with the mandatory legend. Multi-select for a persona
/// física, single-select for a persona moral.
class RegimesStep extends ConsumerWidget {
  /// The step of [flow].
  const new({required this.flow, super.key});

  /// The onboarding.
  final OnboardingFlow flow;

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
      AsyncData(:final value) => _RegimesForm(flow: flow, catalog: value),
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
}

/// The step's header over a load status.
class _Status extends StatelessWidget {
  const new({required this.flow, required this.child});

  final OnboardingFlow flow;
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
  const new({required this.flow, required this.catalog});

  final OnboardingFlow flow;
  final List<Regime> catalog;

  @override
  State<_RegimesForm> createState() => _RegimesFormState();
}

class _RegimesFormState extends State<_RegimesForm> {
  final AppFormController _form = AppFormController();
  RegimeSelectionField? _selectionField;

  OnboardingFlow get _flow => widget.flow;
  RegimeSelectionField get _selection => _selectionField!;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_selectionField != null) return;
    final l10n = AppLocalizations.of(context);
    _selectionField = _form.add(
      RegimeSelectionField(
        label: l10n.onboardingRegimesLabel,
        single: _flow.personType == PersonType.moral,
        missingError: CustomFieldError(l10n.onboardingRegimesRequired),
        conflictError: CustomFieldError(l10n.onboardingRegimesConflict),
        initial: _flow.regimes,
      ),
    );
  }

  @override
  void dispose() {
    _form.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isMoral = _flow.personType == PersonType.moral;
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
              ? l10n.onboardingCreate
              : l10n.onboardingNext,
          children: [
            StepHeader(
              flow: _flow,
              title: l10n.onboardingRegimesTitle,
              intro: isMoral
                  ? l10n.onboardingRegimesIntroMoral
                  : l10n.onboardingRegimesIntroFisica,
            ),
            if (notice != null)
              FormNotice(message: describeNotice(notice, l10n)),
            const EstimateLegend(),
            RegimeList(catalog: widget.catalog, selection: _selection),
          ],
        );
      },
    );
  }
}
