import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/income.dart';
import 'package:control_contable/api/generated/models/warning.dart';
import 'package:control_contable/api/generated/models/warning_code.dart';
import 'package:control_contable/app/app.dart';
import 'package:control_contable/core/records/stacked_undo.dart';
import 'package:control_contable/features/expenses/presentation/expense_save_feedback.dart';
import 'package:control_contable/features/incomes/presentation/income_save_feedback.dart';
import 'package:control_contable/l10n/app_localizations.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'expenses/expenses_fakes.dart';
import 'incomes/incomes_fakes.dart';

const _duplicate = Warning(
  code: WarningCode.duplicateCfdiUuid,
  field: 'cfdi_uuid',
  message: 'Folio repetido.',
);
const _differs = Warning(
  code: WarningCode.withholdingDiffersFromRate,
  field: 'isr_withheld_cents',
  message: 'Retención distinta.',
);
const _cash = Warning(
  code: WarningCode.cashOverThreshold,
  field: 'payment_method',
  message: 'Efectivo.',
);
const _noCfdi = Warning(
  code: WarningCode.noCfdi,
  field: 'has_cfdi',
  message: 'Sin factura.',
);

/// Tech-lead gate M3-12/M4-13: the save confirmation of both record forms
/// follows one rule. A pending Undo is never closed by it (house rule: the
/// Undo toast does not auto-dismiss; incomes used to close it), and the
/// heading counts its warnings (QA M3/M4 D-3: "con un aviso" for two).
void main() {
  final l10n = lookupAppLocalizations(appLocale);
  late ScaffoldMessengerState messenger;

  Future<void> pump(WidgetTester tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Builder(
            builder: (context) {
              messenger = ScaffoldMessenger.of(context);
              return const SizedBox.shrink();
            },
          ),
        ),
      ),
    );
  }

  void pushUndo<T>(StackedUndo<T> undo, T item) => undo.push(
    item,
    messenger: messenger,
    message: (count) => '$count eliminados',
    undoLabel: 'Deshacer',
    onUndo: (_) {},
  );

  group('a pending Undo outlives the save confirmation', () {
    testWidgets('incomes', (tester) async {
      await pump(tester);
      final undo = StackedUndo<Income>();
      pushUndo(undo, income());
      await tester.pumpAndSettle();

      showIncomeSavedToast(messenger, l10n, const [], undo: undo);
      await tester.pumpAndSettle();

      expect(find.text('Deshacer'), findsOneWidget);
      expect(undo.pending, hasLength(1));
    });

    testWidgets('expenses', (tester) async {
      await pump(tester);
      final undo = StackedUndo<Expense>();
      pushUndo(undo, expense());
      await tester.pumpAndSettle();

      showExpenseSavedToast(messenger, l10n, const [], undo: undo);
      await tester.pumpAndSettle();

      expect(find.text('Deshacer'), findsOneWidget);
      expect(undo.pending, hasLength(1));
    });
  });

  group('the heading counts the warnings', () {
    testWidgets('incomes: one, then two', (tester) async {
      await pump(tester);
      final undo = StackedUndo<Income>();

      showIncomeSavedToast(messenger, l10n, const [_duplicate], undo: undo);
      await tester.pumpAndSettle();
      expect(find.textContaining('Ingreso guardado, con un aviso:'), findsOne);

      showIncomeSavedToast(
        messenger,
        l10n,
        const [_duplicate, _differs],
        undo: undo,
        updated: true,
      );
      await tester.pumpAndSettle();
      expect(
        find.textContaining('Ingreso actualizado, con 2 avisos:'),
        findsOne,
      );
    });

    testWidgets('expenses: one, then two', (tester) async {
      await pump(tester);
      final undo = StackedUndo<Expense>();

      showExpenseSavedToast(messenger, l10n, const [_noCfdi], undo: undo);
      await tester.pumpAndSettle();
      expect(find.textContaining('Gasto guardado, con un aviso:'), findsOne);

      showExpenseSavedToast(
        messenger,
        l10n,
        const [_cash, _noCfdi],
        undo: undo,
        updated: true,
      );
      await tester.pumpAndSettle();
      expect(find.textContaining('Gasto actualizado, con 2 avisos:'), findsOne);
    });
  });
}
