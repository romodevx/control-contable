// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Expense {

 Uuid get id;@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(name: 'category_code') String get categoryCode;@JsonKey(name: 'amount_cents') PositiveCents get amountCents;@JsonKey(name: 'iva_rate') IvaRate get ivaRate;@JsonKey(name: 'iva_cents') NonNegativeCents get ivaCents;@JsonKey(name: 'payment_method') PaymentMethod get paymentMethod;@JsonKey(name: 'paid_on') BusinessDate get paidOn;@JsonKey(name: 'has_cfdi') bool get hasCfdi;@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid;/// `null` for a persona física taxpayer.
@JsonKey(includeIfNull: true, name: 'supplier_kind') SupplierKind? get supplierKind;/// `null` unless `supplier_kind: pf_with_activity`.
@JsonKey(includeIfNull: true, name: 'supplier_regime') SupplierRegime? get supplierRegime;/// `null` for a persona física taxpayer.
@JsonKey(includeIfNull: true, name: 'operation_kind') OperationKind? get operationKind;@JsonKey(name: 'residential_rent') bool get residentialRent;/// Read-only mirror of `withholding.isr_cents` (kept for list totals and the estimate).
@JsonKey(name: 'isr_withheld_cents') NonNegativeCents get isrWithheldCents;/// Read-only mirror of `withholding.iva_cents` (kept for list totals and the estimate).
@JsonKey(name: 'iva_withheld_cents') NonNegativeCents get ivaWithheldCents;/// The stored derivation; `null` only for a record stored before automatic derivation that the server has not derived.
@JsonKey(includeIfNull: true) Withholding? get withholding; RecordDescription? get description;/// Effective value ("the user claims it lowers ISR"): the override if any, else the default.
 bool get deductible;@JsonKey(name: 'deductible_source') DeductibleSource get deductibleSource;/// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes (append-only): `cash_over_threshold` (`payment_method`), `no_cfdi` (`has_cfdi`), `not_isr_deductible_in_regime` and `category_not_deductible` (`deductible`), `duplicate_cfdi_uuid` (`cfdi_uuid`), `iva_not_creditable_cash_payment` (`payment_method`), `provisional_parameters` (`paid_on`), `deduction_requires_remittance` (`withholding`), `withholding_cpa_unverified` (`withholding`), `supplier_regime_unknown` (`supplier_regime`), `foreign_counterparty_not_supported` (`supplier_kind`).
 List<Warning> get warnings; Legend get legend;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;
/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseCopyWith<Expense> get copyWith => _$ExpenseCopyWithImpl<Expense>(this as Expense, _$identity);

  /// Serializes this Expense to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Expense;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Expense&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.categoryCode, _this.categoryCode) || other.categoryCode == _this.categoryCode)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.hasCfdi, _this.hasCfdi) || other.hasCfdi == _this.hasCfdi)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid)&&(identical(other.supplierKind, _this.supplierKind) || other.supplierKind == _this.supplierKind)&&(identical(other.supplierRegime, _this.supplierRegime) || other.supplierRegime == _this.supplierRegime)&&(identical(other.operationKind, _this.operationKind) || other.operationKind == _this.operationKind)&&(identical(other.residentialRent, _this.residentialRent) || other.residentialRent == _this.residentialRent)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.withholding, _this.withholding) || other.withholding == _this.withholding)&&(identical(other.description, _this.description) || other.description == _this.description)&&(identical(other.deductible, _this.deductible) || other.deductible == _this.deductible)&&(identical(other.deductibleSource, _this.deductibleSource) || other.deductibleSource == _this.deductibleSource)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings)&&(identical(other.legend, _this.legend) || other.legend == _this.legend)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Expense;
  return Object.hashAll([runtimeType,_this.id,_this.taxpayerId,_this.regimeCode,_this.categoryCode,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.paymentMethod,_this.paidOn,_this.hasCfdi,_this.cfdiUuid,_this.supplierKind,_this.supplierRegime,_this.operationKind,_this.residentialRent,_this.isrWithheldCents,_this.ivaWithheldCents,_this.withholding,_this.description,_this.deductible,_this.deductibleSource,const DeepCollectionEquality().hash(_this.warnings),_this.legend,_this.createdAt,_this.updatedAt]);
}

@override
String toString() {
  final _this = this as Expense;
  return 'Expense(id: ${_this.id}, taxpayerId: ${_this.taxpayerId}, regimeCode: ${_this.regimeCode}, categoryCode: ${_this.categoryCode}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, paymentMethod: ${_this.paymentMethod}, paidOn: ${_this.paidOn}, hasCfdi: ${_this.hasCfdi}, cfdiUuid: ${_this.cfdiUuid}, supplierKind: ${_this.supplierKind}, supplierRegime: ${_this.supplierRegime}, operationKind: ${_this.operationKind}, residentialRent: ${_this.residentialRent}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, withholding: ${_this.withholding}, description: ${_this.description}, deductible: ${_this.deductible}, deductibleSource: ${_this.deductibleSource}, warnings: ${_this.warnings}, legend: ${_this.legend}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $ExpenseCopyWith<$Res>  {
  factory $ExpenseCopyWith(Expense value, $Res Function(Expense) _then) = _$ExpenseCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'category_code') String categoryCode,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'payment_method') PaymentMethod paymentMethod,@JsonKey(name: 'paid_on') BusinessDate paidOn,@JsonKey(name: 'has_cfdi') bool hasCfdi,@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: true, name: 'supplier_kind') SupplierKind? supplierKind,@JsonKey(includeIfNull: true, name: 'supplier_regime') SupplierRegime? supplierRegime,@JsonKey(includeIfNull: true, name: 'operation_kind') OperationKind? operationKind,@JsonKey(name: 'residential_rent') bool residentialRent,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents ivaWithheldCents,@JsonKey(includeIfNull: true) Withholding? withholding, RecordDescription? description, bool deductible,@JsonKey(name: 'deductible_source') DeductibleSource deductibleSource, List<Warning> warnings, Legend legend,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});


$WithholdingCopyWith<$Res>? get withholding;

}
/// @nodoc
class _$ExpenseCopyWithImpl<$Res>
    implements $ExpenseCopyWith<$Res> {
  _$ExpenseCopyWithImpl(this._self, this._then);

  final Expense _self;
  final $Res Function(Expense) _then;

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? taxpayerId = null,Object? regimeCode = null,Object? categoryCode = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? paymentMethod = null,Object? paidOn = null,Object? hasCfdi = null,Object? cfdiUuid = freezed,Object? supplierKind = freezed,Object? supplierRegime = freezed,Object? operationKind = freezed,Object? residentialRent = null,Object? isrWithheldCents = null,Object? ivaWithheldCents = null,Object? withholding = freezed,Object? description = freezed,Object? deductible = null,Object? deductibleSource = null,Object? warnings = null,Object? legend = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(Expense(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,categoryCode: null == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as BusinessDate,hasCfdi: null == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,supplierKind: freezed == supplierKind ? _self.supplierKind : supplierKind // ignore: cast_nullable_to_non_nullable
as SupplierKind?,supplierRegime: freezed == supplierRegime ? _self.supplierRegime : supplierRegime // ignore: cast_nullable_to_non_nullable
as SupplierRegime?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: null == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaWithheldCents: null == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholding: freezed == withholding ? _self.withholding : withholding // ignore: cast_nullable_to_non_nullable
as Withholding?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,deductible: null == deductible ? _self.deductible : deductible // ignore: cast_nullable_to_non_nullable
as bool,deductibleSource: null == deductibleSource ? _self.deductibleSource : deductibleSource // ignore: cast_nullable_to_non_nullable
as DeductibleSource,warnings: null == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}
/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$WithholdingCopyWith<$Res>? get withholding {
    if (_self.withholding == null) {
    return null;
  }

  return $WithholdingCopyWith<$Res>(_self.withholding!, (value) {
    return _then(_self.copyWith(withholding: value));
  });
}
}


/// Adds pattern-matching-related methods to [Expense].
extension ExpensePatterns on Expense {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Expense value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Expense() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Expense value)  $default,){
final _that = this;
switch (_that) {
case _Expense():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Expense value)?  $default,){
final _that = this;
switch (_that) {
case _Expense() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'category_code')  String categoryCode, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod paymentMethod, @JsonKey(name: 'paid_on')  BusinessDate paidOn, @JsonKey(name: 'has_cfdi')  bool hasCfdi, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: true, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: true, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: true, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents ivaWithheldCents, @JsonKey(includeIfNull: true)  Withholding? withholding,  RecordDescription? description,  bool deductible, @JsonKey(name: 'deductible_source')  DeductibleSource deductibleSource,  List<Warning> warnings,  Legend legend, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Expense() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.cfdiUuid,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.isrWithheldCents,_that.ivaWithheldCents,_that.withholding,_that.description,_that.deductible,_that.deductibleSource,_that.warnings,_that.legend,_that.createdAt,_that.updatedAt);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'category_code')  String categoryCode, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod paymentMethod, @JsonKey(name: 'paid_on')  BusinessDate paidOn, @JsonKey(name: 'has_cfdi')  bool hasCfdi, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: true, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: true, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: true, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents ivaWithheldCents, @JsonKey(includeIfNull: true)  Withholding? withholding,  RecordDescription? description,  bool deductible, @JsonKey(name: 'deductible_source')  DeductibleSource deductibleSource,  List<Warning> warnings,  Legend legend, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)  $default,) {final _that = this;
switch (_that) {
case _Expense():
return $default(_that.id,_that.taxpayerId,_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.cfdiUuid,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.isrWithheldCents,_that.ivaWithheldCents,_that.withholding,_that.description,_that.deductible,_that.deductibleSource,_that.warnings,_that.legend,_that.createdAt,_that.updatedAt);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'category_code')  String categoryCode, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod paymentMethod, @JsonKey(name: 'paid_on')  BusinessDate paidOn, @JsonKey(name: 'has_cfdi')  bool hasCfdi, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: true, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: true, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: true, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents ivaWithheldCents, @JsonKey(includeIfNull: true)  Withholding? withholding,  RecordDescription? description,  bool deductible, @JsonKey(name: 'deductible_source')  DeductibleSource deductibleSource,  List<Warning> warnings,  Legend legend, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _Expense() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.cfdiUuid,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.isrWithheldCents,_that.ivaWithheldCents,_that.withholding,_that.description,_that.deductible,_that.deductibleSource,_that.warnings,_that.legend,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Expense implements Expense {
  const _Expense({required this.id, @JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'category_code') required this.categoryCode, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'iva_rate') required this.ivaRate, @JsonKey(name: 'iva_cents') required this.ivaCents, @JsonKey(name: 'payment_method') required this.paymentMethod, @JsonKey(name: 'paid_on') required this.paidOn, @JsonKey(name: 'has_cfdi') required this.hasCfdi, @JsonKey(includeIfNull: true, name: 'cfdi_uuid') required this.cfdiUuid, @JsonKey(includeIfNull: true, name: 'supplier_kind') required this.supplierKind, @JsonKey(includeIfNull: true, name: 'supplier_regime') required this.supplierRegime, @JsonKey(includeIfNull: true, name: 'operation_kind') required this.operationKind, @JsonKey(name: 'residential_rent') required this.residentialRent, @JsonKey(name: 'isr_withheld_cents') required this.isrWithheldCents, @JsonKey(name: 'iva_withheld_cents') required this.ivaWithheldCents, @JsonKey(includeIfNull: true) required this.withholding, required this.description, required this.deductible, @JsonKey(name: 'deductible_source') required this.deductibleSource, required  List<Warning> warnings, required this.legend, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt}): _warnings = warnings;
  factory _Expense.fromJson(Map<String, dynamic> json) => _$ExpenseFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override@JsonKey(name: 'category_code') final  String categoryCode;
@override@JsonKey(name: 'amount_cents') final  PositiveCents amountCents;
@override@JsonKey(name: 'iva_rate') final  IvaRate ivaRate;
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents ivaCents;
@override@JsonKey(name: 'payment_method') final  PaymentMethod paymentMethod;
@override@JsonKey(name: 'paid_on') final  BusinessDate paidOn;
@override@JsonKey(name: 'has_cfdi') final  bool hasCfdi;
@override@JsonKey(includeIfNull: true, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;
/// `null` for a persona física taxpayer.
@override@JsonKey(includeIfNull: true, name: 'supplier_kind') final  SupplierKind? supplierKind;
/// `null` unless `supplier_kind: pf_with_activity`.
@override@JsonKey(includeIfNull: true, name: 'supplier_regime') final  SupplierRegime? supplierRegime;
/// `null` for a persona física taxpayer.
@override@JsonKey(includeIfNull: true, name: 'operation_kind') final  OperationKind? operationKind;
@override@JsonKey(name: 'residential_rent') final  bool residentialRent;
/// Read-only mirror of `withholding.isr_cents` (kept for list totals and the estimate).
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents isrWithheldCents;
/// Read-only mirror of `withholding.iva_cents` (kept for list totals and the estimate).
@override@JsonKey(name: 'iva_withheld_cents') final  NonNegativeCents ivaWithheldCents;
/// The stored derivation; `null` only for a record stored before automatic derivation that the server has not derived.
@override@JsonKey(includeIfNull: true) final  Withholding? withholding;
@override final  RecordDescription? description;
/// Effective value ("the user claims it lowers ISR"): the override if any, else the default.
@override final  bool deductible;
@override@JsonKey(name: 'deductible_source') final  DeductibleSource deductibleSource;
/// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes (append-only): `cash_over_threshold` (`payment_method`), `no_cfdi` (`has_cfdi`), `not_isr_deductible_in_regime` and `category_not_deductible` (`deductible`), `duplicate_cfdi_uuid` (`cfdi_uuid`), `iva_not_creditable_cash_payment` (`payment_method`), `provisional_parameters` (`paid_on`), `deduction_requires_remittance` (`withholding`), `withholding_cpa_unverified` (`withholding`), `supplier_regime_unknown` (`supplier_regime`), `foreign_counterparty_not_supported` (`supplier_kind`).
 final  List<Warning> _warnings;
/// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes (append-only): `cash_over_threshold` (`payment_method`), `no_cfdi` (`has_cfdi`), `not_isr_deductible_in_regime` and `category_not_deductible` (`deductible`), `duplicate_cfdi_uuid` (`cfdi_uuid`), `iva_not_creditable_cash_payment` (`payment_method`), `provisional_parameters` (`paid_on`), `deduction_requires_remittance` (`withholding`), `withholding_cpa_unverified` (`withholding`), `supplier_regime_unknown` (`supplier_regime`), `foreign_counterparty_not_supported` (`supplier_kind`).
@override List<Warning> get warnings {
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_warnings);
}

@override final  Legend legend;
@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseCopyWith<_Expense> get copyWith => __$ExpenseCopyWithImpl<_Expense>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Expense&&(identical(other.id, id) || other.id == id)&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.categoryCode, categoryCode) || other.categoryCode == categoryCode)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.hasCfdi, hasCfdi) || other.hasCfdi == hasCfdi)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid)&&(identical(other.supplierKind, supplierKind) || other.supplierKind == supplierKind)&&(identical(other.supplierRegime, supplierRegime) || other.supplierRegime == supplierRegime)&&(identical(other.operationKind, operationKind) || other.operationKind == operationKind)&&(identical(other.residentialRent, residentialRent) || other.residentialRent == residentialRent)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.withholding, withholding) || other.withholding == withholding)&&(identical(other.description, description) || other.description == description)&&(identical(other.deductible, deductible) || other.deductible == deductible)&&(identical(other.deductibleSource, deductibleSource) || other.deductibleSource == deductibleSource)&&const DeepCollectionEquality().equals(other.warnings, _warnings)&&(identical(other.legend, legend) || other.legend == legend)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hashAll([runtimeType,id,taxpayerId,regimeCode,categoryCode,amountCents,ivaRate,ivaCents,paymentMethod,paidOn,hasCfdi,cfdiUuid,supplierKind,supplierRegime,operationKind,residentialRent,isrWithheldCents,ivaWithheldCents,withholding,description,deductible,deductibleSource,const DeepCollectionEquality().hash(_warnings),legend,createdAt,updatedAt]);
}

@override
String toString() {
    return 'Expense(id: $id, taxpayerId: $taxpayerId, regimeCode: $regimeCode, categoryCode: $categoryCode, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, paymentMethod: $paymentMethod, paidOn: $paidOn, hasCfdi: $hasCfdi, cfdiUuid: $cfdiUuid, supplierKind: $supplierKind, supplierRegime: $supplierRegime, operationKind: $operationKind, residentialRent: $residentialRent, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, withholding: $withholding, description: $description, deductible: $deductible, deductibleSource: $deductibleSource, warnings: $warnings, legend: $legend, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$ExpenseCopyWith<$Res> implements $ExpenseCopyWith<$Res> {
  factory _$ExpenseCopyWith(_Expense value, $Res Function(_Expense) _then) = __$ExpenseCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'category_code') String categoryCode,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'payment_method') PaymentMethod paymentMethod,@JsonKey(name: 'paid_on') BusinessDate paidOn,@JsonKey(name: 'has_cfdi') bool hasCfdi,@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: true, name: 'supplier_kind') SupplierKind? supplierKind,@JsonKey(includeIfNull: true, name: 'supplier_regime') SupplierRegime? supplierRegime,@JsonKey(includeIfNull: true, name: 'operation_kind') OperationKind? operationKind,@JsonKey(name: 'residential_rent') bool residentialRent,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents ivaWithheldCents,@JsonKey(includeIfNull: true) Withholding? withholding, RecordDescription? description, bool deductible,@JsonKey(name: 'deductible_source') DeductibleSource deductibleSource, List<Warning> warnings, Legend legend,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});


@override $WithholdingCopyWith<$Res>? get withholding;

}
/// @nodoc
class __$ExpenseCopyWithImpl<$Res>
    implements _$ExpenseCopyWith<$Res> {
  __$ExpenseCopyWithImpl(this._self, this._then);

  final _Expense _self;
  final $Res Function(_Expense) _then;

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? taxpayerId = null,Object? regimeCode = null,Object? categoryCode = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? paymentMethod = null,Object? paidOn = null,Object? hasCfdi = null,Object? cfdiUuid = freezed,Object? supplierKind = freezed,Object? supplierRegime = freezed,Object? operationKind = freezed,Object? residentialRent = null,Object? isrWithheldCents = null,Object? ivaWithheldCents = null,Object? withholding = freezed,Object? description = freezed,Object? deductible = null,Object? deductibleSource = null,Object? warnings = null,Object? legend = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(_Expense(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,categoryCode: null == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as BusinessDate,hasCfdi: null == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,supplierKind: freezed == supplierKind ? _self.supplierKind : supplierKind // ignore: cast_nullable_to_non_nullable
as SupplierKind?,supplierRegime: freezed == supplierRegime ? _self.supplierRegime : supplierRegime // ignore: cast_nullable_to_non_nullable
as SupplierRegime?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: null == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaWithheldCents: null == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholding: freezed == withholding ? _self.withholding : withholding // ignore: cast_nullable_to_non_nullable
as Withholding?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,deductible: null == deductible ? _self.deductible : deductible // ignore: cast_nullable_to_non_nullable
as bool,deductibleSource: null == deductibleSource ? _self.deductibleSource : deductibleSource // ignore: cast_nullable_to_non_nullable
as DeductibleSource,warnings: null == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

/// Create a copy of Expense
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$WithholdingCopyWith<$Res>? get withholding {
    if (_self.withholding == null) {
    return null;
  }

  return $WithholdingCopyWith<$Res>(_self.withholding!, (value) {
    return _then(_self.copyWith(withholding: value));
  });
}
}

// dart format on
