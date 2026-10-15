// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

import 'package:dio/dio.dart';
import 'package:retrofit/retrofit.dart';

import '../models/annual_breakdown.dart';
import '../models/annual_computed_section_kind.dart';
import '../models/annual_estimate.dart';
import '../models/annual_inputs.dart';
import '../models/annual_inputs_update.dart';
import '../models/deduction_category.dart';
import '../models/filed_payments.dart';
import '../models/filed_payments_update.dart';
import '../models/fiscal_year.dart';
import '../models/personal_deduction.dart';
import '../models/personal_deduction_create.dart';
import '../models/personal_deduction_list.dart';
import '../models/personal_deduction_update.dart';
import '../models/regime_code.dart';
import '../models/salary_statement.dart';
import '../models/salary_statement_create.dart';
import '../models/salary_statement_list.dart';
import '../models/salary_statement_update.dart';
import '../models/uuid.dart';

part 'annual_client.g.dart';

@RestApi()
abstract class AnnualClient {
  factory AnnualClient(Dio dio, {String? baseUrl}) = _AnnualClient;

  /// Annual return estimate (declaración anual) of one taxpayer and fiscal year.
  ///
  /// Roughly how much the taxpayer will owe, or get back, in the annual return of `fiscalYear`, and by when. Derived from the current records, annual inputs and tax inputs on every read, never stored (ADR-0002 §5, ADR-0009); same inputs, parameters and engine version → byte-identical JSON. It never files and never changes a monthly figure.
  ///
  /// **Sections** (MODULE-annual §2.2, fixed order `pf_art152`, `pf_resico_check`, `pm_601`, `pm_resico`, then one `explanation` per other regime held in the year, in catalog order). A persona física's 605, 606, 612 and 625 non-definitive incomes are aggregated in **one** `pf_art152` section (LISR 152); RESICO PF (626) is an informative 113-F check that never has a payable or refundable balance (rule 3.13.7); PM 601 and RESICO PM are estimated per regime; every other regime is an `explanation` section with its deadline and no figure. A regime counts when its window intersects the fiscal year, archived regimes included for their active months. `headline_sections` names the sections whose `tax.balance` is the headline: PF the `pf_art152` section, PM its 601 / RESICO PM sections (each shown on its own, never netted); informational sections never feed it.
  ///
  /// **Status per section** (precedence `informational` > `incomplete` > `complete`): `incomplete` is computed with the documented fallback and lists `missing_inputs` (whose `target` is the screen the "Completa tus datos" CTA opens). A month with no filed payment uses the app's own monthly estimate (`payments_estimated` disclaimer), which is **not** missing unless that month's estimate was itself incomplete.
  ///
  /// **Year in progress.** When `fiscalYear` is the current year (America/Mexico_City), `year_status` is `in_progress`, every figure uses the records through `data_through` (the last closed month, `null` in January) and the `year_in_progress` disclaimer is present; December is never extrapolated (OQ-A3).
  ///
  /// `legend` is the mandatory constant; the client renders the server field. Rules a CPA has not validated still run with their documented default, set `cpa_review_pending`, list `cpa_topics` and add the `cpa_review_pending` disclaimer (severity `strong`). Reads are allowed for a read-only user and for an archived taxpayer (no `402`). Budget: `estimate_reads` (ADR-0005), one read plus one more when the 12-month estimate range is needed for default payments; spent → `429` with `Retry-After`.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscalYear` → `out_of_range` (a year after the current one). `422 business_rule_violation` details: `fiscalYear` → `parameters_unavailable` (before 2026, the first fiscal year with transcribed parameters).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @GET('/v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}')
  Future<AnnualEstimate> getAnnualEstimate({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('fiscalYear') required FiscalYear fiscalYear,
  });

  /// Step-by-step computation of one annual section, with legal basis.
  ///
  /// The computation trace of `section` in `fiscalYear` (ADR-0002 §1: strategies compute only by appending lines, so the explanation cannot drift from the math). `fiscalYear` follows the rules of `getAnnualEstimate`. `lines` are in `seq` order; every line except `info` and `rounding_adjustment` cites at least one legal basis, and `source` says where its figure comes from (each provisional payment line is `filed_payment`, `tax_input:prior_provisional_payments` or `estimated_by_app`, §2.5). Every cap that binds is an `info` line "Tope aplicado: {concepto}, de {x} a {y}"; every excluded deduction (cash in a restricted category, donations without the prior year's income) is an `info` line too. One `rounding_adjustment` line reconciles displayed lines with a rounded result, never a hidden centavo. `disclaimers` are the estimate's disclaimers that bear on this section, same order. Reads are allowed for a read-only user and an archived taxpayer. Budget: `estimate_reads`, as `getAnnualEstimate`.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscalYear` → `out_of_range` (a future year). `422 business_rule_violation` details: `fiscalYear` → `parameters_unavailable` (before 2026); `section` → `section_not_present` (the estimate of that year has no such section).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  ///
  /// [section] - A computed section kind of the estimate (`explanation` sections have no trace).
  @GET('/v1/taxpayers/{taxpayerId}/annual-estimates/{fiscalYear}/breakdown')
  Future<AnnualBreakdown> getAnnualBreakdown({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('fiscalYear') required FiscalYear fiscalYear,
    @Query('section') required AnnualComputedSectionKind section,
  });

  /// Per-year annual facts of a taxpayer, plus the annual tax inputs of its regimes (read-only).
  ///
  /// The taxpayer's annual facts for `fiscalYear` (MODULE-annual §2.6) and, read-only, the annual-relevant M5 tax inputs of every regime held in that year (`regime_tax_inputs`: `prior_year_losses`, `ptu_paid`, `annual_depreciation`, `interest_withholding`, as the regime's catalog defines them). Those stay owned by `getTaxInputs` / `putTaxInputs` (one source per fact, ADR-0009); the app links each to that screen. A year with nothing saved answers the defaults with `updated_at: null`. Reads are allowed for a read-only user and an archived taxpayer. Budget: `annual_reads` (600 per hour, ADR-0005).
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscalYear` → `out_of_range` (a future year). `422 business_rule_violation` details: `fiscalYear` → `parameters_unavailable` (before 2026).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @GET('/v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}')
  Future<AnnualInputs> getAnnualInputs({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('fiscalYear') required FiscalYear fiscalYear,
  });

  /// Replace the per-year annual facts of a persona física.
  ///
  /// Replaces the taxpayer's annual facts for `fiscalYear` (idempotent: the same body twice leaves the same state). `prior_year_accumulable_income_cents: null` means unknown, not zero: donations are then excluded and the section is `incomplete` (OQ-A5). The regime tax inputs are not written here (`putTaxInputs`).
  ///
  /// **Order of checks:** schema (`400`) → read-only access (`402`, OQ-A9) → archived taxpayer, person type and fiscal year (`422`) → upsert. Never uses a trial record slot. Budget: `annual_writes` (300 per hour, ADR-0005). Audited as `annual_inputs.updated` with user, taxpayer and fiscal year only, never the values.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscalYear` → `out_of_range` (a future year). `422 business_rule_violation` details: `taxpayerId` → `taxpayer_archived` or `person_type_not_allowed` (a persona moral: these facts are personas físicas' only); `fiscalYear` → `parameters_unavailable`.
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @PUT('/v1/taxpayers/{taxpayerId}/annual-inputs/{fiscalYear}')
  Future<AnnualInputs> putAnnualInputs({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('fiscalYear') required FiscalYear fiscalYear,
    @Body() required AnnualInputsUpdate body,
  });

  /// Salary constancias (605) of one fiscal year, with totals.
  ///
  /// Every salary constancia of the taxpayer for `fiscal_year` (at most 20, so no pagination), ordered by `created_at`, then `id`. `totals` sum them. A persona moral answers `items: []`. Reads are allowed for a read-only user and an archived taxpayer. Budget: `annual_reads`.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscal_year` → `out_of_range` (a future year). `422 business_rule_violation` details: `fiscal_year` → `parameters_unavailable` (before 2026).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @GET('/v1/taxpayers/{taxpayerId}/salary-statements')
  Future<SalaryStatementList> listSalaryStatements({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('fiscal_year') required FiscalYear fiscalYear,
  });

  /// Record a salary constancia (idempotent client id).
  ///
  /// Records one employer's constancia de sueldos for a fiscal year, as its totals (accumulable, exempt, ISR withheld). The app never recomputes art. 93 exemptions or the employment subsidy, and no employer RFC is taken (`employer_label` is an optional alias). Allowed for any persona física, also without regime 605 (OQ-A2): then the warning `salary_regime_not_held` on `fiscal_year` suggests adding it.
  ///
  /// **Idempotency (double submit)** as `createIncome`: the client sends a UUID `id` once per form session; a retry with the same fields → `200` with the stored row; the same `id` with other fields or another taxpayer → `409` (`id` / `id_reused`); an `id` of another user → `409` with no `details`.
  ///
  /// **Order of checks:** schema (`400`) → idempotency (`200`/`409`) → read-only access (`402`, OQ-A9; never a trial slot) → archived taxpayer, person type, fiscal year, list bound (`422`) → insert. Budget: `annual_writes`. Figures and labels are never logged.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscal_year` → `out_of_range` (a future year); `employer_label` → `invalid` (blank, untrimmed or control characters; stored NFC, ADR-0004). `422 business_rule_violation` details: `taxpayerId` → `taxpayer_archived` or `person_type_not_allowed`; `fiscal_year` → `parameters_unavailable`; `body` → `limit_reached` (20 constancias in that year).
  @POST('/v1/taxpayers/{taxpayerId}/salary-statements')
  Future<SalaryStatement> createSalaryStatement({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required SalaryStatementCreate body,
  });

  /// Edit a salary constancia (merge-patch; null clears the label).
  ///
  /// Merge-patch on top-level fields: absent = unchanged, `employer_label: null` clears it. `id` cannot be sent. Every rule of `createSalaryStatement` is re-run on the merged row, with the same codes and warnings. A read-only user gets `402`; an archived taxpayer `422 taxpayer_archived`. Last write wins. Budget: `annual_writes`.
  @PATCH('/v1/taxpayers/{taxpayerId}/salary-statements/{statementId}')
  Future<SalaryStatement> updateSalaryStatement({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('statementId') required Uuid statementId,
    @Body() required SalaryStatementUpdate body,
  });

  /// Delete a salary constancia.
  ///
  /// Deletes the constancia (hard delete; the client confirms first, it cannot be undone). Deleting an unknown or already deleted constancia answers `404`. Allowed for a read-only user (removing their own data, as records). An archived taxpayer answers `422 taxpayer_archived`. Budget: `annual_writes`.
  @DELETE('/v1/taxpayers/{taxpayerId}/salary-statements/{statementId}')
  Future<void> deleteSalaryStatement({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('statementId') required Uuid statementId,
  });

  /// Live personal deductions of one fiscal year, with registered totals per category.
  ///
  /// Live (not deleted) personal deductions whose `paid_on` falls in `fiscal_year` (at most 500, so no pagination), ordered by category (the `DeductionCategory` order), then `paid_on` descending, then `id`. `category` narrows items and totals. `totals` are the **registered** amounts per category (rows excluded for cash included); the **applied** amounts after caps are in the annual estimate's breakdown. A persona moral answers `items: []`. `warnings` is always `[]` here. Reads are allowed for a read-only user and an archived taxpayer. Budget: `annual_reads`.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscal_year` → `out_of_range` (a future year). `422 business_rule_violation` details: `fiscal_year` → `parameters_unavailable` (before 2026).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @GET('/v1/taxpayers/{taxpayerId}/personal-deductions')
  Future<PersonalDeductionList> listPersonalDeductions({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Query('fiscal_year') required FiscalYear fiscalYear,
    @Query('category') DeductionCategory? category,
  });

  /// Record a personal deduction (idempotent client id).
  ///
  /// Records one personal deduction of a persona física (LISR 151 fr. I–VIII and the colegiaturas decree). The fiscal year is the year of `paid_on`. Category-specific fields: `donee_type` is required for `donation` and must be absent or `null` otherwise; `student_label`, `tuition_level` and `scholarship_cents` are required for `tuition` and absent or `null` otherwise. `cash` is accepted and stored, and for `medical`, `medical_disability`, `school_transport` and `tuition` it is **not deductible**: the row is saved with `excluded: cash_not_deductible` and the warning `cash_not_deductible` on `payment_method`, and the estimate shows it excluded. A `beneficiary` other than `self` means the user attested that the dependant's income is below one UMA anual (the form's checkbox; not verified ⚠️). `cfdi_uuid` is format-checked only and never sent to the SAT.
  ///
  /// **Idempotency (double submit)** as `createIncome`: retry with the same `id` and fields → `200`; the same `id` with other fields, another taxpayer or a deleted row → `409` (`id` / `id_reused`); an `id` of another user → `409` with no `details`.
  ///
  /// **Order of checks:** schema (`400`) → idempotency (`200`/`409`) → read-only access (`402`, OQ-A9; never a trial slot) → archived taxpayer, person type, list bound (`422`) → field rules (`400`) → insert. Budget: `annual_writes`. Amounts, labels and the beneficiary are never logged.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `paid_on` → `out_of_range` (before 2026-01-01 or after today, America/Mexico_City, as `BusinessDate`); `donee_type`, `student_label`, `tuition_level`, `scholarship_cents` → `required` (missing for their category) or `not_applicable` (sent for another category); `student_label` → `invalid` (blank, untrimmed or control characters; stored NFC). `422 business_rule_violation` details: `taxpayerId` → `taxpayer_archived` or `person_type_not_allowed`; `body` → `limit_reached` (500 live deductions in that year).
  @POST('/v1/taxpayers/{taxpayerId}/personal-deductions')
  Future<PersonalDeduction> createPersonalDeduction({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Body() required PersonalDeductionCreate body,
  });

  /// Edit a personal deduction (merge-patch; null clears a nullable field).
  ///
  /// Merge-patch on top-level fields: absent = unchanged, `null` clears a nullable one (`donee_type`, `student_label`, `tuition_level`, `scholarship_cents`, `cfdi_uuid`). `id` cannot be sent. Every rule of `createPersonalDeduction` is re-run on the merged row (changing `category` must also send or clear its category fields), with the same codes and `warnings`; moving `paid_on` to another year moves the row to that fiscal year. A deleted row answers `404`. A read-only user gets `402`; an archived taxpayer `422 taxpayer_archived`. Last write wins. Budget: `annual_writes`.
  @PATCH('/v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}')
  Future<PersonalDeduction> updatePersonalDeduction({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('deductionId') required Uuid deductionId,
    @Body() required PersonalDeductionUpdate body,
  });

  /// Soft-delete a personal deduction (Undo via restore).
  ///
  /// Sets `deleted_at`: the deduction leaves the list and the estimate at once and can be restored for 30 days (then it is purged), so the app offers Undo instead of a confirmation. Deleting an already deleted deduction answers `204`. Allowed for a read-only user. An archived taxpayer answers `422 taxpayer_archived`. Budget: `annual_writes`.
  @DELETE('/v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}')
  Future<void> deletePersonalDeduction({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('deductionId') required Uuid deductionId,
  });

  /// Undo a personal-deduction delete.
  ///
  /// Clears `deleted_at`. Allowed for a read-only user (no access check, never `402`): Undo gives back the user's own row and creates nothing, as `restoreIncome`. The list bound is not re-checked. Restoring a live deduction answers `200` unchanged; a purged one, or another user's, `404`. An archived taxpayer answers `422 taxpayer_archived`. `warnings` are computed as on create. Budget: `annual_writes`.
  @POST('/v1/taxpayers/{taxpayerId}/personal-deductions/{deductionId}/restore')
  Future<PersonalDeduction> restorePersonalDeduction({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('deductionId') required Uuid deductionId,
  });

  /// The 12 provisional ISR payments of one regime and year, with their source.
  ///
  /// One row per month (January .. December) with the provisional ISR payment the annual estimate credits and where it comes from (MODULE-annual §2.5, OQ-A1): **(1)** `filed_cents`, the amount the user confirmed or typed here → **(2)** `tax_input_cents`, the M5 tax input `prior_provisional_payments` of that month → **(3)** `app_estimate_cents`, the app's own monthly estimate for the regime (`max(0, isr)`). `effective_cents` and `source` are the first present. A month outside the regime's window, or not closed yet in the current year, is `in_window: false` or carries `source: none`. The screen prefills the 12 inline rows with `effective_cents`; "Confirmar todos" sends them to `putFiledPayments`.
  ///
  /// Only regimes with provisional ISR payments: 606, 612, 625 (without the definitive option), 626, 601. Reads are allowed for a read-only user and an archived taxpayer. Budget: `estimate_reads` (one read plus one for the 12-month estimate range).
  ///
  /// `FieldError.code` values beyond JSON Schema keywords: `fiscalYear` → `out_of_range` (a future year). `422 business_rule_violation` details: `fiscalYear` → `parameters_unavailable`; `regimeCode` → `regime_not_active` (not held in any month of the year) or `payments_not_applicable` (a held regime without provisional ISR payments, e.g. 605 or 625 with the definitive option).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @GET(
    '/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}',
  )
  Future<FiledPayments> getFiledPayments({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Path('fiscalYear') required FiscalYear fiscalYear,
  });

  /// Replace the confirmed provisional ISR payments of one regime and year.
  ///
  /// Replaces **all** confirmed months of `(taxpayer, regime, fiscalYear)` with `months` (an empty array clears them; an absent month is "not confirmed", so the estimate falls back to the tax input or the app estimate). Idempotent. At most one entry per month; `amount_cents >= 0` is the amount actually paid (`0` is a confirmed zero). Answers the same body as `getFiledPayments`.
  ///
  /// **Order of checks:** schema (`400`) → read-only access (`402`, OQ-A9) → archived taxpayer, fiscal year, regime and window (`422`) → month rules (`400`) → replace. Budget: `annual_writes` plus the `estimate_reads` of the answer. Audited as `annual_filed_payments.updated` with user, taxpayer, regime and fiscal year only.
  ///
  /// `FieldError.code` values beyond JSON Schema keywords (`field` is `months.<index>[.<property>]`): `invalid` (a repeated month); `out_of_range` (a month not closed yet in the current year); `fiscalYear` → `out_of_range` (a future year). `422 business_rule_violation` details: `taxpayerId` → `taxpayer_archived`; `fiscalYear` → `parameters_unavailable`; `regimeCode` → `regime_not_active` or `payments_not_applicable`; `months.<index>.month` → `regime_not_active` (a month outside the regime's window).
  ///
  /// [fiscalYear] - Fiscal year (2026 .. the current year, America/Mexico_City). Before 2026 → `422`; after the current year → `400 out_of_range`.
  @PUT(
    '/v1/taxpayers/{taxpayerId}/regimes/{regimeCode}/filed-payments/{fiscalYear}',
  )
  Future<FiledPayments> putFiledPayments({
    @Path('taxpayerId') required Uuid taxpayerId,
    @Path('regimeCode') required RegimeCode regimeCode,
    @Path('fiscalYear') required FiscalYear fiscalYear,
    @Body() required FiledPaymentsUpdate body,
  });
}
