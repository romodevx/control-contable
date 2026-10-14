import 'package:control_contable/api/generated/models/expense_category_list.dart';
import 'package:control_contable/features/expenses/data/expenses_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Which catalog to read: a taxpayer's categories for one regime, on a
/// date (`YYYY-MM-DD`; today when `null`).
typedef ExpenseCatalogKey = ({String taxpayerId, String regime, String? on});

/// The category catalog of one [ExpenseCatalogKey] (names for the list
/// rows). No automatic retry: a row falls back to the code.
final FutureProviderFamily<ExpenseCategoryList, ExpenseCatalogKey>
expenseCatalogProvider = FutureProvider.autoDispose
    .family<ExpenseCategoryList, ExpenseCatalogKey>(
      (ref, key) => ref
          .read(expensesRepositoryProvider)
          .categories(
            taxpayerId: key.taxpayerId,
            regime: key.regime,
            on: key.on,
          ),
      name: 'expenseCatalogProvider',
      retry: (retryCount, error) => null,
    );
