import 'package:control_contable/api/generated/models/history_change_reason.dart';
import 'package:control_contable/api/generated/models/history_month.dart';
import 'package:control_contable/api/generated/models/history_tax.dart';
import 'package:control_contable/api/generated/models/history_tax_iva_outcome.dart';
import 'package:control_contable/api/generated/models/taxpayer_estimate_status.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/tax_engine/domain/money_format.dart';
import 'package:control_contable/features/taxpayers/domain/periods.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';

/// Every text of one History month, shown and spoken, in es-MX. Amounts
/// are formatted, never computed or rounded (MODULE-tax-engine §2.7).
class HistoryMonthLabels {
  /// The labels of [month] in [context]'s locale.
  new(BuildContext context, this.month)
    : _l10n = AppLocalizations.of(context),
      _locale = Localizations.localeOf(context).toString();

  /// The month described.
  final HistoryMonth month;

  final AppLocalizations _l10n;
  final String _locale;

  /// "Febrero".
  String get name {
    final word = _monthWord;
    return word[0].toUpperCase() + word.substring(1);
  }

  /// "Febrero de 2026".
  String get title => _l10n.historyMonthOfYear(name, '$_year');

  /// "Ver ingresos de febrero de 2026": the incomes link, spoken.
  String get incomesLink => _l10n.historyViewIncomesOf(
    _l10n.historyMonthOfYear(_monthWord, '$_year'),
  );

  /// "Ver gastos de febrero de 2026": the expenses link, spoken.
  String get expensesLink => _l10n.historyViewExpensesOf(
    _l10n.historyMonthOfYear(_monthWord, '$_year'),
  );

  /// What a screen reader hears on coming back from the month
  /// (MODULE-history §6): "Febrero de 2026. Visto el 20 mar.", or its
  /// marker when it still differs; `null` when it carries neither.
  String? get returnState {
    final state = marker ?? viewed;
    return state == null ? null : [title, state].map(_sentence).join(' ');
  }

  String get _monthWord =>
      DateFormat.MMMM(_locale).format(periodMonth(month.period));

  int get _year => periodMonth(month.period).year;

  /// "ISR $400.00 · IVA $3,100.00", or the IVA as saldo a favor.
  String taxLine(HistoryTax tax) {
    final isr = formatMoney(tax.isrPayableRoundedCents);
    return tax.ivaOutcome == HistoryTaxIvaOutcome.inFavor
        ? _l10n.historyTaxLineInFavor(isr, formatMoney(tax.ivaInFavorCents))
        : _l10n.historyTaxLine(isr, formatMoney(tax.ivaPayableRoundedCents));
  }

  /// "Retenciones a enterar $10,333.33".
  String withholdings(HistoryTax tax) =>
      _l10n.historyWithholdings(formatMoney(tax.withholdingsToRemitCents));

  /// The status word; `null` for a month with no regime.
  String? get status => switch (month.status) {
    TaxpayerEstimateStatus.complete => _l10n.estimateComplete,
    TaxpayerEstimateStatus.incomplete => _l10n.estimateIncomplete,
    TaxpayerEstimateStatus.partial => _l10n.estimatePartial,
    TaxpayerEstimateStatus.informational ||
    TaxpayerEstimateStatus.$unknown => _l10n.estimateInformational,
    TaxpayerEstimateStatus.none => null,
  };

  /// Why there is no figure, or no record, in the month; `null` otherwise.
  String? get gap {
    if (month.status == TaxpayerEstimateStatus.none) {
      return _l10n.historyNoRegime;
    }
    return month.hasRecords ? null : _l10n.historyNoRecords;
  }

  /// "Vence el 17 de marzo de 2026".
  String get dueDate => _l10n.estimateDueDate(parseBusinessDate(month.dueDate));

  /// The changed marker in plain words (MODULE-history §2.4); `null` when
  /// unchanged. The current month is live and never marked.
  String? get marker {
    if (!month.changedSinceViewed || month.isCurrent) return null;
    return month.changeReason == HistoryChangeReason.engine
        ? _l10n.historyChangedEngine
        : _l10n.historyChangedRecords;
  }

  /// "Visto el 14 feb" for a viewed month that did not change.
  String? get viewed {
    final viewedAt = month.viewedAt;
    if (viewedAt == null || month.isCurrent || marker != null) return null;
    return _l10n.historyViewedOn(businessToday(viewedAt));
  }

  /// The whole row as one sentence, in reading order (MODULE-history §6):
  /// "Febrero de 2026. Total a pagar: 3,500 pesos. Completo. Cambió …".
  String get spoken {
    final tax = month.tax;
    return [
      if (month.isCurrent) _l10n.historyMonthCurrentSpoken(title) else title,
      if (tax != null) _spokenTotal(tax.totalPayableRoundedCents),
      ?status,
      ?gap,
      ?marker,
      ?viewed,
    ].map(_sentence).join(' ');
  }

  String _spokenTotal(int centavos) =>
      '${_l10n.estimateTotalPayable}: ${spokenMoney(centavos)}';

  static String _sentence(String part) => part.endsWith('.') ? part : '$part.';
}
