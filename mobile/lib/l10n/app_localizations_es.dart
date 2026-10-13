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

  @override
  String get formErrorNotAccepted => 'Marca esta casilla para continuar.';

  @override
  String get formShowPassword => 'Mostrar contraseña';

  @override
  String get formHidePassword => 'Ocultar contraseña';

  @override
  String get homeCreateAccount => 'Crear cuenta';

  @override
  String get authRegisterTitle => 'Crear cuenta';

  @override
  String get authRegisterIntro =>
      'Con una cuenta, los ingresos, gastos y estimados quedan guardados de forma segura.';

  @override
  String get authRegisterSubmit => 'Crear cuenta';

  @override
  String get authNameLabel => 'Nombre';

  @override
  String get authEmailLabel => 'Correo electrónico';

  @override
  String get authPasswordLabel => 'Contraseña';

  @override
  String get authPasswordHelp =>
      'Al menos 12 caracteres. Una frase fácil de recordar funciona bien.';

  @override
  String get authEmailInvalid => 'Escribe un correo como nombre@ejemplo.mx.';

  @override
  String get authNameTooLong => 'Escribe un nombre de hasta 120 caracteres.';

  @override
  String get authPasswordTooShort => 'Usa al menos 12 caracteres.';

  @override
  String get authPasswordTooLong => 'Usa como máximo 128 caracteres.';

  @override
  String get authPasswordPolicy =>
      'Esta contraseña es muy común o apareció en una filtración. Elige otra.';

  @override
  String get authFieldRejected => 'Revisa este dato.';

  @override
  String get authFormRejected => 'Revisa los datos e intenta de nuevo.';

  @override
  String get authConsentsHeading => 'Documentos legales';

  @override
  String get authConsentPrivacy => 'Acepto el aviso de privacidad';

  @override
  String get authConsentTerms => 'Acepto los términos y condiciones';

  @override
  String get authReadPrivacy => 'Leer el aviso de privacidad';

  @override
  String get authReadTerms => 'Leer los términos y condiciones';

  @override
  String get authDocumentsLoading => 'Cargando documentos legales…';

  @override
  String get authDocumentsFailed =>
      'No se pudieron cargar los documentos legales.';

  @override
  String get authDocumentsNotLoaded =>
      'Los documentos legales no se han cargado. Reintenta para continuar.';

  @override
  String get authDocumentsStale =>
      'Los documentos se actualizaron. Revísalos y acéptalos de nuevo.';

  @override
  String get authDocumentOpenFailed =>
      'No se pudo abrir el documento. Intenta de nuevo.';

  @override
  String get authRetry => 'Reintentar';

  @override
  String authErrorRateLimited(int minutes) {
    String _temp0 = intl.Intl.pluralLogic(
      minutes,
      locale: localeName,
      other: 'Demasiados intentos. Espera $minutes min y vuelve a intentarlo.',
      one: 'Demasiados intentos. Espera 1 min y vuelve a intentarlo.',
    );
    return '$_temp0';
  }

  @override
  String get authErrorRateLimitedNoWait =>
      'Demasiados intentos. Espera unos minutos y vuelve a intentarlo.';

  @override
  String get authErrorUnreachable =>
      'No hay conexión con el servidor. Revisa la conexión a internet e intenta de nuevo.';

  @override
  String get authErrorUnavailable =>
      'Este servicio no está disponible por ahora.';

  @override
  String get authErrorUnexpected =>
      'Algo salió mal. Intenta de nuevo en unos minutos.';

  @override
  String get authVerifyTitle => 'Confirmar correo';

  @override
  String authVerifyIntro(String email) {
    return 'Enviamos un código de 6 dígitos a $email. Vence en 15 minutos.';
  }

  @override
  String get authVerifySpamHint =>
      'Si no aparece, conviene revisar la carpeta de correo no deseado.';

  @override
  String get authCodeLabel => 'Código de verificación';

  @override
  String get authCodeFormat => 'Escribe los 6 dígitos del código.';

  @override
  String get authCodeRejected =>
      'El código no es válido o ya venció. Revisa el correo o pide uno nuevo.';

  @override
  String get authVerifySubmit => 'Confirmar';

  @override
  String get authResend => 'Enviar otro código';

  @override
  String authResendWait(String time) {
    return 'Otro código disponible en $time.';
  }

  @override
  String get authResendSent =>
      'Si hay un registro pendiente con ese correo, enviamos un código nuevo.';

  @override
  String get authChangeEmail => 'Cambiar correo';

  @override
  String get authVerifyNoPending => 'No hay un registro en curso.';
}

/// The translations for Spanish Castilian, as used in Mexico (`es_MX`).
class AppLocalizationsEsMx extends AppLocalizationsEs {
  AppLocalizationsEsMx() : super('es_MX');
}
