// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/expense.dart';
import '../models/expense_category_list.dart';
import '../models/expense_create.dart';
import '../models/expense_page.dart';
import '../models/expense_update.dart';
import '../models/period.dart';
import '../models/regime_code.dart';
import '../models/uuid.dart';

part 'expenses_client.g.dart';

@RestApi()
abstract class ExpensesClient {
  factory ExpensesClient(Dio dio, {String? baseUrl}) = _ExpensesClient;

  /// Expense categories with deductibility resolved for the taxpayer's regime and options.
  ///
  /// Replaces the global `GET /v1/expense-categories` (OQ-E1): the hint depends on the taxpayer's person type (626 PF vs 626 PM) and regime options (606 `deduction_mode`, 625 `definitive_option`), so the catalog is read per taxpayer. Every category comes with `deductibility` (the hint shown), `default_deductible` (the boolean a new expense starts with; conservative for partial cases), the `note` and the article-level `legal_basis`, ordered by the catalog's `sort_order`. `iva_rates_allowed` and `cash_payment_limit_cents` are dated parameter data on `on` (default: today, America/ Mexico_City). `legend` is the mandatory text the form shows under the category. The catalog is pending CPA review (⚠️). Allowed for a read-only user and for an archived taxpayer. `regime` must be held by the taxpayer on `on` (`422 regime_not_active`).
  ///
  /// [on] - Date the dated parameters are read on (America/Mexico_City); defaults to today. `2022-01-01` … today (`out_of_range`).
  @GET('/v1/taxpayers/{taxpayerId}/expense-categories')
  Future<ExpenseCategoryList> listExpenseCategories({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('regime') required RegimeCode regime,
    @Query('on') String? onValue,
  });

  /// List a month's expenses (keyset pagination) with period totals.
  ///
  /// Live (not deleted) expenses of the taxpayer whose `paid_on` (the booking date for every regime) falls in `period`, ordered by `paid_on`, then `id`, both descending; pagination, `regime` filter and totals as `listIncomes`. `legend` is the mandatory text of the list header. `warnings` is always `[]` here.
  ///
  /// [cursor] - Opaque keyset cursor from `next_cursor`.
  @GET('/v1/taxpayers/{taxpayerId}/expenses')
  Future<ExpensePage> listExpenses({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') required Period period,
    @Query('limit') int? limit = 50,
    @Query('regime') RegimeCode? regime,
    @Query('cursor') String? cursor,
  });

  /// Record an expense (idempotent client id; trial cap enforced).
  ///
  /// Records an expense of the active taxpayer. Idempotency, order of checks, regime window, IVA rate validity and the `iva_cents` tolerance are those of `createIncome` (same `200` / `409 id_reused` rules). Trial cap: 5 expenses per user across all taxpayers, independent of incomes; a spent cap answers `402` with `{field: expenses_remaining, code: trial_cap_reached}`.
  ///
  /// **Deductibility.** `deductible` omitted or `null` → the default: the category rule for the regime, person type and options (MODULE-expenses §2.2–2.3) AND `has_cfdi` AND NOT a cash payment over the limit (`payment_method: cash` with `amount_cents + iva_cents` above `cash_payment_limit_cents`, or any cash `combustible`). A boolean is the user's override (`deductible_source: user`). Non-deductibility is never an error.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `category_code` → `unknown_category`; `cfdi_uuid` → `has_cfdi_mismatch` (a folio fiscal with `has_cfdi: false`); `paid_on` → `out_of_range` (before 2022-01-01 or after today). `422 business_rule_violation` details: `regime_code` → `regime_not_active`; `taxpayerId` → `taxpayer_archived`; `iva_rate` → `rate_not_valid_on_date`; `iva_cents` → `iva_amount_mismatch`; `isr_withheld_cents` / `iva_withheld_cents` → `withholding_not_applicable` (only a persona moral withholds from its suppliers), `withholding_exceeds_base`.
  ///
  /// Success bodies carry non-blocking `warnings`: `cash_over_threshold` on `payment_method`, `no_cfdi` on `has_cfdi`, `not_isr_deductible_in_regime` and `category_not_deductible` on `deductible` (an override to `true` against the hint), `duplicate_cfdi_uuid` on `cfdi_uuid`, `iva_not_creditable_cash_payment` on `payment_method` (a cash payment over the limit claimed as deductible: the claim stands for ISR, its IVA is not creditable, LIVA 5 fr. II ⚠️; added in 0.6.0, append-only).
  @POST('/v1/taxpayers/{taxpayerId}/expenses')
  Future<Expense> createExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required ExpenseCreate body,
  });

  /// Get an expense.
  ///
  /// A live expense of the taxpayer. A soft-deleted, purged, foreign or unknown expense, or a malformed id, answers `404`. `warnings` is always `[]` here.
  @GET('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}')
  Future<Expense> getExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
  });

  /// Edit an expense (merge-patch; null clears a nullable field or the override).
  ///
  /// Merge-patch on top-level fields as `updateIncome`: absent is unchanged, `null` clears `cfdi_uuid` or `description`, and `deductible: null` clears the user's override (the default applies again). Without an override the default `deductible` is recomputed when the category, regime, payment method, amounts, `has_cfdi` or `paid_on` change; an override is kept as is. Every rule of `createExpense` is re-run on the merged expense with the same `details[].code` values and `warnings`. A read-only user gets `402`; an archived taxpayer `422 taxpayer_archived`.
  @PATCH('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}')
  Future<Expense> updateExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
    @Body() required ExpenseUpdate body,
  });

  /// Soft-delete an expense (Undo via restore).
  ///
  /// As `deleteIncome`: sets `deleted_at`, restorable for 30 days, `204` also when already deleted, allowed for a read-only user, never frees a trial slot; an archived taxpayer answers `422 taxpayer_archived`.
  @DELETE('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}')
  Future<void> deleteExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
  });

  /// Undo a delete (never consumes a trial slot).
  ///
  /// As `restoreIncome`: clears `deleted_at` without consuming a trial slot (Undo works at 5/5), read-only check only (`402`), regime window re-checked (`422 regime_not_active`), archived taxpayer `422 taxpayer_archived`, a live expense `200` unchanged, a purged one `404`. `deductible` is not recomputed; `warnings` are computed as on create.
  @POST('/v1/taxpayers/{taxpayerId}/expenses/{expenseId}/restore')
  Future<Expense> restoreExpense({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('expenseId') required Uuid expenseId,
  });
}
