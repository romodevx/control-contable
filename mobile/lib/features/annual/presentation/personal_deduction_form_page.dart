import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_leave_guard.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/core/time/clock.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/application/personal_deduction_form_model.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/annual/presentation/personal_deduction_fields.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The new personal deduction form (ANN-19), on its own screen (up to 7
/// fields, ui-forms-and-modals.md): category first, then its fields. The
/// server is the authority; `402` opens the paywall over the kept data.
class PersonalDeductionFormPage extends ConsumerStatefulWidget {
  /// The form of [taxpayerId]'s deductions in [fiscalYear].
  const new({required this.taxpayerId, required this.fiscalYear, super.key});

  /// Whose deduction.
  final String taxpayerId;

  /// The year it belongs to.
  final int fiscalYear;

  @override
  ConsumerState<PersonalDeductionFormPage> createState() =>
      _PersonalDeductionFormPageState();
}

class _PersonalDeductionFormPageState
    extends ConsumerState<PersonalDeductionFormPage> {
  late final AppLocalizations _l10n = AppLocalizations.of(context);
  late final PersonalDeductionFormModel _model = PersonalDeductionFormModel(
    fiscalYear: widget.fiscalYear,
    today: businessToday(ref.read(clockProvider)()),
    copy: _copy(_l10n, widget.fiscalYear),
    newId: ref.read(clientIdProvider),
  );
  bool _moreOpen = false;
  bool _isLeaving = false;
  String? _notice;

  Future<void> _save() async {
    if (blockedAsReadOnly(context, ref)) return;
    setState(() => _notice = null);
    final outcome = await annualSaveOutcomeOf(
      () => ref
          .read(annualInputWritesProvider)
          .createDeduction(widget.taxpayerId, _model.toCreate()),
    );
    if (!mounted) return;
    switch (outcome) {
      case RecordSaved():
        ref
            .read(annualDeductionsUndoProvider)
            .notify(ScaffoldMessenger.of(context), _l10n.annualDeductionSaved);
        _leave();
      case RecordNeedsPlan():
        showAnnualPaywall(context);
      case RecordRejected(:final details):
        _showVerdict(details);
      case RecordNotSaved(:final notice):
        setState(() => _notice = annualSaveNoticeText(notice, _l10n));
    }
  }

  void _showVerdict(List<api.FieldError> details) {
    FieldError describe(api.FieldError detail) => describeAnnualDetail(
      detail,
      _l10n,
      limitReached: _l10n.annualDeductionsFullOnSave,
      outOfRange: _l10n.annualDeductionDateOutOfYear('${widget.fiscalYear}'),
    );
    final verdict = matchServerFieldErrors(
      details,
      fieldsByPath: _model.fieldsByPath,
      describe: describe,
    );
    _model.form.showServerErrors(verdict.matched);
    if (verdict.unmatched.isEmpty) return;
    setState(() {
      _notice = verdict.unmatched
          .map((detail) => describeFieldError(describe(detail), _l10n))
          .toSet()
          .join(' ');
    });
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
          title: _l10n.annualDeductionDiscardTitle,
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
    child: Scaffold(
      appBar: AppBar(title: Text(_l10n.annualDeductionFormTitle)),
      body: ListenableBuilder(
        listenable: _model.form,
        builder: (context, _) => AppForm(
          controller: _model.form,
          onSave: _save,
          onCancel: () => unawaited(_requestLeave()),
          children: [
            if (_notice case final notice?) FormNotice(message: notice),
            ...personalDeductionFields(
              context,
              model: _model,
              moreOpen: _moreOpen,
              onToggleMore: () => setState(() => _moreOpen = !_moreOpen),
            ),
          ],
        ),
      ),
    ),
  );
}

PersonalDeductionFormCopy _copy(AppLocalizations l10n, int year) => (
  category: l10n.annualDeductionCategory,
  amount: l10n.annualDeductionAmount,
  paidOn: l10n.annualDeductionPaidOn,
  electronic: l10n.annualDeductionElectronic,
  beneficiary: l10n.annualDeductionBeneficiary,
  attest: l10n.annualDeductionAttest('$year'),
  donee: l10n.annualDeductionDonee,
  student: l10n.annualDeductionStudent,
  studentInvalid: l10n.annualDeductionStudentInvalid,
  level: l10n.annualDeductionLevel,
  scholarship: l10n.annualDeductionScholarship,
  cfdi: l10n.annualDeductionCfdi,
  cfdiInvalid: l10n.annualDeductionCfdiInvalid,
);
