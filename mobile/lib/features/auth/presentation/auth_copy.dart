import 'dart:math' as math;

import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/l10n/app_localizations.dart';

const int _secondsPerMinute = 60;

/// The form-level es-MX message for [failure]. Field verdicts
/// (`ApiValidationFailure` details) are shown on their fields instead.
String describeAuthFailure(ApiFailure failure, AppLocalizations l10n) =>
    switch (failure) {
      ApiRateLimited(:final retryAfter?) => l10n.authErrorRateLimited(
        _minutesRoundedUp(retryAfter),
      ),
      ApiRateLimited() => l10n.authErrorRateLimitedNoWait,
      ApiUnreachable() => l10n.authErrorUnreachable,
      ApiNotFound() => l10n.authErrorUnavailable,
      ApiValidationFailure() => l10n.authFormRejected,
      ApiUnauthorized() || ApiUnexpected() => l10n.authErrorUnexpected,
    };

/// A countdown as `m:ss` (`1:00`, `0:45`, `12:05`).
String formatCountdown(Duration remaining) {
  final seconds = remaining.inSeconds % _secondsPerMinute;
  return '${remaining.inMinutes}:${seconds.toString().padLeft(2, '0')}';
}

int _minutesRoundedUp(Duration wait) =>
    math.max(1, (wait.inSeconds / _secondsPerMinute).ceil());
