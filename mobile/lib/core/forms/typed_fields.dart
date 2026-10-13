import 'package:control_contable/core/forms/app_text_field.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

/// A peso amount: decimal keyboard, `$` prefix; accepts `1,234.50` or
/// `1234.5` and reads back as centavos ([MoneyFieldController.centavos]).
class AppMoneyField extends StatelessWidget {
  /// A money field for [field].
  const new({required this.field, required this.textInputAction, super.key});

  /// The amount field.
  final MoneyFieldController field;

  /// The IME action key.
  final TextInputAction textInputAction;

  @override
  Widget build(BuildContext context) => AppTextField(
    field: field,
    keyboardType: const TextInputType.numberWithOptions(decimal: true),
    textInputAction: textInputAction,
    autofillHints: const [],
    prefixText: r'$ ',
    inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[\d.,$ ]'))],
  );
}

/// A calendar date typed as day/month/year in any common form
/// ([DateFieldController.date]).
class AppDateField extends StatelessWidget {
  /// A date field for [field].
  const new({required this.field, required this.textInputAction, super.key});

  /// The date field.
  final DateFieldController field;

  /// The IME action key.
  final TextInputAction textInputAction;

  @override
  Widget build(BuildContext context) => AppTextField(
    field: field,
    keyboardType: TextInputType.datetime,
    textInputAction: textInputAction,
    autofillHints: const [],
    hintText: AppLocalizations.of(context).formDateHint,
    inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[\d/\-. ]'))],
  );
}
