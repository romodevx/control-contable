// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:freezed_annotation/freezed_annotation.dart';

import 'decimal_rate.dart';
import 'non_negative_cents.dart';
import 'withholder.dart';
import 'withholding_basis.dart';
import 'withholding_explanation.dart';
import 'withholding_rule_id.dart';

part 'withholding.freezed.dart';
part 'withholding.g.dart';

/// **Derived, never entered** (ADR-0008, MODULE-withholdings §3). The server computes it with the kernel's decision table from the facts of the record (person type and regime of the taxpayer, `counterparty_kind` or the supplier fields, `operation_kind`, `residential_rent`, `via_platform`, amounts, dates) and stores it with the `table_version` that produced it. No request accepts it. `isr_cents` and `iva_cents` are the record's `isr_withheld_cents` / `iva_withheld_cents`. Amounts: `isr_cents = round(amount_cents × isr_rate)`; `iva_cents` is the `iva_basis` share of the record's IVA, or 4 % of `amount_cents`; one rounding to the centavo, half away from zero. `net_cents` = subtotal + IVA − both withholdings: what the taxpayer receives (income) or pays the supplier (expense).
@Freezed()
abstract class Withholding with _$Withholding {
  const factory Withholding({
    @JsonKey(name: 'isr_cents') required NonNegativeCents isrCents,
    @JsonKey(name: 'iva_cents') required NonNegativeCents ivaCents,

    /// Fraction of `amount_cents` (RESICO `0.0125`, fees and rent `0.10`); `'0'` when no ISR is withheld.
    @JsonKey(name: 'isr_rate') required DecimalRate isrRate,
    @JsonKey(name: 'iva_basis') required WithholdingBasis ivaBasis,
    required Withholder withholder,
    @JsonKey(name: 'net_cents') required NonNegativeCents netCents,

    /// Expenses with `withholder: you` and an amount above 0: the 17th of the month after `paid_on`, when the taxpayer must remit both withholdings (never netted against IVA). `null` otherwise.
    @JsonKey(includeIfNull: true, name: 'remit_due_on')
    required String? remitDueOn,
    @JsonKey(name: 'rule_ids') required List<WithholdingRuleId> ruleIds,

    /// One line per entry of `rule_ids`, same order.
    required List<WithholdingExplanation> explanations,

    /// Version of the decision table and rates that produced the figures (e.g. `2026.1`). A CPA answer or a new year's rates bumps it, and `cmd/backfill-withholdings` re-derives older records (MODULE-withholdings §5.3).
    @JsonKey(name: 'table_version') required String tableVersion,

    /// A ⚠️ rule a CPA has not confirmed matched (warning `withholding_cpa_unverified`).
    @JsonKey(name: 'cpa_unverified') required bool cpaUnverified,
  }) = _Withholding;

  factory Withholding.fromJson(Map<String, Object?> json) =>
      _$WithholdingFromJson(json);
}
