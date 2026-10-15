import 'package:control_contable/api/generated/models/expense.dart';
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

/// The categories whose default operation is not "business services",
/// mirroring the seed's `category_regime_rules` (MODULE-withholdings §2.2)
/// so E3 starts at the category's operation. The answer the user sees is
/// the one sent; the seed only decides an omitted one.
const Map<String, OperationKind> _categoryOperations = {
  'servicios_profesionales': OperationKind.professionalServices,
  'renta_local': OperationKind.rentRealEstate,
  'arrendamiento_equipo': OperationKind.rentMovable,
  'fletes': OperationKind.groundFreight,
  'comisiones': OperationKind.commission,
  'mercancia': OperationKind.saleOfGoods,
};

/// The operation a [category] stands for.
OperationKind categoryDefaultOperation(String category) =>
    _categoryOperations[category] ?? OperationKind.businessServices;

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
        initialValue:
            (stored?.operationKind ??
                    (stored == null
                        ? null
                        : categoryDefaultOperation(stored.categoryCode)))
                ?.json,
      ),
    );
    residentialRent = form.add(
      ChoiceFieldController(
        label: labels.residentialRent,
        isRequired: false,
        initialValue: (stored?.residentialRent ?? false).choice,
      ),
    );
  }

  /// Whether the taxpayer is a persona moral (only they withhold).
  final bool isMoral;

  final SupplierMemory? Function(String category) _lastUsedFor;
  bool _opensItself;

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

  /// The category changed to [category]: every answer moves to its
  /// defaults (the supplier last used for it, else a company, except for
  /// fees, rent and freight, where the user must choose).
  void followCategory(String? category) {
    if (!isMoral) return;
    final remembered = category == null ? null : _lastUsedFor(category);
    final kind =
        remembered?.kind ??
        (category == null || withholdingCategories.contains(category)
            ? null
            : SupplierKind.pm);
    _opensItself = kind != SupplierKind.pm;
    supplierKind.value = kind?.json;
    supplierRegime.value = (remembered?.regime ?? SupplierRegime.unknown).json;
    operationKind.value = category == null
        ? null
        : categoryDefaultOperation(category).json;
    residentialRent.value = no;
  }
}
