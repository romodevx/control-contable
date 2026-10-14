import 'dart:async';

import 'package:control_contable/api/generated/models/expense_category_list.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/expenses/application/expense_form_model.dart';
import 'package:flutter/foundation.dart';

/// What the catalog depends on (`listExpenseCategories` query).
typedef ExpenseCatalogQuery = ({String regime, String? on});

/// Asks for the category catalog whenever the regime or the payment date
/// of [_model] changes, and hands each answer to the model. Only the
/// latest request lands; a failure keeps the last catalog ([failed]).
class ExpenseCatalogLoader {
  /// Starts loading for [_model] through [_fetch].
  new(this._model, this._fetch) {
    _model.addListener(_check);
    _check();
  }

  final ExpenseFormModel _model;
  final Future<ExpenseCategoryList> Function(ExpenseCatalogQuery query) _fetch;
  ExpenseCatalogQuery? _query;
  int _generation = 0;
  bool _isDisposed = false;

  /// Whether the last request failed.
  final ValueNotifier<bool> failed = ValueNotifier(false);

  void _check() {
    final day = _model.paidDay;
    // A half-typed date keeps the catalog of the last valid one.
    if (day == null && _query != null) return;
    final query = (regime: _model.regimeCode, on: day);
    if (query == _query) return;
    _query = query;
    unawaited(_load(query, ++_generation));
  }

  /// Asks again for the current query (after a failure).
  void retry() {
    final query = _query;
    if (query != null) unawaited(_load(query, ++_generation));
  }

  Future<void> _load(ExpenseCatalogQuery query, int generation) async {
    ExpenseCategoryList? catalog;
    try {
      catalog = await _fetch(query);
    } on ApiFailure {
      catalog = null;
    }
    if (_isDisposed || generation != _generation) return;
    failed.value = catalog == null;
    if (catalog != null) _model.applyCatalog(catalog);
  }

  /// Stops listening; answers still in flight are dropped.
  void dispose() {
    _isDisposed = true;
    _model.removeListener(_check);
    failed.dispose();
  }
}
