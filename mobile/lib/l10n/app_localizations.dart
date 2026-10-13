import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_es.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations)!;
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('es'),
    Locale('es', 'MX'),
  ];

  /// App name shown by the OS task switcher.
  ///
  /// In es, this message translates to:
  /// **'Control Contable'**
  String get appTitle;

  /// Bottom bar item and title of the home screen.
  ///
  /// In es, this message translates to:
  /// **'Inicio'**
  String get navHome;

  /// Bottom bar item: income list.
  ///
  /// In es, this message translates to:
  /// **'Ingresos'**
  String get navIncome;

  /// Bottom bar item: expense list.
  ///
  /// In es, this message translates to:
  /// **'Gastos'**
  String get navExpenses;

  /// Bottom bar item: monthly history.
  ///
  /// In es, this message translates to:
  /// **'Historial'**
  String get navHistory;

  /// Bottom bar item: settings.
  ///
  /// In es, this message translates to:
  /// **'Ajustes'**
  String get navSettings;

  /// Scaffold placeholder on the home screen until the dashboard module ships.
  ///
  /// In es, this message translates to:
  /// **'Aquí aparecerá el estimado del mes.'**
  String get homePlaceholder;

  /// Mandatory legend under every estimate. Never hidden, never dismissible.
  ///
  /// In es, this message translates to:
  /// **'Esto es un aproximado. No todos los gastos aplican para tu giro comercial.'**
  String get estimateLegend;

  /// Title of the screen shown for an unknown route.
  ///
  /// In es, this message translates to:
  /// **'Página no encontrada'**
  String get notFoundTitle;

  /// Button on the not-found screen that returns to home.
  ///
  /// In es, this message translates to:
  /// **'Ir al inicio'**
  String get notFoundAction;

  /// Splash shown while a deep link waits for the module flags.
  ///
  /// In es, this message translates to:
  /// **'Cargando…'**
  String get bootLoading;

  /// Text mark after the label of an optional field (forms mark the fewer kind).
  ///
  /// In es, this message translates to:
  /// **'(opcional)'**
  String get formMarkOptional;

  /// Text mark after the label of a required field (forms mark the fewer kind).
  ///
  /// In es, this message translates to:
  /// **'(obligatorio)'**
  String get formMarkRequired;

  /// Inline error: a required field was left empty. Shown only after submit.
  ///
  /// In es, this message translates to:
  /// **'Escribe este dato.'**
  String get formErrorRequired;

  /// Inline error: the text is not a peso amount.
  ///
  /// In es, this message translates to:
  /// **'Escribe un monto, por ejemplo 1,234.50.'**
  String get formErrorMoney;

  /// Inline error: the text is not a real calendar date (day/month/year).
  ///
  /// In es, this message translates to:
  /// **'Escribe una fecha como 31/01/2026.'**
  String get formErrorDate;

  /// Heading of the error summary shown at the top of a form after a failed submit.
  ///
  /// In es, this message translates to:
  /// **'{count, plural, =1{Revisa 1 dato antes de guardar} other{Revisa {count} datos antes de guardar}}'**
  String formErrorSummaryTitle(int count);

  /// One link in the error summary; tapping it moves focus to the field.
  ///
  /// In es, this message translates to:
  /// **'{label}: {error}'**
  String formErrorSummaryItem(String label, String error);

  /// Primary form button.
  ///
  /// In es, this message translates to:
  /// **'Guardar'**
  String get formSave;

  /// Screen reader label of the primary button while saving.
  ///
  /// In es, this message translates to:
  /// **'Guardando…'**
  String get formSaving;

  /// Secondary form button on repetitive entry: saves, keeps context fields, clears the rest.
  ///
  /// In es, this message translates to:
  /// **'Guardar y crear otro'**
  String get formSaveAndCreateAnother;

  /// Text button that leaves a form without saving.
  ///
  /// In es, this message translates to:
  /// **'Cancelar'**
  String get formCancel;

  /// Hint inside an empty date field (never a label).
  ///
  /// In es, this message translates to:
  /// **'dd/mm/aaaa'**
  String get formDateHint;

  /// Inline error: a required checkbox (a consent) was left unchecked. Shown only after submit.
  ///
  /// In es, this message translates to:
  /// **'Marca esta casilla para continuar.'**
  String get formErrorNotAccepted;

  /// Tooltip and screen reader label of the toggle that reveals a password.
  ///
  /// In es, this message translates to:
  /// **'Mostrar contraseña'**
  String get formShowPassword;

  /// Tooltip and screen reader label of the toggle that hides a revealed password.
  ///
  /// In es, this message translates to:
  /// **'Ocultar contraseña'**
  String get formHidePassword;

  /// Home entry point to the sign-up form (shown only while module_auth is on).
  ///
  /// In es, this message translates to:
  /// **'Crear cuenta'**
  String get homeCreateAccount;

  /// Title of the sign-up screen.
  ///
  /// In es, this message translates to:
  /// **'Crear cuenta'**
  String get authRegisterTitle;

  /// One-line reason to sign up, at the top of the sign-up form.
  ///
  /// In es, this message translates to:
  /// **'Con una cuenta, los ingresos, gastos y estimados quedan guardados de forma segura.'**
  String get authRegisterIntro;

  /// Primary button of the sign-up form.
  ///
  /// In es, this message translates to:
  /// **'Crear cuenta'**
  String get authRegisterSubmit;

  /// Label of the optional name field on the sign-up form.
  ///
  /// In es, this message translates to:
  /// **'Nombre'**
  String get authNameLabel;

  /// Label of the email field.
  ///
  /// In es, this message translates to:
  /// **'Correo electrónico'**
  String get authEmailLabel;

  /// Label of the password field.
  ///
  /// In es, this message translates to:
  /// **'Contraseña'**
  String get authPasswordLabel;

  /// Help under the new-password field: the rule plus a gentle tip.
  ///
  /// In es, this message translates to:
  /// **'Al menos 12 caracteres. Una frase fácil de recordar funciona bien.'**
  String get authPasswordHelp;

  /// Inline error: the email is not a valid address. Shown on blur.
  ///
  /// In es, this message translates to:
  /// **'Escribe un correo como nombre@ejemplo.mx.'**
  String get authEmailInvalid;

  /// Inline error: the name is longer than the contract allows.
  ///
  /// In es, this message translates to:
  /// **'Escribe un nombre de hasta 120 caracteres.'**
  String get authNameTooLong;

  /// Inline error: the password is shorter than 12 characters.
  ///
  /// In es, this message translates to:
  /// **'Usa al menos 12 caracteres.'**
  String get authPasswordTooShort;

  /// Inline error: the password is longer than 128 characters.
  ///
  /// In es, this message translates to:
  /// **'Usa como máximo 128 caracteres.'**
  String get authPasswordTooLong;

  /// Inline error from the server: the password is common or breached.
  ///
  /// In es, this message translates to:
  /// **'Esta contraseña es muy común o apareció en una filtración. Elige otra.'**
  String get authPasswordPolicy;

  /// Inline error from the server for a field without a more specific message.
  ///
  /// In es, this message translates to:
  /// **'Revisa este dato.'**
  String get authFieldRejected;

  /// Form-level message when the server rejects something no field shows.
  ///
  /// In es, this message translates to:
  /// **'Revisa los datos e intenta de nuevo.'**
  String get authFormRejected;

  /// Heading above the consent checkboxes on the sign-up form.
  ///
  /// In es, this message translates to:
  /// **'Documentos legales'**
  String get authConsentsHeading;

  /// Consent checkbox for the privacy notice (required to sign up).
  ///
  /// In es, this message translates to:
  /// **'Acepto el aviso de privacidad'**
  String get authConsentPrivacy;

  /// Consent checkbox for the terms and conditions (required to sign up).
  ///
  /// In es, this message translates to:
  /// **'Acepto los términos y condiciones'**
  String get authConsentTerms;

  /// Link under the privacy consent that opens the current privacy notice.
  ///
  /// In es, this message translates to:
  /// **'Leer el aviso de privacidad'**
  String get authReadPrivacy;

  /// Link under the terms consent that opens the current terms.
  ///
  /// In es, this message translates to:
  /// **'Leer los términos y condiciones'**
  String get authReadTerms;

  /// Shown while the legal documents load.
  ///
  /// In es, this message translates to:
  /// **'Cargando documentos legales…'**
  String get authDocumentsLoading;

  /// Shown when the legal documents could not be loaded; a retry button follows.
  ///
  /// In es, this message translates to:
  /// **'No se pudieron cargar los documentos legales.'**
  String get authDocumentsFailed;

  /// Form-level message: submit pressed before the legal documents loaded.
  ///
  /// In es, this message translates to:
  /// **'Los documentos legales no se han cargado. Reintenta para continuar.'**
  String get authDocumentsNotLoaded;

  /// Error on the consents: the server published new versions; they must be accepted again.
  ///
  /// In es, this message translates to:
  /// **'Los documentos se actualizaron. Revísalos y acéptalos de nuevo.'**
  String get authDocumentsStale;

  /// Snack bar: a legal document link could not be opened.
  ///
  /// In es, this message translates to:
  /// **'No se pudo abrir el documento. Intenta de nuevo.'**
  String get authDocumentOpenFailed;

  /// Button that retries loading something that failed.
  ///
  /// In es, this message translates to:
  /// **'Reintentar'**
  String get authRetry;

  /// Form-level message for 429 with a Retry-After, rounded up to minutes.
  ///
  /// In es, this message translates to:
  /// **'{minutes, plural, =1{Demasiados intentos. Espera 1 min y vuelve a intentarlo.} other{Demasiados intentos. Espera {minutes} min y vuelve a intentarlo.}}'**
  String authErrorRateLimited(int minutes);

  /// Form-level message for 429 without a Retry-After.
  ///
  /// In es, this message translates to:
  /// **'Demasiados intentos. Espera unos minutos y vuelve a intentarlo.'**
  String get authErrorRateLimitedNoWait;

  /// Form-level message: offline or timeout.
  ///
  /// In es, this message translates to:
  /// **'No hay conexión con el servidor. Revisa la conexión a internet e intenta de nuevo.'**
  String get authErrorUnreachable;

  /// Form-level message for 404 (the module may be off on the server).
  ///
  /// In es, this message translates to:
  /// **'Este servicio no está disponible por ahora.'**
  String get authErrorUnavailable;

  /// Form-level message for a server error or an unreadable answer.
  ///
  /// In es, this message translates to:
  /// **'Algo salió mal. Intenta de nuevo en unos minutos.'**
  String get authErrorUnexpected;

  /// Title of the verify-email screen.
  ///
  /// In es, this message translates to:
  /// **'Confirmar correo'**
  String get authVerifyTitle;

  /// Where the code went and how long it lasts.
  ///
  /// In es, this message translates to:
  /// **'Enviamos un código de 6 dígitos a {email}. Vence en 15 minutos.'**
  String authVerifyIntro(String email);

  /// Help under the verify intro: check the spam folder.
  ///
  /// In es, this message translates to:
  /// **'Si no aparece, conviene revisar la carpeta de correo no deseado.'**
  String get authVerifySpamHint;

  /// Label of the 6-digit code field.
  ///
  /// In es, this message translates to:
  /// **'Código de verificación'**
  String get authCodeLabel;

  /// Inline error: the code is not 6 digits. Shown on blur.
  ///
  /// In es, this message translates to:
  /// **'Escribe los 6 dígitos del código.'**
  String get authCodeFormat;

  /// Inline error: the server rejected the code (one generic message: wrong, expired, used up).
  ///
  /// In es, this message translates to:
  /// **'El código no es válido o ya venció. Revisa el correo o pide uno nuevo.'**
  String get authCodeRejected;

  /// Primary button of the verify-email screen.
  ///
  /// In es, this message translates to:
  /// **'Confirmar'**
  String get authVerifySubmit;

  /// Button that asks for a new code once the cooldown is over.
  ///
  /// In es, this message translates to:
  /// **'Enviar otro código'**
  String get authResend;

  /// Countdown before a new code can be asked for, as m:ss.
  ///
  /// In es, this message translates to:
  /// **'Otro código disponible en {time}.'**
  String authResendWait(String time);

  /// Generic confirmation after a resend (never confirms that the email exists).
  ///
  /// In es, this message translates to:
  /// **'Si hay un registro pendiente con ese correo, enviamos un código nuevo.'**
  String get authResendSent;

  /// Text button that returns to the sign-up form to fix the email.
  ///
  /// In es, this message translates to:
  /// **'Cambiar correo'**
  String get authChangeEmail;

  /// Verify screen opened without a sign-up in progress.
  ///
  /// In es, this message translates to:
  /// **'No hay un registro en curso.'**
  String get authVerifyNoPending;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['es'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when language+country codes are specified.
  switch (locale.languageCode) {
    case 'es':
      {
        switch (locale.countryCode) {
          case 'MX':
            return AppLocalizationsEsMx();
        }
        break;
      }
  }

  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'es':
      return AppLocalizationsEs();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}
