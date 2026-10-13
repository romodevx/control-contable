// ignore: unused_import
import 'package:intl/intl.dart' as intl;

import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appTitle => 'Control Contable';

  @override
  String get navHome => 'Inicio';

  @override
  String get navIncome => 'Ingresos';

  @override
  String get navExpenses => 'Gastos';

  @override
  String get navHistory => 'Historial';

  @override
  String get navSettings => 'Ajustes';

  @override
  String get homePlaceholder => 'Aquí aparecerá el estimado del mes.';

  @override
  String get estimateLegend =>
      'Esto es un aproximado. No todos los gastos aplican para tu giro comercial.';

  @override
  String get notFoundTitle => 'Página no encontrada';

  @override
  String get notFoundAction => 'Ir al inicio';

  @override
  String get bootLoading => 'Cargando…';

  @override
  String get formMarkOptional => '(opcional)';

  @override
  String get formMarkRequired => '(obligatorio)';

  @override
  String get formErrorRequired => 'Escribe este dato.';

  @override
  String get formErrorMoney => 'Escribe un monto, por ejemplo 1,234.50.';

  @override
  String get formErrorDate => 'Escribe una fecha como 31/01/2026.';

  @override
  String formErrorSummaryTitle(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Revisa $count datos antes de guardar',
      one: 'Revisa 1 dato antes de guardar',
    );
    return '$_temp0';
  }

  @override
  String formErrorSummaryItem(String label, String error) {
    return '$label: $error';
  }

  @override
  String get formSave => 'Guardar';

  @override
  String get formSaving => 'Guardando…';

  @override
  String get formSaveAndCreateAnother => 'Guardar y crear otro';

  @override
  String get formCancel => 'Cancelar';

  @override
  String get formDateHint => 'dd/mm/aaaa';
}

/// The translations for Spanish Castilian, as used in Mexico (`es_MX`).
class AppLocalizationsEsMx extends AppLocalizationsEs {
  AppLocalizationsEsMx() : super('es_MX');
}
