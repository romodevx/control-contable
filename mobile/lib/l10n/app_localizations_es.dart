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
}

/// The translations for Spanish Castilian, as used in Mexico (`es_MX`).
class AppLocalizationsEsMx extends AppLocalizationsEs {
  AppLocalizationsEsMx() : super('es_MX');
}
