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

  /// Button next to a failed load (a list, a catalog, a document) that loads it again.
  ///
  /// In es, this message translates to:
  /// **'Reintentar'**
  String get loadRetry;

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

  /// Home app-bar action that ends the session (shown only while module_auth is on). Works offline: the device forgets the session even when the server cannot be told.
  ///
  /// In es, this message translates to:
  /// **'Cerrar sesión'**
  String get authSignOut;

  /// Title of the login screen, where a signed-out user is sent.
  ///
  /// In es, this message translates to:
  /// **'Iniciar sesión'**
  String get authLoginTitle;

  /// Primary button of the login form.
  ///
  /// In es, this message translates to:
  /// **'Entrar'**
  String get authLoginSubmit;

  /// Login 401: one generic message for a wrong password, an unknown email or an unverified sign-up (no account enumeration).
  ///
  /// In es, this message translates to:
  /// **'Correo o contraseña incorrectos.'**
  String get authLoginRejected;

  /// Lead-in before the link from the login form to the sign-up form.
  ///
  /// In es, this message translates to:
  /// **'¿Aún no tienes cuenta?'**
  String get authLoginNoAccount;

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

  /// Title of the account (profile) screen and tooltip of its entry on home.
  ///
  /// In es, this message translates to:
  /// **'Mi cuenta'**
  String get accountTitle;

  /// Status while the profile loads.
  ///
  /// In es, this message translates to:
  /// **'Cargando tu cuenta…'**
  String get accountLoading;

  /// The profile could not be loaded; a retry button follows.
  ///
  /// In es, this message translates to:
  /// **'No se pudo cargar tu cuenta.'**
  String get accountLoadFailed;

  /// Note under the read-only email of the account screen.
  ///
  /// In es, this message translates to:
  /// **'El correo no se puede cambiar.'**
  String get accountEmailReadOnly;

  /// Primary button of the account screen (saves the name).
  ///
  /// In es, this message translates to:
  /// **'Guardar cambios'**
  String get accountSave;

  /// Confirmation after the name is saved.
  ///
  /// In es, this message translates to:
  /// **'Guardamos tu nombre.'**
  String get accountSaved;

  /// Heading of the account-deletion section.
  ///
  /// In es, this message translates to:
  /// **'Eliminar cuenta'**
  String get accountDeleteHeading;

  /// One line under the account-deletion heading.
  ///
  /// In es, this message translates to:
  /// **'Borra tu cuenta y todos sus datos de forma permanente.'**
  String get accountDeleteIntro;

  /// Button that opens the account-deletion confirmation.
  ///
  /// In es, this message translates to:
  /// **'Eliminar mi cuenta'**
  String get accountDeleteOpen;

  /// Title of the account-deletion confirmation dialog.
  ///
  /// In es, this message translates to:
  /// **'¿Eliminar tu cuenta?'**
  String get accountDeleteTitle;

  /// What the deletion removes; it is irreversible.
  ///
  /// In es, this message translates to:
  /// **'Se borrarán para siempre tu cuenta y todo lo que guardaste: contribuyentes, ingresos, gastos, estimados y recordatorios. No se puede deshacer.'**
  String get accountDeleteWhat;

  /// The trial registry keeps the email hash: no second free trial.
  ///
  /// In es, this message translates to:
  /// **'La prueba gratuita no se reinicia: si vuelves a registrarte con este correo, no tendrás otra.'**
  String get accountDeleteTrial;

  /// A store subscription is not cancelled by the deletion.
  ///
  /// In es, this message translates to:
  /// **'Si pagas una suscripción, cancélala en la tienda de tu teléfono (App Store o Google Play): eliminar la cuenta no la cancela.'**
  String get accountDeleteSubscription;

  /// Password re-entered to confirm the deletion.
  ///
  /// In es, this message translates to:
  /// **'Contraseña actual'**
  String get accountDeletePasswordLabel;

  /// Destructive primary button of the deletion dialog.
  ///
  /// In es, this message translates to:
  /// **'Eliminar cuenta'**
  String get accountDeleteConfirm;

  /// Busy state of the deletion button, read by screen readers.
  ///
  /// In es, this message translates to:
  /// **'Eliminando cuenta…'**
  String get accountDeleting;

  /// The re-entered password was refused; nothing was deleted.
  ///
  /// In es, this message translates to:
  /// **'La contraseña no es correcta.'**
  String get accountDeleteWrongPassword;

  /// Countdown after a 429 (Retry-After) on login or account deletion, as m:ss; a tap during it sends nothing.
  ///
  /// In es, this message translates to:
  /// **'Podrás intentar de nuevo en {time}.'**
  String authRetryIn(String time);

  /// Title of the screen shown after the account is deleted.
  ///
  /// In es, this message translates to:
  /// **'Cuenta eliminada'**
  String get goodbyeTitle;

  /// Body of the goodbye screen.
  ///
  /// In es, this message translates to:
  /// **'Eliminamos tu cuenta y todos sus datos. Gracias por usar Control Contable.'**
  String get goodbyeBody;

  /// Button from the goodbye screen to the login form.
  ///
  /// In es, this message translates to:
  /// **'Ir a iniciar sesión'**
  String get goodbyeToLogin;

  /// Sign-out could not remove the stored session token, even by wiping the secure store.
  ///
  /// In es, this message translates to:
  /// **'No pudimos borrar la sesión guardada en este teléfono. Cierra la app y vuelve a abrirla; si el aviso sigue, borra los datos de la app desde los ajustes del teléfono.'**
  String get sessionTokenLeftOnDevice;

  /// Inline error: a required choice (radio or checkbox group) was left empty. Shown only after submit.
  ///
  /// In es, this message translates to:
  /// **'Elige una opción.'**
  String get formErrorChoose;

  /// Home, when the active taxpayer has no regime in effect this month; followed by the add-regime button.
  ///
  /// In es, this message translates to:
  /// **'Agrega un régimen para ver tu estimado.'**
  String get homeNoRegime;

  /// Home button that starts the taxpayer onboarding (module_taxpayers).
  ///
  /// In es, this message translates to:
  /// **'Registrar contribuyente'**
  String get homeAddTaxpayer;

  /// App bar title of every onboarding step.
  ///
  /// In es, this message translates to:
  /// **'Nuevo contribuyente'**
  String get onboardingTitle;

  /// Onboarding progress, read before the step title.
  ///
  /// In es, this message translates to:
  /// **'Paso {step} de {total}'**
  String onboardingProgress(int step, int total);

  /// Primary button that moves to the next onboarding step.
  ///
  /// In es, this message translates to:
  /// **'Siguiente'**
  String get onboardingNext;

  /// Primary button of the last onboarding step before the taxpayer is created.
  ///
  /// In es, this message translates to:
  /// **'Crear contribuyente'**
  String get onboardingCreate;

  /// Primary button after the taxpayer was created.
  ///
  /// In es, this message translates to:
  /// **'Continuar'**
  String get onboardingContinue;

  /// Step 1 title: the person type is chosen before anything else (OD-3).
  ///
  /// In es, this message translates to:
  /// **'¿Persona física o persona moral?'**
  String get onboardingPersonTypeTitle;

  /// Step 1 explanation.
  ///
  /// In es, this message translates to:
  /// **'Elige cómo está dado de alta el contribuyente ante el SAT. No se puede cambiar después.'**
  String get onboardingPersonTypeIntro;

  /// Step 1 choice card title.
  ///
  /// In es, this message translates to:
  /// **'Persona física'**
  String get onboardingFisicaTitle;

  /// Step 1 plain-language explanation of persona física.
  ///
  /// In es, this message translates to:
  /// **'Una persona que trabaja por su cuenta: honorarios, rentas, ventas por plataformas digitales o un negocio propio. Su RFC tiene 13 caracteres.'**
  String get onboardingFisicaBody;

  /// Step 1 choice card title.
  ///
  /// In es, this message translates to:
  /// **'Persona moral'**
  String get onboardingMoralTitle;

  /// Step 1 plain-language explanation of persona moral.
  ///
  /// In es, this message translates to:
  /// **'Una empresa o sociedad (S.A., S. de R.L., A.C. y otras). Su RFC tiene 12 caracteres.'**
  String get onboardingMoralBody;

  /// Step 2 title.
  ///
  /// In es, this message translates to:
  /// **'RFC y nombre'**
  String get onboardingIdentityTitle;

  /// Step 2 field: the taxpayer display name.
  ///
  /// In es, this message translates to:
  /// **'Nombre para identificarlo'**
  String get onboardingDisplayNameLabel;

  /// Help under the display name field.
  ///
  /// In es, this message translates to:
  /// **'Así aparecerá en la app; por ejemplo, tu nombre o «Mi negocio».'**
  String get onboardingDisplayNameHelp;

  /// Display name over 80 characters.
  ///
  /// In es, this message translates to:
  /// **'Escribe un nombre de hasta 80 caracteres.'**
  String get onboardingDisplayNameTooLong;

  /// Step 2 field: the RFC.
  ///
  /// In es, this message translates to:
  /// **'RFC'**
  String get onboardingRfcLabel;

  /// Help under the RFC field for a persona física.
  ///
  /// In es, this message translates to:
  /// **'13 caracteres, como en la constancia de situación fiscal. Puede llevar espacios o guiones.'**
  String get onboardingRfcHelpFisica;

  /// Help under the RFC field for a persona moral.
  ///
  /// In es, this message translates to:
  /// **'12 caracteres, como en la constancia de situación fiscal. Puede llevar espacios o guiones.'**
  String get onboardingRfcHelpMoral;

  /// RFC error: wrong length for a persona física.
  ///
  /// In es, this message translates to:
  /// **'El RFC de una persona física tiene 13 caracteres.'**
  String get onboardingRfcLengthFisica;

  /// RFC error: wrong length for a persona moral.
  ///
  /// In es, this message translates to:
  /// **'El RFC de una persona moral tiene 12 caracteres.'**
  String get onboardingRfcLengthMoral;

  /// RFC error: characters outside the RFC shape.
  ///
  /// In es, this message translates to:
  /// **'Revisa el RFC: primero las letras, luego los 6 números de la fecha y al final la homoclave.'**
  String get onboardingRfcFormat;

  /// RFC error: the date part is not a real date.
  ///
  /// In es, this message translates to:
  /// **'Los 6 números del RFC no forman una fecha válida (año, mes y día).'**
  String get onboardingRfcDate;

  /// RFC error: XAXX010101000 or XEXX010101000.
  ///
  /// In es, this message translates to:
  /// **'Ese es un RFC genérico del SAT. Escribe el RFC propio del contribuyente.'**
  String get onboardingRfcGeneric;

  /// RFC error: 409 rfc_already_registered.
  ///
  /// In es, this message translates to:
  /// **'Ya registraste un contribuyente con este RFC. Si lo archivaste, puedes reactivarlo en Ajustes.'**
  String get onboardingRfcDuplicate;

  /// Non-blocking warning: the RFC check digit does not match (D-RFC).
  ///
  /// In es, this message translates to:
  /// **'Revisa que tu RFC esté bien escrito'**
  String get onboardingRfcCheckWarning;

  /// Explanation under the check-digit warning.
  ///
  /// In es, this message translates to:
  /// **'El último carácter no coincide con el que calcula el SAT. Algunos RFC reales no coinciden: si el tuyo es igual al de tu constancia, continúa. Si no, corrígelo aquí.'**
  String get onboardingRfcCheckBody;

  /// Step 3 title.
  ///
  /// In es, this message translates to:
  /// **'Regímenes fiscales'**
  String get onboardingRegimesTitle;

  /// Step 3 explanation for a persona física (multi-select).
  ///
  /// In es, this message translates to:
  /// **'Elige todos los regímenes en los que está dado de alta. Aparecen en la constancia de situación fiscal.'**
  String get onboardingRegimesIntroFisica;

  /// Step 3 explanation for a persona moral (single-select).
  ///
  /// In es, this message translates to:
  /// **'Elige el régimen de la empresa. Una persona moral tiene un solo régimen a la vez.'**
  String get onboardingRegimesIntroMoral;

  /// Name of the regime selection in the error summary.
  ///
  /// In es, this message translates to:
  /// **'Regímenes'**
  String get onboardingRegimesLabel;

  /// Step 3 error: nothing selected.
  ///
  /// In es, this message translates to:
  /// **'Elige al menos un régimen.'**
  String get onboardingRegimesRequired;

  /// Step 3 error summary item: an incompatible pair is selected.
  ///
  /// In es, this message translates to:
  /// **'Hay regímenes que no se pueden combinar.'**
  String get onboardingRegimesConflict;

  /// Inline error under a regime incompatible with another selected one.
  ///
  /// In es, this message translates to:
  /// **'No se puede combinar con el {code}: {reason}'**
  String onboardingRegimeConflict(String code, String reason);

  /// Status while the regime catalog loads.
  ///
  /// In es, this message translates to:
  /// **'Cargando regímenes…'**
  String get onboardingRegimesLoading;

  /// The regime catalog failed to load.
  ///
  /// In es, this message translates to:
  /// **'No se pudieron cargar los regímenes.'**
  String get onboardingRegimesLoadFailed;

  /// The regime catalog came back empty.
  ///
  /// In es, this message translates to:
  /// **'No hay regímenes disponibles para este tipo de persona.'**
  String get onboardingRegimesEmpty;

  /// Server answered 422 invalid_regime_combination.
  ///
  /// In es, this message translates to:
  /// **'Esta combinación de regímenes no se puede registrar. Revisa tu selección.'**
  String get onboardingRegimeCombinationRejected;

  /// Server rejected a regime code (400 on regimes.N.regime_code).
  ///
  /// In es, this message translates to:
  /// **'Uno de los regímenes elegidos no se aceptó. Revisa tu selección.'**
  String get onboardingRegimeRejected;

  /// Regime group: the app estimates ISR and IVA.
  ///
  /// In es, this message translates to:
  /// **'Cálculo completo'**
  String get onboardingSupportFullTitle;

  /// Plain explanation of the full support level.
  ///
  /// In es, this message translates to:
  /// **'La app estima cada mes el ISR y el IVA a pagar con los ingresos y gastos que registres.'**
  String get onboardingSupportFullBody;

  /// Regime group: the app estimates part of the taxes.
  ///
  /// In es, this message translates to:
  /// **'Cálculo parcial'**
  String get onboardingSupportPartialTitle;

  /// Plain explanation of the partial support level.
  ///
  /// In es, this message translates to:
  /// **'La app estima solo una parte de los impuestos (por ejemplo, solo el IVA). Lo demás conviene revisarlo con un contador.'**
  String get onboardingSupportPartialBody;

  /// Regime group: no estimate, explanation only.
  ///
  /// In es, this message translates to:
  /// **'Solo explicación'**
  String get onboardingSupportInfoTitle;

  /// Plain explanation of the informational support level.
  ///
  /// In es, this message translates to:
  /// **'La app no calcula impuestos para estos regímenes: explica en palabras sencillas qué obligaciones tienen.'**
  String get onboardingSupportInfoBody;

  /// Step 4 title.
  ///
  /// In es, this message translates to:
  /// **'Opciones del régimen'**
  String get onboardingOptionsTitle;

  /// Step 4 heading: which selected regime these options belong to.
  ///
  /// In es, this message translates to:
  /// **'Régimen {current} de {total}: {code} · {name}'**
  String onboardingOptionsRegime(
    int current,
    int total,
    String code,
    String name,
  );

  /// Server rejected an option value (regimes.N.options.key).
  ///
  /// In es, this message translates to:
  /// **'Revisa esta opción.'**
  String get onboardingOptionRejected;

  /// Help under a rate field with bounds.
  ///
  /// In es, this message translates to:
  /// **'Un número entre {min} y {max}, con punto decimal.'**
  String onboardingRateHelp(String min, String max);

  /// Rate error: not a decimal number.
  ///
  /// In es, this message translates to:
  /// **'Escribe un número con punto decimal, por ejemplo 0.25.'**
  String get onboardingRateInvalid;

  /// Rate error: outside the option bounds.
  ///
  /// In es, this message translates to:
  /// **'Escribe un número entre {min} y {max}.'**
  String onboardingRateOutOfRange(String min, String max);

  /// Option value blind_35 (606 deduction mode).
  ///
  /// In es, this message translates to:
  /// **'Deducción ciega: 35 % de los ingresos'**
  String get optionValueBlind35;

  /// Option value actual (606 deduction mode).
  ///
  /// In es, this message translates to:
  /// **'Gastos reales con factura'**
  String get optionValueActual;

  /// Option value transport_delivery (625 activities).
  ///
  /// In es, this message translates to:
  /// **'Transporte de pasajeros o entrega de bienes'**
  String get optionValueTransportDelivery;

  /// Option value lodging (625 activities).
  ///
  /// In es, this message translates to:
  /// **'Hospedaje'**
  String get optionValueLodging;

  /// Option value goods_services (625 activities).
  ///
  /// In es, this message translates to:
  /// **'Venta de bienes y prestación de servicios'**
  String get optionValueGoodsServices;

  /// Option value of rif_year (621).
  ///
  /// In es, this message translates to:
  /// **'Año {year}'**
  String optionValueRifYear(String year);

  /// Form notice: the server could not be reached.
  ///
  /// In es, this message translates to:
  /// **'No hay conexión. Revisa tu internet e inténtalo de nuevo.'**
  String get onboardingErrorUnreachable;

  /// Form notice: 429.
  ///
  /// In es, this message translates to:
  /// **'Hubo demasiados cambios seguidos. Espera unos minutos e inténtalo de nuevo.'**
  String get onboardingErrorRateLimited;

  /// Form notice: 422 taxpayer_limit_reached.
  ///
  /// In es, this message translates to:
  /// **'Ya tienes 10 contribuyentes, el máximo por cuenta. Elimina uno que ya no uses para registrar otro.'**
  String get onboardingErrorLimit;

  /// Form notice: any other failure.
  ///
  /// In es, this message translates to:
  /// **'Algo salió mal. Inténtalo de nuevo.'**
  String get onboardingErrorUnexpected;

  /// Read-only user (402): title.
  ///
  /// In es, this message translates to:
  /// **'Tu cuenta es de solo consulta'**
  String get onboardingPaywallTitle;

  /// Read-only user (402): explanation.
  ///
  /// In es, this message translates to:
  /// **'Tu periodo de prueba terminó y no hay un plan activo. Puedes consultar, archivar y eliminar tu información, pero registrar un contribuyente requiere un plan.'**
  String get onboardingPaywallBody;

  /// Onboarding opened at the cap of taxpayers (OD-2): title.
  ///
  /// In es, this message translates to:
  /// **'Llegaste al máximo de contribuyentes'**
  String get onboardingLimitTitle;

  /// Onboarding opened at the cap of taxpayers (OD-2): why and how to free a slot.
  ///
  /// In es, this message translates to:
  /// **'Ya tienes {cap} contribuyentes, el máximo por cuenta; los archivados también cuentan. Para registrar otro, elimina uno que ya no uses.'**
  String onboardingLimitBody(int cap);

  /// Joins the last item of a list: '606, 612 y 621'.
  ///
  /// In es, this message translates to:
  /// **'{head} y {last}'**
  String listAnd(String head, String last);

  /// Read-only user (402): leaves onboarding.
  ///
  /// In es, this message translates to:
  /// **'Volver al inicio'**
  String get onboardingPaywallAction;

  /// Title after the taxpayer was created with warnings.
  ///
  /// In es, this message translates to:
  /// **'Revisa antes de continuar'**
  String get onboardingReviewTitle;

  /// Confirmation after create; rfc is the masked RFC.
  ///
  /// In es, this message translates to:
  /// **'Registramos a «{name}», RFC {rfc}.'**
  String onboardingReviewCreated(String name, String rfc);

  /// Non-blocking warning: regime_combination_needs_cpa_review naming the pair (each «code · name», joined with listAnd); the server reason follows.
  ///
  /// In es, this message translates to:
  /// **'Registraste juntos {regimes}. {reason}'**
  String onboardingCpaWarningBody(String regimes, String reason);

  /// Non-blocking warning: regime_combination_needs_cpa_review; the server reason follows.
  ///
  /// In es, this message translates to:
  /// **'Combinación pendiente de revisión por un contador'**
  String get onboardingCpaWarningTitle;

  /// Optional step 5 title (module_tax_engine on).
  ///
  /// In es, this message translates to:
  /// **'Datos para el cálculo'**
  String get onboardingInputsTitle;

  /// Optional step 5 explanation.
  ///
  /// In es, this message translates to:
  /// **'Opcional. Con estos datos de {year} el estimado es más preciso; puedes agregarlos después.'**
  String onboardingInputsIntro(String year);

  /// Skips the optional tax inputs of one regime.
  ///
  /// In es, this message translates to:
  /// **'Omitir'**
  String get onboardingInputsSkip;

  /// Screen-reader label of the top-bar active-taxpayer switcher.
  ///
  /// In es, this message translates to:
  /// **'Contribuyente activo: {name}. Cambiar'**
  String switcherLabel(String name);

  /// Title of the taxpayer switcher sheet.
  ///
  /// In es, this message translates to:
  /// **'Contribuyentes'**
  String get switcherTitle;

  /// Switch in the taxpayer switcher that lists archived taxpayers too.
  ///
  /// In es, this message translates to:
  /// **'Mostrar archivados'**
  String get switcherShowArchived;

  /// Badge on an archived taxpayer in the switcher.
  ///
  /// In es, this message translates to:
  /// **'Archivado'**
  String get taxpayerArchivedBadge;

  /// Screen-reader reading of a masked RFC (the middle is hidden).
  ///
  /// In es, this message translates to:
  /// **'RFC {start}, oculto, {end}'**
  String taxpayerRfcSemantics(String start, String end);

  /// Tooltip of the settings button on a switcher row.
  ///
  /// In es, this message translates to:
  /// **'Ajustes de {name}'**
  String taxpayerSettingsTooltip(String name);

  /// App bar title of the taxpayer settings screen.
  ///
  /// In es, this message translates to:
  /// **'Ajustes del contribuyente'**
  String get settingsTitle;

  /// Taxpayer list loading.
  ///
  /// In es, this message translates to:
  /// **'Cargando contribuyentes…'**
  String get settingsLoading;

  /// Taxpayer list failed to load.
  ///
  /// In es, this message translates to:
  /// **'No se pudieron cargar tus contribuyentes.'**
  String get settingsLoadFailed;

  /// Settings of a taxpayer that is gone.
  ///
  /// In es, this message translates to:
  /// **'Este contribuyente ya no existe.'**
  String get settingsNotFound;

  /// Settings section: name and RFC.
  ///
  /// In es, this message translates to:
  /// **'Datos'**
  String get settingsDataTitle;

  /// Opens the inline name editor.
  ///
  /// In es, this message translates to:
  /// **'Editar nombre'**
  String get settingsEditName;

  /// Opens the inline RFC correction.
  ///
  /// In es, this message translates to:
  /// **'Corregir RFC'**
  String get settingsCorrectRfc;

  /// Helper text of the RFC correction field.
  ///
  /// In es, this message translates to:
  /// **'Escribe el RFC completo, como en la constancia. Nunca lo mostramos completo.'**
  String get settingsRfcHelp;

  /// Confirmation after a rename.
  ///
  /// In es, this message translates to:
  /// **'Guardamos el nombre.'**
  String get settingsNameSaved;

  /// Confirmation after an RFC correction with no warning.
  ///
  /// In es, this message translates to:
  /// **'Corregimos el RFC.'**
  String get settingsRfcCorrected;

  /// A write was refused because the taxpayer is archived.
  ///
  /// In es, this message translates to:
  /// **'Este contribuyente está archivado. Desarchívalo para hacer cambios.'**
  String get settingsArchivedWrite;

  /// Settings section: the regimes.
  ///
  /// In es, this message translates to:
  /// **'Regímenes'**
  String get settingsRegimesTitle;

  /// Period of an active regime.
  ///
  /// In es, this message translates to:
  /// **'Desde {start}'**
  String settingsRegimeSince(String start);

  /// Period of an ended regime.
  ///
  /// In es, this message translates to:
  /// **'De {start} a {end}'**
  String settingsRegimeRange(String start, String end);

  /// Badge of an ended regime.
  ///
  /// In es, this message translates to:
  /// **'Terminado'**
  String get settingsRegimeEnded;

  /// Opens the add-regime flow; also its primary.
  ///
  /// In es, this message translates to:
  /// **'Agregar régimen'**
  String get settingsAddRegime;

  /// Opens a regime's options.
  ///
  /// In es, this message translates to:
  /// **'Editar opciones'**
  String get settingsEditOptions;

  /// Opens the inline end-regime panel.
  ///
  /// In es, this message translates to:
  /// **'Terminar régimen'**
  String get settingsEndRegime;

  /// Month picker of the end-regime panel.
  ///
  /// In es, this message translates to:
  /// **'Último mes en este régimen'**
  String get settingsEndRegimeLabel;

  /// Primary of the end-regime panel.
  ///
  /// In es, this message translates to:
  /// **'Terminar'**
  String get settingsEndRegimeConfirm;

  /// 409 regime_has_records: asks to confirm.
  ///
  /// In es, this message translates to:
  /// **'Este régimen tiene ingresos o gastos registrados. Se conservan y siguen contando hasta el mes que elijas.'**
  String get settingsEndRegimeHasRecords;

  /// Confirms ending a regime that has records.
  ///
  /// In es, this message translates to:
  /// **'Terminar y conservar registros'**
  String get settingsEndRegimeKeep;

  /// 422 records_after_end_period.
  ///
  /// In es, this message translates to:
  /// **'Hay registros después de ese mes. Elige el mes del último registro o uno posterior.'**
  String get settingsEndRegimeRecordsAfter;

  /// Why the last active regime cannot be ended.
  ///
  /// In es, this message translates to:
  /// **'Es su único régimen activo. Para dejar de usarlo, archiva al contribuyente.'**
  String get settingsLastRegime;

  /// Confirmation after ending a regime.
  ///
  /// In es, this message translates to:
  /// **'Terminamos el régimen {code}.'**
  String settingsRegimeEndedDone(String code);

  /// Settings section: archive and delete.
  ///
  /// In es, this message translates to:
  /// **'Archivar o eliminar'**
  String get settingsManageTitle;

  /// Archives the taxpayer.
  ///
  /// In es, this message translates to:
  /// **'Archivar'**
  String get settingsArchive;

  /// What archiving does.
  ///
  /// In es, this message translates to:
  /// **'Lo ocultas de tu lista. Conservas su información y puedes desarchivarlo.'**
  String get settingsArchiveHelp;

  /// Read-only user (OQ-8): confirmation before an archive they cannot undo; title names the taxpayer.
  ///
  /// In es, this message translates to:
  /// **'¿Archivar a «{name}»?'**
  String settingsArchiveOneWayTitle(String name);

  /// Read-only user (OQ-8): why the archive is one-way.
  ///
  /// In es, this message translates to:
  /// **'Tu cuenta es de solo consulta. Podrás seguir consultándolo, pero no podrás desarchivarlo hasta tener un plan.'**
  String get settingsArchiveOneWayBody;

  /// Read-only user (OQ-8): confirms the one-way archive.
  ///
  /// In es, this message translates to:
  /// **'Archivar de todos modos'**
  String get settingsArchiveOneWayConfirm;

  /// Undo toast after archiving.
  ///
  /// In es, this message translates to:
  /// **'Archivaste a «{name}».'**
  String settingsArchived(String name);

  /// Undo action of the archive toast.
  ///
  /// In es, this message translates to:
  /// **'Deshacer'**
  String get settingsUndo;

  /// Unarchives the taxpayer.
  ///
  /// In es, this message translates to:
  /// **'Desarchivar'**
  String get settingsUnarchive;

  /// Confirmation after unarchiving.
  ///
  /// In es, this message translates to:
  /// **'«{name}» volvió a tu lista.'**
  String settingsUnarchived(String name);

  /// Banner on an archived taxpayer's settings.
  ///
  /// In es, this message translates to:
  /// **'Este contribuyente está archivado. Puedes consultarlo; para hacer cambios, desarchívalo.'**
  String get settingsArchivedBanner;

  /// Opens the delete confirmation.
  ///
  /// In es, this message translates to:
  /// **'Eliminar'**
  String get settingsDelete;

  /// What deleting does.
  ///
  /// In es, this message translates to:
  /// **'Borra al contribuyente y todos sus registros. No se puede deshacer.'**
  String get settingsDeleteHelp;

  /// Title of the delete confirmation dialog.
  ///
  /// In es, this message translates to:
  /// **'¿Eliminar a «{name}»?'**
  String settingsDeleteTitle(String name);

  /// Body of the delete confirmation dialog.
  ///
  /// In es, this message translates to:
  /// **'Se borrarán sus ingresos y gastos, sus datos para el cálculo y su historial. No se puede deshacer. Si solo quieres ocultarlo, archívalo.'**
  String get settingsDeleteBody;

  /// Destructive primary of the delete dialog.
  ///
  /// In es, this message translates to:
  /// **'Eliminar definitivamente'**
  String get settingsDeleteConfirm;

  /// Confirmation after deleting.
  ///
  /// In es, this message translates to:
  /// **'Eliminaste a «{name}».'**
  String settingsDeleted(String name);

  /// Paywall message on a refused settings write (402).
  ///
  /// In es, this message translates to:
  /// **'Tu periodo de prueba terminó y no hay un plan activo. Puedes consultar, archivar y eliminar tu información, pero hacer cambios requiere un plan.'**
  String get paywallWriteBody;

  /// Closes the paywall message.
  ///
  /// In es, this message translates to:
  /// **'Entendido'**
  String get paywallClose;

  /// Intro of the add-regime step.
  ///
  /// In es, this message translates to:
  /// **'Elige el régimen que vas a agregar. Cuenta a partir de este mes.'**
  String get addRegimeIntro;

  /// Nothing left to add.
  ///
  /// In es, this message translates to:
  /// **'Ya tiene todos los regímenes disponibles para su tipo de persona.'**
  String get addRegimeNoneLeft;

  /// Confirmation after adding a regime.
  ///
  /// In es, this message translates to:
  /// **'Agregamos el régimen {code}.'**
  String addRegimeAdded(String code);

  /// Regime added with a CPA-review warning.
  ///
  /// In es, this message translates to:
  /// **'Agregamos el régimen {code}. Un contador debe confirmar la combinación: {reason}'**
  String addRegimeCpaReview(String code, String reason);

  /// Regime added with a CPA-review warning, naming the held regimes it pairs with (e.g. '606 y 612').
  ///
  /// In es, this message translates to:
  /// **'Agregamos el régimen {code}. Un contador debe confirmar su combinación con {partners}: {reason}'**
  String addRegimeCpaReviewWith(String code, String partners, String reason);

  /// OQ-7 warning on the options page.
  ///
  /// In es, this message translates to:
  /// **'El cambio aplica a todos los meses de este régimen, también a los anteriores.'**
  String get optionsEditWarning;

  /// Confirmation after saving options.
  ///
  /// In es, this message translates to:
  /// **'Guardamos las opciones.'**
  String get optionsSaved;

  /// App bar title of the add-regime screen.
  ///
  /// In es, this message translates to:
  /// **'Nuevo régimen'**
  String get addRegimeTitle;

  /// Tooltip of the month switcher's back button (incomes, expenses).
  ///
  /// In es, this message translates to:
  /// **'Mes anterior'**
  String get recordsPreviousMonth;

  /// Tooltip of the month switcher's forward button.
  ///
  /// In es, this message translates to:
  /// **'Mes siguiente'**
  String get recordsNextMonth;

  /// Filter chip that shows every regime.
  ///
  /// In es, this message translates to:
  /// **'Todos'**
  String get recordsFilterAll;

  /// Label of the regime filter chips.
  ///
  /// In es, this message translates to:
  /// **'Régimen'**
  String get recordsFilterLabel;

  /// Action of the persistent toast after deleting records.
  ///
  /// In es, this message translates to:
  /// **'Deshacer'**
  String get recordsUndo;

  /// Confirm button of a one-way delete (read-only account).
  ///
  /// In es, this message translates to:
  /// **'Eliminar'**
  String get recordsDelete;

  /// Income list when the user has no active taxpayer.
  ///
  /// In es, this message translates to:
  /// **'Registra un contribuyente para llevar sus ingresos.'**
  String get incomesNoTaxpayer;

  /// Empty income list for a month.
  ///
  /// In es, this message translates to:
  /// **'Sin ingresos en {month}'**
  String incomesEmptyTitle(String month);

  /// Warm copy under the empty income list.
  ///
  /// In es, this message translates to:
  /// **'Registra tu primer ingreso y te mostramos cuánto podrías pagar.'**
  String get incomesEmptyBody;

  /// Opens the new-income form (list action and empty state).
  ///
  /// In es, this message translates to:
  /// **'Registrar ingreso'**
  String get incomesRecord;

  /// Income list failed to load; a retry button follows.
  ///
  /// In es, this message translates to:
  /// **'No se pudieron cargar los ingresos.'**
  String get incomesLoadFailed;

  /// Income list is loading.
  ///
  /// In es, this message translates to:
  /// **'Cargando ingresos…'**
  String get incomesLoading;

  /// How many incomes the month (filtered) holds.
  ///
  /// In es, this message translates to:
  /// **'{count, plural, =1{1 ingreso} other{{count} ingresos}}'**
  String incomesCount(int count);

  /// Heading of the month totals card.
  ///
  /// In es, this message translates to:
  /// **'Total del mes con IVA'**
  String get incomesTotalTitle;

  /// Totals card line: subtotal before IVA.
  ///
  /// In es, this message translates to:
  /// **'Sin IVA'**
  String get incomesTotalSubtotal;

  /// Totals card line: IVA charged.
  ///
  /// In es, this message translates to:
  /// **'IVA'**
  String get incomesTotalIva;

  /// Totals card line: ISR and IVA withheld by payers.
  ///
  /// In es, this message translates to:
  /// **'Retenciones'**
  String get incomesTotalWithheld;

  /// Badge of an accrual income not yet collected.
  ///
  /// In es, this message translates to:
  /// **'Por cobrar'**
  String get incomesPendingCollection;

  /// Tooltip and label of a row's delete button.
  ///
  /// In es, this message translates to:
  /// **'Eliminar ingreso: {title}'**
  String incomesDeleteTooltip(String title);

  /// Screen-reader label of a row: tapping it edits the income.
  ///
  /// In es, this message translates to:
  /// **'Editar ingreso: {title}'**
  String incomesEditLabel(String title);

  /// Persistent toast after deleting incomes; deletes stack.
  ///
  /// In es, this message translates to:
  /// **'{count, plural, =1{Ingreso eliminado} other{{count} ingresos eliminados}}'**
  String incomesDeleted(int count);

  /// Toast after Undo restored the deleted incomes.
  ///
  /// In es, this message translates to:
  /// **'{count, plural, =1{Ingreso restaurado} other{{count} ingresos restaurados}}'**
  String incomesRestored(int count);

  /// Toast when a delete failed; the row is back.
  ///
  /// In es, this message translates to:
  /// **'No se pudo eliminar el ingreso.'**
  String get incomesDeleteFailed;

  /// Toast when Undo failed.
  ///
  /// In es, this message translates to:
  /// **'No se pudo restaurar el ingreso.'**
  String get incomesRestoreFailed;

  /// Toast when Undo failed with 422 regime_not_active.
  ///
  /// In es, this message translates to:
  /// **'No se pudo restaurar: el régimen ya no cubre la fecha del ingreso.'**
  String get incomesRestoreRegimeEnded;

  /// Who paid: an individual.
  ///
  /// In es, this message translates to:
  /// **'Persona física'**
  String get counterpartyPersonaFisica;

  /// Who paid: a company.
  ///
  /// In es, this message translates to:
  /// **'Persona moral'**
  String get counterpartyPersonaMoral;

  /// Who paid: the general public (no RFC).
  ///
  /// In es, this message translates to:
  /// **'Público en general'**
  String get counterpartyPublicoGeneral;

  /// Who paid: a digital platform (Uber, Airbnb, Mercado Libre).
  ///
  /// In es, this message translates to:
  /// **'Plataforma digital'**
  String get counterpartyPlataforma;

  /// IVA rate option.
  ///
  /// In es, this message translates to:
  /// **'16 %'**
  String get ivaRate16;

  /// IVA rate option: border region decree.
  ///
  /// In es, this message translates to:
  /// **'8 % (región fronteriza)'**
  String get ivaRate8;

  /// IVA rate option.
  ///
  /// In es, this message translates to:
  /// **'0 %'**
  String get ivaRate0;

  /// IVA rate option: exempt.
  ///
  /// In es, this message translates to:
  /// **'Exento'**
  String get ivaRateExempt;

  /// 625 platform activity.
  ///
  /// In es, this message translates to:
  /// **'Transporte o entrega'**
  String get activityTransportDelivery;

  /// 625 platform activity.
  ///
  /// In es, this message translates to:
  /// **'Hospedaje'**
  String get activityLodging;

  /// 625 platform activity.
  ///
  /// In es, this message translates to:
  /// **'Venta de bienes o servicios'**
  String get activityGoodsServices;

  /// App bar title of the new-income page.
  ///
  /// In es, this message translates to:
  /// **'Nuevo ingreso'**
  String get incomeFormTitle;

  /// App bar title of the edit-income page (same form, prefilled).
  ///
  /// In es, this message translates to:
  /// **'Editar ingreso'**
  String get incomeEditTitle;

  /// Toast after an edit was saved.
  ///
  /// In es, this message translates to:
  /// **'Ingreso actualizado'**
  String get incomeUpdated;

  /// Start of the toast after an edit saved with non-blocking warnings.
  ///
  /// In es, this message translates to:
  /// **'Ingreso actualizado, con un aviso:'**
  String get incomeUpdatedWithWarnings;

  /// Date the income was collected.
  ///
  /// In es, this message translates to:
  /// **'Fecha de cobro'**
  String get incomeFieldCollectedOn;

  /// Date the CFDI was issued.
  ///
  /// In es, this message translates to:
  /// **'Fecha de factura'**
  String get incomeFieldInvoicedOn;

  /// Free text describing the income.
  ///
  /// In es, this message translates to:
  /// **'Descripción'**
  String get incomeFieldDescription;

  /// Subtotal before IVA.
  ///
  /// In es, this message translates to:
  /// **'Monto sin IVA'**
  String get incomeFieldAmount;

  /// IVA rate choice.
  ///
  /// In es, this message translates to:
  /// **'Tasa de IVA'**
  String get incomeFieldIvaRate;

  /// Counterparty type choice.
  ///
  /// In es, this message translates to:
  /// **'¿Quién pagó?'**
  String get incomeFieldCounterparty;

  /// Regime the income belongs to (only with more than one).
  ///
  /// In es, this message translates to:
  /// **'Régimen'**
  String get incomeFieldRegime;

  /// 625 activity (only with more than one).
  ///
  /// In es, this message translates to:
  /// **'Actividad en la plataforma'**
  String get incomeFieldActivity;

  /// IVA amount, prefilled from the rate.
  ///
  /// In es, this message translates to:
  /// **'IVA cobrado'**
  String get incomeFieldIvaAmount;

  /// ISR the payer withheld.
  ///
  /// In es, this message translates to:
  /// **'ISR retenido'**
  String get incomeFieldIsrWithheld;

  /// IVA the payer withheld.
  ///
  /// In es, this message translates to:
  /// **'IVA retenido'**
  String get incomeFieldIvaWithheld;

  /// The CFDI's folio fiscal.
  ///
  /// In es, this message translates to:
  /// **'Folio fiscal (UUID)'**
  String get incomeFieldCfdiUuid;

  /// Collapsed section with the withholdings.
  ///
  /// In es, this message translates to:
  /// **'Retenciones'**
  String get incomeSectionWithholdings;

  /// Collapsed section: IVA amount, the other date, folio fiscal.
  ///
  /// In es, this message translates to:
  /// **'Más detalles'**
  String get incomeSectionMore;

  /// Help under the withholdings section.
  ///
  /// In es, this message translates to:
  /// **'Prellenadas con la tasa sugerida. Corrígelas con lo que diga el CFDI.'**
  String get incomeWithholdingsHelp;

  /// Live total under the amount.
  ///
  /// In es, this message translates to:
  /// **'Total con IVA {total}'**
  String incomeSummaryTotal(String total);

  /// Live breakdown under the amount.
  ///
  /// In es, this message translates to:
  /// **'IVA {iva} · Retenciones {withheld} · Neto {net}'**
  String incomeSummaryNet(String iva, String withheld, String net);

  /// Notice for an income in an informational regime (OQ-4).
  ///
  /// In es, this message translates to:
  /// **'Este régimen es informativo: el ingreso no cambia tu estimación.'**
  String get incomeInformationalRegime;

  /// Format error of the folio fiscal.
  ///
  /// In es, this message translates to:
  /// **'Escribe el folio como aparece en el CFDI: 32 letras y números en grupos de 8-4-4-4-12.'**
  String get incomeCfdiInvalid;

  /// Defaults endpoint failed; the form still works.
  ///
  /// In es, this message translates to:
  /// **'No se pudieron cargar las tasas sugeridas. Puedes capturar los montos directamente.'**
  String get incomeDefaultsFailed;

  /// Server detail out_of_range on a date.
  ///
  /// In es, this message translates to:
  /// **'La fecha debe estar entre el 1 de enero de 2022 y hoy.'**
  String get incomeErrorOutOfRange;

  /// Server detail iva_amount_mismatch.
  ///
  /// In es, this message translates to:
  /// **'El IVA no corresponde a la tasa elegida. Revísalo en el CFDI.'**
  String get incomeErrorIvaMismatch;

  /// Server detail rate_not_valid_on_date.
  ///
  /// In es, this message translates to:
  /// **'Esta tasa no está vigente en la fecha del ingreso.'**
  String get incomeErrorRateNotValid;

  /// Server detail withholding_not_applicable.
  ///
  /// In es, this message translates to:
  /// **'Con este tipo de cliente no hay esta retención.'**
  String get incomeErrorWithholdingNotApplicable;

  /// Server detail withholding_exceeds_base.
  ///
  /// In es, this message translates to:
  /// **'La retención no puede ser mayor que su base.'**
  String get incomeErrorWithholdingExceedsBase;

  /// Server detail regime_not_active.
  ///
  /// In es, this message translates to:
  /// **'El contribuyente no tenía este régimen en esa fecha.'**
  String get incomeErrorRegimeNotActive;

  /// Server detail activity_not_held.
  ///
  /// In es, this message translates to:
  /// **'Esta actividad no está registrada en el régimen.'**
  String get incomeErrorActivityNotHeld;

  /// Any other server verdict on a field.
  ///
  /// In es, this message translates to:
  /// **'Revisa este dato.'**
  String get incomeErrorInvalid;

  /// Save failed offline; the retry is idempotent.
  ///
  /// In es, this message translates to:
  /// **'No se pudo guardar por falta de conexión. Vuelve a tocar Guardar: no se duplicará.'**
  String get incomeSaveUnreachable;

  /// 409 id_reused on create.
  ///
  /// In es, this message translates to:
  /// **'Este ingreso ya se había guardado con otros datos. Revisa la lista de ingresos.'**
  String get incomeSaveIdReused;

  /// 404 on an edit: the income was deleted elsewhere.
  ///
  /// In es, this message translates to:
  /// **'Este ingreso ya no existe: se eliminó en otro lugar.'**
  String get incomeSaveGone;

  /// Any other save failure.
  ///
  /// In es, this message translates to:
  /// **'No se pudo guardar el ingreso. Intenta de nuevo.'**
  String get incomeSaveFailed;

  /// Toast after saving an income.
  ///
  /// In es, this message translates to:
  /// **'Ingreso guardado'**
  String get incomeSaved;

  /// Heading before the non-blocking warnings of a saved income.
  ///
  /// In es, this message translates to:
  /// **'Ingreso guardado, con un aviso:'**
  String get incomeSavedWithWarnings;

  /// Warning duplicate_cfdi_uuid (non-blocking).
  ///
  /// In es, this message translates to:
  /// **'Ya hay otro ingreso con este folio fiscal. Si es un pago parcial del mismo CFDI, está bien.'**
  String get incomeWarningDuplicateCfdi;

  /// Warning withholding_differs_from_rate (non-blocking).
  ///
  /// In es, this message translates to:
  /// **'La retención es distinta de la tasa sugerida. Se guardó lo que escribiste: el CFDI manda.'**
  String get incomeWarningWithholdingDiffers;

  /// Paywall title on 402 trial_cap_reached for incomes.
  ///
  /// In es, this message translates to:
  /// **'Llegaste al límite de ingresos de prueba'**
  String get incomeTrialCapTitle;

  /// Paywall body on 402 trial_cap_reached for incomes.
  ///
  /// In es, this message translates to:
  /// **'La prueba incluye 5 ingresos y ya los registraste. Para registrar más ingresos se necesita un plan. Lo que capturaste sigue en el formulario.'**
  String get incomeTrialCapBody;

  /// Leaving the form with unsaved changes.
  ///
  /// In es, this message translates to:
  /// **'¿Descartar este ingreso?'**
  String get incomeDiscardTitle;

  /// Body of the discard confirmation.
  ///
  /// In es, this message translates to:
  /// **'Lo que escribiste se perderá.'**
  String get incomeDiscardBody;

  /// Stays on the form.
  ///
  /// In es, this message translates to:
  /// **'Seguir editando'**
  String get incomeDiscardKeep;

  /// Leaves the form, losing the changes.
  ///
  /// In es, this message translates to:
  /// **'Descartar'**
  String get incomeDiscardConfirm;
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
