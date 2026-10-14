import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/features/incomes/application/income_defaults_loader.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:control_contable/features/incomes/application/income_form_request.dart';
import 'package:control_contable/features/incomes/application/income_records.dart';
import 'package:control_contable/features/incomes/application/income_save.dart';
import 'package:control_contable/features/incomes/data/incomes_repository.dart';
import 'package:control_contable/features/incomes/presentation/income_form_copy.dart';
import 'package:control_contable/features/incomes/presentation/income_form_fields.dart';
import 'package:control_contable/features/incomes/presentation/income_leave_guard.dart';
import 'package:control_contable/features/incomes/presentation/income_save_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The new-income form of one taxpayer: defaults prefilled and refreshed
/// as the context changes, *Guardar* (back to the list) and *Guardar y
/// crear otro* (stay, keep the context), idempotent retries, server
/// verdicts on their fields, non-blocking warnings, the paywall on `402`.
class IncomeFormPage extends ConsumerStatefulWidget {
  /// The form for [taxpayerId] over the model [createModel] builds.
  const new({required this.taxpayerId, required this.createModel, super.key});

  /// Whose income.
  final String taxpayerId;

  /// Builds the form's model once.
  final IncomeFormModel Function() createModel;

  @override
  ConsumerState<IncomeFormPage> createState() => _IncomeFormPageState();
}

class _IncomeFormPageState extends ConsumerState<IncomeFormPage> {
  late final IncomeFormModel _model = widget.createModel();
  late final IncomeDefaultsLoader _defaults = IncomeDefaultsLoader(
    _model,
    (key) => ref
        .read(incomesRepositoryProvider)
        .defaults(
          taxpayerId: widget.taxpayerId,
          regime: key.regime,
          counterparty: key.payer,
          on: key.on,
          platformActivity: key.activity,
        ),
  );
  bool? _withholdingsOpen;
  bool _moreOpen = false;
  String? _notice;
  bool _isLeaving = false;

  @override
  void initState() {
    super.initState();
    // Starts the first defaults request.
    _defaults.failed.addListener(_refresh);
  }

  void _refresh() => setState(() {});

  AppLocalizations get _l10n => AppLocalizations.of(context);

  /// Sends the form; the stored income, or `null` when it was not saved
  /// (the page already says why).
  Future<Income?> _persist() async {
    final missing = _model.missingChoices;
    if (missing.isNotEmpty) {
      _model.form.showServerErrors(missing);
      return null;
    }
    setState(() => _notice = null);
    final outcome = await saveIncome(
      ref.read(incomeRecordsProvider),
      widget.taxpayerId,
      _model.toCreate(),
    );
    if (!mounted) return null;
    switch (outcome) {
      case IncomeSaved(:final income):
        return income;
      case IncomeNeedsPlan(:final isTrialCap):
        showIncomePaywall(context, isTrialCap: isTrialCap);
      case IncomeRejected(:final details):
        _showVerdict(details);
      case IncomeNotSaved(:final notice):
        setState(() => _notice = incomeSaveNoticeText(notice, _l10n));
    }
    return null;
  }

  void _showVerdict(List<api.FieldError> details) {
    final verdict = matchServerFieldErrors(
      details,
      fieldsByPath: _model.fieldsByPath,
      describe: (detail) => describeIncomeDetail(detail, _l10n),
    );
    _model.form.showServerErrors(verdict.matched);
    if (verdict.unmatched.isEmpty) return;
    setState(() {
      _notice = verdict.unmatched
          .map((d) => describeFieldError(describeIncomeDetail(d, _l10n), _l10n))
          .toSet()
          .join(' ');
    });
  }

  Future<void> _save() async {
    final income = await _persist();
    if (income == null || !mounted) return;
    showIncomeSavedToast(ScaffoldMessenger.of(context), _l10n, income.warnings);
    _leave();
  }

  Future<bool> _saveForNext() async {
    final income = await _persist();
    if (income == null || !mounted) return false;
    _model.startNextSession();
    showIncomeSavedToast(ScaffoldMessenger.of(context), _l10n, income.warnings);
    return true;
  }

  void _leave() {
    if (_isLeaving || !mounted) return;
    _isLeaving = true;
    Navigator.of(context).pop();
  }

  Future<void> _requestLeave() async {
    if (_model.form.isDirty && !await confirmDiscardIncome(context)) return;
    _leave();
  }

  @override
  void dispose() {
    _defaults.dispose();
    _model.dispose();
    super.dispose();
  }

  IncomeSections get _sections => incomeSections(
    _model,
    withholdingsOpen: _withholdingsOpen,
    moreOpen: _moreOpen,
  );

  @override
  Widget build(BuildContext context) => IncomeLeaveGuard(
    onLeave: _requestLeave,
    child: Scaffold(
      appBar: AppBar(title: Text(_l10n.incomeFormTitle)),
      body: ListenableBuilder(
        listenable: _model,
        builder: (context, _) => AppForm(
          controller: _model.form,
          onSave: _save,
          onSaveForNext: _saveForNext,
          allowCreateAnother: true,
          onCancel: () => unawaited(_requestLeave()),
          children: [
            if (_notice case final notice?) FormNotice(message: notice),
            if (_defaults.failed.value)
              FormNotice(
                message: _l10n.incomeDefaultsFailed,
                tone: FormNoticeTone.info,
              ),
            ...incomeFormFields(
              context,
              model: _model,
              open: _sections,
              onToggleWithholdings: () =>
                  setState(() => _withholdingsOpen = !_sections.withholdings),
              onToggleMore: () => setState(() => _moreOpen = !_moreOpen),
            ),
          ],
        ),
      ),
    ),
  );
}
