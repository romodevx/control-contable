import 'dart:async';

import 'package:control_contable/api/generated/models/tax_inputs.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/tax_engine/application/breakdown.dart';
import 'package:control_contable/features/tax_engine/application/estimate.dart';
import 'package:control_contable/features/tax_engine/application/tax_inputs.dart';
import 'package:control_contable/features/tax_engine/application/tax_inputs_form_model.dart';
import 'package:control_contable/features/tax_engine/data/tax_engine_repository.dart';
import 'package:control_contable/features/tax_engine/presentation/monthly_payments_section.dart';
import 'package:control_contable/features/tax_engine/presentation/tax_input_fields.dart';
import 'package:control_contable/features/tax_engine/presentation/tax_inputs_feedback.dart';
import 'package:control_contable/features/tax_engine/presentation/tax_inputs_header.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/features/taxpayers/presentation/option_fields.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Where the field opened at lands in the viewport (0 top, 1 bottom): high,
/// clear of the keyboard and the sticky footer.
const double _focusAlignment = 0.2;

/// The tax-inputs form of one regime and fiscal year (M5-27), a page of its
/// own (more than six fields for 612 and 601). Opens at the input the
/// estimate found missing; saves every value at once (PUT, idempotent),
/// then goes back to a refreshed estimate. A read-only account sees its
/// values; saving shows the paywall and keeps what was typed.
class TaxInputsFormPage extends ConsumerStatefulWidget {
  /// The form for [query] over the stored [inputs], opened at [focusKey].
  const new({
    required this.query,
    required this.inputs,
    this.focusKey,
    super.key,
  });

  /// Whose inputs: taxpayer, regime and fiscal year.
  final TaxInputsQuery query;

  /// What the server holds.
  final TaxInputs inputs;

  /// The input to open the form at (the estimate's first missing key).
  final String? focusKey;

  @override
  ConsumerState<TaxInputsFormPage> createState() => _TaxInputsFormPageState();
}

class _TaxInputsFormPageState extends ConsumerState<TaxInputsFormPage> {
  late final TaxInputsFormModel _model;
  bool _isModelReady = false;
  int? _openGroup;
  String? _notice;
  bool _isLeaving = false;

  AppLocalizations get _l10n => AppLocalizations.of(context);

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (_isModelReady) return;
    _isModelReady = true;
    final year = widget.inputs.fiscalYear;
    _model = TaxInputsFormModel(
      widget.inputs,
      current: currentPeriod(ref.read(clockProvider)()),
      monthLabel: (month) =>
          _l10n.taxInputsMonthLabel(monthName(context, year, month)),
      checkRate: (value) => checkRate(value, _l10n, min: rateMin, max: rateMax),
    );
    _openGroup = _model.groupOf(_model.focusTarget(widget.focusKey)) ?? 0;
    WidgetsBinding.instance.addPostFrameCallback((_) => _focusTarget());
  }

  void _focusTarget() {
    final focus = _model.focusTarget(widget.focusKey)?.focusNode;
    if (!mounted || focus == null) return;
    focus.requestFocus();
    if (focus.context case final target?) {
      Scrollable.ensureVisible(target, alignment: _focusAlignment);
    }
  }

  Future<void> _save() async {
    final plan = _model.toUpdate();
    setState(() => _notice = null);
    final outcome = await saveTaxInputs(
      ref.read(taxEngineRepositoryProvider),
      widget.query,
      plan.values,
    );
    if (!mounted) return;
    switch (outcome) {
      case TaxInputsSaved():
        _afterSave();
      case TaxInputsNeedPlan():
        ref.read(readOnlyAccessProvider.notifier).markReadOnly();
        showTaxInputsPaywall(context);
      case TaxInputsRejected():
        _showVerdict(outcome, plan);
      case TaxInputsNotSaved(:final notice):
        setState(() => _notice = taxInputsNoticeText(notice, _l10n));
    }
  }

  void _showVerdict(TaxInputsRejected rejection, TaxInputsUpdatePlan plan) {
    final verdict = _model.showVerdict(
      rejection.details,
      plan,
      describe: (detail) => describeTaxInputDetail(detail, _l10n),
    );
    setState(() {
      _openGroup = verdict.group ?? _openGroup;
      if (verdict.hasUnmatched) _notice = _l10n.taxInputsSaveFailed;
    });
  }

  /// Every estimate and breakdown on the stack reads again with the new
  /// inputs; then back to where the form was opened from.
  void _afterSave() {
    ref
      ..invalidate(estimateProvider)
      ..invalidate(breakdownProvider);
    showTaxInputsSavedToast(context);
    _leave();
  }

  void _leave() {
    if (_isLeaving || !mounted) return;
    _isLeaving = true;
    leaveTaxInputs(context);
  }

  Future<void> _requestLeave() async {
    if (_model.form.isDirty && !await confirmDiscardTaxInputs(context)) return;
    _leave();
  }

  void _toggleGroup(int group) =>
      setState(() => _openGroup = _openGroup == group ? null : group);

  @override
  void dispose() {
    _model.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => TaxInputsLeaveGuard(
    onLeave: _requestLeave,
    child: Scaffold(
      appBar: AppBar(title: Text(_l10n.taxInputsTitle)),
      body: ListenableBuilder(
        listenable: _model.form,
        builder: (context, _) => AppForm(
          controller: _model.form,
          onSave: _save,
          onCancel: () => unawaited(_requestLeave()),
          children: [
            TaxInputsHeader(inputs: widget.inputs),
            if (_notice case final notice?) FormNotice(message: notice),
            for (final entry in _model.entries)
              ?taxInputField(
                context,
                entry,
                fiscalYear: widget.inputs.fiscalYear,
                openGroup: _openGroup,
                onToggleGroup: _toggleGroup,
              ),
          ],
        ),
      ),
    ),
  );
}
