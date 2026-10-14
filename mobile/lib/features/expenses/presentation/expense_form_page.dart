import 'dart:async';

import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_leave_guard.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/core/theme/app_spacing.dart';
import 'package:control_contable/core/widgets/load_status.dart';
import 'package:control_contable/features/expenses/application/expense_catalog_loader.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:control_contable/features/expenses/application/expense_form_request.dart';
import 'package:control_contable/features/expenses/application/expense_records.dart';
import 'package:control_contable/features/expenses/application/expense_save.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:control_contable/features/expenses/presentation/expense_copy.dart';
import 'package:control_contable/features/expenses/presentation/expense_form_fields.dart';
import 'package:control_contable/features/expenses/presentation/expense_save_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The expense form of one taxpayer, new or edit (the model says which):
/// the category catalog loaded for the regime and date, *Guardar* (back
/// to the list) and, for a new expense, *Guardar y crear otro* (stay, keep
/// the context); idempotent retries (a client id on create, a merge-patch
/// of only the changed fields on edit), server verdicts on their fields,
/// non-blocking warnings, the paywall on `402`.
class ExpenseFormPage extends ConsumerStatefulWidget {
  /// The form for [taxpayerId] over the model [createModel] builds.
  const new({required this.taxpayerId, required this.createModel, super.key});

  /// Whose expense.
  final String taxpayerId;

  /// Builds the form's model once.
  final ExpenseFormModel Function() createModel;

  @override
  ConsumerState<ExpenseFormPage> createState() => _ExpenseFormPageState();
}

class _ExpenseFormPageState extends ConsumerState<ExpenseFormPage> {
  late final ExpenseFormModel _model = widget.createModel();
  late final ExpenseCatalogLoader _catalog = ExpenseCatalogLoader(
    _model,
    (query) => ref
        .read(expensesRepositoryProvider)
        .categories(
          taxpayerId: widget.taxpayerId,
          regime: query.regime,
          on: query.on,
        ),
  );
  bool _moreOpen = false;
  bool? _withholdingsOpen;
  String? _notice;
  bool _isLeaving = false;

  @override
  void initState() {
    super.initState();
    // Starts the first catalog request.
    _catalog.failed.addListener(_refresh);
  }

  void _refresh() => setState(() {});

  AppLocalizations get _l10n => AppLocalizations.of(context);

  /// Sends the form; the stored expense, or `null` when it was not saved
  /// (the page already says why). An edit that changed nothing sends
  /// nothing and answers the stored expense.
  Future<Expense?> _persist() async {
    final records = ref.read(expenseRecordsProvider);
    final ExpenseSaveOutcome outcome;
    if (_model.editing case final stored?) {
      final patch = _model.toPatch(stored);
      if (patch.isEmpty) return stored;
      setState(() => _notice = null);
      outcome = await updateExpense(
        records,
        widget.taxpayerId,
        stored.id,
        patch,
      );
    } else {
      setState(() => _notice = null);
      outcome = await saveExpense(
        records,
        widget.taxpayerId,
        _model.toCreate(),
      );
    }
    if (!mounted) return null;
    switch (outcome) {
      case RecordSaved(record: final expense):
        return expense;
      case RecordNeedsPlan(:final isTrialCap):
        showExpensePaywall(context, isTrialCap: isTrialCap);
      case RecordRejected(:final details):
        _showVerdict(details);
      case RecordNotSaved(:final notice):
        setState(() => _notice = expenseSaveNoticeText(notice, _l10n));
    }
    return null;
  }

  void _showVerdict(List<api.FieldError> details) {
    final verdict = matchServerFieldErrors(
      details,
      fieldsByPath: _model.fieldsByPath,
      describe: (detail) => describeExpenseDetail(detail, _l10n),
    );
    _model.form.showServerErrors(verdict.matched);
    if (verdict.unmatched.isEmpty) return;
    setState(() {
      _notice = verdict.unmatched
          .map(
            (d) => describeFieldError(describeExpenseDetail(d, _l10n), _l10n),
          )
          .toSet()
          .join(' ');
    });
  }

  void _toast(Expense expense, {bool updated = false, bool overForm = false}) =>
      showExpenseSavedToast(
        ScaffoldMessenger.of(context),
        _l10n,
        expense.warnings,
        undo: ref.read(expensesUndoProvider),
        updated: updated,
        overForm: overForm,
      );

  Future<void> _save() async {
    final expense = await _persist();
    if (expense == null || !mounted) return;
    if (!identical(expense, _model.editing)) {
      _toast(expense, updated: _model.isEditing);
    }
    _leave();
  }

  Future<bool> _saveForNext() async {
    final expense = await _persist();
    if (expense == null || !mounted) return false;
    _model.startNextSession();
    setState(() => _withholdingsOpen = null);
    _toast(expense, overForm: true);
    return true;
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
          title: _l10n.expenseDiscardTitle,
        )) {
      return;
    }
    _leave();
  }

  @override
  void dispose() {
    _catalog.dispose();
    _model.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => FormLeaveGuard(
    onLeave: _requestLeave,
    child: Scaffold(
      appBar: AppBar(
        title: Text(
          _model.isEditing ? _l10n.expenseEditTitle : _l10n.expenseFormTitle,
        ),
      ),
      body: ListenableBuilder(listenable: _model, builder: _body),
    ),
  );

  Widget _body(BuildContext context, Widget? _) {
    final catalog = _model.catalog;
    if (catalog == null) {
      return Padding(
        padding: const EdgeInsets.all(AppSpacing.gutter),
        child: _catalog.failed.value
            ? LoadFailed(
                message: _l10n.expenseCategoriesLoadFailed,
                onRetry: _catalog.retry,
              )
            : LoadStatus(message: _l10n.expenseCategoriesLoading, busy: true),
      );
    }
    final open = expenseSections(
      _model,
      moreOpen: _moreOpen,
      withholdingsOpen: _withholdingsOpen,
    );
    return AppForm(
      controller: _model.form,
      onSave: _save,
      onSaveForNext: _model.isEditing ? null : _saveForNext,
      allowCreateAnother: !_model.isEditing,
      onCancel: () => unawaited(_requestLeave()),
      children: [
        if (_notice case final notice?) FormNotice(message: notice),
        if (_catalog.failed.value)
          FormNotice(
            message: _l10n.expenseCategoriesLoadFailed,
            tone: FormNoticeTone.info,
          ),
        ...expenseFormFields(
          context,
          model: _model,
          legend: catalog.legend,
          categoryUsage: ref.watch(categoryUsageProvider),
          open: open,
          onToggleMore: () => setState(() => _moreOpen = !open.more),
          onToggleWithholdings: () =>
              setState(() => _withholdingsOpen = !open.withholdings),
        ),
      ],
    );
  }
}
