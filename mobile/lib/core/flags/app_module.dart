/// Placeholder: implemented after the failing test (TDD).
enum AppModule {
  auth,
  taxpayers,
  incomes,
  expenses,
  taxEngine,
  dashboard,
  subscriptions,
  history,
  reminders;

  String get key => throw UnimplementedError();

  static AppModule? fromKey(String key) => throw UnimplementedError();
}
