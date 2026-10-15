import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// A peso amount: decimal keyboard, `$` prefix; accepts `1,234.50` or
/// `1234.5` and reads back as centavos ([MoneyFieldController.centavos]).
class AppMoneyField extends StatelessWidget {
  /// A money field for [field].
  const new({
    required this.field,
    required this.textInputAction,
    this.helperText,
    super.key,
  });

  /// The amount field.
  final MoneyFieldController field;

  /// The IME action key.
  final TextInputAction textInputAction;

  /// A short explanation under the field.
  final String? helperText;

  @override
  Widget build(BuildContext context) => AppTextField(
    field: field,
    keyboardType: const TextInputType.numberWithOptions(decimal: true),
    textInputAction: textInputAction,
    autofillHints: const [],
    prefixText: r'$ ',
    helperText: helperText,
    inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[\d.,$ ]'))],
  );
}

/// A calendar date typed as day/month/year in any common form
/// ([DateFieldController.date]).
class AppDateField extends StatelessWidget {
  /// A date field for [field].
  const new({
    required this.field,
    required this.textInputAction,
    this.helperText,
    super.key,
  });

  /// The date field.
  final DateFieldController field;

  /// The IME action key.
  final TextInputAction textInputAction;

  /// A short explanation under the field.
  final String? helperText;

  @override
  Widget build(BuildContext context) => AppTextField(
    field: field,
    keyboardType: TextInputType.datetime,
    textInputAction: textInputAction,
    autofillHints: const [],
    hintText: AppLocalizations.of(context).formDateHint,
    helperText: helperText,
    inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[\d/\-. ]'))],
  );
}

/// A password: hidden text with a labeled show/hide toggle, no suggestions.
/// [autofillHints] tells the platform whether it is a new password (sign-up)
/// or the current one (login), so password managers offer the right thing.
class AppPasswordField extends StatefulWidget {
  /// A password field for [field].
  const new({
    required this.field,
    required this.textInputAction,
    required this.autofillHints,
    this.helperText,
    super.key,
  });

  /// The password field.
  final AppFieldController field;

  /// The IME action key.
  final TextInputAction textInputAction;

  /// `AutofillHints.newPassword` or `AutofillHints.password`.
  final Iterable<String> autofillHints;

  /// The password rule shown under the field.
  final String? helperText;

  @override
  State<AppPasswordField> createState() => _AppPasswordFieldState();
}

class _AppPasswordFieldState extends State<AppPasswordField> {
  bool _isHidden = true;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AppTextField(
      field: widget.field,
      keyboardType: TextInputType.visiblePassword,
      textInputAction: widget.textInputAction,
      autofillHints: widget.autofillHints,
      helperText: widget.helperText,
      obscureText: _isHidden,
      suffixIcon: IconButton(
        tooltip: _isHidden ? l10n.formShowPassword : l10n.formHidePassword,
        icon: Icon(
          _isHidden ? Icons.visibility_outlined : Icons.visibility_off_outlined,
        ),
        onPressed: () => setState(() => _isHidden = !_isHidden),
      ),
    );
  }
}
