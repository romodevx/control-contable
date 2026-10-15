import 'dart:async';

import 'package:control_contable/api/generated/models/filed_payment_month.dart';
import 'package:control_contable/api/generated/models/filed_payments.dart';
import 'package:control_contable/api/generated/models/filed_payments_update.dart';
import 'package:control_contable/core/forms/app_form.dart';
import 'package:control_contable/core/forms/form_leave_guard.dart';
import 'package:control_contable/core/forms/form_notice.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/features/annual/application/annual_input_data.dart';
import 'package:control_contable/features/annual/application/annual_input_writes.dart';
import 'package:control_contable/features/annual/application/filed_payments_form_model.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/annual/presentation/filed_payment_row.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:intl/intl.dart';

/// The twelve months of [payments] as inline rows, with "Confirmar todos"
/// above them and *Guardar* in the sticky footer. Saving stays here: the
/// list reloads with the confirmed months.
class FiledPaymentsForm extends ConsumerStatefulWidget {
  /// The form of [payments] for [query].
  const new({required this.query, required this.payments, super.key});

  /// Whose payments, which year and regime.
  final FiledPaymentsQuery query;

  /// The months as the server answered them.
  final FiledPayments payments;

  @override
  ConsumerState<FiledPaymentsForm> createState() => _FiledPaymentsFormState();
}

class _FiledPaymentsFormState extends ConsumerState<FiledPaymentsForm> {
  late final AppLocalizations _l10n = AppLocalizations.of(context);
  late final FiledPaymentsFormModel _model = FiledPaymentsFormModel(
    payments: widget.payments,
    monthLabel: _monthName,
  );
  String? _notice;

  String _monthName(FiledPaymentMonth month) {
    final name = DateFormat.MMMM(Localizations.localeOf(context).toString())
        .format(DateTime(widget.payments.fiscalYear, month.month));
    return '${name[0].toUpperCase()}${name.substring(1)}';
  }

  Future<void> _put(FiledPaymentsUpdate Function() body) async {
    if (blockedAsReadOnly(context, ref)) return;
    final update = body();
    setState(() => _notice = null);
    final outcome = await annualSaveOutcomeOf(
      () => ref
          .read(annualInputWritesProvider)
          .putFiledPayments(widget.query, update),
    );
    if (!mounted) return;
    switch (outcome) {
      case RecordSaved():
        ScaffoldMessenger.of(context)
          ..hideCurrentSnackBar()
          ..showSnackBar(
            SnackBar(
              content: Text(_l10n.annualPaymentsSaved(update.months.length)),
            ),
          );
      case RecordNeedsPlan():
        showAnnualPaywall(context);
      case RecordRejected(:final details):
        setState(
          () => _notice = showAnnualVerdict(
            _model.form,
            details,
            fieldsByPath: const {},
            describe: (detail) => describeAnnualDetail(
              detail,
              _l10n,
              limitReached: _l10n.annualWriteFailed,
            ),
            l10n: _l10n,
          ),
        );
      case RecordNotSaved(:final notice):
        setState(() => _notice = annualSaveNoticeText(notice, _l10n));
    }
  }

  Future<void> _requestLeave() async {
    if (_model.form.isDirty &&
        !await confirmDiscardRecord(
          context,
          title: _l10n.annualFactsDiscardTitle,
        )) {
      return;
    }
    if (mounted) Navigator.of(context).pop();
  }

  @override
  void dispose() {
    _model.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final text = Theme.of(context).textTheme;
    return FormLeaveGuard(
      onLeave: _requestLeave,
      child: AppForm(
        controller: _model.form,
        onSave: () => _put(_model.toUpdate),
        children: [
          if (_notice case final notice?) FormNotice(message: notice),
          Text(_l10n.annualPaymentsIntro, style: text.bodyMedium),
          OutlinedButton(
            onPressed: () =>
                unawaited(_model.form.submit(() => _put(_model.confirmAll))),
            child: Text(_l10n.annualPaymentsConfirmAll),
          ),
          Text(_l10n.annualPaymentsConfirmAllHelp, style: text.bodySmall),
          Text(
            _l10n.annualPaymentsTotal(
              formatPesos(widget.payments.totalEffectiveCents),
            ),
            style: text.titleMedium,
          ),
          for (final row in _model.rows) FiledPaymentRowView(row: row),
        ],
      ),
    );
  }
}
