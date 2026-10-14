import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/core/flags/app_module.dart';
import 'package:control_contable/core/router/feature_module.dart';
import 'package:control_contable/features/expenses/presentation/expense_edit_screen.dart';
import 'package:control_contable/features/expenses/presentation/expense_form_screen.dart';
import 'package:control_contable/features/expenses/presentation/expenses_screen.dart';
import 'package:go_router/go_router.dart';

/// Locations owned by `module_expenses`. They scope to the active
/// taxpayer, so no taxpayer id travels in the location.
abstract final class ExpensesRoutes {
  /// The month list (M4-09).
  static const list = '/gastos';

  /// The new-expense page (M4-10).
  static const create = '/gastos/nuevo';

  /// The edit page of expense [expenseId] (M4-10, US-E2).
  static String editOf(String expenseId) =>
      '$list/${Uri.encodeComponent(expenseId)}/editar';
}

/// The expenses slice: unreachable while `module_expenses` is off (the
/// server's `404` is the security boundary).
final FeatureModule expensesFeature = FeatureModule(
  module: AppModule.expenses,
  routes: [
    GoRoute(
      path: ExpensesRoutes.list,
      builder: (context, state) => const ExpensesScreen(),
      routes: [
        GoRoute(
          path: 'nuevo',
          builder: (context, state) => const ExpenseFormScreen(),
        ),
        GoRoute(
          path: ':expenseId/editar',
          builder: (context, state) => ExpenseEditScreen(
            expenseId: state.pathParameters['expenseId']!,
            initial: state.extra is Expense ? state.extra! as Expense : null,
          ),
        ),
      ],
    ),
  ],
);
