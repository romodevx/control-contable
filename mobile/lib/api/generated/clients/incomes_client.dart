// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/counterparty_kind.dart';
import '../models/income.dart';
import '../models/income_create.dart';
import '../models/income_defaults.dart';
import '../models/income_page.dart';
import '../models/income_update.dart';
import '../models/iva_rate.dart';
import '../models/non_negative_cents.dart';
import '../models/operation_kind.dart';
import '../models/period.dart';
import '../models/platform_activity.dart';
import '../models/positive_cents.dart';
import '../models/regime_code.dart';
import '../models/uuid.dart';
import '../models/withholding_preview.dart';

part 'incomes_client.g.dart';

@RestApi()
abstract class IncomesClient {
  factory IncomesClient(Dio dio, {String? baseUrl}) = _IncomesClient;

  /// List a month's incomes (keyset pagination) with period totals.
  ///
  /// Live (not deleted) incomes of the taxpayer whose **booking date** falls in `period`: `invoiced_on` for accrual regimes (PM Título II), `collected_on` for cash regimes (MODULE-incomes §2.2). Ordered by booking date, then `id`, both descending; `cursor` is the opaque `next_cursor` of the previous page (`limit` 1–100, default 50). `regime` narrows items and totals. `totals` cover the whole filtered period, not the page. Accrual incomes with `collected_on: null` come with `pending_collection: true` ("Por cobrar"). Archived taxpayers can be listed. Reads are allowed for a read-only user. `warnings` is always `[]` here (warnings are computed per write, never stored).
  ///
  /// [cursor] - Opaque keyset cursor from `next_cursor`.
  @GET('/v1/taxpayers/{taxpayerId}/incomes')
  Future<IncomePage> listIncomes({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('period') required Period period,
    @Query('limit') int? limit = 50,
    @Query('regime') RegimeCode? regime,
    @Query('cursor') String? cursor,
  });

  /// Record an income (idempotent client id; trial cap enforced).
  ///
  /// Records an income of the active taxpayer. Amounts are MXN centavos (`int64`), the amount is the subtotal **before IVA**; there is no currency field.
  ///
  /// **Idempotency (double submit).** The client generates `id` (UUIDv7 recommended) once per form session and resends it on every retry; the server generates one when it is absent. Before any access check: a new `id` → `201`; an `id` the caller already owns for the same taxpayer, not deleted, with the same persisted fields (compared after normalization) → `200` with the stored income, no second record and no trial slot (even at 5/5); the same `id` with other fields, another taxpayer, or a soft-deleted income → `409 conflict` with `{field: id, code: id_reused}`; an `id` that belongs to another user → `409 conflict` with no `details` (it reveals no ownership).
  ///
  /// **Order of checks** (one transaction): schema (`400`) → idempotency (`200`/`409`) → read-only access (`402`) → regime window and archive (`422`) → domain rules (`400`/`422`) → trial slot (`402`, atomic) → insert. A failed insert gives the slot back.
  ///
  /// **Trial cap.** 5 incomes per user across all taxpayers, independent of the 5 expenses. A spent cap answers `402 entitlement_required` with `{field: incomes_remaining, code: trial_cap_reached}`; deleting never frees a slot.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords (`required`, `minimum`, `enum`, `format`, `pattern`, `maxLength`, `additionalProperties`, …): `collected_on` / `invoiced_on` → `required` (the date the regime's basis needs: `collected_on` for cash regimes, `invoiced_on` for accrual), `out_of_range` (before 2026-01-01, the first fiscal year with a tax parameter set, or after today, America/Mexico_City; see `BusinessDate`); `platform_activity` → `required` (625 with more than one activity), `not_for_regime` (sent for a regime other than 625); `amount_cents` → `minimum` (0 is refused at the edge; `must_be_positive` is the domain's backstop code for the same rule). `422 business_rule_violation` details: `regime_code` → `regime_not_active` (the regime is not held on one of the income's dates); `taxpayerId` → `taxpayer_archived`; `iva_rate` → `rate_not_valid_on_date` (rate not allowed on the IVA date, e.g. `rate_8` after the border decree ends); `iva_cents` → `iva_amount_mismatch` (`rate_0`/`exempt` with IVA > 0, or more than 100 centavos away from `amount_cents × rate`); `platform_activity` → `activity_not_held` (not among the taxpayer's 625 activities); `iva_rate` → `residential_rent_exempt` (`residential_rent: true` with a rate other than `exempt`). `via_platform: true` outside 625 and persona moral regimes → `400` `not_for_regime`.
  ///
  /// **Withholdings are derived (ADR-0008, MODULE-withholdings §3).** The server derives ISR and IVA withheld from the taxpayer's person type and regime, `counterparty_kind`, `operation_kind`, `residential_rent`, `via_platform` and the amounts, stores them with the rule ids and `table_version`, and returns them read-only in `withholding` (mirrored in `isr_withheld_cents` / `iva_withheld_cents`). There is no manual override. `counterparty_kind` is required (`400 required`). **0.12.0 (WH-11, ADR-0008 amendment):** `isr_withheld_cents`, `iva_withheld_cents` and `counterparty_type` are no longer in the schema; a body that carries any of them answers `400 validation_failed` with `additionalProperties` on that field.
  ///
  /// Success bodies carry non-blocking `warnings` (MODULE-incomes §2.8, MODULE-withholdings §6): `duplicate_cfdi_uuid` on `cfdi_uuid` (another live income of the taxpayer has the same folio fiscal: legit for partial collections); `provisional_parameters` on `collected_on` / `invoiced_on` (the date reads a provisional parameter set, see `WarningCode`); `withholding_cpa_unverified` on `withholding`; `foreign_counterparty_not_supported` on `counterparty_kind`. `withholding_differs_from_rate` is retired (never sent since 0.11.3). Messages never carry amounts, descriptions or UUIDs.
  @POST('/v1/taxpayers/{taxpayerId}/incomes')
  Future<Income> createIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required IncomeCreate body,
  });

  /// IVA rates and last-used withholding facts for a new income (prefills the form).
  ///
  /// Resolved from the rate catalog (`RateCatalog`, dated parameter data) for the taxpayer's person type, the regime and its options (625 `activities`) and the date `on` (default: today, America/Mexico_City). The app never hard-codes a rate. `iva_rates_allowed` lists the rates valid on `on` (`rate_8` disappears when the border decree ends). For 625, `platform_activity` defaults to the taxpayer's only activity and stays `null` when there are several and none is sent. Reads are allowed for a read-only user and for an archived taxpayer. `regime` must be held by the taxpayer on `on` (`422 regime_not_active`).
  ///
  /// `last_used` carries the withholding facts to preselect (MODULE-withholdings §2.1). Withholdings are derived (ADR-0008), so there is no rate to suggest: the form shows `previewIncomeWithholdings`. **0.12.0 (WH-11, ADR-0008 amendment):** the `counterparty_type` parameter and the four `*_withholding_*` fields are removed; `last_used` is required.
  ///
  /// [on] - Date the dated parameters are read on (America/Mexico_City); defaults to today. `2026-01-01` (the first fiscal year with a tax parameter set) … today (`out_of_range`).
  ///
  /// [platformActivity] - 625 only (`not_for_regime` otherwise); must be one of the taxpayer's activities (`activity_not_held`).
  @GET('/v1/taxpayers/{taxpayerId}/incomes/defaults')
  Future<IncomeDefaults> getIncomeDefaults({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('regime') required RegimeCode regime,
    @Query('on') String? onValue,
    @Query('platform_activity') PlatformActivity? platformActivity,
  });

  /// Derive the withholdings of an income without storing it (live card of the form).
  ///
  /// Runs the same derivation `createIncome` stores (ADR-0008, MODULE-withholdings §3) on the facts in the query and stores nothing; the form calls it debounced as the user types (§2.3). Same validation as `createIncome` for the fields it takes: `400` for malformed or out-of-range input (`on` before 2026-01-01 or after today → `out_of_range`; `via_platform: true` outside 625 and persona moral regimes or `platform_activity` outside 625 → `not_for_regime`), `422` for `regime_not_active`, `rate_not_valid_on_date`, `iva_amount_mismatch`, `residential_rent_exempt`, `activity_not_held`. `on` is the income's withholding date (`collected_on`, else `invoiced_on`; default today, America/Mexico_City): it picks the dated rates. Reads are allowed for a read-only user and an archived taxpayer (no `402`). Per-user budget `withholding_previews`, 1200 per hour (burst 1200, ADR-0005), shared with `previewExpenseWithholdings` and separate from the record budgets; spent → `429 rate_limited` with `Retry-After`. The query carries amounts, never a name, an RFC or a description; the server does not log it.
  ///
  /// [regime] - The record's `regime_code`; must be held by the taxpayer on `on` (`422 regime_not_active`).
  ///
  /// [on] - Date the dated parameters are read on (America/Mexico_City); defaults to today. `2026-01-01` (the first fiscal year with a tax parameter set) … today (`out_of_range`).
  ///
  /// [amountCents] - Subtotal before IVA, MXN centavos (as `amount_cents` of the record).
  ///
  /// [ivaCents] - The IVA of the record, the base of the IVA shares (same tolerance rule as the record).
  ///
  /// [operationKind] - Omitted → the default the matching create would store.
  ///
  /// [residentialRent] - Default `false`; `true` requires `iva_rate=exempt` (`422 residential_rent_exempt`).
  ///
  /// [viaPlatform] - Default `true` in 625, `false` otherwise.
  ///
  /// [platformActivity] - 625 only; defaults to the taxpayer's only activity.
  @GET('/v1/taxpayers/{taxpayerId}/incomes/withholding-preview')
  Future<WithholdingPreview> previewIncomeWithholdings({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('regime') required RegimeCode regime,
    @Query('amount_cents') required PositiveCents amountCents,
    @Query('iva_rate') required IvaRate ivaRate,
    @Query('iva_cents') required NonNegativeCents ivaCents,
    @Query('counterparty_kind') required CounterpartyKind counterpartyKind,
    @Query('on') String? onValue,
    @Query('operation_kind') OperationKind? operationKind,
    @Query('residential_rent') bool? residentialRent,
    @Query('via_platform') bool? viaPlatform,
    @Query('platform_activity') PlatformActivity? platformActivity,
  });

  /// Get an income.
  ///
  /// A live income of the taxpayer. A soft-deleted, purged, foreign or unknown income, or a malformed id, answers `404`. `warnings` is always `[]` here.
  @GET('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}')
  Future<Income> getIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });

  /// Edit an income (merge-patch; null clears a nullable field).
  ///
  /// Partial update with merge-patch semantics on top-level fields: an absent field is unchanged; `null` clears a nullable one (`invoiced_on`, `collected_on`, `platform_activity`, `cfdi_uuid`, `description`). `id` cannot be sent (`additionalProperties`). Every rule of `createIncome` is re-run on the merged income, with the same `details[].code` values and `warnings`; moving it to another `regime_code` re-checks the regime window. Setting `collected_on` on an accrual income clears `pending_collection`. Any change re-derives `withholding` from the merged facts (ADR-0008; e.g. `counterparty_kind: pf_with_activity` zeroes it, US-W9). A read-only user gets `402`; an archived taxpayer `422 taxpayer_archived`. Last write wins (no ETag in v1).
  @PATCH('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}')
  Future<Income> updateIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
    @Body() required IncomeUpdate body,
  });

  /// Soft-delete an income (Undo via restore).
  ///
  /// Sets `deleted_at`: the income leaves lists, totals and the estimate at once and can be restored for 30 days (then it is purged). Deleting an already deleted income answers `204`. Allowed for a read-only user. Never frees a trial slot. An archived taxpayer answers `422 taxpayer_archived` (archive freezes the data).
  @DELETE('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}')
  Future<void> deleteIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });

  /// Undo a delete (never consumes a trial slot).
  ///
  /// Clears `deleted_at`. **Never consumes a trial slot**: the slot was used once at creation and never freed, so Undo works at 5/5 with no paywall. **Allowed for a read-only user** (no access check, never `402`): Undo gives back the user's own record and creates nothing (0.6.1, MODULE-incomes §2.7). The regime window is re-checked (`422 regime_not_active` when the regime no longer covers the income's dates); an archived taxpayer answers `422 taxpayer_archived`. Restoring a live income answers `200` unchanged; a purged one, or another user's, `404`. `warnings` are computed as on create. A restored income whose `withholding.table_version` is older than the current table is re-derived (ADR-0008), like the backfill would.
  @POST('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore')
  Future<Income> restoreIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });
}
