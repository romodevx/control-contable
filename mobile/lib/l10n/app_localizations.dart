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
