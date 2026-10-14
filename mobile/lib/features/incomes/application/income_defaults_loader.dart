import 'dart:async';

import 'package:control_contable/api/generated/models/counterparty_type.dart';
import 'package:control_contable/api/generated/models/income_defaults.dart';
import 'package:control_contable/api/generated/models/platform_activity.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/incomes/application/income_form_model.dart';
import 'package:flutter/foundation.dart';

/// What the defaults depend on (`getIncomeDefaults` query).
typedef IncomeDefaultsKey = ({
  String regime,
  CounterpartyType payer,
  String? on,
  PlatformActivity? activity,
});

/// Asks for the defaults (IVA rates, suggested withholdings) whenever the
/// regime, the payer, the booking date or the 625 activity of [_model]
/// changes, and hands each answer to the model. Only the latest request
/// lands. A failure leaves the form usable without suggestions ([failed]).
class IncomeDefaultsLoader {
  /// Starts loading for [_model] through [_fetch].
  new(this._model, this._fetch) {
    _model.addListener(_check);
    _check();
  }

  final IncomeFormModel _model;
  final Future<IncomeDefaults> Function(IncomeDefaultsKey key) _fetch;
  IncomeDefaultsKey? _key;
  int _generation = 0;
  bool _isDisposed = false;

  /// Whether the last request failed.
  final ValueNotifier<bool> failed = ValueNotifier(false);

  void _check() {
    final payer = _model.payer;
    if (payer == CounterpartyType.$unknown) return;
    final key = (
      regime: _model.regimeCode,
      payer: payer,
      on: _model.bookingDay,
      activity: _model.platformActivity,
    );
    if (key == _key) return;
    _key = key;
    unawaited(_load(key, ++_generation));
  }

  Future<void> _load(IncomeDefaultsKey key, int generation) async {
    IncomeDefaults? defaults;
    try {
      defaults = await _fetch(key);
    } on ApiFailure {
      defaults = null;
    }
    if (_isDisposed || generation != _generation) return;
    failed.value = defaults == null;
    _model.applyDefaults(defaults);
  }

  /// Stops listening; answers still in flight are dropped.
  void dispose() {
    _isDisposed = true;
    _model.removeListener(_check);
    failed.dispose();
  }
}
