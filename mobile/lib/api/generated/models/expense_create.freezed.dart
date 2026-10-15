// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpenseCreate {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;/// A code of `listExpenseCategories` (`unknown_category` otherwise).
@JsonKey(name: 'category_code') String get categoryCode;/// Subtotal before IVA, MXN centavos.
@JsonKey(name: 'amount_cents') PositiveCents get amountCents;@JsonKey(name: 'iva_rate') IvaRate get ivaRate;/// IVA paid; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
@JsonKey(name: 'iva_cents') NonNegativeCents get ivaCents;@JsonKey(name: 'payment_method') PaymentMethod get paymentMethod;/// The only date (paid basis for every regime in scope); the booking date.
@JsonKey(name: 'paid_on') BusinessDate get paidOn;@JsonKey(name: 'has_cfdi') bool get hasCfdi;/// Client-generated idempotency key (UUIDv7 recommended), once per form session.
@JsonKey(includeIfNull: false) Uuid? get id;/// Requires `has_cfdi: true` (`has_cfdi_mismatch`).
@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid;/// The user's override; omitted or `null` uses the default.
@JsonKey(includeIfNull: false) bool? get deductible;/// Persona moral taxpayers only; required for them since 0.11.3 (`400 required`).
@JsonKey(includeIfNull: false, name: 'supplier_kind') SupplierKind? get supplierKind;/// Only with `supplier_kind: pf_with_activity` (omitted → `unknown`); with any other kind it answers `422` `supplier_not_applicable`.
@JsonKey(includeIfNull: false, name: 'supplier_regime') SupplierRegime? get supplierRegime;/// Persona moral taxpayers only; omitted → the category's default operation.
@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? get operationKind;/// Persona moral taxpayers only, `rent_real_estate` only (default `false`). `true` requires `iva_rate: exempt` and `iva_cents: 0`, else `422` `residential_rent_exempt` on `iva_rate`.
@JsonKey(includeIfNull: false, name: 'residential_rent') bool? get residentialRent;@JsonKey(includeIfNull: false) RecordDescription? get description;
/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseCreateCopyWith<ExpenseCreate> get copyWith => _$ExpenseCreateCopyWithImpl<ExpenseCreate>(this as ExpenseCreate, _$identity);

  /// Serializes this ExpenseCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpenseCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpenseCreate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.categoryCode, _this.categoryCode) || other.categoryCode == _this.categoryCode)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.hasCfdi, _this.hasCfdi) || other.hasCfdi == _this.hasCfdi)&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid)&&(identical(other.deductible, _this.deductible) || other.deductible == _this.deductible)&&(identical(other.supplierKind, _this.supplierKind) || other.supplierKind == _this.supplierKind)&&(identical(other.supplierRegime, _this.supplierRegime) || other.supplierRegime == _this.supplierRegime)&&(identical(other.operationKind, _this.operationKind) || other.operationKind == _this.operationKind)&&(identical(other.residentialRent, _this.residentialRent) || other.residentialRent == _this.residentialRent)&&(identical(other.description, _this.description) || other.description == _this.description));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpenseCreate;
  return Object.hash(runtimeType,_this.regimeCode,_this.categoryCode,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.paymentMethod,_this.paidOn,_this.hasCfdi,_this.id,_this.cfdiUuid,_this.deductible,_this.supplierKind,_this.supplierRegime,_this.operationKind,_this.residentialRent,_this.description);
}

@override
String toString() {
  final _this = this as ExpenseCreate;
  return 'ExpenseCreate(regimeCode: ${_this.regimeCode}, categoryCode: ${_this.categoryCode}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, paymentMethod: ${_this.paymentMethod}, paidOn: ${_this.paidOn}, hasCfdi: ${_this.hasCfdi}, id: ${_this.id}, cfdiUuid: ${_this.cfdiUuid}, deductible: ${_this.deductible}, supplierKind: ${_this.supplierKind}, supplierRegime: ${_this.supplierRegime}, operationKind: ${_this.operationKind}, residentialRent: ${_this.residentialRent}, description: ${_this.description})';
}


}

/// @nodoc
abstract mixin class $ExpenseCreateCopyWith<$Res>  {
  factory $ExpenseCreateCopyWith(ExpenseCreate value, $Res Function(ExpenseCreate) _then) = _$ExpenseCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'category_code') String categoryCode,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'payment_method') PaymentMethod paymentMethod,@JsonKey(name: 'paid_on') BusinessDate paidOn,@JsonKey(name: 'has_cfdi') bool hasCfdi,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: false) bool? deductible,@JsonKey(includeIfNull: false, name: 'supplier_kind') SupplierKind? supplierKind,@JsonKey(includeIfNull: false, name: 'supplier_regime') SupplierRegime? supplierRegime,@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? operationKind,@JsonKey(includeIfNull: false, name: 'residential_rent') bool? residentialRent,@JsonKey(includeIfNull: false) RecordDescription? description
});




}
/// @nodoc
class _$ExpenseCreateCopyWithImpl<$Res>
    implements $ExpenseCreateCopyWith<$Res> {
  _$ExpenseCreateCopyWithImpl(this._self, this._then);

  final ExpenseCreate _self;
  final $Res Function(ExpenseCreate) _then;

/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? categoryCode = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? paymentMethod = null,Object? paidOn = null,Object? hasCfdi = null,Object? id = freezed,Object? cfdiUuid = freezed,Object? deductible = freezed,Object? supplierKind = freezed,Object? supplierRegime = freezed,Object? operationKind = freezed,Object? residentialRent = freezed,Object? description = freezed,}) {
  return _then(ExpenseCreate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,categoryCode: null == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as BusinessDate,hasCfdi: null == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,deductible: freezed == deductible ? _self.deductible : deductible // ignore: cast_nullable_to_non_nullable
as bool?,supplierKind: freezed == supplierKind ? _self.supplierKind : supplierKind // ignore: cast_nullable_to_non_nullable
as SupplierKind?,supplierRegime: freezed == supplierRegime ? _self.supplierRegime : supplierRegime // ignore: cast_nullable_to_non_nullable
as SupplierRegime?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: freezed == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,
  ));
}

}


/// Adds pattern-matching-related methods to [ExpenseCreate].
extension ExpenseCreatePatterns on ExpenseCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpenseCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpenseCreate value)  $default,){
final _that = this;
switch (_that) {
case _ExpenseCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpenseCreate value)?  $default,){
final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'category_code')  String categoryCode, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod paymentMethod, @JsonKey(name: 'paid_on')  BusinessDate paidOn, @JsonKey(name: 'has_cfdi')  bool hasCfdi, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  bool? deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false)  RecordDescription? description)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
return $default(_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.id,_that.cfdiUuid,_that.deductible,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.description);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'category_code')  String categoryCode, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod paymentMethod, @JsonKey(name: 'paid_on')  BusinessDate paidOn, @JsonKey(name: 'has_cfdi')  bool hasCfdi, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  bool? deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false)  RecordDescription? description)  $default,) {final _that = this;
switch (_that) {
case _ExpenseCreate():
return $default(_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.id,_that.cfdiUuid,_that.deductible,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.description);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'category_code')  String categoryCode, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'payment_method')  PaymentMethod paymentMethod, @JsonKey(name: 'paid_on')  BusinessDate paidOn, @JsonKey(name: 'has_cfdi')  bool hasCfdi, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  bool? deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false)  RecordDescription? description)?  $default,) {final _that = this;
switch (_that) {
case _ExpenseCreate() when $default != null:
return $default(_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.id,_that.cfdiUuid,_that.deductible,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.description);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpenseCreate implements ExpenseCreate {
  const _ExpenseCreate({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'category_code') required this.categoryCode, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'iva_rate') required this.ivaRate, @JsonKey(name: 'iva_cents') required this.ivaCents, @JsonKey(name: 'payment_method') required this.paymentMethod, @JsonKey(name: 'paid_on') required this.paidOn, @JsonKey(name: 'has_cfdi') required this.hasCfdi, @JsonKey(includeIfNull: false) this.id, @JsonKey(includeIfNull: false, name: 'cfdi_uuid') this.cfdiUuid, @JsonKey(includeIfNull: false) this.deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind') this.supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime') this.supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind') this.operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent') this.residentialRent, @JsonKey(includeIfNull: false) this.description});
  factory _ExpenseCreate.fromJson(Map<String, dynamic> json) => _$ExpenseCreateFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
/// A code of `listExpenseCategories` (`unknown_category` otherwise).
@override@JsonKey(name: 'category_code') final  String categoryCode;
/// Subtotal before IVA, MXN centavos.
@override@JsonKey(name: 'amount_cents') final  PositiveCents amountCents;
@override@JsonKey(name: 'iva_rate') final  IvaRate ivaRate;
/// IVA paid; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents ivaCents;
@override@JsonKey(name: 'payment_method') final  PaymentMethod paymentMethod;
/// The only date (paid basis for every regime in scope); the booking date.
@override@JsonKey(name: 'paid_on') final  BusinessDate paidOn;
@override@JsonKey(name: 'has_cfdi') final  bool hasCfdi;
/// Client-generated idempotency key (UUIDv7 recommended), once per form session.
@override@JsonKey(includeIfNull: false) final  Uuid? id;
/// Requires `has_cfdi: true` (`has_cfdi_mismatch`).
@override@JsonKey(includeIfNull: false, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;
/// The user's override; omitted or `null` uses the default.
@override@JsonKey(includeIfNull: false) final  bool? deductible;
/// Persona moral taxpayers only; required for them since 0.11.3 (`400 required`).
@override@JsonKey(includeIfNull: false, name: 'supplier_kind') final  SupplierKind? supplierKind;
/// Only with `supplier_kind: pf_with_activity` (omitted → `unknown`); with any other kind it answers `422` `supplier_not_applicable`.
@override@JsonKey(includeIfNull: false, name: 'supplier_regime') final  SupplierRegime? supplierRegime;
/// Persona moral taxpayers only; omitted → the category's default operation.
@override@JsonKey(includeIfNull: false, name: 'operation_kind') final  OperationKind? operationKind;
/// Persona moral taxpayers only, `rent_real_estate` only (default `false`). `true` requires `iva_rate: exempt` and `iva_cents: 0`, else `422` `residential_rent_exempt` on `iva_rate`.
@override@JsonKey(includeIfNull: false, name: 'residential_rent') final  bool? residentialRent;
@override@JsonKey(includeIfNull: false) final  RecordDescription? description;

/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseCreateCopyWith<_ExpenseCreate> get copyWith => __$ExpenseCreateCopyWithImpl<_ExpenseCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpenseCreate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.categoryCode, categoryCode) || other.categoryCode == categoryCode)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.hasCfdi, hasCfdi) || other.hasCfdi == hasCfdi)&&(identical(other.id, id) || other.id == id)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid)&&(identical(other.deductible, deductible) || other.deductible == deductible)&&(identical(other.supplierKind, supplierKind) || other.supplierKind == supplierKind)&&(identical(other.supplierRegime, supplierRegime) || other.supplierRegime == supplierRegime)&&(identical(other.operationKind, operationKind) || other.operationKind == operationKind)&&(identical(other.residentialRent, residentialRent) || other.residentialRent == residentialRent)&&(identical(other.description, description) || other.description == description));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,categoryCode,amountCents,ivaRate,ivaCents,paymentMethod,paidOn,hasCfdi,id,cfdiUuid,deductible,supplierKind,supplierRegime,operationKind,residentialRent,description);
}

@override
String toString() {
    return 'ExpenseCreate(regimeCode: $regimeCode, categoryCode: $categoryCode, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, paymentMethod: $paymentMethod, paidOn: $paidOn, hasCfdi: $hasCfdi, id: $id, cfdiUuid: $cfdiUuid, deductible: $deductible, supplierKind: $supplierKind, supplierRegime: $supplierRegime, operationKind: $operationKind, residentialRent: $residentialRent, description: $description)';
}


}

/// @nodoc
abstract mixin class _$ExpenseCreateCopyWith<$Res> implements $ExpenseCreateCopyWith<$Res> {
  factory _$ExpenseCreateCopyWith(_ExpenseCreate value, $Res Function(_ExpenseCreate) _then) = __$ExpenseCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'category_code') String categoryCode,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'payment_method') PaymentMethod paymentMethod,@JsonKey(name: 'paid_on') BusinessDate paidOn,@JsonKey(name: 'has_cfdi') bool hasCfdi,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: false) bool? deductible,@JsonKey(includeIfNull: false, name: 'supplier_kind') SupplierKind? supplierKind,@JsonKey(includeIfNull: false, name: 'supplier_regime') SupplierRegime? supplierRegime,@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? operationKind,@JsonKey(includeIfNull: false, name: 'residential_rent') bool? residentialRent,@JsonKey(includeIfNull: false) RecordDescription? description
});




}
/// @nodoc
class __$ExpenseCreateCopyWithImpl<$Res>
    implements _$ExpenseCreateCopyWith<$Res> {
  __$ExpenseCreateCopyWithImpl(this._self, this._then);

  final _ExpenseCreate _self;
  final $Res Function(_ExpenseCreate) _then;

/// Create a copy of ExpenseCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? categoryCode = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? paymentMethod = null,Object? paidOn = null,Object? hasCfdi = null,Object? id = freezed,Object? cfdiUuid = freezed,Object? deductible = freezed,Object? supplierKind = freezed,Object? supplierRegime = freezed,Object? operationKind = freezed,Object? residentialRent = freezed,Object? description = freezed,}) {
  return _then(_ExpenseCreate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,categoryCode: null == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,paymentMethod: null == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod,paidOn: null == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as BusinessDate,hasCfdi: null == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,deductible: freezed == deductible ? _self.deductible : deductible // ignore: cast_nullable_to_non_nullable
as bool?,supplierKind: freezed == supplierKind ? _self.supplierKind : supplierKind // ignore: cast_nullable_to_non_nullable
as SupplierKind?,supplierRegime: freezed == supplierRegime ? _self.supplierRegime : supplierRegime // ignore: cast_nullable_to_non_nullable
as SupplierRegime?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: freezed == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,
  ));
}


}

// dart format on
