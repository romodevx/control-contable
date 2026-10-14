// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/period.dart';
import '../models/person_type.dart';
import '../models/regime_code.dart';
import '../models/regime_list.dart';
import '../models/taxpayer.dart';
import '../models/taxpayer_create.dart';
import '../models/taxpayer_list.dart';
import '../models/taxpayer_regime.dart';
import '../models/taxpayer_regime_create.dart';
import '../models/taxpayer_regime_update.dart';
import '../models/taxpayer_update.dart';
import '../models/uuid.dart';

part 'taxpayers_client.g.dart';

@RestApi()
abstract class TaxpayersClient {
  factory TaxpayersClient(Dio dio, {String? baseUrl}) = _TaxpayersClient;

  /// Regime catalog for one person type (support level, options, inputs, compatibility).
  ///
  /// Every SAT `c_RegimenFiscal` key selectable for `person_type` (catalog key is `(code, person_type)`: 626 and 610 exist for both). `treatment` is the support level: `full`, `partial` or `informational` (an informational or partial regime carries a plain-language `explanation`, shown at selection time). `option_defs` drive the per-regime options form and their server-side validation; `required_inputs` are the per-fiscal-year tax-input keys the estimate needs (M5, not options); `incompatible_with` lists the regimes it cannot overlap in time, with the reason. Global, read-only data: a read-only user may list it.
  @GET('/v1/regimes')
  Future<RegimeList> listRegimes({
    @Query('person_type') required PersonType personType,
  });

  /// Taxpayers managed by the user (feeds the active-taxpayer switcher).
  ///
  /// The caller's taxpayers ordered by `created_at`, then `id`. Archived taxpayers are left out unless `include_archived=true`. At most 10 exist per user (OD-2), so the list is not paginated.
  @GET('/v1/taxpayers')
  Future<TaxpayerList> listTaxpayers({
    @Query('include_archived') bool? includeArchived = false,
  });

  /// Create a taxpayer with its regimes and options in one transaction (onboarding).
  ///
  /// Onboarding order (OD-3): person type, then RFC and display name, then regimes, then per-regime options. The taxpayer, every regime and every option are written in one transaction: on any failure nothing is persisted. `regimes` is required (at least one); year-to-date tax inputs are not part of this call (M5).
  ///
  /// The RFC is normalized (trim, upper-case, NFC, inner spaces and `-` removed), then validated: length for the person type (13 física, 12 moral), alphabet, real calendar date; the generic RFCs `XAXX010101000` and `XEXX010101000` are refused. The SAT módulo 11 check digit does NOT reject (D-RFC: some SAT-issued RFCs fail the published algorithm): a mismatch is accepted and the response carries the non-blocking warning `{code: rfc_check_digit_mismatch, field: rfc}` so the app asks the user to double-check. It is never checked against the SAT. The full RFC is never returned or echoed in an error: the response carries only `rfc_masked`.
  ///
  /// Order of checks: schema (400) → access policy (`402 entitlement_required` for a read-only user) → idempotent retry → cap → business rules. Retrying with a client `id` the caller already owns answers `200` with the existing taxpayer, even at the cap or for a read-only user; an `id` that exists outside the caller's scope answers `409 conflict`. The cap is 10 taxpayers per user for every plan, archived included, counted under a row lock so concurrent creates cannot overshoot (`422 taxpayer_limit_reached`, not 402: paying does not lift it).
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `rfc` → `rfc_length_mismatch`, `rfc_format`, `rfc_date`, `rfc_generic`; `regimes.N.regime_code` → `not_for_person_type`, `duplicate` (same code twice in the body); `regimes.N.options.<key>` → `unknown_option`, `invalid_option`, `required`; `regimes.N.start_period` → `out_of_range` (allowed `2026-01`, the first fiscal year with tax parameters, … the current month, America/Mexico_City).
  ///
  /// Regime compatibility (`regime_compatibility`): an overlapping pair the catalog marks not allowed answers `422 invalid_regime_combination`; an overlapping pair whose verdict still needs CPA review is allowed and the response carries one warning per such pair, `{code: regime_combination_needs_cpa_review, field: regimes.N.regime_code}` (N = the later of the two in the body), with the pair's `reason` as `message`.
  @POST('/v1/taxpayers')
  Future<Taxpayer> createTaxpayer({@Body() required TaxpayerCreate body});

  /// Get a taxpayer (archived ones too)
  @GET('/v1/taxpayers/{taxpayerId}')
  Future<Taxpayer> getTaxpayer({@Path('taxpayerId') required Uuid taxpayerId});

  /// Rename a taxpayer or correct its RFC (person type is immutable).
  ///
  /// `display_name` is trimmed and NFC-normalized (1–80 characters). `rfc` corrects a mistyped RFC of the same person type (OQ-1): it is normalized and validated as in `createTaxpayer` (same `details[].code` values; a check-digit mismatch is accepted with the `rfc_check_digit_mismatch` warning), re-encrypted, re-hashed and checked for uniqueness in the caller's account. `person_type` or any other field answers `400` (`additionalProperties`). A read-only user gets `402`; an archived taxpayer `422 business_rule_violation` (`taxpayer_archived`).
  @PATCH('/v1/taxpayers/{taxpayerId}')
  Future<Taxpayer> updateTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required TaxpayerUpdate body,
  });

  /// Permanently delete a taxpayer and all its records (irreversible).
  ///
  /// Requires `confirm=true`; without it answers `400 validation_failed` (`confirm` / `confirm_required`) and deletes nothing. Deletes the taxpayer, its regimes, incomes, expenses, tax inputs and estimate history in one transaction. Frees one cap slot and the RFC; trial usage counters are not decremented. Allowed for a read-only user and for an archived taxpayer.
  ///
  /// [confirm] - Must be `true` to perform a destructive or data-affecting action.
  @DELETE('/v1/taxpayers/{taxpayerId}')
  Future<void> deleteTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('confirm') bool? confirm = false,
  });

  /// Archive (hide) a taxpayer, keeping its data.
  ///
  /// Sets `archived_at`. The taxpayer leaves the list and the switcher; reads by id and past estimates still work; writes are refused until it is unarchived. It still counts toward the cap and keeps its RFC reserved. Archiving an archived taxpayer answers `200` unchanged. Allowed for a read-only user.
  @POST('/v1/taxpayers/{taxpayerId}/archive')
  Future<Taxpayer> archiveTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
  });

  /// Unarchive a taxpayer (it returns to the list and the switcher).
  ///
  /// Clears `archived_at` (OQ-2). No cap check: an archived taxpayer already counts. Unarchiving an active taxpayer answers `200` unchanged. A read-only user gets `402`.
  @POST('/v1/taxpayers/{taxpayerId}/unarchive')
  Future<Taxpayer> unarchiveTaxpayer({
    @Path('taxpayerId') required Uuid taxpayerId,
  });

  /// Add a regime to a taxpayer (or reopen an ended one).
  ///
  /// `start_period` defaults to the current month (America/Mexico_City); allowed `2026-01` (the first fiscal year with tax parameters) … the current month (`out_of_range`). Options are validated against the regime's `option_defs` (omitted options with a default are filled; same `details[].code` values as `createTaxpayer`, with `options.<key>` paths). A code not in the catalog for the taxpayer's person type answers `400` (`regime_code` / `not_for_person_type`). A regime whose period window overlaps an incompatible one answers `422 invalid_regime_combination`; regimes that do not overlap in time never conflict; an overlapping pair whose compatibility still needs CPA review is allowed and the response carries the warning `{code: regime_combination_needs_cpa_review, field: regime_code}` (one per such pair, `message` = the pair's `reason`). Adding a regime that is already active answers `409 conflict`; adding an ended one reopens it (`end_period` cleared, original `start_period` kept, body options applied) and answers `201`. A read-only user gets `402`; an archived taxpayer `422 business_rule_violation` (`taxpayer_archived`).
  @POST('/v1/taxpayers/{taxpayerId}/regimes')
  Future<TaxpayerRegime> addTaxpayerRegime({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required TaxpayerRegimeCreate body,
  });

  /// Change a regime's options (applies to every period of the regime).
  ///
  /// Merge semantics: each key sent replaces its value, keys not sent keep theirs. Validated against the regime's `option_defs` (`options.<key>` / `unknown_option`, `invalid_option`). Options are not versioned in time (OQ-7): the change applies to every period of the regime. A regime the taxpayer does not hold answers `404`. A read-only user gets `402`; an archived taxpayer `422 business_rule_violation` (`taxpayer_archived`).
  @PATCH('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}')
  Future<TaxpayerRegime> updateTaxpayerRegime({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Body() required TaxpayerRegimeUpdate body,
  });

  /// End a regime at `end_period` (records are kept; never a hard delete).
  ///
  /// Sets `end_period`: `start_period` ≤ `end_period` ≤ the current month (`end_period` / `out_of_range`), and not before the latest period holding a record in the regime (`422`, `records_after_end_period`). Records are kept and still count for periods up to `end_period`. When the regime has records and `confirm` is not `true` the call answers `409 regime_has_records` and changes nothing. The last active regime cannot be ended (`422`, `last_active_regime`): archive the taxpayer instead. A regime already ended answers `409 conflict`. Allowed for a read-only user; an archived taxpayer answers `422` (`taxpayer_archived`).
  ///
  /// [confirm] - Must be `true` to perform a destructive or data-affecting action.
  @DELETE('/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}')
  Future<TaxpayerRegime> archiveTaxpayerRegime({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Query('end_period') required Period endPeriod,
    @Query('confirm') bool? confirm = false,
  });
}
