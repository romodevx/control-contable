import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/salary_statement.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/form_leave_guard.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/application/salary_statement_form_model.dart';
import 'package:control_contable/features/annual/presentation/annual_form_sheet_frame.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/annual/presentation/salary_statement_fields.dart';
import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Opens the constancia form as a sheet that fits its content (≤ 6 fields,
/// ui-forms-and-modals.md): new, or an edit of [editing].
Future<void> showSalaryStatementSheet(
  BuildContext context, {
  required String taxpayerId,
  required int fiscalYear,
  SalaryStatement? editing,
}) => showModalBottomSheet<void>(
  context: context,
  isScrollControlled: true,
  useSafeArea: true,
  // Leaving with changes asks first; a drag would skip that question.
  enableDrag: false,
  builder: (_) => SalaryStatementSheet(
    taxpayerId: taxpayerId,
    fiscalYear: fiscalYear,
    editing: editing,
  ),
);

/// The constancia form inside its sheet: totals as on the CFDI de
/// retenciones, the special situations under a disclosure, the server's
/// verdict on its fields, the paywall on `402`.
class SalaryStatementSheet extends ConsumerStatefulWidget {
  /// The form of [taxpayerId]'s [fiscalYear], editing [editing] if given.
  const new({
    required this.taxpayerId,
    required this.fiscalYear,
    this.editing,
    super.key = const ValueKey('salary-statement-sheet'),
  });

  /// Whose constancia.
  final String taxpayerId;

  /// Its fiscal year.
  final int fiscalYear;

  /// The constancia edited; `null` for a new one.
  final SalaryStatement? editing;

  @override
  ConsumerState<SalaryStatementSheet> createState() =>
      _SalaryStatementSheetState();
}

class _SalaryStatementSheetState extends ConsumerState<SalaryStatementSheet> {
  late final AppLocalizations _l10n = AppLocalizations.of(context);
  late final SalaryStatementFormModel _model = SalaryStatementFormModel(
    copy: salaryStatementCopy(_l10n),
    editing: widget.editing,
  );
  late final String _id = ref.read(clientIdProvider)();
  late bool _moreOpen = _model.hasSpecialSituation;
  bool _isLeaving = false;
  String? _notice;

  Future<void> _save() async {
    if (blockedAsReadOnly(context, ref)) return;
    final writes = ref.read(annualInputWritesProvider);
    final stored = widget.editing;
    final patch = stored == null ? null : _model.toPatch(stored);
    if (patch != null && patch.isEmpty) return _leave();
    setState(() => _notice = null);
    final outcome = await annualSaveOutcomeOf(
      () => stored == null
          ? writes.createStatement(
              widget.taxpayerId,
              _model.toCreate(fiscalYear: widget.fiscalYear, id: _id),
            )
          : writes.updateStatement(widget.taxpayerId, stored.id, patch!),
      notFound: stored == null
          ? RecordSaveNotice.failed
          : RecordSaveNotice.gone,
    );
    if (!mounted) return;
    switch (outcome) {
      case RecordSaved(record: final statement):
        showSalarySavedToast(
          ScaffoldMessenger.of(context),
          _l10n,
          statement,
          updated: stored != null,
        );
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
    final notice = showAnnualVerdict(
      _model.form,
      details,
      fieldsByPath: _model.fieldsByPath,
      describe: (detail) => describeAnnualDetail(
        detail,
        _l10n,
        limitReached: _l10n.annualSalaryFull,
      ),
      l10n: _l10n,
    );
    if (notice != null) setState(() => _notice = notice);
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
          title: _l10n.annualSalaryDiscardTitle,
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
    child: AnnualFormSheetFrame(
      title: widget.editing == null
          ? _l10n.annualSalaryFormTitle
          : _l10n.annualSalaryEditTitle,
      form: ListenableBuilder(
        listenable: _model.form,
        builder: (context, _) => AppForm(
          controller: _model.form,
          onSave: _save,
          onCancel: () => unawaited(_requestLeave()),
          fitContent: true,
          children: [
            if (_notice case final notice?) FormNotice(message: notice),
            ...salaryStatementFields(
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
