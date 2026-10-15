import 'package:control_contable/api/generated/models/field_error.dart' as api;
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/server_field_errors.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/annual/presentation/annual_write_feedback.dart';
import 'package:control_contable/features/auth/presentation/auth_copy.dart';
import 'package:control_contable/features/expenses/presentation/expense_copy.dart';
import 'package:control_contable/features/incomes/presentation/income_form_copy.dart';
import 'package:control_contable/features/taxpayers/application/onboarding_failures.dart';
import 'package:control_contable/features/taxpayers/presentation/onboarding_copy.dart';
import 'package:control_contable/l10n/app_localizations_es.dart';
import 'package:flutter_test/flutter_test.dart';

// Contract 0.15.0: every user-text field answers `400 validation_failed`
// with the detail code `invisible_characters` (ADR-0004 Amendment 1). Each
// form must say so on the field, never a misleading "too long" or a
// generic "could not save".
void main() {
  final l10n = AppLocalizationsEs();
  final expected = isA<CustomFieldError>().having(
    (e) => e.message,
    'message',
    l10n.fieldInvisibleCharacters,
  );

  api.FieldError invisible(String field) => api.FieldError(
    field: field,
    code: invisibleCharactersCode,
    message: 'Remove invisible or formatting characters.',
  );

  test('the code is the contract one', () {
    expect(invisibleCharactersCode, 'invisible_characters');
  });

  test('the copy names the problem in es-MX', () {
    expect(l10n.fieldInvisibleCharacters, contains('invisibles'));
  });

  final describers = <String, FieldError Function(api.FieldError)>{
    'sign-up name': (d) => describeSignUpDetail(d, l10n),
    'profile name': (d) => describeProfileNameDetail(d, l10n),
    'income description': (d) => describeIncomeDetail(d, l10n),
    'expense description': (d) => describeExpenseDetail(d, l10n),
    'annual label': (d) => describeAnnualDetail(d, l10n, limitReached: 'limit'),
  };
  final fields = {
    'sign-up name': 'name',
    'profile name': 'name',
    'income description': 'description',
    'expense description': 'description',
    'annual label': 'employer_label',
  };
  for (final MapEntry(key: name, value: describe) in describers.entries) {
    test('$name: invisible_characters reads its own message', () {
      expect(describe(invisible(fields[name]!)), expected);
    });
  }

  group('taxpayer display name', () {
    test('sortDetails tells invisible characters from length', () {
      expect(
        sortDetails([
          invisible('display_name'),
        ], sentRegimes: const []).displayName,
        DisplayNameRejection.invisibleCharacters,
      );
      expect(
        sortDetails([
          const api.FieldError(
            field: 'display_name',
            code: 'maxLength',
            message: 'too long',
          ),
        ], sentRegimes: const []).displayName,
        DisplayNameRejection.length,
      );
    });

    test('each rejection reads its own message', () {
      expect(
        describeDisplayNameRejection(
          DisplayNameRejection.invisibleCharacters,
          l10n,
        ),
        expected,
      );
      expect(
        describeDisplayNameRejection(DisplayNameRejection.length, l10n),
        isA<CustomFieldError>().having(
          (e) => e.message,
          'message',
          l10n.onboardingDisplayNameTooLong,
        ),
      );
    });

    test('a settings rename failure maps to the display name verdict', () {
      expect(
        displayNameRejectionOf(
          ApiValidationFailure([invisible('display_name')]),
        ),
        DisplayNameRejection.invisibleCharacters,
      );
      expect(displayNameRejectionOf(const ApiUnreachable()), isNull);
    });
  });
}
