import 'dart:async';

import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/access/paywall_sheet.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/features/tax_engine/application/tax_inputs.dart';
import 'package:control_contable/features/tax_engine/presentation/tax_input_fields.dart';
import 'package:control_contable/features/tax_engine/tax_engine_routes.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// The form-level message of a save that did not go through.
String taxInputsNoticeText(TaxInputsSaveNotice notice, AppLocalizations l10n) =>
    switch (notice) {
      TaxInputsSaveNotice.unreachable => l10n.estimateUnreachable,
      TaxInputsSaveNotice.archived => l10n.taxInputsArchived,
      TaxInputsSaveNotice.regimeNotActive => l10n.taxInputsRegimeNotActive,
      TaxInputsSaveNotice.parametersUnavailable =>
        l10n.estimateParametersUnavailable,
      TaxInputsSaveNotice.gone => l10n.estimateNotFound,
      TaxInputsSaveNotice.failed => l10n.taxInputsSaveFailed,
    };

/// The copy of one server verdict on a value (the server's English
/// `message` is never shown).
FieldError describeTaxInputDetail(
  api.FieldError detail,
  AppLocalizations l10n,
) => CustomFieldError(switch (detail.code) {
  'unknown_input' => l10n.taxInputsUnknownInput,
  'out_of_range' ||
  'maximum' => l10n.onboardingRateOutOfRange(rateMin, rateMax),
  'minimum' => l10n.taxInputsNegative,
  _ => l10n.taxInputsInvalid,
});

/// Asks whether to drop what was typed; `true` to leave.
Future<bool> confirmDiscardTaxInputs(BuildContext context) async {
  final l10n = AppLocalizations.of(context);
  final discard = await showDialog<bool>(
    context: context,
    builder: (dialogContext) => AlertDialog(
      title: Text(l10n.taxInputsDiscardTitle),
      content: Text(l10n.taxInputsDiscardBody),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(dialogContext).pop(false),
          child: Text(l10n.taxInputsDiscardKeep),
        ),
        FilledButton(
          onPressed: () => Navigator.of(dialogContext).pop(true),
          child: Text(l10n.taxInputsDiscardConfirm),
        ),
      ],
    ),
  );
  return discard ?? false;
}

/// The paywall over the form (what was typed stays): a read-only account
/// can see its inputs but not change them.
void showTaxInputsPaywall(BuildContext context) {
  final l10n = AppLocalizations.of(context);
  unawaited(
    showPaywallSheet(
      context,
      title: l10n.onboardingPaywallTitle,
      body: l10n.paywallWriteBody,
    ),
  );
}

/// Confirms a save; the estimate behind already reads the new inputs.
void showTaxInputsSavedToast(BuildContext context) =>
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(content: Text(AppLocalizations.of(context).taxInputsSaved)),
      );

/// Back to where the form was opened from (the estimate or a breakdown),
/// or to the estimate when it was opened by a link.
void leaveTaxInputs(BuildContext context) {
  final navigator = Navigator.of(context);
  if (navigator.canPop()) {
    navigator.pop();
  } else {
    context.go(TaxEngineRoutes.estimate);
  }
}

/// Routes every way out of the form (system back, the app bar's back)
/// through [onLeave], which asks first when there are unsaved changes.
class TaxInputsLeaveGuard extends StatelessWidget {
  /// Guards [child].
  const new({required this.onLeave, required this.child, super.key});

  /// Leaves, asking first when needed.
  final Future<void> Function() onLeave;

  /// The page.
  final Widget child;

  @override
  Widget build(BuildContext context) => PopScope(
    canPop: false,
    onPopInvokedWithResult: (didPop, _) {
      if (!didPop) unawaited(onLeave());
    },
    child: child,
  );
}
