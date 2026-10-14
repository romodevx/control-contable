// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/counterparty_type.dart';
import '../models/income.dart';
import '../models/income_create.dart';
import '../models/income_defaults.dart';
import '../models/income_page.dart';
import '../models/income_update.dart';
import '../models/period.dart';
import '../models/platform_activity.dart';
import '../models/regime_code.dart';
import '../models/uuid.dart';

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
  /// `FieldError.code` values beyond JSON Schema keywords (`required`, `minimum`, `enum`, `format`, `pattern`, `maxLength`, `additionalProperties`, …): `collected_on` / `invoiced_on` → `required` (the date the regime's basis needs: `collected_on` for cash regimes, `invoiced_on` for accrual), `out_of_range` (before 2022-01-01 or after today, America/Mexico_City); `platform_activity` → `required` (625 with more than one activity), `not_for_regime` (sent for a regime other than 625); `amount_cents` → `minimum` (0 is refused at the edge; `must_be_positive` is the domain's backstop code for the same rule). `422 business_rule_violation` details: `regime_code` → `regime_not_active` (the regime is not held on one of the income's dates); `taxpayerId` → `taxpayer_archived`; `iva_rate` → `rate_not_valid_on_date` (rate not allowed on the IVA date, e.g. `rate_8` after the border decree ends); `iva_cents` → `iva_amount_mismatch` (`rate_0`/`exempt` with IVA > 0, or more than 100 centavos away from `amount_cents × rate`); `isr_withheld_cents` / `iva_withheld_cents` → `withholding_not_applicable` (the person type, regime and `counterparty_type` allow no such withholding), `withholding_exceeds_base` (ISR above `amount_cents`, IVA above `iva_cents`); `platform_activity` → `activity_not_held` (not among the taxpayer's 625 activities).
  ///
  /// Success bodies carry non-blocking `warnings` (MODULE-incomes §2.8): `duplicate_cfdi_uuid` on `cfdi_uuid` (another live income of the taxpayer has the same folio fiscal: legit for partial collections) and `withholding_differs_from_rate` on `isr_withheld_cents` / `iva_withheld_cents` (the CFDI wins). Messages never carry amounts, descriptions or UUIDs.
  @POST('/v1/taxpayers/{taxpayerId}/incomes')
  Future<Income> createIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required IncomeCreate body,
  });

  /// IVA rates and suggested withholding rates for a new income (prefills the form).
  ///
  /// Resolved from the rate catalog (`RateCatalog`, dated parameter data) for the taxpayer's person type, the regime and its options (625 `activities`, `rfc_given_to_platform`), the counterparty and the date `on` (default: today, America/Mexico_City). The app never hard-codes a rate. A suggestion is UX only: the stored value is what the user enters (the CFDI wins). `iva_rates_allowed` lists the rates valid on `on` (`rate_8` disappears when the border decree ends). A withholding that is not allowed has `*_allowed: false` and a `null` rate (the value must be 0); an allowed one without a suggestion has `*_allowed: true` and a `null` rate. The ISR rate applies to `amount_cents`, the IVA rate to `iva_cents`. For 625, `platform_activity` selects the ISR rate; it defaults to the taxpayer's only activity and stays `null` when there are several and none is sent. Reads are allowed for a read-only user and for an archived taxpayer. `regime` must be held by the taxpayer on `on` (`422 regime_not_active`).
  ///
  /// [on] - Date the dated parameters are read on (America/Mexico_City); defaults to today. `2022-01-01` … today (`out_of_range`).
  ///
  /// [platformActivity] - 625 only (`not_for_regime` otherwise); must be one of the taxpayer's activities (`activity_not_held`).
  @GET('/v1/taxpayers/{taxpayerId}/incomes/defaults')
  Future<IncomeDefaults> getIncomeDefaults({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('regime') required RegimeCode regime,
    @Query('counterparty_type') required CounterpartyType counterpartyType,
    @Query('on') String? onValue,
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
  /// Partial update with merge-patch semantics on top-level fields: an absent field is unchanged; `null` clears a nullable one (`invoiced_on`, `collected_on`, `platform_activity`, `cfdi_uuid`, `description`). `id` cannot be sent (`additionalProperties`). Every rule of `createIncome` is re-run on the merged income, with the same `details[].code` values and `warnings`; moving it to another `regime_code` re-checks the regime window. Setting `collected_on` on an accrual income clears `pending_collection`. A read-only user gets `402`; an archived taxpayer `422 taxpayer_archived`. Last write wins (no ETag in v1).
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
  /// Clears `deleted_at`. **Never consumes a trial slot**: the slot was used once at creation and never freed, so Undo works at 5/5 with no paywall. **Allowed for a read-only user** (no access check, never `402`): Undo gives back the user's own record and creates nothing (0.6.1, MODULE-incomes §2.7). The regime window is re-checked (`422 regime_not_active` when the regime no longer covers the income's dates); an archived taxpayer answers `422 taxpayer_archived`. Restoring a live income answers `200` unchanged; a purged one, or another user's, `404`. `warnings` are computed as on create.
  @POST('/v1/taxpayers/{taxpayerId}/incomes/{incomeId}/restore')
  Future<Income> restoreIncome({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('incomeId') required Uuid incomeId,
  });
}
