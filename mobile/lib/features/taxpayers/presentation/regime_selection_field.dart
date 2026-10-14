import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/api/generated/models/regime_incompatibility.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/features/taxpayers/domain/regime_compatibility.dart';
import 'package:flutter/foundation.dart';

/// The regimes chosen in step 3, in selection order, under the form
/// policy: nothing chosen is an error on submit; an incompatible pair is
/// an error too (its reason shows inline at once, under the later regime).
/// [single] makes it single-select (a persona moral holds one regime).
/// `held` are regimes the taxpayer already holds: a chosen regime that
/// conflicts with one of them is an error too.
class RegimeSelectionField extends AppFormFieldModel {
  /// The selection labeled [label], starting with [initial].
  new({
    required super.label,
    required this.single,
    required this.missingError,
    required this.conflictError,
    List<Regime> initial = const [],
    List<Regime> held = const [],
  }) : _held = List.unmodifiable(held),
       _initial = List.unmodifiable(initial),
       _selected = [...initial];

  /// Whether choosing one regime drops the others.
  final bool single;

  /// Shown when nothing is chosen.
  final FieldError missingError;

  /// Shown when an incompatible pair is chosen.
  final FieldError conflictError;

  final List<Regime> _held;
  List<Regime> _initial;
  final List<Regime> _selected;

  /// The chosen regimes, in selection order.
  List<Regime> get selected => List.unmodifiable(_selected);

  /// Incompatible pairs among the held and chosen ones, keyed by the
  /// chosen code (held regimes come first, so they are never the key).
  Map<String, RegimeIncompatibility> get conflicts =>
      regimeConflicts([..._held, ..._selected]);

  /// Whether [regime] is chosen.
  bool isChosen(Regime regime) =>
      _selected.any((chosen) => chosen.code == regime.code);

  /// Chooses [regime], or drops it when it was chosen (multi-select).
  void toggle(Regime regime) {
    final wasChosen = isChosen(regime);
    if (single) _selected.clear();
    _selected.removeWhere((chosen) => chosen.code == regime.code);
    if (single || !wasChosen) _selected.add(regime);
    notifyListeners();
    if (error != null) setError(check());
  }

  @override
  bool get isDirty => !listEquals(_selected, _initial);

  @override
  void markPristine() => _initial = List.unmodifiable(_selected);

  @override
  void reset() {
    setError(null);
    _selected.clear();
    notifyListeners();
  }

  @override
  @protected
  FieldError? check() {
    if (_selected.isEmpty) return isRequired ? missingError : null;
    return conflicts.isEmpty ? null : conflictError;
  }
}
