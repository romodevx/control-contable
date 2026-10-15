import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/records/record_save.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Annual writes never consume a trial slot (OQ-A9): no `402` names a cap.
const String _noTrialCap = '';

/// Runs an annual write and names how it ended; never throws an
/// `ApiFailure`. A `404` becomes [notFound].
Future<RecordSaveOutcome<T>> annualSaveOutcomeOf<T>(
  Future<T> Function() write, {
  RecordSaveNotice notFound = RecordSaveNotice.failed,
}) => recordSaveOutcomeOf(write, notFound: notFound, trialCap: _noTrialCap);

/// Opens the read-only paywall (its data stays behind it).
void showAnnualPaywall(BuildContext context) {
  final l10n = AppLocalizations.of(context);
  unawaited(
    showPaywallSheet(
      context,
      title: l10n.onboardingPaywallTitle,
      body: l10n.paywallWriteBody,
    ),
  );
}

/// When the session already knows the user is read-only, opens the paywall
/// and answers `true`: the write is not attempted (UX only; the server's
/// `402` stays the boundary).
bool blockedAsReadOnly(BuildContext context, WidgetRef ref) {
  if (!ref.read(readOnlyAccessProvider)) return false;
  showAnnualPaywall(context);
  return true;
}

/// The words of a form-level [notice].
String annualSaveNoticeText(RecordSaveNotice notice, AppLocalizations l10n) =>
    switch (notice) {
      RecordSaveNotice.unreachable => l10n.annualWriteUnreachable,
      RecordSaveNotice.idReused => l10n.annualWriteIdReused,
      RecordSaveNotice.gone => l10n.annualWriteGone,
      RecordSaveNotice.failed => l10n.annualWriteFailed,
    };

/// The user-facing error of one server [detail]: [limitReached] for the
/// list bound, [outOfRange] for a date or year out of range.
FieldError describeAnnualDetail(
  api.FieldError detail,
  AppLocalizations l10n, {
  required String limitReached,
  String? outOfRange,
}) => CustomFieldError(switch (detail.code) {
  'limit_reached' => limitReached,
  'out_of_range' => outOfRange ?? l10n.annualFutureYear,
  'taxpayer_archived' => l10n.settingsArchivedWrite,
  'person_type_not_allowed' => l10n.annualWritePersonType,
  'parameters_unavailable' => l10n.annualParametersUnavailable,
  'regime_not_active' => l10n.annualWriteRegimeNotActive,
  'invalid' => l10n.annualWriteInvalidText,
  _ => l10n.annualWriteFailed,
});

/// Why a delete or restore was refused, in words.
String refusalText(
  ApiFailure failure,
  AppLocalizations l10n, {
  required String fallback,
}) => switch (failure) {
  ApiRejected(:final details)
      when details.any((d) => d.code == 'taxpayer_archived') =>
    l10n.settingsArchivedWrite,
  _ => fallback,
};
