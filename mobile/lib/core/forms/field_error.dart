import 'package:control_contable/l10n/app_localizations.dart';

/// Why a field's value is not accepted. Kept as data, not text, so the
/// validation policy runs without a `BuildContext` and the copy stays in the
/// ARB files.
sealed class FieldError {
  const new();
}

/// A required field is empty.
final class MissingValue extends FieldError {
  /// The error.
  const new();
}

/// The text is not a peso amount (see `parseCentavos`).
final class InvalidMoney extends FieldError {
  /// The error.
  const new();
}

/// The text is not a real calendar date (see `parseFlexibleDate`).
final class InvalidDate extends FieldError {
  /// The error.
  const new();
}

/// A screen-specific rule, with copy the screen already localized.
final class CustomFieldError extends FieldError {
  /// The error with its user-facing [message].
  const new(this.message);

  /// What the user reads.
  final String message;
}

/// Checks a non-empty, trimmed value; `null` when it is accepted.
typedef FieldValidator = FieldError? Function(String value);

/// The user-facing copy of [error].
String describeFieldError(FieldError error, AppLocalizations l10n) =>
    switch (error) {
      MissingValue() => l10n.formErrorRequired,
      InvalidMoney() => l10n.formErrorMoney,
      InvalidDate() => l10n.formErrorDate,
      CustomFieldError(:final message) => message,
    };
