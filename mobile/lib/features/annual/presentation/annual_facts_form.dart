import 'dart:async';

import 'package:control_contable/api/generated/models/annual_inputs.dart';
import 'package:control_contable/core/forms/app_checkbox_field.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/form_leave_guard.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/typed_fields.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/features/annual/application/annual_estimate.dart';
import 'package:control_contable/features/annual/application/annual_facts_form_model.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/annual/presentation/regime_tax_inputs_list.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// A persona física's annual facts (3 fields, a page like the rest of the
/// annual inputs), then the regimes' tax inputs, read-only with links.
class AnnualFactsForm extends ConsumerStatefulWidget {
  /// The form of [facts] for [query].
  const new({required this.query, required this.facts, super.key});

  /// Whose facts and which year.
  final AnnualQuery query;

  /// As the server answered them.
  final AnnualInputs facts;

  @override
  ConsumerState<AnnualFactsForm> createState() => _AnnualFactsFormState();
}

class _AnnualFactsFormState extends ConsumerState<AnnualFactsForm> {
  late final AppLocalizations _l10n = AppLocalizations.of(context);
  late final AnnualFactsFormModel _model = AnnualFactsFormModel(
    facts: widget.facts,
    copy: (
      priorYear: _l10n.annualFactsPriorYear('${widget.query.fiscalYear - 1}'),
      otherExempt: _l10n.annualFactsOtherExempt,
      toldEmployer: _l10n.annualFactsToldEmployer,
    ),
  );
  bool _isLeaving = false;
  String? _notice;

  Future<void> _save() async {
    if (blockedAsReadOnly(context, ref)) return;
    setState(() => _notice = null);
    final outcome = await annualSaveOutcomeOf(
      () => ref
          .read(annualInputWritesProvider)
          .putFacts(widget.query, _model.toUpdate()),
    );
    if (!mounted) return;
    switch (outcome) {
      case RecordSaved():
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(SnackBar(content: Text(_l10n.annualFactsSaved)));
        _leave();
      case RecordNeedsPlan():
        showAnnualPaywall(context);
      case RecordRejected(:final details):
        final notice = showAnnualVerdict(
          _model.form,
          details,
          fieldsByPath: _model.fieldsByPath,
          describe: (detail) => describeAnnualDetail(
            detail,
            _l10n,
            limitReached: _l10n.annualWriteFailed,
          ),
          l10n: _l10n,
        );
        if (notice != null) setState(() => _notice = notice);
      case RecordNotSaved(:final notice):
        setState(() => _notice = annualSaveNoticeText(notice, _l10n));
    }
  }

  void _leave() {
    if (_isLeaving || !mounted) return;
    _isLeaving = true;
    Navigator.of(context).pop();
  }

  Future<void> _requestLeave() async {
    if (_model.form.isDirty &&
        !await confirmDiscardRecord(
          context,
          title: _l10n.annualFactsDiscardTitle,
        )) {
      return;
    }
    _leave();
  }

  @override
  void dispose() {
    _model.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => FormLeaveGuard(
    onLeave: _requestLeave,
    child: AppForm(
      controller: _model.form,
      onSave: _save,
      onCancel: () => unawaited(_requestLeave()),
      children: [
        if (_notice case final notice?) FormNotice(message: notice),
        AppMoneyField(
          field: _model.priorYear,
          textInputAction: TextInputAction.next,
          helperText: _l10n.annualFactsPriorYearHelp,
        ),
        AppMoneyField(
          field: _model.otherExempt,
          textInputAction: TextInputAction.done,
          helperText: _l10n.annualFactsOtherExemptHelp,
        ),
        AppCheckboxField(field: _model.toldEmployer),
        RegimeTaxInputsList(
          query: widget.query,
          inputs: widget.facts.regimeTaxInputs,
        ),
      ],
    ),
  );
}
