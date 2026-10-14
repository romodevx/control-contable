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
  String get loadRetry => 'Reintentar';

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
  String get authSignOut => 'Cerrar sesión';

  @override
  String get authLoginTitle => 'Iniciar sesión';

  @override
  String get authLoginSubmit => 'Entrar';

  @override
  String get authLoginRejected => 'Correo o contraseña incorrectos.';

  @override
  String get authLoginNoAccount => '¿Aún no tienes cuenta?';

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

  @override
  String get accountTitle => 'Mi cuenta';

  @override
  String get accountLoading => 'Cargando tu cuenta…';

  @override
  String get accountLoadFailed => 'No se pudo cargar tu cuenta.';

  @override
  String get accountEmailReadOnly => 'El correo no se puede cambiar.';

  @override
  String get accountSave => 'Guardar cambios';

  @override
  String get accountSaved => 'Guardamos tu nombre.';

  @override
  String get accountDeleteHeading => 'Eliminar cuenta';

  @override
  String get accountDeleteIntro =>
      'Borra tu cuenta y todos sus datos de forma permanente.';

  @override
  String get accountDeleteOpen => 'Eliminar mi cuenta';

  @override
  String get accountDeleteTitle => '¿Eliminar tu cuenta?';

  @override
  String get accountDeleteWhat =>
      'Se borrarán para siempre tu cuenta y todo lo que guardaste: contribuyentes, ingresos, gastos, estimados y recordatorios. No se puede deshacer.';

  @override
  String get accountDeleteTrial =>
      'La prueba gratuita no se reinicia: si vuelves a registrarte con este correo, no tendrás otra.';

  @override
  String get accountDeleteSubscription =>
      'Si pagas una suscripción, cancélala en la tienda de tu teléfono (App Store o Google Play): eliminar la cuenta no la cancela.';

  @override
  String get accountDeletePasswordLabel => 'Contraseña actual';

  @override
  String get accountDeleteConfirm => 'Eliminar cuenta';

  @override
  String get accountDeleting => 'Eliminando cuenta…';

  @override
  String get accountDeleteWrongPassword => 'La contraseña no es correcta.';

  @override
  String authRetryIn(String time) {
    return 'Podrás intentar de nuevo en $time.';
  }

  @override
  String get goodbyeTitle => 'Cuenta eliminada';

  @override
  String get goodbyeBody =>
      'Eliminamos tu cuenta y todos sus datos. Gracias por usar Control Contable.';

  @override
  String get goodbyeToLogin => 'Ir a iniciar sesión';

  @override
  String get sessionTokenLeftOnDevice =>
      'No pudimos borrar la sesión guardada en este teléfono. Cierra la app y vuelve a abrirla; si el aviso sigue, borra los datos de la app desde los ajustes del teléfono.';

  @override
  String get formErrorChoose => 'Elige una opción.';

  @override
  String get homeNoRegime => 'Agrega un régimen para ver tu estimado.';

  @override
  String get homeAddTaxpayer => 'Registrar contribuyente';

  @override
  String get onboardingTitle => 'Nuevo contribuyente';

  @override
  String onboardingProgress(int step, int total) {
    return 'Paso $step de $total';
  }

  @override
  String get onboardingNext => 'Siguiente';

  @override
  String get onboardingCreate => 'Crear contribuyente';

  @override
  String get onboardingContinue => 'Continuar';

  @override
  String get onboardingPersonTypeTitle => '¿Persona física o persona moral?';

  @override
  String get onboardingPersonTypeIntro =>
      'Elige cómo está dado de alta el contribuyente ante el SAT. No se puede cambiar después.';

  @override
  String get onboardingFisicaTitle => 'Persona física';

  @override
  String get onboardingFisicaBody =>
      'Una persona que trabaja por su cuenta: honorarios, rentas, ventas por plataformas digitales o un negocio propio. Su RFC tiene 13 caracteres.';

  @override
  String get onboardingMoralTitle => 'Persona moral';

  @override
  String get onboardingMoralBody =>
      'Una empresa o sociedad (S.A., S. de R.L., A.C. y otras). Su RFC tiene 12 caracteres.';

  @override
  String get onboardingIdentityTitle => 'RFC y nombre';

  @override
  String get onboardingDisplayNameLabel => 'Nombre para identificarlo';

  @override
  String get onboardingDisplayNameHelp =>
      'Así aparecerá en la app; por ejemplo, tu nombre o «Mi negocio».';

  @override
  String get onboardingDisplayNameTooLong =>
      'Escribe un nombre de hasta 80 caracteres.';

  @override
  String get onboardingRfcLabel => 'RFC';

  @override
  String get onboardingRfcHelpFisica =>
      '13 caracteres, como en la constancia de situación fiscal. Puede llevar espacios o guiones.';

  @override
  String get onboardingRfcHelpMoral =>
      '12 caracteres, como en la constancia de situación fiscal. Puede llevar espacios o guiones.';

  @override
  String get onboardingRfcLengthFisica =>
      'El RFC de una persona física tiene 13 caracteres.';

  @override
  String get onboardingRfcLengthMoral =>
      'El RFC de una persona moral tiene 12 caracteres.';

  @override
  String get onboardingRfcFormat =>
      'Revisa el RFC: primero las letras, luego los 6 números de la fecha y al final la homoclave.';

  @override
  String get onboardingRfcDate =>
      'Los 6 números del RFC no forman una fecha válida (año, mes y día).';

  @override
  String get onboardingRfcGeneric =>
      'Ese es un RFC genérico del SAT. Escribe el RFC propio del contribuyente.';

  @override
  String get onboardingRfcDuplicate =>
      'Ya registraste un contribuyente con este RFC. Si lo archivaste, puedes reactivarlo en Ajustes.';

  @override
  String get onboardingRfcCheckWarning => 'Revisa que tu RFC esté bien escrito';

  @override
  String get onboardingRfcCheckBody =>
      'El último carácter no coincide con el que calcula el SAT. Algunos RFC reales no coinciden: si el tuyo es igual al de tu constancia, continúa. Si no, corrígelo aquí.';

  @override
  String get onboardingRegimesTitle => 'Regímenes fiscales';

  @override
  String get onboardingRegimesIntroFisica =>
      'Elige todos los regímenes en los que está dado de alta. Aparecen en la constancia de situación fiscal.';

  @override
  String get onboardingRegimesIntroMoral =>
      'Elige el régimen de la empresa. Una persona moral tiene un solo régimen a la vez.';

  @override
  String get onboardingStartPeriodLabelSingle =>
      '¿Desde qué mes tributas en este régimen?';

  @override
  String get onboardingStartPeriodLabelMany =>
      '¿Desde qué mes tributas en estos regímenes?';

  @override
  String get onboardingStartPeriodHint =>
      'Solo podrás registrar ingresos y gastos desde el mes que elijas. Si ya tributabas antes de este mes, elige el mes en que empezaste, o enero de 2026 si empezaste antes de 2026.';

  @override
  String get onboardingStartPeriodRejected =>
      'Ese mes no se puede elegir. Elige un mes entre enero de 2026 y el mes actual.';

  @override
  String get onboardingRegimesLabel => 'Regímenes';

  @override
  String get onboardingRegimesRequired => 'Elige al menos un régimen.';

  @override
  String get onboardingRegimesConflict =>
      'Hay regímenes que no se pueden combinar.';

  @override
  String onboardingRegimeConflict(String code, String reason) {
    return 'No se puede combinar con el $code: $reason';
  }

  @override
  String get onboardingRegimesLoading => 'Cargando regímenes…';

  @override
  String get onboardingRegimesLoadFailed =>
      'No se pudieron cargar los regímenes.';

  @override
  String get onboardingRegimesEmpty =>
      'No hay regímenes disponibles para este tipo de persona.';

  @override
  String get onboardingRegimeCombinationRejected =>
      'Esta combinación de regímenes no se puede registrar. Revisa tu selección.';

  @override
  String get onboardingRegimeRejected =>
      'Uno de los regímenes elegidos no se aceptó. Revisa tu selección.';

  @override
  String get onboardingSupportFullTitle => 'Cálculo completo';

  @override
  String get onboardingSupportFullBody =>
      'La app estima cada mes el ISR y el IVA a pagar con los ingresos y gastos que registres.';

  @override
  String get onboardingSupportPartialTitle => 'Cálculo parcial';

  @override
  String get onboardingSupportPartialBody =>
      'La app estima solo una parte de los impuestos (por ejemplo, solo el IVA). Lo demás conviene revisarlo con un contador.';

  @override
  String get onboardingSupportInfoTitle => 'Solo explicación';

  @override
  String get onboardingSupportInfoBody =>
      'La app no calcula impuestos para estos regímenes: explica en palabras sencillas qué obligaciones tienen.';

  @override
  String get onboardingOptionsTitle => 'Opciones del régimen';

  @override
  String onboardingOptionsRegime(
    int current,
    int total,
    String code,
    String name,
  ) {
    return 'Régimen $current de $total: $code · $name';
  }

  @override
  String get onboardingOptionRejected => 'Revisa esta opción.';

  @override
  String onboardingRateHelp(String min, String max) {
    return 'Un número entre $min y $max, con punto decimal.';
  }

  @override
  String get onboardingRateInvalid =>
      'Escribe un número con punto decimal, por ejemplo 0.25.';

  @override
  String onboardingRateOutOfRange(String min, String max) {
    return 'Escribe un número entre $min y $max.';
  }

  @override
  String get optionValueBlind35 => 'Deducción ciega: 35 % de los ingresos';

  @override
  String get optionValueActual => 'Gastos reales con factura';

  @override
  String get optionValueTransportDelivery =>
      'Transporte de pasajeros o entrega de bienes';

  @override
  String get optionValueLodging => 'Hospedaje';

  @override
  String get optionValueGoodsServices =>
      'Venta de bienes y prestación de servicios';

  @override
  String optionValueRifYear(String year) {
    return 'Año $year';
  }

  @override
  String get onboardingErrorUnreachable =>
      'No hay conexión. Revisa tu internet e inténtalo de nuevo.';

  @override
  String get onboardingErrorRateLimited =>
      'Hubo demasiados cambios seguidos. Espera unos minutos e inténtalo de nuevo.';

  @override
  String get onboardingErrorLimit =>
      'Ya tienes 10 contribuyentes, el máximo por cuenta. Elimina uno que ya no uses para registrar otro.';

  @override
  String get onboardingErrorUnexpected => 'Algo salió mal. Inténtalo de nuevo.';

  @override
  String get onboardingPaywallTitle => 'Tu cuenta es de solo consulta';

  @override
  String get onboardingPaywallBody =>
      'Tu periodo de prueba terminó y no hay un plan activo. Puedes consultar, archivar y eliminar tu información, pero registrar un contribuyente requiere un plan.';

  @override
  String get onboardingLimitTitle => 'Llegaste al máximo de contribuyentes';

  @override
  String onboardingLimitBody(int cap) {
    return 'Ya tienes $cap contribuyentes, el máximo por cuenta; los archivados también cuentan. Para registrar otro, elimina uno que ya no uses.';
  }

  @override
  String listAnd(String head, String last) {
    return '$head y $last';
  }

  @override
  String get onboardingPaywallAction => 'Volver al inicio';

  @override
  String get onboardingReviewTitle => 'Revisa antes de continuar';

  @override
  String onboardingReviewCreated(String name, String rfc) {
    return 'Registramos a «$name», RFC $rfc.';
  }

  @override
  String onboardingCpaWarningBody(String regimes, String reason) {
    return 'Registraste juntos $regimes. $reason';
  }

  @override
  String get onboardingCpaWarningTitle =>
      'Combinación pendiente de revisión por un contador';

  @override
  String get onboardingInputsTitle => 'Datos para el cálculo';

  @override
  String onboardingInputsIntro(String year) {
    return 'Opcional. Con estos datos de $year el estimado es más preciso; puedes agregarlos después.';
  }

  @override
  String get onboardingInputsSkip => 'Omitir';

  @override
  String switcherLabel(String name) {
    return 'Contribuyente activo: $name. Cambiar';
  }

  @override
  String get switcherTitle => 'Contribuyentes';

  @override
  String get switcherShowArchived => 'Mostrar archivados';

  @override
  String get taxpayerArchivedBadge => 'Archivado';

  @override
  String taxpayerRfcSemantics(String start, String end) {
    return 'RFC $start, oculto, $end';
  }

  @override
  String taxpayerSettingsTooltip(String name) {
    return 'Ajustes de $name';
  }

  @override
  String get settingsTitle => 'Ajustes del contribuyente';

  @override
  String get settingsLoading => 'Cargando contribuyentes…';

  @override
  String get settingsLoadFailed => 'No se pudieron cargar tus contribuyentes.';

  @override
  String get settingsNotFound => 'Este contribuyente ya no existe.';

  @override
  String get settingsDataTitle => 'Datos';

  @override
  String get settingsEditName => 'Editar nombre';

  @override
  String get settingsCorrectRfc => 'Corregir RFC';

  @override
  String get settingsRfcHelp =>
      'Escribe el RFC completo, como en la constancia. Nunca lo mostramos completo.';

  @override
  String get settingsNameSaved => 'Guardamos el nombre.';

  @override
  String get settingsRfcCorrected => 'Corregimos el RFC.';

  @override
  String get settingsArchivedWrite =>
      'Este contribuyente está archivado. Desarchívalo para hacer cambios.';

  @override
  String get settingsRegimesTitle => 'Regímenes';

  @override
  String settingsRegimeSince(String start) {
    return 'Desde $start';
  }

  @override
  String settingsRegimeRange(String start, String end) {
    return 'De $start a $end';
  }

  @override
  String get settingsRegimeEnded => 'Terminado';

  @override
  String get settingsAddRegime => 'Agregar régimen';

  @override
  String get settingsEditOptions => 'Editar opciones';

  @override
  String get settingsEndRegime => 'Terminar régimen';

  @override
  String get settingsEndRegimeLabel => 'Último mes en este régimen';

  @override
  String get settingsEndRegimeConfirm => 'Terminar';

  @override
  String get settingsEndRegimeHasRecords =>
      'Este régimen tiene ingresos o gastos registrados. Se conservan y siguen contando hasta el mes que elijas.';

  @override
  String get settingsEndRegimeKeep => 'Terminar y conservar registros';

  @override
  String get settingsEndRegimeRecordsAfter =>
      'Hay registros después de ese mes. Elige el mes del último registro o uno posterior.';

  @override
  String get settingsLastRegime =>
      'Es su único régimen activo. Para dejar de usarlo, archiva al contribuyente.';

  @override
  String settingsRegimeEndedDone(String code) {
    return 'Terminamos el régimen $code.';
  }

  @override
  String get settingsManageTitle => 'Archivar o eliminar';

  @override
  String get settingsArchive => 'Archivar';

  @override
  String get settingsArchiveHelp =>
      'Lo ocultas de tu lista. Conservas su información y puedes desarchivarlo.';

  @override
  String settingsArchiveOneWayTitle(String name) {
    return '¿Archivar a «$name»?';
  }

  @override
  String get settingsArchiveOneWayBody =>
      'Tu cuenta es de solo consulta. Podrás seguir consultándolo, pero no podrás desarchivarlo hasta tener un plan.';

  @override
  String get settingsArchiveOneWayConfirm => 'Archivar de todos modos';

  @override
  String settingsArchived(String name) {
    return 'Archivaste a «$name».';
  }

  @override
  String get settingsUndo => 'Deshacer';

  @override
  String get settingsUnarchive => 'Desarchivar';

  @override
  String settingsUnarchived(String name) {
    return '«$name» volvió a tu lista.';
  }

  @override
  String get settingsArchivedBanner =>
      'Este contribuyente está archivado. Puedes consultarlo; para hacer cambios, desarchívalo.';

  @override
  String get settingsDelete => 'Eliminar';

  @override
  String get settingsDeleteHelp =>
      'Borra al contribuyente y todos sus registros. No se puede deshacer.';

  @override
  String settingsDeleteTitle(String name) {
    return '¿Eliminar a «$name»?';
  }

  @override
  String get settingsDeleteBody =>
      'Se borrarán sus ingresos y gastos, sus datos para el cálculo y su historial. No se puede deshacer. Si solo quieres ocultarlo, archívalo.';

  @override
  String get settingsDeleteConfirm => 'Eliminar definitivamente';

  @override
  String settingsDeleted(String name) {
    return 'Eliminaste a «$name».';
  }

  @override
  String get paywallWriteBody =>
      'Tu periodo de prueba terminó y no hay un plan activo. Puedes consultar, archivar y eliminar tu información, pero hacer cambios requiere un plan.';

  @override
  String get paywallClose => 'Entendido';

  @override
  String get addRegimeIntro =>
      'Elige el régimen que vas a agregar y el mes desde el que tributas en él.';

  @override
  String get addRegimeNoneLeft =>
      'Ya tiene todos los regímenes disponibles para su tipo de persona.';

  @override
  String addRegimeReopensFrom(String month) {
    return 'Lo reactivamos desde $month, el mes en que empezó.';
  }

  @override
  String addRegimeAdded(String code) {
    return 'Agregamos el régimen $code.';
  }

  @override
  String addRegimeCpaReview(String code, String reason) {
    return 'Agregamos el régimen $code. Un contador debe confirmar la combinación: $reason';
  }

  @override
  String addRegimeCpaReviewWith(String code, String partners, String reason) {
    return 'Agregamos el régimen $code. Un contador debe confirmar su combinación con $partners: $reason';
  }

  @override
  String get optionsEditWarning =>
      'El cambio aplica a todos los meses de este régimen, también a los anteriores.';

  @override
  String get optionsSaved => 'Guardamos las opciones.';

  @override
  String get addRegimeTitle => 'Nuevo régimen';

  @override
  String get recordsPreviousMonth => 'Mes anterior';

  @override
  String get recordsNextMonth => 'Mes siguiente';

  @override
  String get recordsFilterAll => 'Todos';

  @override
  String get recordsFilterLabel => 'Régimen';

  @override
  String get recordsUndo => 'Deshacer';

  @override
  String get recordsDelete => 'Eliminar';

  @override
  String get incomesNoTaxpayer =>
      'Registra un contribuyente para llevar sus ingresos.';

  @override
  String incomesEmptyTitle(String month) {
    return 'Sin ingresos en $month';
  }

  @override
  String get incomesEmptyBody =>
      'Registra tu primer ingreso y te mostramos cuánto podrías pagar.';

  @override
  String get incomesRecord => 'Registrar ingreso';

  @override
  String get incomesLoadFailed => 'No se pudieron cargar los ingresos.';

  @override
  String get incomesLoading => 'Cargando ingresos…';

  @override
  String incomesCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count ingresos',
      one: '1 ingreso',
    );
    return '$_temp0';
  }

  @override
  String get incomesTotalTitle => 'Total del mes con IVA';

  @override
  String get incomesTotalSubtotal => 'Sin IVA';

  @override
  String get incomesTotalIva => 'IVA';

  @override
  String get incomesTotalWithheld => 'Retenciones';

  @override
  String get incomesPendingCollection => 'Por cobrar';

  @override
  String incomesDeleteTooltip(String title) {
    return 'Eliminar ingreso: $title';
  }

  @override
  String incomesEditLabel(String title) {
    return 'Editar ingreso: $title';
  }

  @override
  String incomesDeleted(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count ingresos eliminados',
      one: 'Ingreso eliminado',
    );
    return '$_temp0';
  }

  @override
  String incomesRestored(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count ingresos restaurados',
      one: 'Ingreso restaurado',
    );
    return '$_temp0';
  }

  @override
  String incomesRestoredSome(int restored, int total) {
    String _temp0 = intl.Intl.pluralLogic(
      restored,
      locale: localeName,
      other: 'Se restauraron $restored de $total ingresos.',
      one: 'Se restauró 1 de $total ingresos.',
      zero: 'No se restauró ninguno de los $total ingresos.',
    );
    return '$_temp0';
  }

  @override
  String get incomesDeleteFailed => 'No se pudo eliminar el ingreso.';

  @override
  String get incomesRestoreFailed => 'No se pudo restaurar el ingreso.';

  @override
  String get incomesRestoreRegimeEnded =>
      'No se pudo restaurar: el régimen ya no cubre la fecha del ingreso.';

  @override
  String get counterpartyPersonaFisica => 'Persona física';

  @override
  String get counterpartyPersonaMoral => 'Persona moral';

  @override
  String get counterpartyPublicoGeneral => 'Público en general';

  @override
  String get counterpartyPlataforma => 'Plataforma digital';

  @override
  String get ivaRate16 => '16 %';

  @override
  String get ivaRate8 => '8 % (región fronteriza)';

  @override
  String get ivaRate0 => '0 %';

  @override
  String get ivaRateExempt => 'Exento';

  @override
  String get activityTransportDelivery => 'Transporte o entrega';

  @override
  String get activityLodging => 'Hospedaje';

  @override
  String get activityGoodsServices => 'Venta de bienes o servicios';

  @override
  String get incomeFormTitle => 'Nuevo ingreso';

  @override
  String get incomeEditTitle => 'Editar ingreso';

  @override
  String get incomeUpdated => 'Ingreso actualizado';

  @override
  String incomeUpdatedWithWarnings(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Ingreso actualizado, con $count avisos:',
      one: 'Ingreso actualizado, con un aviso:',
    );
    return '$_temp0';
  }

  @override
  String get incomeFieldCollectedOn => 'Fecha de cobro';

  @override
  String get incomeFieldInvoicedOn => 'Fecha de factura';

  @override
  String get incomeFieldDescription => 'Descripción';

  @override
  String get incomeFieldAmount => 'Monto sin IVA';

  @override
  String get incomeFieldIvaRate => 'Tasa de IVA';

  @override
  String get incomeFieldCounterparty => '¿Quién pagó?';

  @override
  String get incomeFieldRegime => 'Régimen';

  @override
  String get incomeFieldActivity => 'Actividad en la plataforma';

  @override
  String get incomeFieldIvaAmount => 'IVA cobrado';

  @override
  String get incomeFieldIsrWithheld => 'ISR retenido';

  @override
  String get incomeFieldIvaWithheld => 'IVA retenido';

  @override
  String get incomeFieldCfdiUuid => 'Folio fiscal (UUID)';

  @override
  String get incomeSectionWithholdings => 'Retenciones';

  @override
  String get incomeSectionMore => 'Más detalles';

  @override
  String get incomeWithholdingsHelp =>
      'Prellenadas con la tasa sugerida. Corrígelas con lo que diga el CFDI.';

  @override
  String incomeSummaryTotal(String total) {
    return 'Total con IVA $total';
  }

  @override
  String incomeSummaryNet(String iva, String withheld, String net) {
    return 'IVA $iva · Retenciones $withheld · Neto $net';
  }

  @override
  String get incomeInformationalRegime =>
      'Este régimen es informativo: el ingreso no cambia tu estimación.';

  @override
  String get incomeCfdiInvalid =>
      'Escribe el folio como aparece en el CFDI: 32 letras y números en grupos de 8-4-4-4-12.';

  @override
  String get incomeDefaultsFailed =>
      'No se pudieron cargar las tasas sugeridas. Puedes capturar los montos directamente.';

  @override
  String get incomeErrorOutOfRange =>
      'La fecha debe estar entre el 1 de enero de 2022 y hoy.';

  @override
  String get incomeErrorIvaMismatch =>
      'El IVA no corresponde a la tasa elegida. Revísalo en el CFDI.';

  @override
  String get incomeErrorRateNotValid =>
      'Esta tasa no está vigente en la fecha del ingreso.';

  @override
  String get incomeErrorWithholdingNotApplicable =>
      'Con este tipo de cliente no hay esta retención.';

  @override
  String get incomeErrorWithholdingExceedsBase =>
      'La retención no puede ser mayor que su base.';

  @override
  String get incomeErrorRegimeNotActive =>
      'El contribuyente no tenía este régimen en esa fecha.';

  @override
  String get incomeErrorActivityNotHeld =>
      'Esta actividad no está registrada en el régimen.';

  @override
  String get incomeErrorInvalid => 'Revisa este dato.';

  @override
  String get incomeSaveUnreachable =>
      'No se pudo guardar por falta de conexión. Vuelve a tocar Guardar: no se duplicará.';

  @override
  String get incomeSaveIdReused =>
      'Este ingreso ya se había guardado con otros datos. Revisa la lista de ingresos.';

  @override
  String get incomeSaveGone =>
      'Este ingreso ya no existe: se eliminó en otro lugar.';

  @override
  String get incomeSaveFailed =>
      'No se pudo guardar el ingreso. Intenta de nuevo.';

  @override
  String get incomeSaved => 'Ingreso guardado';

  @override
  String incomeSavedWithWarnings(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Ingreso guardado, con $count avisos:',
      one: 'Ingreso guardado, con un aviso:',
    );
    return '$_temp0';
  }

  @override
  String get incomeWarningDuplicateCfdi =>
      'Ya hay otro ingreso con este folio fiscal. Si es un pago parcial del mismo CFDI, está bien.';

  @override
  String get incomeWarningWithholdingDiffers =>
      'La retención es distinta de la tasa sugerida. Se guardó lo que escribiste: el CFDI manda.';

  @override
  String get incomeTrialCapTitle => 'Llegaste al límite de ingresos de prueba';

  @override
  String get incomeTrialCapBody =>
      'La prueba incluye 5 ingresos y ya los registraste. Para registrar más ingresos se necesita un plan. Lo que capturaste sigue en el formulario.';

  @override
  String get incomeDiscardTitle => '¿Descartar este ingreso?';

  @override
  String get recordDiscardBody => 'Lo que escribiste se perderá.';

  @override
  String get recordDiscardKeep => 'Seguir editando';

  @override
  String get recordDiscardConfirm => 'Descartar';

  @override
  String get homeOpenEstimate => 'Ver estimado del mes';

  @override
  String get estimateTitle => 'Estimado del mes';

  @override
  String get estimatePeriodLabel => 'Mes';

  @override
  String get estimatePeriodHelp =>
      'Se puede estimar desde enero de 2026 hasta el mes en curso. Los meses futuros aún no se pueden estimar y antes de 2026 no hay tarifas cargadas.';

  @override
  String get estimateCurrentMonth =>
      'Mes en curso: se calcula con lo registrado hasta hoy.';

  @override
  String get estimateLoading => 'Cargando el estimado…';

  @override
  String get estimateLoadFailed => 'No se pudo cargar el estimado.';

  @override
  String get estimateUnreachable =>
      'No hay conexión. Revisa tu internet e inténtalo de nuevo.';

  @override
  String get estimateParametersUnavailable =>
      'No hay tarifas cargadas para ese mes. Elige un mes desde enero de 2026.';

  @override
  String get estimateFutureMonth =>
      'Ese mes todavía no llega. Elige el mes en curso o uno anterior.';

  @override
  String get estimateNotFound =>
      'No se encontró este contribuyente. Puede haberse eliminado en otro dispositivo.';

  @override
  String get estimateNoTaxpayer =>
      'Agrega un contribuyente para ver su estimado.';

  @override
  String estimateNoRegimes(String month) {
    return 'Este contribuyente no tiene regímenes en $month. No hay nada que estimar.';
  }

  @override
  String get estimateIncomplete => 'Incompleto';

  @override
  String get estimateComplete => 'Completo';

  @override
  String get estimatePartial => 'Parcial';

  @override
  String get estimateInformational => 'Informativo';

  @override
  String get estimateTotalsIncomplete =>
      'Faltan datos en algún régimen; las cifras usan ceros donde faltan.';

  @override
  String get estimateIsrPayable => 'ISR a pagar';

  @override
  String get estimateIsrInFavor => 'ISR: saldo a favor';

  @override
  String get estimateIvaPayable => 'IVA a pagar';

  @override
  String get estimateIvaInFavor => 'IVA: saldo a favor';

  @override
  String get estimateWithholdingsToRemit => 'Retenciones a enterar';

  @override
  String get estimateTotalPayable => 'Total a pagar';

  @override
  String estimateDueDate(DateTime date) {
    final intl.DateFormat dateDateFormat = intl.DateFormat.yMMMMd(localeName);
    final String dateString = dateDateFormat.format(date);

    return 'Vence el $dateString';
  }

  @override
  String get estimateDueDateNote => 'Fecha sujeta a confirmación.';

  @override
  String estimateIvaConsolidated(int count) {
    return 'IVA consolidado de $count regímenes: se declara uno solo por RFC.';
  }

  @override
  String get estimateCpaBannerTitle => 'Requiere validación de un contador';

  @override
  String estimateRegimeCode(String code) {
    return 'Régimen $code';
  }

  @override
  String estimateIncompleteAmount(String label) {
    return '$label (incompleto)';
  }

  @override
  String get estimateIncompleteFigures =>
      'Cifras con datos faltantes: se usaron ceros.';

  @override
  String get estimateMissingTitle => 'Falta:';

  @override
  String get estimateCompleteInputs => 'Completa tus datos';

  @override
  String get estimateViewBreakdown => 'Ver desglose';

  @override
  String estimateCpaTopics(String topics) {
    return 'Pendiente de validar por un contador: $topics.';
  }

  @override
  String get estimateIvaCharged => 'IVA trasladado cobrado';

  @override
  String get estimateIvaCreditable => 'IVA acreditable pagado';

  @override
  String get estimateIvaWithheld => 'IVA retenido por clientes';

  @override
  String get estimateIvaExcluded => 'IVA excluido por pago definitivo';

  @override
  String get estimateIvaContributes =>
      'Se suma al IVA del mes; no se paga por separado.';

  @override
  String get cpaTopicRifReduction => 'la reducción del RIF por año';

  @override
  String get cpaTopicRifBimonthlyIva => 'el IVA bimestral del RIF';

  @override
  String get cpaTopicPlatformDefinitiveIva =>
      'el IVA definitivo de plataformas';

  @override
  String get cpaTopicCrossRegimeIvaNetting =>
      'la compensación del IVA entre regímenes';

  @override
  String get cpaTopicResicoPmInterestWithholding =>
      'la retención de intereses en RESICO de personas morales';

  @override
  String get cpaTopicOther => 'otras reglas';

  @override
  String get breakdownTitle => 'Desglose';

  @override
  String get taxInputsTitle => 'Datos fiscales';

  @override
  String get recordErrorInvalid => 'Revisa este dato.';

  @override
  String get expensesNoTaxpayer =>
      'Registra un contribuyente para llevar sus gastos.';

  @override
  String expensesEmptyTitle(String month) {
    return 'Sin gastos en $month';
  }

  @override
  String get expensesEmptyBody =>
      'Registra tus gastos y te mostramos cuáles podrían bajar tus impuestos.';

  @override
  String get expensesRecord => 'Registrar gasto';

  @override
  String get expensesLoadFailed => 'No se pudieron cargar los gastos.';

  @override
  String get expensesLoading => 'Cargando gastos…';

  @override
  String expensesCount(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count gastos',
      one: '1 gasto',
    );
    return '$_temp0';
  }

  @override
  String get expensesTotalTitle => 'Total del mes con IVA';

  @override
  String get expensesTotalSubtotal => 'Sin IVA';

  @override
  String get expensesTotalIva => 'IVA pagado';

  @override
  String get expensesTotalWithheld => 'Retenciones a enterar';

  @override
  String get expensesNotDeductible => 'No deducible';

  @override
  String expensesDeleteTooltip(String title) {
    return 'Eliminar gasto: $title';
  }

  @override
  String expensesEditLabel(String title) {
    return 'Editar gasto: $title';
  }

  @override
  String expensesDeleted(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count gastos eliminados',
      one: 'Gasto eliminado',
    );
    return '$_temp0';
  }

  @override
  String expensesRestored(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: '$count gastos restaurados',
      one: 'Gasto restaurado',
    );
    return '$_temp0';
  }

  @override
  String expensesRestoredSome(int restored, int total) {
    String _temp0 = intl.Intl.pluralLogic(
      restored,
      locale: localeName,
      other: 'Se restauraron $restored de $total gastos.',
      one: 'Se restauró 1 de $total gastos.',
      zero: 'No se restauró ninguno de los $total gastos.',
    );
    return '$_temp0';
  }

  @override
  String get expensesDeleteFailed => 'No se pudo eliminar el gasto.';

  @override
  String get expensesRestoreFailed => 'No se pudo restaurar el gasto.';

  @override
  String get expensesRestoreRegimeEnded =>
      'No se pudo restaurar: el régimen ya no cubre la fecha del gasto.';

  @override
  String get expenseFormTitle => 'Nuevo gasto';

  @override
  String get expenseEditTitle => 'Editar gasto';

  @override
  String get expenseFieldCategory => 'Categoría';

  @override
  String get expenseCategoryNone => 'Elige una categoría';

  @override
  String get expenseCategoryPickerTitle => 'Elige la categoría del gasto';

  @override
  String get expenseCategorySearch => 'Buscar categoría';

  @override
  String get expenseCategoryNoMatch =>
      'Ninguna categoría coincide con la búsqueda.';

  @override
  String get expenseCategoriesLoading => 'Cargando categorías…';

  @override
  String get expenseCategoriesLoadFailed =>
      'No se pudieron cargar las categorías.';

  @override
  String get expenseFieldAmount => 'Monto sin IVA';

  @override
  String get expenseFieldIvaRate => 'Tasa de IVA';

  @override
  String get expenseFieldPaidOn => 'Fecha de pago';

  @override
  String get expenseFieldPaymentMethod => 'Forma de pago';

  @override
  String get expenseFieldHasCfdi => '¿Tienes factura?';

  @override
  String get expenseHasCfdiYes => 'Sí';

  @override
  String get expenseHasCfdiNo => 'No';

  @override
  String get expenseFieldRegime => 'Régimen';

  @override
  String get expenseFieldIvaAmount => 'IVA pagado';

  @override
  String get expenseFieldDeductible => '¿Es deducible?';

  @override
  String get expenseDeductibleSuggested =>
      'Valor sugerido para esta categoría y forma de pago';

  @override
  String get expenseDeductibleUser => 'Lo elegiste tú';

  @override
  String get expenseDeductibleReset => 'Usar el valor sugerido';

  @override
  String get expenseFieldCfdiUuid => 'Folio fiscal (UUID)';

  @override
  String get expenseFieldDescription => 'Descripción';

  @override
  String get expenseFieldIsrWithheld => 'ISR retenido';

  @override
  String get expenseFieldIvaWithheld => 'IVA retenido';

  @override
  String get expenseSectionMore => 'Más detalles';

  @override
  String get expenseSectionWithholdings => 'Retenciones a tu proveedor';

  @override
  String get expenseWithholdingsHelp =>
      'Lo que le retuviste al proveedor, como dice el CFDI. Lo tienes que enterar al SAT.';

  @override
  String expenseSummaryTotal(String total) {
    return 'Total con IVA $total';
  }

  @override
  String get expenseHintDeductible => 'Normalmente deducible para tu régimen.';

  @override
  String get expenseHintPartial => 'Deducible sólo en parte para tu régimen.';

  @override
  String get expenseHintNotDeductible =>
      'Normalmente no es deducible para tu régimen.';

  @override
  String get expenseHintNotIsr => 'No reduce tu ISR en este régimen.';

  @override
  String get expenseHintResico => 'No reduce tu ISR en RESICO.';

  @override
  String get expenseHintResicoIva =>
      'En RESICO tus gastos no reducen tu ISR, pero el IVA que pagaste sí puede acreditarse.';

  @override
  String get expenseHintBlind35 =>
      'Con la deducción ciega del 35 % este gasto no cuenta.';

  @override
  String expenseHintBasis(String basis) {
    return 'Fundamento: $basis';
  }

  @override
  String expenseCashOverLimit(String limit) {
    return 'Los pagos en efectivo de más de $limit no son deducibles.';
  }

  @override
  String get expenseCashFuel =>
      'El combustible pagado en efectivo no es deducible.';

  @override
  String get expenseCashIva =>
      'El IVA de este pago en efectivo no es acreditable.';

  @override
  String get expenseNoCfdi => 'Sin factura, el gasto no es deducible.';

  @override
  String get expenseWarningNotIsrInRegime =>
      'En tu régimen este gasto no reduce el ISR, aunque lo marcaste como deducible.';

  @override
  String get expenseWarningCategoryNotDeductible =>
      'Esta categoría normalmente no es deducible; se guardó como la marcaste.';

  @override
  String get expenseWarningDuplicateCfdi =>
      'Ya hay otro gasto con este folio fiscal.';

  @override
  String get expenseWarningProvisional =>
      'Se usaron los parámetros fiscales del año anterior: los de este año aún no se publican.';

  @override
  String get expenseSaved => 'Gasto guardado';

  @override
  String expenseSavedWithWarnings(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Gasto guardado, con $count avisos:',
      one: 'Gasto guardado, con un aviso:',
    );
    return '$_temp0';
  }

  @override
  String get expenseUpdated => 'Gasto actualizado';

  @override
  String expenseUpdatedWithWarnings(int count) {
    String _temp0 = intl.Intl.pluralLogic(
      count,
      locale: localeName,
      other: 'Gasto actualizado, con $count avisos:',
      one: 'Gasto actualizado, con un aviso:',
    );
    return '$_temp0';
  }

  @override
  String get expenseSaveUnreachable =>
      'No se pudo guardar por falta de conexión. Vuelve a tocar Guardar: no se duplicará.';

  @override
  String get expenseSaveIdReused =>
      'Este gasto ya se había guardado con otros datos. Revisa la lista de gastos.';

  @override
  String get expenseSaveGone =>
      'Este gasto ya no existe: se eliminó en otro lugar.';

  @override
  String get expenseSaveFailed =>
      'No se pudo guardar el gasto. Intenta de nuevo.';

  @override
  String get expenseTrialCapTitle => 'Llegaste al límite de gastos de prueba';

  @override
  String get expenseTrialCapBody =>
      'La prueba incluye 5 gastos y ya los registraste. Para registrar más gastos se necesita un plan. Lo que capturaste sigue en el formulario.';

  @override
  String get expenseDiscardTitle => '¿Descartar este gasto?';

  @override
  String get expenseErrorUnknownCategory => 'Elige una categoría de la lista.';

  @override
  String get expenseErrorHasCfdiMismatch =>
      'Si capturas el folio fiscal, indica que tienes factura.';

  @override
  String get expenseErrorRateNotValid =>
      'Esta tasa no está vigente en la fecha del gasto.';

  @override
  String get expenseErrorWithholdingNotApplicable =>
      'Sólo una persona moral retiene a sus proveedores.';

  @override
  String get paymentCash => 'Efectivo';

  @override
  String get paymentTransfer => 'Transferencia';

  @override
  String get paymentCardDebit => 'Tarjeta de débito';

  @override
  String get paymentCardCredit => 'Tarjeta de crédito';

  @override
  String get paymentCheck => 'Cheque';

  @override
  String get paymentEWallet => 'Monedero electrónico';

  @override
  String get paymentOther => 'Otro';

  @override
  String get breakdownLoading => 'Cargando el desglose…';

  @override
  String breakdownSubtitle(String code, String month) {
    return 'Régimen $code · $month';
  }

  @override
  String breakdownVersions(String year, String version, String engine) {
    return 'Parámetros $year v$version · Motor $engine';
  }

  @override
  String get breakdownGroupIsr => 'ISR';

  @override
  String get breakdownGroupIva => 'IVA';

  @override
  String get breakdownWithholdingsHelp =>
      'Lo que retuviste a otros y debes enterar al SAT. Se paga aparte de tus impuestos.';

  @override
  String get breakdownIvaHelp =>
      'El IVA del mes es uno solo: sumamos lo cobrado y lo pagado en todos tus regímenes y resulta un solo pago (LIVA 5-D).';

  @override
  String get breakdownIvaNettingUnverified =>
      'Esta suma entre regímenes aún no la valida un contador; tómala con reserva.';

  @override
  String breakdownRate(String rate) {
    return 'Tasa: $rate';
  }

  @override
  String get breakdownSourceRecords => 'De tus registros';

  @override
  String get breakdownSourceComputed => 'Calculado';

  @override
  String get breakdownSourceInput => 'Dato que capturaste';

  @override
  String get breakdownSourceParams => 'Tarifa oficial';

  @override
  String get breakdownCpaLine => 'Sin validar por un contador';

  @override
  String breakdownLegalBasis(String references) {
    return 'Fundamento: $references';
  }

  @override
  String breakdownReadLaw(String law) {
    return 'Leer: $law';
  }

  @override
  String get breakdownRoundingHelp =>
      'Cada línea se muestra redondeada al centavo; este ajuste hace que la suma coincida con el resultado. No es un cobro extra.';

  @override
  String get breakdownNoLines =>
      'Este régimen no tiene desglose: no estimamos su impuesto.';

  @override
  String get breakdownRegimeNotActive => 'No tenías este régimen en ese mes.';

  @override
  String get taxInputsLoading => 'Cargando tus datos fiscales…';

  @override
  String get taxInputsLoadFailed => 'No se pudieron cargar tus datos fiscales.';

  @override
  String taxInputsSubtitle(String code, String year) {
    return 'Régimen $code · Ejercicio $year';
  }

  @override
  String get taxInputsIntro =>
      'Datos del ejercicio que tu estimado necesita y no puede sacar de tus registros. Lo que no captures se calcula con cero y el estimado queda Incompleto.';

  @override
  String get taxInputsNeeded => 'Tu estimado lo necesita.';

  @override
  String get taxInputsMonthlyHelp =>
      'Lo que pagaste cada mes, no el estimado. Escribe 0 si no pagaste; un mes vacío cuenta como dato faltante.';

  @override
  String get taxInputsNoMonthsYet =>
      'Aún no hay meses pagados en este ejercicio.';

  @override
  String taxInputsSuggestion(String amount) {
    return 'Estimado por la app: $amount — confírmalo';
  }

  @override
  String get taxInputsUseSuggestion => 'Usar este monto';

  @override
  String taxInputsMonthLabel(String month) {
    return 'Pago de $month';
  }

  @override
  String taxInputsMonthsRange(String first, String last) {
    return '$first a $last';
  }

  @override
  String get taxInputsSaved => 'Guardamos tus datos. Tu estimado ya los usa.';

  @override
  String get taxInputsSaveFailed =>
      'No se pudieron guardar tus datos. Inténtalo de nuevo.';

  @override
  String get taxInputsArchived =>
      'Este contribuyente está archivado: sus datos fiscales no se pueden cambiar.';

  @override
  String get taxInputsRegimeNotActive =>
      'No tienes este régimen en ese ejercicio.';

  @override
  String get taxInputsUnknownInput => 'Este dato no aplica a este régimen.';

  @override
  String get taxInputsNegative => 'El monto no puede ser negativo.';

  @override
  String get taxInputsInvalid => 'Revisa este dato.';

  @override
  String get taxInputsDiscardTitle => '¿Descartar los cambios?';

  @override
  String get taxInputsDiscardBody => 'Lo que escribiste se perderá.';

  @override
  String get taxInputsDiscardKeep => 'Seguir editando';

  @override
  String get taxInputsDiscardConfirm => 'Descartar';

  @override
  String get dashboardLoading => 'Cargando el inicio…';

  @override
  String get dashboardLoadFailed => 'No se pudo cargar el inicio.';

  @override
  String get dashboardNoTaxpayer =>
      'Registre un contribuyente para ver su inicio.';

  @override
  String get dashboardCompleteOnboarding => 'Completar registro';

  @override
  String get dashboardEmpty =>
      'Aún no hay ingresos ni gastos registrados. El estimado aparece con el primer registro.';

  @override
  String get dashboardRegisterIncome => 'Registrar ingreso';

  @override
  String get dashboardRegisterExpense => 'Registrar gasto';

  @override
  String get dashboardSeeDetail => 'Ver detalle';

  @override
  String get dashboardCompleteData => 'Completa tus datos';

  @override
  String get dashboardInformational =>
      'Este régimen es informativo: la app no estima su impuesto';

  @override
  String get dashboardMonthTitle => 'Este mes';

  @override
  String get dashboardIncomeCollected => 'Ingresos cobrados';

  @override
  String get dashboardExpensePaid => 'Gastos pagados';

  @override
  String get dashboardDifference => 'Diferencia';

  @override
  String get dashboardBeforeIva =>
      'Montos antes de IVA, por fecha de cobro y de pago.';

  @override
  String dashboardTrendTitle(int count) {
    return 'Últimos $count meses';
  }

  @override
  String get dashboardViewTaxes => 'Impuestos';

  @override
  String get dashboardViewIncomeExpense => 'Ingresos y gastos';

  @override
  String get dashboardSeriesIva => 'IVA';

  @override
  String get dashboardSeriesIsr => 'ISR';

  @override
  String get dashboardSeriesWithholdings => 'Retenciones';

  @override
  String get dashboardTotal => 'Total';

  @override
  String get dashboardIvaInFavor => 'IVA: saldo a favor';

  @override
  String get dashboardMarkerKey => '▲ Mes incompleto o parcial';

  @override
  String get dashboardShowTable => 'Ver como tabla';

  @override
  String get dashboardShowChart => 'Ver como gráfica';

  @override
  String dashboardTaxSummary(
    int count,
    String month,
    String amount,
    String status,
  ) {
    return 'Impuestos estimados, últimos $count meses. $month: $amount$status.';
  }

  @override
  String dashboardTaxPeak(String month, String amount) {
    return ' Mes más alto: $month, $amount.';
  }

  @override
  String dashboardIncomeExpenseSummary(
    int count,
    String month,
    String current,
    String window,
  ) {
    return 'Ingresos y gastos, últimos $count meses. $month: $current. En el periodo: $window.';
  }

  @override
  String dashboardTrialBanner(int days) {
    String _temp0 = intl.Intl.pluralLogic(
      days,
      locale: localeName,
      other: 'quedan $days días',
      one: 'queda 1 día',
    );
    return 'Prueba gratuita: $_temp0';
  }

  @override
  String dashboardTrialRecords(int incomes, int expenses) {
    String _temp0 = intl.Intl.pluralLogic(
      incomes,
      locale: localeName,
      other: '$incomes ingresos',
      one: '1 ingreso',
    );
    String _temp1 = intl.Intl.pluralLogic(
      expenses,
      locale: localeName,
      other: '$expenses gastos',
      one: '1 gasto',
    );
    return ' · $_temp0 y $_temp1';
  }

  @override
  String get dashboardTrialScope =>
      'Los registros disponibles se cuentan entre todos sus contribuyentes.';

  @override
  String get dashboardReadOnlyBanner =>
      'Modo lectura: los datos siguen visibles. Para registrar, contrate un plan.';

  @override
  String get dashboardSeePlans => 'Ver planes';

  @override
  String get dashboardPlansTitle => 'Planes';

  @override
  String get dashboardPlansBody =>
      'La contratación de planes estará disponible próximamente.';
}

/// The translations for Spanish Castilian, as used in Mexico (`es_MX`).
class AppLocalizationsEsMx extends AppLocalizationsEs {
  AppLocalizationsEsMx() : super('es_MX');
}
