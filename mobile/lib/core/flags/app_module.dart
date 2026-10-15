/// The product modules a flag can enable (`docs/project-specs/MODULES.md`).
///
/// The app asks for modules only through this enum, so a flag key is never a
/// loose string. Keys the server sends that are not listed here are ignored.
enum AppModule {
  /// M1 · Auth & Account.
  auth('module_auth'),

  /// M2 · Taxpayers & Onboarding.
  taxpayers('module_taxpayers'),

  /// M3 · Incomes.
  incomes('module_incomes'),

  /// M4 · Expenses.
  expenses('module_expenses'),

  /// M5 · Tax Engine (estimate + breakdown).
  taxEngine('module_tax_engine'),

  /// M6 · Dashboard.
  dashboard('module_dashboard'),

  /// M7 · Subscriptions (RevenueCat).
  subscriptions('module_subscriptions'),

  /// M8 · History.
  history('module_history'),

  /// M9 · Reminders (push).
  reminders('module_reminders'),

  /// Annual estimate (declaración anual, MODULE-annual).
  annual('module_annual');

  new(this.key);

  /// The flag key on the wire (`ModuleKey` in the contract).
  final String key;

  static final Map<String, AppModule> _byKey = {
    for (final module in values) module.key: module,
  };

  /// The module for [key], or `null` when this build does not know it.
  static AppModule? fromKey(String key) => _byKey[key];
}
