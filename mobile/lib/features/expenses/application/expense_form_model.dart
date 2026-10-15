import 'package:control_contable/api/generated/models/deductible_source.dart';
import 'package:control_contable/api/generated/models/expense.dart';
import 'package:control_contable/api/generated/models/expense_category.dart';
import 'package:control_contable/api/generated/models/expense_category_list.dart';
import 'package:control_contable/api/generated/models/iva_rate.dart';
import 'package:control_contable/api/generated/models/payment_method.dart';
import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/taxpayer.dart';
import 'package:control_contable/api/generated/models/taxpayer_regime.dart';
import 'package:control_contable/core/fiscal/housing_rent_rate.dart';
import 'package:control_contable/core/fiscal/iva_math.dart';
import 'package:control_contable/core/forms/choice_controllers.dart';
import 'package:control_contable/core/forms/field_controller.dart';
import 'package:control_contable/core/forms/field_error.dart';
import 'package:control_contable/core/forms/flexible_date.dart';
import 'package:control_contable/core/forms/form_controller.dart';
import 'package:control_contable/core/forms/money.dart';
import 'package:control_contable/core/forms/money_suggestions.dart';
import 'package:control_contable/core/forms/typed_field_controllers.dart';
import 'package:control_contable/core/forms/yes_no.dart';
import 'package:control_contable/core/records/record_month.dart';
import 'package:control_contable/features/expenses/application/expense_supplier_facts.dart';
import 'package:control_contable/features/expenses/domain/expense_preview_query.dart';
import 'package:control_contable/features/expenses/domain/expense_rules.dart';
import 'package:flutter/foundation.dart';

export 'package:control_contable/core/forms/yes_no.dart' show no, yes;

/// A folio fiscal: 8-4-4-4-12 hexadecimal digits (contract `CfdiUuid`).
final RegExp _cfdiUuid = RegExp(
  r'^[0-9A-Fa-f]{8}-([0-9A-Fa-f]{4}-){3}[0-9A-Fa-f]{12}$',
);

/// The visible labels the model needs (read from the ARB files by the
/// screen, so this layer has no `BuildContext`).
typedef ExpenseFormCopy = ({
  String category,
  String amount,
  String ivaRate,
  String paidOn,
  String paymentMethod,
  String hasCfdi,
  String regime,
  String ivaAmount,
  String deductible,
  String cfdiUuid,
  String cfdiUuidInvalid,
  String description,
  SupplierFactLabels supplier,
});

/// What the expense form page holds (M4-10, WH-12): its fields in reading
/// order, "Tu proveedor" for a persona moral (the server derives what to
/// withhold, ADR-0008), the category catalog of the chosen regime and
/// date, the live deductibility preview, and the idempotency id of the
/// form session.
///
/// New expense: the context fields (date, payment method, "¿Tienes
/// factura?", IVA rate, regime) survive "Guardar y crear otro"; category,
/// amounts, folio, description and the override are cleared and a new
/// session id starts. Edit ([editing]): every field starts at the stored
/// expense; a stored override shows as the user's.
class ExpenseFormModel extends ChangeNotifier {
  /// The form of [taxpayer] for [regimes] (those in effect on the date),
  /// starting on [today] paid with [paymentMethod], or on [editing].
  new({
    required this.taxpayer,
    required this._regimes,
    required DateTime today,
    required PaymentMethod paymentMethod,
    required this.copy,
    required String Function() newId,
    SupplierMemory? Function(String category) lastSupplierFor = _noMemory,
    this.editing,
  }) : _newId = newId,
       _sessionId = newId() {
    _build(today, paymentMethod, lastSupplierFor);
    final stored = editing;
    if (stored == null ||
        stored.ivaCents == ivaCentsFor(stored.amountCents, stored.ivaRate)) {
      _suggestions.adopt(ivaAmount);
    }
    for (final field in [amount, ivaAmount]) {
      field.text.addListener(_recompute);
    }
    ivaRate.addListener(_recompute);
    category.addListener(() => supplier.followCategory(chosenCategory));
    form.addListener(_followFacts);
    // Text fields notify the form only when their error changes: the date
    // feeds the withholding preview (its remittance deadline), so each new
    // date must reach listeners.
    paidOn.text.addListener(notifyListeners);
  }

  /// Whose expense (its person type decides the withholdings).
  final Taxpayer taxpayer;

  /// The copy the fields were labeled with.
  final ExpenseFormCopy copy;

  /// The stored expense this form edits; `null` for a new one.
  final Expense? editing;

  final List<TaxpayerRegime> _regimes;
  final String Function() _newId;
  final MoneySuggestions _suggestions = MoneySuggestions();
  late final HousingRentRate _housing = HousingRentRate(
    ivaRate,
    () => IvaRate.rate16,
  );
  String _sessionId;
  ExpenseCategoryList? _catalog;
  bool _isApplying = false;

  /// The fields, in reading order.
  final AppFormController form = AppFormController();

  /// The category code.
  late final ChoiceFieldController category;

  /// E1–E4: what the supplier withholdings are derived from.
  late final ExpenseSupplierFacts supplier;

  /// Subtotal before IVA.
  late final MoneyFieldController amount;

  /// The IVA rate (`IvaRate.json`).
  late final ChoiceFieldController ivaRate;

  /// The payment date (the only date of an expense).
  late final DateFieldController paidOn;

  /// How it was paid (`PaymentMethod.json`).
  late final ChoiceFieldController paymentMethod;

  /// Whether there is a CFDI ([yes] or [no]).
  late final ChoiceFieldController hasCfdi;

  /// The regime; shown only with more than one in effect.
  late final ChoiceFieldController regime;

  /// IVA paid; follows the rate until the user writes the CFDI's.
  late final MoneyFieldController ivaAmount;

  /// The user's deductibility override ([yes] / [no]); `null` uses the
  /// suggested value.
  late final ChoiceFieldController deductible;

  /// The CFDI's folio fiscal.
  late final AppFieldController cfdiUuid;

  /// What the expense was.
  late final AppFieldController description;

  static SupplierMemory? _noMemory(String category) => null;

  void _build(
    DateTime today,
    PaymentMethod method,
    SupplierMemory? Function(String category) lastSupplierFor,
  ) {
    final stored = editing;
    category = form.add(
      ChoiceFieldController(
        label: copy.category,
        initialValue: stored?.categoryCode,
      ),
    );
    supplier = ExpenseSupplierFacts(
      form: form,
      labels: copy.supplier,
      isMoral: isMoral,
      lastUsedFor: lastSupplierFor,
      stored: stored,
    );
    amount = form.moneyField(
      label: copy.amount,
      initialText: stored == null ? '' : formatCentavos(stored.amountCents),
    );
    ivaRate = form.add(
      ChoiceFieldController(
        label: copy.ivaRate,
        keepForNext: true,
        initialValue: (stored?.ivaRate ?? IvaRate.rate16).json,
      ),
    );
    paidOn = form.dateField(
      label: copy.paidOn,
      keepForNext: true,
      initialText: formatDate(
        stored == null ? today : parseBusinessDate(stored.paidOn),
      ),
    );
    paymentMethod = form.add(
      ChoiceFieldController(
        label: copy.paymentMethod,
        keepForNext: true,
        initialValue: (stored?.paymentMethod ?? method).json,
      ),
    );
    hasCfdi = form.add(
      ChoiceFieldController(
        label: copy.hasCfdi,
        keepForNext: true,
        initialValue: (stored?.hasCfdi ?? true) ? yes : no,
      ),
    );
    regime = form.add(
      ChoiceFieldController(
        label: copy.regime,
        keepForNext: true,
        initialValue: stored?.regimeCode ?? _regimes.first.regimeCode,
      ),
    );
    ivaAmount = form.moneyField(
      label: copy.ivaAmount,
      isRequired: false,
      initialText: stored == null ? '' : formatCentavos(stored.ivaCents),
    );
    deductible = form.add(
      ChoiceFieldController(
        label: copy.deductible,
        isRequired: false,
        initialValue: _storedOverride(stored),
      ),
    );
    cfdiUuid = form.textField(
      label: copy.cfdiUuid,
      isRequired: false,
      validator: (value) => _cfdiUuid.hasMatch(value)
          ? null
          : CustomFieldError(copy.cfdiUuidInvalid),
      canonicalize: (value) => value.toUpperCase(),
      initialText: stored?.cfdiUuid ?? '',
    );
    description = form.textField(
      label: copy.description,
      isRequired: false,
      initialText: stored?.description ?? '',
    );
  }

  static String? _storedOverride(Expense? stored) =>
      stored == null || stored.deductibleSource != DeductibleSource.user
      ? null
      : (stored.deductible ? yes : no);

  /// Whether this form edits a stored expense.
  bool get isEditing => editing != null;

  /// The idempotency id of this form session: every retry resends it.
  String get sessionId => _sessionId;

  /// The regimes the user can choose (more than one shows the field).
  List<TaxpayerRegime> get regimes => List.unmodifiable(_regimes);

  /// The chosen regime's code.
  String get regimeCode => regime.value ?? _regimes.first.regimeCode;

  /// The chosen regime.
  TaxpayerRegime get heldRegime => _regimes.firstWhere(
    (r) => r.regimeCode == regimeCode,
    orElse: () => _regimes.first,
  );

  /// Whether the taxpayer is a persona moral (only they withhold from
  /// suppliers, MODULE-expenses §2.6).
  bool get isMoral => taxpayer.personType == PersonType.moral;

  /// The payment date as the contract's `YYYY-MM-DD`, when valid.
  String? get paidDay {
    final date = paidOn.date;
    return date == null ? null : businessDate(date);
  }

  /// The catalog of the current regime and date; `null` before the first
  /// answer.
  ExpenseCategoryList? get catalog => _catalog;

  /// The catalog's categories, the most used in [usage] first, then in
  /// the catalog's order.
  List<ExpenseCategory> categoriesByUse(Map<String, int> usage) {
    final items = [...?_catalog?.items];
    final order = {for (final (i, c) in items.indexed) c.code: i};
    return items..sort((a, b) {
      final byUse = (usage[b.code] ?? 0).compareTo(usage[a.code] ?? 0);
      return byUse != 0 ? byUse : order[a.code]!.compareTo(order[b.code]!);
    });
  }

  /// The chosen category, once the catalog knows it.
  ExpenseCategory? get chosenCategory {
    final code = category.value;
    return _catalog?.items.where((c) => c.code == code).firstOrNull;
  }

  /// The IVA rates to offer: the allowed ones once known.
  List<IvaRate> get ivaRatesOffered =>
      _catalog?.ivaRatesAllowed ?? IvaRate.$valuesDefined;

  /// How it was paid.
  PaymentMethod get method => PaymentMethod.fromJson(paymentMethod.value ?? '');

  /// Whether there is a CFDI.
  bool get withCfdi => hasCfdi.value != no;

  /// The IVA the user wrote, else the one the rate gives.
  int get ivaCents =>
      ivaAmount.centavos ??
      ivaCentsFor(amount.centavos ?? 0, IvaRate.fromJson(ivaRate.value ?? ''));

  /// Subtotal plus IVA (what was paid).
  int get totalCents => (amount.centavos ?? 0) + ivaCents;

  /// The cash rule this payment breaks, if any (MODULE-expenses §2.5).
  CashRisk? get cashRisk => cashRiskOf(
    method: method,
    category: chosenCategory,
    totalCents: totalCents,
    limitCents: _catalog?.cashPaymentLimitCents,
  );

  /// The deductibility the server would give without an override.
  bool get suggestedDeductible => suggestedDeductibleOf(
    category: chosenCategory,
    hasCfdi: withCfdi,
    cashRisk: cashRisk,
  );

  /// The user's override; `null` uses [suggestedDeductible].
  bool? get deductibleOverride => switch (deductible.value) {
    yes => true,
    no => false,
    _ => null,
  };

  /// What the switch shows.
  bool get effectiveDeductible => deductibleOverride ?? suggestedDeductible;

  /// The plain-language hint of the chosen category for this taxpayer.
  DeductibilityHint? get hint {
    final chosen = chosenCategory;
    return chosen == null
        ? null
        : hintOf(
            chosen.deductibility,
            regime: heldRegime,
            personType: taxpayer.personType,
          );
  }

  /// Whether the IVA rate is asked (housing rent is exempt by law).
  bool get asksIvaRate => !supplier.isResidentialRent;

  /// The facts the server derives the supplier withholding from, or `null`
  /// for a persona física and until the category, the amount, the rate and
  /// the supplier are known.
  ExpensePreviewQuery? get previewQuery {
    final code = category.value;
    final cents = amount.centavos;
    final kind = supplier.kind;
    final rate = IvaRate.fromJson(ivaRate.value ?? '');
    if (code == null || kind == null || rate == IvaRate.$unknown) return null;
    if (cents == null || cents <= 0) return null;
    return (
      regime: regimeCode,
      on: paidDay,
      amountCents: cents,
      ivaRate: rate,
      ivaCents: ivaCents,
      categoryCode: code,
      supplierKind: kind,
      supplierRegime: supplier.regime,
      operationKind: supplier.operation,
      residentialRent: supplier.residentialRentSent,
    );
  }

  /// Housing rent carries no IVA: the rate is held at exempt meanwhile.
  void _followFacts() {
    _housing.follow(isHousing: supplier.isResidentialRent);
    notifyListeners();
  }

  /// The catalog for the current regime and date arrived.
  void applyCatalog(ExpenseCategoryList catalog) {
    _catalog = catalog;
    final chosen = IvaRate.fromJson(ivaRate.value ?? '');
    if (!catalog.ivaRatesAllowed.contains(chosen)) ivaRate.value = null;
    if (isEditing) supplier.showServerDefault(chosenCategory);
    notifyListeners();
  }

  /// Sets the override to [value]; `null` goes back to the suggestion.
  set deductibleOverride(bool? value) => deductible.value = switch (value) {
    true => yes,
    false => no,
    null => null,
  };

  /// The amount or the rate changed: the IVA suggestion follows.
  void _recompute() {
    // Writing the suggestion notifies its field, which asks again.
    if (_isApplying) return;
    _isApplying = true;
    try {
      final base = amount.centavos;
      _suggestions.suggest(
        ivaAmount,
        base == null
            ? null
            : ivaCentsFor(base, IvaRate.fromJson(ivaRate.value ?? '')),
      );
    } finally {
      _isApplying = false;
    }
    notifyListeners();
  }

  /// "Guardar y crear otro" saved: a new session (a new id) starts. The
  /// form clears its non-context fields itself.
  void startNextSession() {
    _sessionId = _newId();
    _suggestions.forget();
  }

  @override
  void dispose() {
    form.dispose();
    super.dispose();
  }
}
