import 'package:control_contable/core/flags/app_module.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('every module key follows the contract ModuleKey pattern', () {
    final pattern = RegExp(r'^module_[a-z_]+$');

    for (final module in AppModule.values) {
      expect(module.key, matches(pattern));
    }
  });

  test('the modules are the nine product modules of MODULES.md', () {
    expect(AppModule.values.map((m) => m.key), [
      'module_auth',
      'module_taxpayers',
      'module_incomes',
      'module_expenses',
      'module_tax_engine',
      'module_dashboard',
      'module_subscriptions',
      'module_history',
      'module_reminders',
    ]);
  });

  test('a key maps back to its module; an unknown key maps to nothing', () {
    expect(AppModule.fromKey('module_tax_engine'), AppModule.taxEngine);
    expect(AppModule.fromKey('module_from_the_future'), isNull);
  });
}
