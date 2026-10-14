import 'package:control_contable/api/generated/models/error_code.dart';
import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/core/network/api_failure.dart';

/// Separator of the server's field paths (`regimes.1.options.rif_year`).
const String _pathSeparator = '.';

/// Why the server refused the RFC (its `details[].code`, or the `409`).
enum RfcRejection {
  /// `rfc_length_mismatch`.
  length,

  /// `rfc_format`, or any other schema verdict on the RFC.
  format,

  /// `rfc_date`.
  date,

  /// `rfc_generic`.
  generic,

  /// `409 rfc_already_registered`: the user already manages this RFC.
  duplicate,
}

/// A form-level message for the current onboarding screen.
enum OnboardingNotice {
  /// The server could not be reached.
  unreachable,

  /// `429`.
  rateLimited,

  /// `422 taxpayer_limit_reached` (10 per user, archived included).
  limitReached,

  /// `422 invalid_regime_combination`.
  combinationRejected,

  /// A `400` on a regime code.
  regimeRejected,

  /// Anything else.
  unexpected,
}

/// The rejected fields of a `400`, sorted by the step that owns them.
typedef OnboardingVerdict = ({
  RfcRejection? rfc,
  bool displayName,
  bool regime,
  Map<String, Set<String>> options,
});

const Map<String, RfcRejection> _rfcCodes = {
  'rfc_length_mismatch': RfcRejection.length,
  'rfc_format': RfcRejection.format,
  'rfc_date': RfcRejection.date,
  'rfc_generic': RfcRejection.generic,
};

/// Sorts [details] of a create whose body listed [sentRegimes] in order
/// (`regimes.N` indexes them).
OnboardingVerdict sortDetails(
  List<api.FieldError> details, {
  required List<Regime> sentRegimes,
}) {
  RfcRejection? rfc;
  var displayName = false;
  var regime = false;
  final options = <String, Set<String>>{};
  for (final detail in details) {
    final path = detail.field.split(_pathSeparator);
    switch (path) {
      case ['rfc', ...]:
        rfc ??= _rfcCodes[detail.code] ?? RfcRejection.format;
      case ['display_name', ...]:
        displayName = true;
      case ['regimes', final index, 'options', final key, ...]:
        final code = _regimeAt(sentRegimes, index);
        if (code == null) {
          regime = true;
        } else {
          (options[code] ??= {}).add(key);
        }
      case ['regimes', ...]:
        regime = true;
    }
  }
  return (rfc: rfc, displayName: displayName, regime: regime, options: options);
}

/// The RFC verdict of a failed RFC write, or `null` when [failure] is not
/// about the RFC.
RfcRejection? rfcRejectionOf(ApiFailure failure) => switch (failure) {
  ApiRejected(code: ErrorCode.rfcAlreadyRegistered) => RfcRejection.duplicate,
  ApiValidationFailure(:final details) => sortDetails(
    details,
    sentRegimes: const [],
  ).rfc,
  _ => null,
};

/// The form-level message for a [failure] no field owns.
OnboardingNotice noticeFor(ApiFailure failure) => switch (failure) {
  ApiUnreachable() => OnboardingNotice.unreachable,
  ApiRateLimited() => OnboardingNotice.rateLimited,
  ApiRejected(code: ErrorCode.taxpayerLimitReached) =>
    OnboardingNotice.limitReached,
  _ => OnboardingNotice.unexpected,
};

String? _regimeAt(List<Regime> regimes, String index) {
  final position = int.tryParse(index);
  if (position == null || position < 0 || position >= regimes.length) {
    return null;
  }
  return regimes[position].code;
}
