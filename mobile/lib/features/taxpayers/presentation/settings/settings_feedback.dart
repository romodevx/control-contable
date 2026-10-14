import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';

/// `details[].code` of a write on an archived taxpayer.
const String taxpayerArchivedDetail = 'taxpayer_archived';

/// Shows [message] in the app's toast (a `SnackBar`: a polite live region,
/// `role=status`). With [action] it stays until the user acts on it or
/// closes it (house rule: an Undo never auto-dismisses).
///
/// Takes the [messenger] (read before any await): the toast outlives the
/// screen that asked for it.
void showSettingsToast(
  ScaffoldMessengerState messenger,
  String message, {
  SnackBarAction? action,
}) {
  messenger
    ..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        content: Text(message),
        action: action,
        persist: action != null,
        showCloseIcon: action != null,
      ),
    );
}

/// The message for a settings write the server refused for a reason no
/// field owns.
String settingsFailureMessage(ApiFailure failure, AppLocalizations l10n) =>
    switch (failure) {
      ApiRejected(code: ErrorCode.businessRuleViolation, :final details)
          when details.any((d) => d.code == taxpayerArchivedDetail) =>
        l10n.settingsArchivedWrite,
      _ => describeNotice(noticeFor(failure), l10n),
    };

/// Opens the read-only message (`402`): why, what still works, and the
/// way out. Announced when it opens.
Future<void> showPaywall(BuildContext context) {
  final l10n = AppLocalizations.of(context);
  return showPaywallSheet(
    context,
    title: l10n.onboardingPaywallTitle,
    body: l10n.paywallWriteBody,
  );
}
