import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/operation_kind.dart';
import 'package:control_contable/api/generated/models/supplier_kind.dart';
import 'package:control_contable/api/generated/models/supplier_regime.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/form_field_model.dart';
import 'package:control_contable/core/forms/yes_no.dart';
import 'package:control_contable/features/expenses/domain/expense_rules.dart';

/// The labels of the four supplier questions.
typedef SupplierFactLabels = ({
  String supplierKind,
  String supplierRegime,
  String operationKind,
  String residentialRent,
});

/// The supplier last used for a category in this session.
typedef SupplierMemory = ({SupplierKind kind, SupplierRegime? regime});

/// "Tu proveedor" (MODULE-withholdings §2.2, E1–E4): what a persona moral
/// must withhold from its supplier is derived from these facts by the
/// server (ADR-0008). A persona física is never asked: the questions exist
/// but nothing of them is sent ([isMoral]).
class ExpenseSupplierFacts {
  /// The questions added to [form], starting at [stored] when editing;
  /// `lastUsedFor` remembers the supplier of each category.
  new({
    required AppFormController form,
    required SupplierFactLabels labels,
    required this.isMoral,
    required this._lastUsedFor,
    Expense? stored,
  }) : _opensItself = stored != null && stored.supplierKind != SupplierKind.pm {
    supplierKind = form.add(
      ChoiceFieldController(
        label: labels.supplierKind,
        isRequired: isMoral,
        initialValue: stored?.supplierKind?.json,
      ),
    );
    supplierRegime = form.add(
      ChoiceFieldController(
        label: labels.supplierRegime,
        isRequired: false,
        initialValue: (stored?.supplierRegime ?? SupplierRegime.unknown).json,
      ),
    );
    operationKind = form.add(
      ChoiceFieldController(
        label: labels.operationKind,
        isRequired: false,
        initialValue: stored?.operationKind?.json,
      ),
    );
    residentialRent = form.add(
      ChoiceFieldController(
        label: labels.residentialRent,
        isRequired: false,
        initialValue: (stored?.residentialRent ?? false).choice,
      ),
    );
    // A new expense starts at suggestions; a stored one at the user's answers.
    if (stored == null) {
      for (final field in _all) {
        _suggested[field] = field.value;
      }
    }
  }

  /// Whether the taxpayer is a persona moral (only they withhold).
  final bool isMoral;

  final SupplierMemory? Function(String category) _lastUsedFor;
  bool _opensItself;

  /// What each question was last moved to by a suggestion: an answer that
  /// differs from it (and is not empty) is the user's, and a suggestion
  /// never replaces it. A stored expense's answers are the user's.
  final Map<ChoiceFieldController, String?> _suggested = {};

  /// E1 "¿Quién te vendió o te dio el servicio?" (`SupplierKind.json`).
  late final ChoiceFieldController supplierKind;

  /// E2 "¿En qué régimen está tu proveedor?" (`SupplierRegime.json`).
  late final ChoiceFieldController supplierRegime;

  /// E3 "¿Qué te vendió…?" (`OperationKind.json`).
  late final ChoiceFieldController operationKind;

  /// E4 "¿Es una casa o departamento para vivir?" ([yes] / [no]).
  late final ChoiceFieldController residentialRent;

  List<ChoiceFieldController> get _all => [
    supplierKind,
    supplierRegime,
    operationKind,
    residentialRent,
  ];

  /// Whether the section opens by itself: a supplier the user must choose,
  /// or one that is not a company.
  bool get opensItself => isMoral && _opensItself;

  /// The questions, for the server verdicts by path.
  Map<String, AppFormFieldModel> get fieldsByPath => {
    if (isMoral) ...{
      'supplier_kind': supplierKind,
      'supplier_regime': supplierRegime,
      'operation_kind': operationKind,
      'residential_rent': residentialRent,
    },
  };

  /// Whether any answer differs from what the form started with.
  bool get isDirty => _all.any((field) => field.isDirty);

  /// Who sold it, once answered (and only for a persona moral).
  SupplierKind? get kind {
    final value = supplierKind.value;
    return isMoral && value != null ? SupplierKind.fromJson(value) : null;
  }

  /// Whether E2 is asked: a persona física supplier.
  bool get asksRegime => kind == SupplierKind.pfWithActivity;

  /// Whether E3 is asked: a domestic supplier.
  bool get asksOperation =>
      kind == SupplierKind.pm || kind == SupplierKind.pfWithActivity;

  /// `supplier_regime` as sent: only where it was asked.
  SupplierRegime? get regime {
    final value = supplierRegime.value;
    return asksRegime && value != null ? SupplierRegime.fromJson(value) : null;
  }

  /// `operation_kind` as sent: only where it was asked.
  OperationKind? get operation {
    final value = operationKind.value;
    return asksOperation && value != null
        ? OperationKind.fromJson(value)
        : null;
  }

  /// Whether E4 is asked: the rent of a property.
  bool get asksResidentialRent => operation == OperationKind.rentRealEstate;

  /// Whether it is housing rent (exempt, LIVA 20 fr. II).
  bool get isResidentialRent =>
      asksResidentialRent && residentialRent.value == yes;

  /// `residential_rent` as sent: only where it was asked.
  bool? get residentialRentSent =>
      asksResidentialRent ? isResidentialRent : null;

  /// The category changed to [category]: each answer the user has not
  /// given moves to its default, the supplier last used for the category,
  /// else a company (except for fees, rent and freight, where the user must
  /// choose), and the operation the server gives the category
  /// (`default_operation_kind`, never a copy of the seed here; absent, it
  /// stays unanswered and is omitted, so the server applies the same one).
  void followCategory(ExpenseCategory? category) {
    if (!isMoral) return;
    final code = category?.code;
    final remembered = code == null ? null : _lastUsedFor(code);
    final kind =
        remembered?.kind ??
        (code == null || withholdingCategories.contains(code)
            ? null
            : SupplierKind.pm);
    _suggest(supplierKind, kind?.json);
    _opensItself = this.kind != SupplierKind.pm;
    _suggest(
      supplierRegime,
      (remembered?.regime ?? SupplierRegime.unknown).json,
    );
    _suggest(operationKind, category?.defaultOperationKind?.json);
    _suggest(residentialRent, no);
  }

  /// Moves [field] to [value] unless the user answered it.
  void _suggest(ChoiceFieldController field, String? value) {
    final current = field.value;
    if (current != null && current != _suggested[field]) return;
    _suggested[field] = value;
    field.value = value;
  }
}
